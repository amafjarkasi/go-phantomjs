package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	phantomjscloud "github.com/amafjarkasi/go-phantomjs"
	"github.com/amafjarkasi/go-phantomjs/ext/blockpolicy"
	"github.com/amafjarkasi/go-phantomjs/ext/proxy"
	"github.com/amafjarkasi/go-phantomjs/ext/useragents"
)

type modeResult struct {
	DurationMs int64
	Attempts   int
	Cost       float64
	StatusCode int
	Blocked    bool
	Err        error
}

type row struct {
	Name   string
	URL    string
	Before modeResult
	After  modeResult
}

func main() {
	key := os.Getenv("PHANTOMJSCLOUD_API_KEY")
	if key == "" {
		fmt.Println("PHANTOMJSCLOUD_API_KEY is required")
		os.Exit(1)
	}

	client := phantomjscloud.NewClient(key, phantomjscloud.WithTimeout(120*time.Second))
	profile := useragents.ChromeWindowsProfile()

	router := proxy.NewHostRouter(phantomjscloud.ProxyAnonUS).
		RouteHost("amazon.com", phantomjscloud.ProxyAnonUS, phantomjscloud.ProxyAnonCA).
		RouteHost("walmart.com", phantomjscloud.ProxyAnonUS, phantomjscloud.ProxyAnonCA).
		RouteHost("target.com", phantomjscloud.ProxyAnonUS, phantomjscloud.ProxyAnonCA)

	targets := []struct {
		Name string
		URL  string
	}{
		{"amazon-home", "https://www.amazon.com"},
		{"amazon-search-page2", "https://www.amazon.com/s?k=laptop&page=2"},
		{"walmart-home", "https://www.walmart.com"},
		{"walmart-search-page2", "https://www.walmart.com/search?q=laptop&page=2"},
		{"target-home", "https://www.target.com"},
		{"target-search-page2", "https://www.target.com/s?searchTerm=laptop&page=2"},
	}

	results := make([]row, 0, len(targets))
	for _, t := range targets {
		base := phantomjscloud.NewPageRequestBuilder(t.URL).
			WithRenderType("html").
			WithOutputAsJson(true).
			WithProfile(profile).
			Build()

		before := runBefore(context.Background(), client, base)
		after := runAfter(context.Background(), client, base, router)

		results = append(results, row{
			Name:   t.Name,
			URL:    t.URL,
			Before: before,
			After:  after,
		})
	}

	printReport(results)
}

func runBefore(ctx context.Context, client *phantomjscloud.Client, req *phantomjscloud.PageRequest) modeResult {
	start := time.Now()
	resp, err := client.DoPageContext(ctx, req)
	ms := time.Since(start).Milliseconds()
	if err != nil {
		return modeResult{DurationMs: ms, Attempts: 1, Err: err}
	}
	code := resp.Metadata.ContentStatusCode
	if code == 0 && len(resp.PageResponses) > 0 {
		code = resp.PageResponses[0].StatusCode
	}
	return modeResult{
		DurationMs: ms,
		Attempts:   1,
		Cost:       resp.Metadata.BillingCreditCost,
		StatusCode: code,
		Blocked:    blockpolicy.LooksBlocked(resp),
	}
}

func runAfter(
	ctx context.Context,
	client *phantomjscloud.Client,
	baseReq *phantomjscloud.PageRequest,
	router proxy.URLProxyFallbackProvider,
) modeResult {
	start := time.Now()
	level := blockpolicy.LevelAggressive
	totalCost := 0.0
	maxAttempts := 3

	for attempt := 0; attempt < maxAttempts; attempt++ {
		req := *baseReq
		blockpolicy.Apply(&req, level)
		req.Proxy = router.GetProxyForURLAttempt(req.URL, attempt)

		resp, err := client.DoPageContext(ctx, &req)
		if err != nil {
			if attempt == maxAttempts-1 {
				return modeResult{
					DurationMs: time.Since(start).Milliseconds(),
					Attempts:   attempt + 1,
					Cost:       totalCost,
					Err:        err,
				}
			}
			level = blockpolicy.NextLevel(level)
			continue
		}

		totalCost += resp.Metadata.BillingCreditCost
		code := resp.Metadata.ContentStatusCode
		if code == 0 && len(resp.PageResponses) > 0 {
			code = resp.PageResponses[0].StatusCode
		}
		blocked := blockpolicy.LooksBlocked(resp)
		if !blocked || attempt == maxAttempts-1 {
			return modeResult{
				DurationMs: time.Since(start).Milliseconds(),
				Attempts:   attempt + 1,
				Cost:       totalCost,
				StatusCode: code,
				Blocked:    blocked,
			}
		}

		level = blockpolicy.NextLevel(level)
	}

	return modeResult{
		DurationMs: time.Since(start).Milliseconds(),
		Attempts:   maxAttempts,
		Cost:       totalCost,
		Err:        fmt.Errorf("after mode failed unexpectedly"),
	}
}

func printReport(rows []row) {
	fmt.Println("A/B Live Test: before vs after (router + adaptive block policy)")
	fmt.Println(strings.Repeat("=", 78))

	var beforeOK, afterOK, beforeBlocked, afterBlocked int
	var beforeMs, afterMs int64
	var beforeCost, afterCost float64
	var afterAttempts int

	for _, r := range rows {
		beforeErr := ""
		if r.Before.Err != nil {
			beforeErr = oneLineErr(r.Before.Err)
		}
		afterErr := ""
		if r.After.Err != nil {
			afterErr = oneLineErr(r.After.Err)
		}

		fmt.Printf(
			"%s\n  BEFORE: status=%d blocked=%v attempts=%d ms=%d cost=%.4f err=%q\n  AFTER : status=%d blocked=%v attempts=%d ms=%d cost=%.4f err=%q\n",
			r.Name,
			r.Before.StatusCode, r.Before.Blocked, r.Before.Attempts, r.Before.DurationMs, r.Before.Cost, beforeErr,
			r.After.StatusCode, r.After.Blocked, r.After.Attempts, r.After.DurationMs, r.After.Cost, afterErr,
		)

		if r.Before.Err == nil {
			beforeOK++
		}
		if r.After.Err == nil {
			afterOK++
		}
		if r.Before.Blocked {
			beforeBlocked++
		}
		if r.After.Blocked {
			afterBlocked++
		}
		beforeMs += r.Before.DurationMs
		afterMs += r.After.DurationMs
		beforeCost += r.Before.Cost
		afterCost += r.After.Cost
		afterAttempts += r.After.Attempts
	}

	n := float64(len(rows))
	fmt.Println(strings.Repeat("-", 78))
	fmt.Printf("SUMMARY\n")
	fmt.Printf("  BEFORE: ok=%d/%d blocked=%d avg_ms=%.1f avg_cost=%.4f\n",
		beforeOK, len(rows), beforeBlocked, float64(beforeMs)/n, beforeCost/n)
	fmt.Printf("  AFTER : ok=%d/%d blocked=%d avg_ms=%.1f avg_cost=%.4f avg_attempts=%.2f\n",
		afterOK, len(rows), afterBlocked, float64(afterMs)/n, afterCost/n, float64(afterAttempts)/n)
	fmt.Printf("  DELTA : blocked=%d -> %d, avg_ms %.1f -> %.1f, avg_cost %.4f -> %.4f\n",
		beforeBlocked, afterBlocked, float64(beforeMs)/n, float64(afterMs)/n, beforeCost/n, afterCost/n)
}

func oneLineErr(err error) string {
	if err == nil {
		return ""
	}
	msg := strings.ReplaceAll(err.Error(), "\r", " ")
	msg = strings.ReplaceAll(msg, "\n", " ")
	if len(msg) > 220 {
		return msg[:220] + "..."
	}
	return msg
}
