package scraper

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	phantomjscloud "github.com/amafjarkasi/go-phantomjs"
)

func TestBatchProcessor_ScrapeSimple(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"pageResponses":[{"content":"test1"},{"content":"test2"}],"status":"success"}`))
	}))
	defer mockServer.Close()

	client := phantomjscloud.NewClient("test-key", phantomjscloud.WithEndpoint(mockServer.URL+"/"))
	processor := NewBatchProcessor(client, 1, 2)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	results, err := processor.ScrapeSimple(ctx, []string{"http://example.com/1", "http://example.com/2"})
	if err != nil {
		t.Fatalf("ScrapeSimple failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	if results[0].Response.Content != "test1" {
		t.Errorf("Expected test1, got %s", results[0].Response.Content)
	}
}
