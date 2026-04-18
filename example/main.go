package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	phantomjscloud "github.com/amafjarkasi/go-phantomjs"
	"github.com/amafjarkasi/go-phantomjs/ext/blocklist"
	"github.com/amafjarkasi/go-phantomjs/ext/scraper"
	"github.com/amafjarkasi/go-phantomjs/ext/useragents"
	"github.com/amafjarkasi/go-phantomjs/ext/viewport"
)

func main() {
	apiKey := os.Getenv("PHANTOMJSCLOUD_API_KEY")
	if apiKey == "" {
		log.Fatal("PHANTOMJSCLOUD_API_KEY environment variable is required")
	}

	// ── 1. Pro Client Setup ──────────────────────────────────────────────────
	// Configure with automatic exponential backoff retries and a logging interceptor.
	fmt.Println("=== 1. Pro Client Setup (Retries + Interceptors) ===")

	logger := func(req *http.Request, next func(*http.Request) (*http.Response, error)) (*http.Response, error) {
		start := time.Now()
		resp, err := next(req)
		duration := time.Since(start)
		if err == nil {
			fmt.Printf("[INTERCEPTOR] %s %s -> %d (%v)\n", req.Method, req.URL, resp.StatusCode, duration)
		}
		return resp, err
	}

	client := phantomjscloud.NewClient(apiKey,
		phantomjscloud.WithRetry(phantomjscloud.DefaultRetryConfig),
		phantomjscloud.WithInterceptor(logger),
	)

	// ── 2. Advanced Interaction ──────────────────────────────────────────────
	// Using the new OverseerScriptBuilder helpers for complex UI tasks.
	fmt.Println("\n=== 2. Advanced Interaction (New Builder Helpers) ===")

	advancedScript := phantomjscloud.NewOverseerScriptBuilder().
		Goto("https://example.com").
		WaitUntilVisible("h1").
		HighlightElement("h1").
		ScrollToElement("footer").
		ClickByText("More information...").
		WaitUntilHidden(".loader").
		RenderScreenshot(true).
		Build()

	fmt.Println("Advanced script sample:\n", advancedScript)

	// ── 3. High-Level Batch Scraper ──────────────────────────────────────────
	// Using the new ext/scraper package to process multiple URLs efficiently.
	fmt.Println("\n=== 3. High-Level Batch Scraper (Concurrency + Batching) ===")

	processor := scraper.NewBatchProcessor(client, 2, 5) // 2 concurrent requests, 5 pages per request

	urls := []string{
		"https://google.com",
		"https://github.com",
		"https://golang.org",
		"https://news.ycombinator.com",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	results := processor.Scrape(ctx, []phantomjscloud.PageRequest{
		*phantomjscloud.NewPageRequestBuilder(urls[0]).WithBlocklist(blocklist.Lightweight()).Build(),
		*phantomjscloud.NewPageRequestBuilder(urls[1]).WithProfile(useragents.ChromeWindowsProfile()).Build(),
		*phantomjscloud.NewPageRequestBuilder(urls[2]).WithViewport(viewport.FHD.Viewport).Build(),
		*phantomjscloud.NewPageRequestBuilder(urls[3]).WithRenderType("plainText").Build(),
	})

	for res := range results {
		if res.Error != nil {
			fmt.Printf("Batch Scrape Error [%s]: %v\n", res.Request.URL, res.Error)
			continue
		}
		// Using new PageResponse helper methods
		fmt.Printf("Batch Scrape Success [%s]: %d bytes, success=%t, cost=%.4f\n",
			res.Request.URL, len(res.Response.GetContent()), res.Response.IsSuccess(), res.Metadata.BillingCreditCost)
	}

	// ── Original Example A: Full Stealth Scrape ──────────────────────────────
	fmt.Println("\n=== 4. Full stealth scrape ===")

	profile := useragents.ChromeWindowsProfile()
	stealthScript := phantomjscloud.NewOverseerScriptBuilder().
		UseProfile(profile).
		ApplyStealth().
		ApplyViewport(viewport.FHD.Viewport).
		Goto("https://bot.sannysoft.com").
		RenderScreenshot(true).
		Build()

	stealthReq := phantomjscloud.NewPageRequestBuilder("about:blank").
		WithRenderType("jpeg").
		WithProfile(profile).
		WithBlocklist(blocklist.Lightweight()).
		WithOverseerScript(stealthScript).
		Build()

	resp, err := client.DoPage(stealthReq)
	if err != nil {
		log.Printf("Stealth scrape error: %v\n", err)
	} else {
		fmt.Printf("Stealth scrape: status=%s cost=%.4f credits\n",
			resp.Status, resp.Metadata.BillingCreditCost)
	}
}
