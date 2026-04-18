package scraper

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	phantomjscloud "github.com/amafjarkasi/go-phantomjs"
	"github.com/amafjarkasi/go-phantomjs/ext/blockpolicy"
	"github.com/amafjarkasi/go-phantomjs/ext/persona"
	"github.com/amafjarkasi/go-phantomjs/ext/proxy"
	"github.com/amafjarkasi/go-phantomjs/ext/session"
	"github.com/amafjarkasi/go-phantomjs/ext/useragents"
	"github.com/amafjarkasi/go-phantomjs/ext/viewport"
)

func TestDoPageWithChallengeOrchestration_UsesPersonaAndSessionAcrossAttempts(t *testing.T) {
	var attempts []phantomjscloud.UserRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req phantomjscloud.UserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		attempts = append(attempts, req)

		if len(attempts) == 1 {
			resp := phantomjscloud.UserResponse{
				Status:  "success",
				Billing: phantomjscloud.Billing{CreditCost: 0, QuotaUsage: 0},
				PageResponses: []phantomjscloud.PageResponse{
					{
						StatusCode: 503,
						Content:    "captcha required",
						Cookies: []phantomjscloud.Cookie{
							{Name: "challenge", Value: "passed", Domain: ".example.com"},
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		resp := phantomjscloud.UserResponse{
			Status:  "success",
			Billing: phantomjscloud.Billing{CreditCost: 0, QuotaUsage: 0},
			PageResponses: []phantomjscloud.PageResponse{
				{StatusCode: 200, Content: "ok"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := phantomjscloud.NewClient("test-key", phantomjscloud.WithEndpoint(server.URL+"/"))
	baseReq := phantomjscloud.NewPageRequestBuilder("https://example.com/products").WithOutputAsJson(true).Build()

	engine := persona.NewEngine().
		Define("desktop-us", persona.Config{
			Proxy:    phantomjscloud.ProxyAnonUS,
			Profile:  useragents.ChromeWindowsProfile(),
			Viewport: viewport.FHD,
		}).
		RouteHost("example.com", "desktop-us")

	store := session.NewStore()
	resp, trace, err := DoPageWithChallengeOrchestration(
		context.Background(),
		client,
		baseReq,
		ChallengeOrchestrationOptions{
			Persona:     engine,
			Session:     store,
			StartLevel:  blockpolicy.LevelAggressive,
			MaxAttempts: 2,
		},
	)
	if err != nil {
		t.Fatalf("expected orchestration success, got error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if len(trace) != 2 {
		t.Fatalf("expected 2 attempts, got %d", len(trace))
	}
	if trace[0].Persona != "desktop-us" {
		t.Fatalf("expected persona desktop-us, got %q", trace[0].Persona)
	}
	if !trace[0].Blocked {
		t.Fatal("expected first attempt to be marked blocked")
	}
	if trace[1].Blocked {
		t.Fatal("expected second attempt to be marked non-blocked")
	}
	pb, ok := trace[0].Proxy.(phantomjscloud.ProxyBuiltin)
	if !ok {
		t.Fatalf("expected trace proxy builtin type, got %#v", trace[0].Proxy)
	}
	if pb.Location != "us" {
		t.Fatalf("expected trace proxy location us, got %#v", pb)
	}

	if len(attempts) != 2 {
		t.Fatalf("expected 2 request attempts, got %d", len(attempts))
	}
	if attempts[0].Pages[0].Proxy != "anon-us" {
		t.Fatalf("expected normalized proxy anon-us in first attempt, got %v", attempts[0].Pages[0].Proxy)
	}
	if attempts[0].Pages[0].RequestSettings.UserAgent == "" {
		t.Fatal("expected persona user agent in first attempt")
	}
	if len(attempts[1].Pages[0].RequestSettings.Cookies) != 1 {
		t.Fatalf("expected persisted cookie in second attempt, got %#v", attempts[1].Pages[0].RequestSettings.Cookies)
	}
	if attempts[1].Pages[0].RequestSettings.Cookies[0].Name != "challenge" {
		t.Fatalf("expected challenge cookie, got %#v", attempts[1].Pages[0].RequestSettings.Cookies)
	}
}

func TestDoPageWithChallengeOrchestration_ReportsRouterHealthBetweenAttempts(t *testing.T) {
	var seenProxies []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req phantomjscloud.UserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if len(req.Pages) == 0 {
			t.Fatal("missing page request")
		}
		if p, ok := req.Pages[0].Proxy.(string); ok {
			seenProxies = append(seenProxies, p)
		} else {
			seenProxies = append(seenProxies, "")
		}

		resp := phantomjscloud.UserResponse{
			Status:  "success",
			Billing: phantomjscloud.Billing{CreditCost: 0, QuotaUsage: 0},
			PageResponses: []phantomjscloud.PageResponse{
				{StatusCode: 503, Content: "captcha required"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := phantomjscloud.NewClient("test-key", phantomjscloud.WithEndpoint(server.URL+"/"))
	baseReq := phantomjscloud.NewPageRequestBuilder("https://example.com/products").WithOutputAsJson(true).Build()
	router := proxy.NewHealthRouter().
		RouteHost("example.com", "anon-us", "anon-ca", "anon-nl")
	// Pre-bias scores so anon-ca and anon-nl are tied healthiest; round-robin picks anon-ca first.
	router.ReportFailure("https://example.com", "anon-us")
	router.ReportFailure("https://example.com", "anon-us")

	_, trace, err := DoPageWithChallengeOrchestration(
		context.Background(),
		client,
		baseReq,
		ChallengeOrchestrationOptions{
			Router:      router,
			StartLevel:  blockpolicy.LevelAggressive,
			MaxAttempts: 2,
		},
	)
	if err == nil {
		t.Fatal("expected blocked error")
	}

	if len(seenProxies) != 2 {
		t.Fatalf("expected 2 proxy observations, got %d", len(seenProxies))
	}
	if seenProxies[0] != "anon-ca" {
		t.Fatalf("expected first attempt to use anon-ca, got %q", seenProxies[0])
	}
	if seenProxies[1] != "anon-nl" {
		t.Fatalf("expected second attempt to reflect health update and use anon-nl, got %q", seenProxies[1])
	}
	if len(trace) != 2 {
		t.Fatalf("expected 2 trace attempts, got %d", len(trace))
	}
	if len(trace[0].Health) != 3 {
		t.Fatalf("expected health snapshot of 3 proxies, got %#v", trace[0].Health)
	}
	if trace[0].Health[0].Proxy != "anon-ca" {
		t.Fatalf("expected healthiest proxy anon-ca in trace[0], got %#v", trace[0].Health[0])
	}
	if len(trace[1].Health) != 3 {
		t.Fatalf("expected health snapshot of 3 proxies in trace[1], got %#v", trace[1].Health)
	}
	if trace[1].Health[0].Proxy != "anon-nl" {
		t.Fatalf("expected healthiest proxy anon-nl in trace[1], got %#v", trace[1].Health[0])
	}
}
