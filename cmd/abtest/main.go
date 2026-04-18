package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	phantomjscloud "github.com/amafjarkasi/go-phantomjs"
	"github.com/amafjarkasi/go-phantomjs/ext/blockpolicy"
	"github.com/amafjarkasi/go-phantomjs/ext/proxy"
	"github.com/amafjarkasi/go-phantomjs/ext/scraper"
	"github.com/amafjarkasi/go-phantomjs/ext/useragents"
)

type modeResult struct {
	DurationMs int64
	Attempts   int
	Cost       float64
	StatusCode int
	Blocked    bool
	Err        error
	DebugLines []string
}

type row struct {
	Name   string
	URL    string
	Before modeResult
	After  modeResult
}

type exportRun struct {
	GeneratedAt string      `json:"generatedAt"`
	Rows        []exportRow `json:"rows"`
}

type exportRow struct {
	Name   string         `json:"name"`
	URL    string         `json:"url"`
	Before exportMode     `json:"before"`
	After  exportMode     `json:"after"`
	Debug  []string       `json:"debug,omitempty"`
}

type exportMode struct {
	StatusCode int    `json:"statusCode"`
	Blocked    bool   `json:"blocked"`
	Attempts   int    `json:"attempts"`
	DurationMs int64  `json:"durationMs"`
	Cost       float64 `json:"cost"`
	Error      string `json:"error,omitempty"`
}

func main() {
	key := os.Getenv("PHANTOMJSCLOUD_API_KEY")
	if key == "" {
		fmt.Println("PHANTOMJSCLOUD_API_KEY is required")
		os.Exit(1)
	}

	client := phantomjscloud.NewClient(key, phantomjscloud.WithTimeout(120*time.Second))
	profile := useragents.ChromeWindowsProfile()

	router := proxy.NewHealthRouter(phantomjscloud.ProxyAnonUS).
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
	if err := exportIfRequested(results); err != nil {
		fmt.Printf("EXPORT ERROR: %s\n", oneLineErr(err))
	}
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
	router *proxy.HealthRouter,
) modeResult {
	start := time.Now()
	resp, attempts, err := scraper.DoPageWithChallengeOrchestration(
		ctx,
		client,
		baseReq,
		scraper.ChallengeOrchestrationOptions{
			Router:      router,
			StartLevel:  blockpolicy.LevelAggressive,
			MaxAttempts: 3,
		},
	)

	totalCost := 0.0
	for i := range attempts {
		if attempts[i].Response != nil {
			totalCost += attempts[i].Response.Metadata.BillingCreditCost
		}
	}
	report := scraper.BuildChallengeDebugReport(attempts)
	debugLines := make([]string, 0, len(report.Attempts))
	for i := range report.Attempts {
		a := report.Attempts[i]
		debugLines = append(debugLines, fmt.Sprintf(
			"attempt=%d proxy=%s blocked=%v err=%v health=%v delta=%v",
			a.Attempt, a.SelectedProxy, a.Blocked, a.HasError, a.Health, a.HealthDelta,
		))
	}

	if err != nil && resp == nil {
		return modeResult{
			DurationMs: time.Since(start).Milliseconds(),
			Attempts:   len(attempts),
			Cost:       totalCost,
			Err:        err,
			DebugLines: debugLines,
		}
	}

	code := 0
	blocked := false
	if resp != nil {
		code = resp.Metadata.ContentStatusCode
		if code == 0 && len(resp.PageResponses) > 0 {
			code = resp.PageResponses[0].StatusCode
		}
		blocked = blockpolicy.LooksBlocked(resp)
	}
	return modeResult{
		DurationMs: time.Since(start).Milliseconds(),
		Attempts:   len(attempts),
		Cost:       totalCost,
		StatusCode: code,
		Blocked:    blocked,
		Err:        err,
		DebugLines: debugLines,
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
		for _, line := range r.After.DebugLines {
			fmt.Printf("    DEBUG: %s\n", line)
		}

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

func exportIfRequested(rows []row) error {
	jsonPath := strings.TrimSpace(os.Getenv("ABTEST_EXPORT_JSON"))
	mdPath := strings.TrimSpace(os.Getenv("ABTEST_EXPORT_MD"))
	if jsonPath == "" && mdPath == "" {
		return nil
	}
	if jsonPath != "" {
		if err := exportJSON(rows, jsonPath); err != nil {
			return err
		}
		fmt.Printf("EXPORTED JSON: %s\n", jsonPath)
	}
	if mdPath != "" {
		if err := exportMarkdown(rows, mdPath); err != nil {
			return err
		}
		fmt.Printf("EXPORTED MD: %s\n", mdPath)
	}
	return nil
}

func exportJSON(rows []row, path string) error {
	payload := exportRun{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Rows:        make([]exportRow, 0, len(rows)),
	}
	for i := range rows {
		r := rows[i]
		payload.Rows = append(payload.Rows, exportRow{
			Name:   r.Name,
			URL:    r.URL,
			Before: toExportMode(r.Before),
			After:  toExportMode(r.After),
			Debug:  append([]string(nil), r.After.DebugLines...),
		})
	}
	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o644)
}

func exportMarkdown(rows []row, path string) error {
	var b strings.Builder
	b.WriteString("# A/B Live Test Report\n\n")
	b.WriteString(fmt.Sprintf("- Generated: %s\n\n", time.Now().UTC().Format(time.RFC3339)))
	b.WriteString("| Name | URL | Before | After |\n")
	b.WriteString("|---|---|---|---|\n")
	for i := range rows {
		r := rows[i]
		before := fmt.Sprintf("status=%d blocked=%v attempts=%d ms=%d cost=%.4f err=%q",
			r.Before.StatusCode, r.Before.Blocked, r.Before.Attempts, r.Before.DurationMs, r.Before.Cost, oneLineErr(r.Before.Err))
		after := fmt.Sprintf("status=%d blocked=%v attempts=%d ms=%d cost=%.4f err=%q",
			r.After.StatusCode, r.After.Blocked, r.After.Attempts, r.After.DurationMs, r.After.Cost, oneLineErr(r.After.Err))
		b.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", r.Name, r.URL, before, after))
	}
	b.WriteString("\n## After Debug Trace\n\n")
	for i := range rows {
		r := rows[i]
		b.WriteString(fmt.Sprintf("### %s\n\n", r.Name))
		if len(r.After.DebugLines) == 0 {
			b.WriteString("- no debug lines\n\n")
			continue
		}
		for _, line := range r.After.DebugLines {
			b.WriteString(fmt.Sprintf("- `%s`\n", line))
		}
		b.WriteString("\n")
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func toExportMode(m modeResult) exportMode {
	return exportMode{
		StatusCode: m.StatusCode,
		Blocked:    m.Blocked,
		Attempts:   m.Attempts,
		DurationMs: m.DurationMs,
		Cost:       m.Cost,
		Error:      oneLineErr(m.Err),
	}
}
