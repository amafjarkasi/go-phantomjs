package phantomjscloud

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_Do(t *testing.T) {
	// Start a mock HTTP server
	mockRes := UserResponse{
		Status:  "success",
		Billing: Billing{CreditCost: 1.0},
	}

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)

		var ur UserRequest
		err := json.Unmarshal(body, &ur)
		if err != nil {
			t.Fatalf("Server failed to unmarshal UserRequest payload: %v", err)
		}

		if len(ur.Pages) != 2 {
			t.Errorf("Expected 2 pages, got %d", len(ur.Pages))
		}

		w.Header().Set("pjsc-billing-credit-cost", "1.00")
		w.Header().Set("Content-Type", "application/json")
		b, _ := json.Marshal(mockRes)
		w.Write(b)
	}))
	defer mockServer.Close()

	client := NewClient("test-key")
	// Note: since our struct hardcodes the baseUrl as a constant, we normally can't override it easily in tests without refactoring.
	// For this test, we simply assume Do works if we can intercept the transport or we can just mock a different baseEndpointUrl.

	// Refactoring client.go slightly to allow variable Endpoint is trivial, but keeping the actual URL const is safer for users.
	// To test marshalling properly without an e2e hit:
	client.httpClient.Transport = &ProxyRoundTripper{TargetUrl: mockServer.URL}

	req := &UserRequest{
		Pages: []PageRequest{
			{URL: "http://example.com/one"},
			{URL: "http://example.com/two"},
		},
		OutputAsJson: true,
		Proxy: ProxyOptions{
			Geolocation: "de",
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Client.Do failed: %v", err)
	}

	if resp.Billing.CreditCost != 1.0 {
		t.Errorf("Expected 1.0 credit cost, got %f", resp.Billing.CreditCost)
	}
}

type ProxyRoundTripper struct {
	TargetUrl string
}

func (p *ProxyRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Re-route the request to our mock server
	newReq, _ := http.NewRequest(req.Method, p.TargetUrl, req.Body)
	newReq.Header = req.Header
	return http.DefaultTransport.RoundTrip(newReq)
}

func TestParseMetadata(t *testing.T) {
	headers := http.Header{}
	headers.Set("pjsc-billing-credit-cost", "2.25")
	headers.Set("pjsc-content-status-code", "201")
	headers.Set("pjsc-content-done-when", "load")

	meta := parseMetadata(headers)

	if meta.BillingCreditCost != 2.25 {
		t.Errorf("Expected 2.25, got %f", meta.BillingCreditCost)
	}
	if meta.ContentStatusCode != 201 {
		t.Errorf("Expected 201, got %d", meta.ContentStatusCode)
	}
	if meta.ContentDoneWhen != "load" {
		t.Errorf("Expected load, got %s", meta.ContentDoneWhen)
	}
}

func TestOverseerScriptBuilder(t *testing.T) {
	b := NewOverseerScriptBuilder()
	script := b.AddScriptTag("http://example.com/script.js").
		Evaluate("() => { return 'done'; }").
		WaitForSelector("body").
		Type("input#name", "test user", 100).
		Build()

	expected := "await page.addScriptTag({url: 'http://example.com/script.js'});\n" +
		"await page.evaluate(() => { return 'done'; });\n" +
		"await page.waitForSelector('body');\n" +
		"await page.type('input#name', 'test user',{delay:100});\n"

	if script != expected {
		t.Errorf("Builder output mismatch.\nExpected:\n%s\nGot:\n%s", expected, script)
	}
}
