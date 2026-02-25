package main

import (
	"fmt"
	"log"
	"os"

	phantomjscloud "github.com/amafjarkasi/go-phantomjs"
	"github.com/amafjarkasi/go-phantomjs/ext/blocklist"
	"github.com/amafjarkasi/go-phantomjs/ext/useragents"
	"github.com/amafjarkasi/go-phantomjs/ext/viewport"
)

func main() {
	apiKey := os.Getenv("PHANTOMJSCLOUD_API_KEY")
	if apiKey == "" {
		log.Fatal("PHANTOMJSCLOUD_API_KEY environment variable is required")
	}
	client := phantomjscloud.NewClient(apiKey)

	// ── Example A: Full Stealth Scrape ───────────────────────────────────────
	// Chrome/Windows fingerprint + stealth evasions + lightweight blocklist.
	// Use this as the starting point for any site with bot detection.
	fmt.Println("=== Example A: Full stealth scrape ===")

	profile := useragents.ChromeWindowsProfile()
	stealthScript := phantomjscloud.NewOverseerScriptBuilder().
		UseProfile(profile).                  // Set realistic UA + headers
		ApplyStealth().                       // Inject 14 fingerprint evasions
		ApplyViewport(viewport.FHD.Viewport). // 1920x1080 desktop
		Goto("https://bot.sannysoft.com").
		WaitForDelay(1500).
		RenderScreenshot(true).
		Build()

	stealthReq := phantomjscloud.NewPageRequestBuilder("about:blank").
		WithRenderType("jpeg").
		WithProfile(profile).
		WithBlocklist(blocklist.Lightweight()). // block ads + trackers + fonts
		WithOverseerScript(stealthScript).
		Build()
	resp, err := client.DoPage(stealthReq)
	if err != nil {
		log.Printf("Example A error: %v\n", err)
	} else {
		fmt.Printf("Example A: status=%s  cost=%.4f credits\n",
			resp.Status, resp.Metadata.BillingCreditCost)
	}

	// ── Example B: Auto-Login ────────────────────────────────────────────────
	// Type credentials, submit, and wait for the backend redirect.
	// ClickAndWaitForNavigation avoids the race condition between click and load.
	fmt.Println("\n=== Example B: Auto-login (demo — not a real site) ===")

	loginScript := phantomjscloud.NewOverseerScriptBuilder().
		WaitForSelector("input#username").
		ClearInput("input#username").
		Type("input#username", "user@example.com", 60).
		ClearInput("input#password").
		Type("input#password", "s3cr3t", 60).
		ClickAndWaitForNavigation("button[type=submit]").
		WaitForSelector(".dashboard-header").
		RenderContent().
		Done().
		Build()

	fmt.Println("Login script:\n", loginScript)

	// Attach it to a real request like this:
	_ = &phantomjscloud.PageRequest{
		URL:            "https://example.com/login",
		RenderType:     "plainText",
		OverseerScript: loginScript,
	}

	// ── Example C: Open Graph Thumbnail (1200×630) ───────────────────────────
	// Generate a social share image using the Thumbnail1200 viewport preset.
	fmt.Println("\n=== Example C: Open Graph thumbnail (1200x630) ===")

	thumbReq := &phantomjscloud.PageRequest{
		URL:            "https://go.dev",
		RenderType:     "jpeg",
		RenderSettings: viewport.Thumbnail1200.AsRenderSettings(),
		RequestSettings: phantomjscloud.RequestSettings{
			ResourceModifier: blocklist.Fonts(), // block fonts for speed
		},
	}
	thumbBytes, err := client.FetchScreenshot(thumbReq.URL, "jpeg", &thumbReq.RenderSettings)
	if err != nil {
		log.Printf("Example C error: %v\n", err)
	} else {
		path := "thumb.jpg"
		if err := os.WriteFile(path, thumbBytes, 0o644); err != nil {
			log.Printf("Example C write error: %v\n", err)
		} else {
			fmt.Printf("Example C: saved %d bytes → %s\n", len(thumbBytes), path)
		}
	}
}
