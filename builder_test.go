package phantomjscloud_test

import (
	"testing"

	phantomjscloud "github.com/jbdt/go-phantomjs"
	"github.com/jbdt/go-phantomjs/ext/useragents"
)

func TestPageRequestBuilder_Defaults(t *testing.T) {
	req := phantomjscloud.NewPageRequestBuilder("https://example.com").Build()

	if req.URL != "https://example.com" {
		t.Errorf("expected URL https://example.com, got %q", req.URL)
	}
}

func TestPageRequestBuilder_WithRenderType(t *testing.T) {
	req := phantomjscloud.NewPageRequestBuilder("https://example.com").
		WithRenderType("jpeg").
		Build()

	if req.RenderType != "jpeg" {
		t.Errorf("expected renderType jpeg, got %q", req.RenderType)
	}
}

func TestPageRequestBuilder_WithProxy(t *testing.T) {
	req := phantomjscloud.NewPageRequestBuilder("https://example.com").
		WithProxy(phantomjscloud.ProxyAnonUS).
		Build()

	if req.Proxy != phantomjscloud.ProxyAnonUS {
		t.Errorf("expected proxy %q, got %v", phantomjscloud.ProxyAnonUS, req.Proxy)
	}
}

func TestPageRequestBuilder_WithProfile(t *testing.T) {
	profile := useragents.ChromeWindowsProfile()
	req := phantomjscloud.NewPageRequestBuilder("https://example.com").
		WithProfile(profile).
		Build()

	if req.RequestSettings.UserAgent != profile.UserAgent {
		t.Errorf("expected UserAgent %q, got %q", profile.UserAgent, req.RequestSettings.UserAgent)
	}
	for k, v := range profile.Headers {
		if req.RequestSettings.CustomHeaders[k] != v {
			t.Errorf("header %q: expected %q, got %q", k, v, req.RequestSettings.CustomHeaders[k])
		}
	}
}

func TestPageRequestBuilder_WithBlocklist(t *testing.T) {
	rules := []phantomjscloud.ResourceModifier{
		{Regex: ".*ads.*", IsBlacklisted: true},
		{Regex: ".*tracker.*", IsBlacklisted: true},
	}
	req := phantomjscloud.NewPageRequestBuilder("https://example.com").
		WithBlocklist(rules).
		Build()

	if len(req.RequestSettings.ResourceModifier) != 2 {
		t.Errorf("expected 2 ResourceModifier rules, got %d", len(req.RequestSettings.ResourceModifier))
	}
}

func TestPageRequestBuilder_WithBlocklist_Appends(t *testing.T) {
	// Calling WithBlocklist twice should accumulate rules, not replace them.
	batch1 := []phantomjscloud.ResourceModifier{{Regex: ".*ads.*", IsBlacklisted: true}}
	batch2 := []phantomjscloud.ResourceModifier{{Regex: ".*fonts.*", IsBlacklisted: true}}

	req := phantomjscloud.NewPageRequestBuilder("https://example.com").
		WithBlocklist(batch1).
		WithBlocklist(batch2).
		Build()

	if len(req.RequestSettings.ResourceModifier) != 2 {
		t.Errorf("expected 2 accumulated rules, got %d", len(req.RequestSettings.ResourceModifier))
	}
}

func TestPageRequestBuilder_WithViewport(t *testing.T) {
	vp := phantomjscloud.Viewport{Width: 1920, Height: 1080}
	req := phantomjscloud.NewPageRequestBuilder("https://example.com").
		WithViewport(vp).
		Build()

	if req.RenderSettings.Viewport == nil {
		t.Fatal("expected non-nil Viewport in RenderSettings")
	}
	if req.RenderSettings.Viewport.Width != 1920 || req.RenderSettings.Viewport.Height != 1080 {
		t.Errorf("expected 1920x1080, got %dx%d",
			req.RenderSettings.Viewport.Width, req.RenderSettings.Viewport.Height)
	}
}

func TestPageRequestBuilder_WithOverseerScriptBuilder(t *testing.T) {
	sb := phantomjscloud.NewOverseerScriptBuilder().
		WaitForSelector("body").
		Goto("https://example.com")

	req := phantomjscloud.NewPageRequestBuilder("https://example.com").
		WithOverseerScriptBuilder(sb).
		Build()

	if req.OverseerScript == "" {
		t.Error("expected non-empty OverseerScript")
	}
}

func TestPageRequestBuilder_WithHeader(t *testing.T) {
	req := phantomjscloud.NewPageRequestBuilder("https://example.com").
		WithHeader("X-Custom", "value").
		WithHeader("Authorization", "Bearer token").
		Build()

	if req.RequestSettings.CustomHeaders["X-Custom"] != "value" {
		t.Error("expected X-Custom header to be set")
	}
	if req.RequestSettings.CustomHeaders["Authorization"] != "Bearer token" {
		t.Error("expected Authorization header to be set")
	}
}

func TestPageRequestBuilder_Build_ReturnsValueCopy(t *testing.T) {
	// Mutating the returned pointer must not affect a subsequent Build() call.
	b := phantomjscloud.NewPageRequestBuilder("https://example.com").WithRenderType("jpeg")
	req1 := b.Build()
	req1.RenderType = "pdf"
	req2 := b.Build()

	if req2.RenderType != "jpeg" {
		t.Errorf("Build() should return an independent copy; got %q", req2.RenderType)
	}
}
