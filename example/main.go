package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	phantomjscloud "github.com/amafjarkasi/go-phantomjs"
	"github.com/amafjarkasi/go-phantomjs/ext/blocklist"
	"github.com/amafjarkasi/go-phantomjs/ext/proxy"
	"github.com/amafjarkasi/go-phantomjs/ext/scraper"
	"github.com/amafjarkasi/go-phantomjs/ext/useragents"
)

func main() {
	apiKey := os.Getenv("PHANTOMJSCLOUD_API_KEY")
	if apiKey == "" {
		log.Fatal("PHANTOMJSCLOUD_API_KEY environment variable is required")
	}

	// ── 1. Advanced Client with Rotation & Logging ───────────────────────────
	fmt.Println("=== 1. Advanced Client Setup ===")

	// Round-robin proxy rotation pool
	proxies := proxy.NewRotatingProxyProvider(
		phantomjscloud.ProxyAnonUS,
		phantomjscloud.ProxyAnonDE,
	)

	client := phantomjscloud.NewClient(apiKey,
		phantomjscloud.WithRetry(phantomjscloud.DefaultRetryConfig),
		phantomjscloud.WithProxyProvider(proxies),
		phantomjscloud.WithLogger(slog.Default()),
		phantomjscloud.WithInterceptor(phantomjscloud.LoggingInterceptor(slog.Default())),
	)

	// ── 2. Declarative Data Extraction ───────────────────────────────────────
	fmt.Println("\n=== 2. Declarative Data Extraction ===")

	extractScript := phantomjscloud.NewOverseerScriptBuilder().
		Goto("https://news.ycombinator.com").
		WaitUntilVisible(".hnname").
		Extract(map[string]string{
			"site_name":   ".hnname",
			"first_story": ".titleline > a",
		}).
		ExtractLinks().
		ExtractMetaTags().
		Build()

	req := phantomjscloud.NewPageRequestBuilder("about:blank").
		WithOverseerScript(extractScript).
		Build()

	resp, err := client.DoPage(req)
	if err != nil {
		log.Printf("Extraction error: %v\n", err)
	} else {
		var data struct {
			SiteName   string `json:"site_name"`
			FirstStory string `json:"first_story"`
		}
		// Using new type-safe extraction helper
		if err := resp.PageResponses[0].GetAutomationResultAs(&data); err == nil {
			fmt.Printf("Extracted: Site=%s, First Story=%s\n", data.SiteName, data.FirstStory)
		}
	}

	// ── 3. High-Level Batch Scraper ──────────────────────────────────────────
	fmt.Println("\n=== 3. High-Level Batch Scraper ===")

	processor := scraper.NewBatchProcessor(client, 2, 5)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	results := processor.Scrape(ctx, []phantomjscloud.PageRequest{
		*phantomjscloud.NewPageRequestBuilder("https://google.com").WithBlocklist(blocklist.Lightweight()).Build(),
		*phantomjscloud.NewPageRequestBuilder("https://github.com").WithProfile(useragents.ChromeWindowsProfile()).Build(),
	})

	for res := range results {
		if res.Error != nil {
			fmt.Printf("Batch Error [%s]: %v\n", res.Request.URL, res.Error)
		} else {
			fmt.Printf("Batch Success [%s]: %d bytes, success=%t\n",
				res.Request.URL, len(res.Response.GetContent()), res.Response.IsSuccess())
		}
	}
}
