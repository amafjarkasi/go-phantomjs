package phantomjscloud

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkClient_Do(b *testing.B) {
	// Setup a mock server that reads the body to simulate network consumption
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "success", "pageResponses": [{"content": "ok"}]}`))
	}))
	defer server.Close()

	// Create a client pointing to the mock server
	c := NewClient("test-key")

	// We need to override the transport to redirect requests to our mock server
	c.httpClient.Transport = &BenchmarkRoundTripper{TargetURL: server.URL}

	// Create a large request
	// We use a large Content string to simulate a heavy payload (e.g. ~10MB)
	payloadSize := 10 * 1024 * 1024 // 10MB
	largeContent := createLargeString(payloadSize)

	req := &UserRequest{
		Pages: []PageRequest{
			{
				URL:     "https://example.com",
				Content: largeContent,
				RenderSettings: RenderSettings{
					Viewport: &Viewport{
						Width:  1920,
						Height: 1080,
					},
				},
			},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := c.Do(req)
		if err != nil {
			b.Fatalf("Do failed: %v", err)
		}
	}
}

func createLargeString(size int) string {
	// Create a string of 'a' of length size
	// Efficiently create it
	b := make([]byte, size)
	for i := range b {
		b[i] = 'a'
	}
	return string(b)
}

// BenchmarkRoundTripper redirects all requests to the TargetURL
type BenchmarkRoundTripper struct {
	TargetURL string
}

func (t *BenchmarkRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Create a new request to the target URL
	// We must not read the body here because it might be a pipe that can only be read once.
	// But http.NewRequest with a body argument wraps it.
	// To preserve the body stream, we should use the existing body.

	targetReq, err := http.NewRequest(req.Method, t.TargetURL, req.Body)
	if err != nil {
		return nil, err
	}
	// Copy headers
	targetReq.Header = req.Header

	// We use DefaultTransport to talk to the httptest server (which is local)
	return http.DefaultTransport.RoundTrip(targetReq)
}
