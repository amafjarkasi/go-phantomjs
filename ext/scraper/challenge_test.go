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

