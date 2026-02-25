package phantomjscloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

const (
	baseEndpointUrl = "https://phantomjscloud.com/api/browser/v2/"
)

// Client is a PhantomJsCloud API client.
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new Client using the provided API key.
// Passing an empty string will use the demo key "a-demo-key-with-low-quota-per-ip-address" (not recommended for production).
func NewClient(apiKey string) *Client {
	if apiKey == "" {
		apiKey = "a-demo-key-with-low-quota-per-ip-address"
	}
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

// UserResponseWithMeta wraps the UserResponse API payload along with the HTTP Response metadata headers.
type UserResponseWithMeta struct {
	UserResponse
	Metadata ResponseMetadata
}

// DoPage is a convenience method that wraps a single PageRequest inside a UserRequest.
func (c *Client) DoPage(req *PageRequest) (*UserResponseWithMeta, error) {
	uReq := &UserRequest{
		Pages: []PageRequest{*req},
	}
	return c.Do(uReq)
}

// Do serializes a UserRequest, performs the HTTP POST to PhantomJsCloud, and parses the response.
func (c *Client) Do(req *UserRequest) (*UserResponseWithMeta, error) {
	endpoint := baseEndpointUrl + c.apiKey + "/"

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer httpResp.Body.Close()

	bodyBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if httpResp.StatusCode >= 400 && httpResp.StatusCode < 600 {
		return nil, fmt.Errorf("phantomjscloud returned HTTP Status %d: %s", httpResp.StatusCode, string(bodyBytes))
	}

	var userResp UserResponse
	if err := json.Unmarshal(bodyBytes, &userResp); err != nil {
		return nil, fmt.Errorf("failed to decode response payload: %w", err)
	}

	// Build the response with metadata
	result := &UserResponseWithMeta{
		UserResponse: userResp,
		Metadata:     parseMetadata(httpResp.Header),
	}

	return result, nil
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
