//go:build integration
// +build integration

package phantomjscloud_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	phantomjscloud "github.com/amafjarkasi/go-phantomjs"
	"github.com/amafjarkasi/go-phantomjs/ext/blocklist"
	"github.com/amafjarkasi/go-phantomjs/ext/scraper"
	"github.com/amafjarkasi/go-phantomjs/ext/useragents"
	"github.com/amafjarkasi/go-phantomjs/ext/viewport"
)

const liveTestURL = "https://example.com"

func liveClient(t *testing.T) *phantomjscloud.Client {
	t.Helper()

	apiKey := os.Getenv("PHANTOMJSCLOUD_API_KEY")
	if apiKey == "" {
		t.Skip("PHANTOMJSCLOUD_API_KEY is required for integration tests")
	}

	return phantomjscloud.NewClient(apiKey, phantomjscloud.WithTimeout(120*time.Second))
}

func contentString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func TestLivePageRenderAndExtensions(t *testing.T) {
	client := liveClient(t)

	req := phantomjscloud.NewPageRequestBuilder(liveTestURL).
		WithRenderType("html").
		WithOutputAsJson(true).
		WithProfile(useragents.ChromeWindowsProfile()).
		WithBlocklist(blocklist.Lightweight()).
		WithDoneWhen([]phantomjscloud.DoneWhen{{Event: "load"}}).
		WithRenderSettings(viewport.FHD.AsRenderSettings()).
		Build()

	resp, err := client.DoPage(req)
	if err != nil {
		t.Fatalf("DoPage failed: %v", err)
	}

	body := contentString(resp.Content.Data)
	if !strings.Contains(body, "Example Domain") {
		t.Fatalf("expected Example Domain in rendered HTML, got %q", body)
	}

	if resp.Metadata.ContentStatusCode != 200 {
		t.Fatalf("expected content status 200, got %d", resp.Metadata.ContentStatusCode)
	}

	if len(resp.PageResponses) == 0 {
		t.Fatal("expected at least one page response")
	}

	if len(resp.PageResponses[0].Events) == 0 {
		t.Fatal("expected live JSON response events")
	}
}

func TestLiveConvenienceFetchers(t *testing.T) {
	client := liveClient(t)

	plainText, err := client.FetchPlainText(liveTestURL)
	if err != nil {
		t.Fatalf("FetchPlainText failed: %v", err)
	}
	if !strings.Contains(plainText, "Example Domain") {
		t.Fatalf("expected plain text content, got %q", plainText)
	}

	shotSettings := viewport.Thumbnail1200.AsRenderSettings()
	screenshot, err := client.FetchScreenshot(liveTestURL, "png", &shotSettings)
	if err != nil {
		t.Fatalf("FetchScreenshot failed: %v", err)
	}
	if len(screenshot) == 0 {
		t.Fatal("expected non-empty screenshot bytes")
	}

	pdfSettings := &phantomjscloud.RenderSettings{EmulateMedia: "print"}
	pdfBytes, err := client.RenderRawHTML(
		`<html><body><h1>Live Integration</h1><p>Rendered through PhantomJsCloud</p></body></html>`,
		"pdf",
		pdfSettings,
	)
	if err != nil {
		t.Fatalf("RenderRawHTML failed: %v", err)
	}
	if len(pdfBytes) == 0 {
		t.Fatal("expected non-empty PDF bytes")
	}
}

func TestLiveAutomation(t *testing.T) {
	client := liveClient(t)

	result, err := client.FetchWithAutomation(
		liveTestURL,
		phantomjscloud.NewOverseerScriptBuilder().
			WaitForSelector("h1").
			Evaluate("() => ({ title: document.title, heading: document.querySelector('h1').textContent })"),
	)
	if err != nil {
		t.Fatalf("FetchWithAutomation failed: %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map result, got %T", result)
	}

	if _, ok := resultMap["errors"]; !ok {
		t.Fatalf("expected automation result map to include errors, got %#v", resultMap)
	}
	if _, ok := resultMap["storage"]; !ok {
		t.Fatalf("expected automation result map to include storage, got %#v", resultMap)
	}
}

func TestLiveClickAndWaitForNavigation(t *testing.T) {
	client := liveClient(t)

	script := phantomjscloud.NewOverseerScriptBuilder().
		WaitForSelector("a").
		ClickAndWaitForNavigation("a").
		WaitForUrl("iana.org").
		Build()

	req := phantomjscloud.NewPageRequestBuilder(liveTestURL).
		WithOutputAsJson(true).
		WithOverseerScript(script).
		Build()

	result, err := client.DoPage(req)
	if err != nil {
		t.Fatalf("DoPage failed: %v", err)
	}

	if len(result.PageResponses) == 0 || result.PageResponses[0].FrameData == nil {
		t.Fatal("expected frame data from navigation request")
	}

	if !strings.Contains(result.PageResponses[0].FrameData.Url, "iana.org") {
		t.Fatalf("expected navigation to iana.org, got %q", result.PageResponses[0].FrameData.Url)
	}
}

func TestLiveManualWait(t *testing.T) {
	client := liveClient(t)

	result, err := client.FetchWithAutomation(
		liveTestURL,
		phantomjscloud.NewOverseerScriptBuilder().
			WaitForSelector("h1").
			ManualWait().
			Done(),
	)
	if err != nil {
		t.Fatalf("FetchWithAutomation failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil automation result")
	}
}

func TestLiveBatchScraper(t *testing.T) {
	client := liveClient(t)
	processor := scraper.NewBatchProcessor(client, 1, 1)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	requests := []phantomjscloud.PageRequest{
		*phantomjscloud.NewPageRequestBuilder(liveTestURL).WithOutputAsJson(true).Build(),
		*phantomjscloud.NewPageRequestBuilder(liveTestURL).WithOutputAsJson(true).Build(),
	}

	results, err := processor.ScrapeAll(ctx, requests)
	if err != nil {
		t.Fatalf("ScrapeAll failed: %v", err)
	}
	if len(results) != len(requests) {
		t.Fatalf("expected %d results, got %d", len(requests), len(results))
	}
	for i, res := range results {
		if res.Error != nil {
			t.Fatalf("result %d returned error: %v", i, res.Error)
		}
	}
}

func TestLivePostData(t *testing.T) {
	client := liveClient(t)

	req := phantomjscloud.NewPageRequestBuilder("https://httpbin.org/anything").
		WithOutputAsJson(true).
		Build()
	req.UrlSettings = &phantomjscloud.UrlSettings{
		Operation: "POST",
		Data:      `{"hello":"world"}`,
	}

	resp, err := client.DoPage(req)
	if err != nil {
		t.Fatalf("DoPage failed: %v", err)
	}

	body := ""
	if resp.PageResponses[0].FrameData != nil {
		body = resp.PageResponses[0].FrameData.Content
	}
	if !strings.Contains(body, `"method": "POST"`) || !strings.Contains(body, `"form": {`) {
		t.Fatalf("expected posted body in echoed response, got %q", body)
	}
}

func TestLiveBasicAuth(t *testing.T) {
	client := liveClient(t)

	req := phantomjscloud.NewPageRequestBuilder("https://httpbin.org/basic-auth/user/pass").
		WithOutputAsJson(true).
		Build()
	req.RequestSettings.Authentication = &phantomjscloud.Authentication{
		UserName: "user",
		Password: "pass",
	}

	resp, err := client.DoPage(req)
	if err != nil {
		t.Fatalf("DoPage failed: %v", err)
	}

	body := ""
	if resp.PageResponses[0].FrameData != nil {
		body = resp.PageResponses[0].FrameData.Content
	}
	if !strings.Contains(body, "\"authenticated\": true") {
		t.Fatalf("expected authenticated=true in response, got %q", body)
	}
}

func TestLiveCookies(t *testing.T) {
	client := liveClient(t)

	req := phantomjscloud.NewPageRequestBuilder("https://httpbin.org/cookies").
		WithOutputAsJson(true).
		Build()
	req.RequestSettings.Cookies = []phantomjscloud.Cookie{
		{Name: "live_cookie", Value: "works", Domain: "httpbin.org"},
	}

	resp, err := client.DoPage(req)
	if err != nil {
		t.Fatalf("DoPage failed: %v", err)
	}

	body := ""
	if resp.PageResponses[0].FrameData != nil {
		body = resp.PageResponses[0].FrameData.Content
	}
	if !strings.Contains(body, `"live_cookie": "works"`) {
		t.Fatalf("expected cookie in echoed response, got %q", body)
	}
}

func TestLiveSuppressJson(t *testing.T) {
	client := liveClient(t)

	req := phantomjscloud.NewPageRequestBuilder(liveTestURL).
		WithOutputAsJson(true).
		WithSuppressJson([]string{"originalRequest"}).
		Build()

	resp, err := client.DoPage(req)
	if err != nil {
		t.Fatalf("DoPage failed: %v", err)
	}

	if original, ok := resp.OriginalRequest.(string); !ok || !strings.Contains(original, "OUTPUT SUPPRESSED") {
		t.Fatalf("expected originalRequest suppression marker, got %#v", resp.OriginalRequest)
	}
	if len(resp.PageResponses) == 0 {
		t.Fatal("expected page responses to remain present")
	}
}
