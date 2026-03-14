package phantomjscloud

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchScreenshot_InvalidBase64(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := UserResponse{
			Status: "success",
			PageResponses: []PageResponse{
				{Content: "invalid-base64-!!!"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	client := NewClient("test-key", WithEndpoint(mockServer.URL+"/"))

	_, err := client.FetchScreenshot("http://example.com", "png", nil)
	if err == nil {
		t.Fatal("expected error for invalid base64 content, got nil")
	}
	if !strings.Contains(err.Error(), "failed to decode base64 content") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestFetchPDF_InvalidBase64(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := UserResponse{
			Status: "success",
			PageResponses: []PageResponse{
				{Content: "invalid-base64-!!!"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	client := NewClient("test-key", WithEndpoint(mockServer.URL+"/"))

	_, err := client.FetchPDF("http://example.com", nil)
	if err == nil {
		t.Fatal("expected error for invalid base64 content, got nil")
	}
	if !strings.Contains(err.Error(), "failed to decode base64 content") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRenderRawHTML_InvalidBase64(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := UserResponse{
			Status: "success",
			PageResponses: []PageResponse{
				{Content: "invalid-base64-!!!"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	client := NewClient("test-key", WithEndpoint(mockServer.URL+"/"))

	_, err := client.RenderRawHTML("<html></html>", "png", nil)
	if err == nil {
		t.Fatal("expected error for invalid base64 content, got nil")
	}
	if !strings.Contains(err.Error(), "failed to decode base64 content") {
		t.Errorf("unexpected error message: %v", err)
	}
}
