package phantomjscloud

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	baseEndpointUrl    = "https://phantomjscloud.com/api/browser/v2/"
	defaultHTTPTimeout = 120 * time.Second
)

// ClientOption is a functional option for configuring a Client.
type ClientOption func(*Client)

// RetryConfig defines the strategy for automatic retries on transient errors.
type RetryConfig struct {
	MaxRetries      int
	InitialInterval time.Duration
	Multiplier      float64
	MaxInterval     time.Duration
}

// DefaultRetryConfig provides a sensible default for most use cases.
var DefaultRetryConfig = RetryConfig{
	MaxRetries:      3,
	InitialInterval: 1 * time.Second,
	Multiplier:      2.0,
	MaxInterval:     10 * time.Second,
}

// Interceptor allows modifying requests before they are sent or responses after they are received.
type Interceptor func(req *http.Request, next func(*http.Request) (*http.Response, error)) (*http.Response, error)

// WithHTTPClient replaces the default HTTP client.
// Use this to set a custom transport, proxy, or TLS config.
func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *Client) { c.httpClient = hc }
}

// WithTimeout sets the HTTP timeout for all requests.
// Default is 120s (PhantomJsCloud renders can be slow on complex pages).
// WithEndpoint allows overriding the default API endpoint.
// This is useful for testing or if you need to use a proxy.
func WithEndpoint(url string) ClientOption {
	return func(c *Client) { c.endpoint = url }
}

func WithTimeout(d time.Duration) ClientOption {
	return func(c *Client) { c.httpClient.Timeout = d }
}

// WithRetry enables automatic retries for transient errors (429, 503, timeouts).
func WithRetry(cfg RetryConfig) ClientOption {
	return func(c *Client) { c.retryConfig = &cfg }
}

// WithInterceptor adds a middleware that can inspect/modify requests and responses.
func WithInterceptor(i Interceptor) ClientOption {
	return func(c *Client) { c.interceptors = append(c.interceptors, i) }
}

// Client is a PhantomJsCloud API client.
type Client struct {
	apiKey       string
	endpoint     string
	httpClient   *http.Client
	retryConfig  *RetryConfig
	interceptors []Interceptor
}

// NewClient creates a new Client using the provided API key.
// The API key is required; without it, requests will fail.
//
//	client := phantomjscloud.NewClient("YOUR_KEY",
//	    phantomjscloud.WithTimeout(60*time.Second),
//	)
func NewClient(apiKey string, opts ...ClientOption) *Client {
	c := &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: defaultHTTPTimeout},
		endpoint:   baseEndpointUrl,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// UserResponseWithMeta wraps the UserResponse API payload along with the HTTP Response metadata headers.
type UserResponseWithMeta struct {
	UserResponse
	Metadata ResponseMetadata
}

// DoPage is a convenience method that wraps a single PageRequest inside a UserRequest.
func (c *Client) DoPage(req *PageRequest) (*UserResponseWithMeta, error) {
	return c.DoPageContext(context.Background(), req)
}

// DoPageContext is like DoPage but honours the provided context for cancellation and deadlines.
func (c *Client) DoPageContext(ctx context.Context, req *PageRequest) (*UserResponseWithMeta, error) {
	uReq := &UserRequest{
		Pages: []PageRequest{*req},
	}
	return c.DoContext(ctx, uReq)
}

// Do serializes a UserRequest, performs the HTTP POST to PhantomJsCloud, and parses the response.
func (c *Client) Do(req *UserRequest) (*UserResponseWithMeta, error) {
	return c.DoContext(context.Background(), req)
}

// DoContext is like Do but honours the provided context for cancellation and deadlines.
// It automatically handles retries if WithRetry was used during client initialization.
func (c *Client) DoContext(ctx context.Context, req *UserRequest) (*UserResponseWithMeta, error) {
	if c.apiKey == "" {
		return nil, errors.New("API key is required")
	}

	if c.retryConfig == nil {
		return c.doSingle(ctx, req)
	}

	var lastErr error
	cfg := c.retryConfig
	interval := cfg.InitialInterval

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		res, err := c.doSingle(ctx, req)
		if err == nil {
			return res, nil
		}

		lastErr = err

		// Only retry on transient errors (429, 503, or network/timeout)
		if !isRetryable(err) || attempt == cfg.MaxRetries {
			break
		}

		// Backoff
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(interval):
			interval = time.Duration(float64(interval) * cfg.Multiplier)
			if interval > cfg.MaxInterval {
				interval = cfg.MaxInterval
			}
		}
	}

	return nil, lastErr
}

func (c *Client) doSingle(ctx context.Context, req *UserRequest) (*UserResponseWithMeta, error) {
	endpoint := c.endpoint + c.apiKey + "/"

	// Since we might retry, we can't use io.Pipe easily if we want to avoid double encoding,
	// but for now, simple encoding into a buffer is safer for retries.
	// In the future, we could optimize this.
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Apply interceptors
	httpClientDo := c.httpClient.Do
	for i := len(c.interceptors) - 1; i >= 0; i-- {
		interceptor := c.interceptors[i]
		currentDo := httpClientDo
		httpClientDo = func(r *http.Request) (*http.Response, error) {
			return interceptor(r, currentDo)
		}
	}

	httpResp, err := httpClientDo(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode >= 400 && httpResp.StatusCode < 600 {
		bodyBytes, err := io.ReadAll(httpResp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
		err = fmt.Errorf("phantomjscloud returned HTTP Status %d: %s", httpResp.StatusCode, string(bodyBytes))
		// Wrap with status code for retry logic
		return nil, &httpError{StatusCode: httpResp.StatusCode, Err: err}
	}

	var userResp UserResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&userResp); err != nil {
		return nil, fmt.Errorf("failed to decode response payload: %w", err)
	}

	return &UserResponseWithMeta{
		UserResponse: userResp,
		Metadata:     parseMetadata(httpResp.Header),
	}, nil
}

type httpError struct {
	StatusCode int
	Err        error
}

func (e *httpError) Error() string { return e.Err.Error() }
func (e *httpError) Unwrap() error { return e.Err }

func isRetryable(err error) bool {
	var hErr *httpError
	if errors.As(err, &hErr) {
		// 429 Too Many Requests, 503 Service Unavailable, 504 Gateway Timeout
		return hErr.StatusCode == 429 || hErr.StatusCode == 503 || hErr.StatusCode == 504
	}
	// Also retry on network timeouts or connection refused
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	errStr := err.Error()
	return strings.Contains(errStr, "timeout") || strings.Contains(errStr, "connection refused") || strings.Contains(errStr, "reset by peer")
}

// FetchPDF is a convenience method that returns the raw base64-decoded PDF bytes for a given URL.
// It simplifies generating PDFs directly without handling the raw JSON wrapper.
func (c *Client) FetchPDF(url string, overrideOptions *PdfOptions) ([]byte, error) {
	req := &PageRequest{
		URL:        url,
		RenderType: "pdf",
	}

	if overrideOptions != nil {
		req.RenderSettings = RenderSettings{
			PdfOptions: overrideOptions,
		}
	}

	res, err := c.DoPage(req)
	if err != nil {
		return nil, err
	}

	if len(res.PageResponses) == 0 {
		return nil, errors.New("no page response returned")
	}

	decoded, err := base64.StdEncoding.DecodeString(res.PageResponses[0].Content)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 content: %w", err)
	}

	return decoded, nil
}

// FetchPlainText is a convenience method that returns the raw text context of the page, stripped of all HTML tags.
// Useful for feeding lightweight text into LLM pipelines or doing semantic analysis.
func (c *Client) FetchPlainText(url string) (string, error) {
	req := &PageRequest{
		URL:        url,
		RenderType: "plainText",
	}

	res, err := c.DoPage(req)
	if err != nil {
		return "", err
	}

	if len(res.PageResponses) == 0 {
		return "", errors.New("no page response returned")
	}

	return res.PageResponses[0].Content, nil
}

// FetchScreenshot is a convenience method that returns the raw base64-decoded image bytes.
func (c *Client) FetchScreenshot(url string, renderType string, renderSettings *RenderSettings) ([]byte, error) {
	if renderType != "png" && renderType != "jpeg" {
		renderType = "png"
	}

	req := &PageRequest{
		URL:        url,
		RenderType: renderType,
	}

	if renderSettings != nil {
		req.RenderSettings = *renderSettings
	}

	res, err := c.DoPage(req)
	if err != nil {
		return nil, err
	}

	if len(res.PageResponses) == 0 {
		return nil, errors.New("no page response returned")
	}

	decoded, err := base64.StdEncoding.DecodeString(res.PageResponses[0].Content)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 content: %w", err)
	}

	return decoded, nil
}

// RenderRawHTML allows you to upload raw dynamic string HTML and render it natively through PhantomJS
// as if it were loaded from an external URL. Ideal for generating reports or templated emails.
func (c *Client) RenderRawHTML(html string, renderType string, renderSettings *RenderSettings) ([]byte, error) {
	if renderType != "png" && renderType != "jpeg" && renderType != "pdf" {
		renderType = "png"
	}

	req := &PageRequest{
		// PhantomJS Cloud expects either a URL or raw Content.
		// "http://localhost/blank" triggers the render engine directly on the supplied content.
		URL:        "http://localhost/blank",
		Content:    html,
		RenderType: renderType,
	}

	if renderSettings != nil {
		req.RenderSettings = *renderSettings
	}

	res, err := c.DoPage(req)
	if err != nil {
		return nil, err
	}

	if len(res.PageResponses) == 0 {
		return nil, errors.New("no page response returned")
	}

	decoded, err := base64.StdEncoding.DecodeString(res.PageResponses[0].Content)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 content: %w", err)
	}

	return decoded, nil
}

// FetchWithAutomation executes a built overseerScript and automatically extracts the underlying arbitrary automationResult payload.
func (c *Client) FetchWithAutomation(url string, builder *OverseerScriptBuilder) (interface{}, error) {
	req := &PageRequest{
		URL:            url,
		RenderType:     "automation",
		OverseerScript: builder.Build(),
		OutputAsJson:   true,
	}

	res, err := c.DoPage(req)
	if err != nil {
		return nil, err
	}

	if len(res.PageResponses) == 0 {
		return nil, errors.New("no page response returned")
	}

	if res.PageResponses[0].AutomationResult != nil {
		return res.PageResponses[0].AutomationResult, nil
	}

	return nil, errors.New("automation result was omitted or empty in response")
}

// parseMetadata extracts specific pjsc headers
func parseMetadata(headers http.Header) ResponseMetadata {
	meta := ResponseMetadata{}

	if costStr := headers.Get("pjsc-billing-credit-cost"); costStr != "" {
		if cost, err := strconv.ParseFloat(costStr, 64); err == nil {
			meta.BillingCreditCost = cost
		}
	}

	if statusCodeStr := headers.Get("pjsc-content-status-code"); statusCodeStr != "" {
		if code, err := strconv.Atoi(statusCodeStr); err == nil {
			meta.ContentStatusCode = code
		}
	}

	if doneWhen := headers.Get("pjsc-content-done-when"); doneWhen != "" {
		meta.ContentDoneWhen = doneWhen
	}

	return meta
}
