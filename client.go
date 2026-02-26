package phantomjscloud

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const (
	baseEndpointUrl    = "https://phantomjscloud.com/api/browser/v2/"
	defaultHTTPTimeout = 120 * time.Second
)

// ClientOption is a functional option for configuring a Client.
type ClientOption func(*Client)

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

// Client is a PhantomJsCloud API client.
type Client struct {
	apiKey     string
	endpoint   string
	httpClient *http.Client
}

// NewClient creates a new Client using the provided API key.
// Passing an empty string will use the demo key "a-demo-key-with-low-quota-per-ip-address" (not recommended for production).
// Use functional options to customise the underlying HTTP client:
//
//	client := phantomjscloud.NewClient("YOUR_KEY",
//	    phantomjscloud.WithTimeout(60*time.Second),
//	)
func NewClient(apiKey string, opts ...ClientOption) *Client {
	if apiKey == "" {
		apiKey = "a-demo-key-with-low-quota-per-ip-address"
	}
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
func (c *Client) DoContext(ctx context.Context, req *UserRequest) (*UserResponseWithMeta, error) {
	endpoint := c.endpoint + c.apiKey + "/"

	// Use io.Pipe to stream the request body instead of buffering it all in memory.
	pr, pw := io.Pipe()

	// Encode JSON in a goroutine
	go func() {
		err := json.NewEncoder(pw).Encode(req)
		pw.CloseWithError(err)
	}()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, pr)
	if err != nil {
		pr.CloseWithError(err)
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		pr.CloseWithError(err)
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode >= 400 && httpResp.StatusCode < 600 {
		bodyBytes, err := io.ReadAll(httpResp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
		return nil, fmt.Errorf("phantomjscloud returned HTTP Status %d: %s", httpResp.StatusCode, string(bodyBytes))
	}

	var userResp UserResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&userResp); err != nil {
		return nil, fmt.Errorf("failed to decode response payload: %w", err)
	}

	result := &UserResponseWithMeta{
		UserResponse: userResp,
		Metadata:     parseMetadata(httpResp.Header),
	}

	return result, nil
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
