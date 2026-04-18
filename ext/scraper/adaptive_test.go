package scraper

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	phantomjscloud "github.com/amafjarkasi/go-phantomjs"
	"github.com/amafjarkasi/go-phantomjs/ext/blockpolicy"
	"github.com/amafjarkasi/go-phantomjs/ext/proxy"
)

func TestDoPageWithAdaptiveBlockPolicy_RelaxesAndSucceeds(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req phantomjscloud.UserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		modCount := len(req.Pages[0].RequestSettings.ResourceModifier)
		statusCode := 200
		content := "ok page"
		if modCount > 55 {
			statusCode = 503
			content = "Robot or human?"
		}

		resp := phantomjscloud.UserResponse{
			Status: "success",
			Billing: phantomjscloud.Billing{
				CreditCost: 0,
				QuotaUsage: 0,
			},
			PageResponses: []phantomjscloud.PageResponse{
				{StatusCode: statusCode, Content: content},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		b, _ := json.Marshal(resp)
		_, _ = w.Write(b)
	}))
	defer server.Close()

	client := phantomjscloud.NewClient("test-key", phantomjscloud.WithEndpoint(server.URL+"/"))
	baseReq := phantomjscloud.NewPageRequestBuilder("https://example.com").WithOutputAsJson(true).Build()

	resp, attempts, err := DoPageWithAdaptiveBlockPolicy(
		context.Background(),
		client,
		baseReq,
		blockpolicy.LevelAggressive,
		2,
	)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected response")
	}
	if len(attempts) != 2 {
		t.Fatalf("expected 2 attempts, got %d", len(attempts))
	}
	if attempts[0].Level != blockpolicy.LevelAggressive || attempts[1].Level != blockpolicy.LevelBalanced {
		t.Fatalf("unexpected level sequence: %+v", attempts)
	}
}

func TestDoPageWithAdaptiveBlockPolicy_StopsWhenStillBlocked(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := phantomjscloud.UserResponse{
			Status: "success",
			Billing: phantomjscloud.Billing{
				CreditCost: 0,
				QuotaUsage: 0,
			},
			PageResponses: []phantomjscloud.PageResponse{
				{StatusCode: 503, Content: "captcha required"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		b, _ := json.Marshal(resp)
		_, _ = w.Write(b)
	}))
	defer server.Close()

	client := phantomjscloud.NewClient("test-key", phantomjscloud.WithEndpoint(server.URL+"/"))
	baseReq := phantomjscloud.NewPageRequestBuilder("https://example.com").WithOutputAsJson(true).Build()

	_, attempts, err := DoPageWithAdaptiveBlockPolicy(
		context.Background(),
		client,
		baseReq,
		blockpolicy.LevelAggressive,
		2,
	)
	if err == nil {
		t.Fatal("expected error when still blocked")
	}
	if len(attempts) != 2 {
		t.Fatalf("expected 2 attempts, got %d", len(attempts))
	}
}

func TestDoPageWithRoutingAndAdaptivePolicy_RoutesByAttemptAndSucceeds(t *testing.T) {
	var proxies []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req phantomjscloud.UserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		p, _ := req.Pages[0].Proxy.(string)
		proxies = append(proxies, p)

		statusCode := 503
		content := "captcha required"
		if len(proxies) == 2 {
			statusCode = 200
			content = "ok"
		}

		resp := phantomjscloud.UserResponse{
			Status: "success",
			Billing: phantomjscloud.Billing{
				CreditCost: 0,
				QuotaUsage: 0,
			},
			PageResponses: []phantomjscloud.PageResponse{
				{StatusCode: statusCode, Content: content},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		b, _ := json.Marshal(resp)
		_, _ = w.Write(b)
	}))
	defer server.Close()

	client := phantomjscloud.NewClient("test-key", phantomjscloud.WithEndpoint(server.URL+"/"))
	baseReq := phantomjscloud.NewPageRequestBuilder("https://example.com").WithOutputAsJson(true).Build()
	router := proxy.NewHostRouter(phantomjscloud.ProxyAnonUS).
		RouteHost("example.com", phantomjscloud.ProxyAnonUS, phantomjscloud.ProxyAnonCA)

	resp, attempts, err := DoPageWithRoutingAndAdaptivePolicy(
		context.Background(),
		client,
		baseReq,
		router,
		blockpolicy.LevelAggressive,
		3,
	)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if len(attempts) != 2 {
		t.Fatalf("expected 2 attempts, got %d", len(attempts))
	}
	if len(proxies) != 2 {
		t.Fatalf("expected 2 routed requests, got %d", len(proxies))
	}
	if proxies[0] != "anon-us" || proxies[1] != "anon-ca" {
		t.Fatalf("expected fallback proxies [anon-us, anon-ca], got %#v", proxies)
	}
}
