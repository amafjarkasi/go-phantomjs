package persona

import (
	"testing"

	phantomjscloud "github.com/amafjarkasi/go-phantomjs"
	"github.com/amafjarkasi/go-phantomjs/ext/blocklist"
	"github.com/amafjarkasi/go-phantomjs/ext/useragents"
	"github.com/amafjarkasi/go-phantomjs/ext/viewport"
)

func TestEngine_ApplyForURLAttempt_HostAndFallback(t *testing.T) {
	engine := NewEngine().
		Define("us-desktop", Config{
			Proxy:    phantomjscloud.ProxyAnonUS,
			Profile:  useragents.ChromeWindowsProfile(),
			Viewport: viewport.FHD,
			Blockers: blocklist.Lightweight(),
		}).
		Define("ca-mobile", Config{
			Proxy:    phantomjscloud.ProxyAnonCA,
			Profile:  useragents.FirefoxWindowsProfile(),
			Viewport: viewport.MobilePortrait,
		}).
		RouteHost("example.com", "us-desktop", "ca-mobile").
		SetDefault("us-desktop")

	req0 := phantomjscloud.NewPageRequestBuilder("https://example.com/search").Build()
	name0 := engine.ApplyForURLAttempt(req0, 0)
	if name0 != "us-desktop" {
		t.Fatalf("expected first persona us-desktop, got %q", name0)
	}
	if req0.Proxy != phantomjscloud.ProxyAnonUS {
		t.Fatalf("expected proxy %q, got %v", phantomjscloud.ProxyAnonUS, req0.Proxy)
	}
	if req0.RequestSettings.UserAgent == "" {
		t.Fatal("expected user agent from profile")
	}
	if req0.RenderSettings.Viewport == nil || req0.RenderSettings.Viewport.Width != viewport.FHD.Viewport.Width {
		t.Fatalf("expected FHD viewport, got %#v", req0.RenderSettings.Viewport)
	}
	if len(req0.RequestSettings.ResourceModifier) == 0 {
		t.Fatal("expected blockers to be applied")
	}

	req1 := phantomjscloud.NewPageRequestBuilder("https://example.com/search").Build()
	name1 := engine.ApplyForURLAttempt(req1, 1)
	if name1 != "ca-mobile" {
		t.Fatalf("expected fallback persona ca-mobile, got %q", name1)
	}
	if req1.Proxy != phantomjscloud.ProxyAnonCA {
		t.Fatalf("expected proxy %q, got %v", phantomjscloud.ProxyAnonCA, req1.Proxy)
	}
	if req1.RenderSettings.Viewport == nil || req1.RenderSettings.Viewport.Width != viewport.MobilePortrait.Viewport.Width {
		t.Fatalf("expected mobile viewport, got %#v", req1.RenderSettings.Viewport)
	}
}

