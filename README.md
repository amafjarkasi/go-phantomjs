# go-phantomjs

[![Go Reference](https://pkg.go.dev/badge/github.com/amafjarkasi/go-phantomjs.svg)](https://pkg.go.dev/github.com/amafjarkasi/go-phantomjs)
[![CI](https://github.com/amafjarkasi/go-phantomjs/actions/workflows/ci.yml/badge.svg)](https://github.com/amafjarkasi/go-phantomjs/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/go-1.21+-blue)](https://go.dev/dl/)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

`go-phantomjs` is a production-ready Go client library for the [PhantomJsCloud](https://phantomjscloud.com/) API. Beyond a thin API wrapper, it ships a full browser-automation scripting layer and a modular extension ecosystem (`ext/`) that composes fingerprint evasions, realistic browser profiles, URL blocklists, and viewport presets into a single fluent API — letting you build sophisticated headless-browser scrapers entirely in Go.

## Features

- **`PageRequestBuilder`**: Fluent builder for `PageRequest` itself — composes `ext/` presets into a complete request without touching nested struct fields. Chain `WithRenderType`, `WithProxy`, `WithProfile`, `WithBlocklist`, `WithRenderSettings`, `WithViewport`, `WithOverseerScriptBuilder`, and more.
- **Fluent Automation Builder**: `OverseerScriptBuilder` generates complex Puppeteer-style automation scripts without string concatenation — click, type, scroll, hover, inject scripts, evaluate JS, take mid-execution screenshots, and manage cookies and headers, all chainable.
- **`FetchWithAutomation()`**: Execute a script and get the native structured return value from any `Evaluate()` call back as a parsed Go `any` — no JSON unwrapping required.
- **Convenience Fetchers**: One-line helpers `FetchPlainText()`, `FetchPDF()`, `FetchScreenshot()`, and `RenderRawHTML()` skip the response envelope entirely — `RenderRawHTML` renders an in-memory HTML string without needing a web server, perfect for PDF invoice and report generation.
- **Full API Type Coverage**: Strongly-typed Go structs for every PhantomJsCloud parameter — `PageRequest`, `RequestSettings`, `RenderSettings`, `ResourceModifier`, `DoneWhen`, `UrlSettings`, `ProxyOptions`, and more — with JSON tags already wired.
- **Stealth Evasions** (`ext/stealth`): One call — `ApplyStealth()` — injects 14 browser fingerprinting evasions ported from [`puppeteer-extra-plugin-stealth`](https://github.com/berstend/puppeteer-extra/tree/master/packages/puppeteer-extra-plugin-stealth), spoofing `navigator.webdriver`, WebGL vendor strings, Chrome runtime APIs, iframe `contentWindow`, media codec lists, outer window dimensions, and more. The evasion payload is embedded at compile time via `//go:embed` — zero runtime deps.
- **Browser Profiles** (`ext/useragents`): 15+ current UA string constants covering Chrome, Firefox, Safari, Edge, and mobile browsers, plus `Profile` structs that bundle a UA with a complete set of matching `Sec-CH-UA`, `Accept`, `Accept-Language`, and `Sec-Fetch-*` headers. Apply in one call with `UseProfile()` on the builder.
- **URL Blocklists** (`ext/blocklist`): Pre-built `ResourceModifier` slices blocking 20 ad networks, 28 analytics beacons, web fonts, and media assets — composable via `Lightweight()` and `Full()` presets. Cuts page request volume by 40–60% on ad-heavy targets, directly reducing API billing cost.
- **Viewport Presets** (`ext/viewport`): Named presets for every common screen — desktop HD/FHD/QHD/4K, laptop, mobile portrait/landscape, tablet, and OG image thumbnails (640×480, 1200×630). Apply to `RenderSettings` via `AsRenderSettings()` or live page emulation via `ApplyViewport()` on the builder.
- **Race-Free Navigation**: `ClickAndWaitForNavigation()` atomically pairs a click with a page-load wait, eliminating the race condition that `Click()` + `WaitForNavigation()` are susceptible to on fast servers.
- **Response Metadata**: Automatically parses `pjsc-*` response headers into a structured `ResponseMetadata` object — billing credit cost, content status code, and done-when event — attached to every response.
- **Proxy Constants**: Named constants for every PhantomJsCloud proxy location (`ProxyAnonUS`, `ProxyGeoUK`, `ProxyGeoDE`, etc.) so you're never hardcoding location strings.
- **CI-Tested**: GitHub Actions runs `go vet`, `go test -race`, `go build`, and `golangci-lint` on every push. All packages have dedicated unit tests including race-detector coverage.

## Installation

```bash
go get github.com/amafjarkasi/go-phantomjs
```

## Quick Start

### Convenience Fetchers (PDFs, Images, Scripts)

If you don't care about the full API metadata, you can use the built in convenience fetchers:

```go
package main

import (
 "os"
 "log"
 "github.com/amafjarkasi/go-phantomjs"
)

func main() {
    client := phantomjscloud.NewClient("") // demo key
    
    // Fetch purely the stripped text of a webpage (Great for LLMs!)
    text, err := client.FetchPlainText("https://example.com")
    if err == nil {
        log.Println(text)
    }

    // Fetch a base64-decoded PDF instantly
    pdfBytes, err := client.FetchPDF("https://example.com", nil)
    if err != nil {
        log.Fatalf("Error: %v", err)
    }
    os.WriteFile("output.pdf", pdfBytes, 0644)
    
    // Evaluate a script block and cleanly parse exactly what it returns
    builder := phantomjscloud.NewOverseerScriptBuilder().
        WaitForNavigation().
        Evaluate("() => { return { title: document.title }; }")
        
    result, err := client.FetchWithAutomation("https://example.com", builder)
    if err != nil {
        log.Fatalf("Error: %v", err)
    }
    
    // Prints map[title:Example Domain]
    log.Println(result) 
}
```

### Advanced HTML Extraction

```go
package main

import (
 "fmt"
 "log"
 "github.com/amafjarkasi/go-phantomjs"
)

func main() {
 // Passing an empty string uses the free demo key (low quota).
 // Replace with your actual PhantomJsCloud API key.
 client := phantomjscloud.NewClient("")

 req := &phantomjscloud.PageRequest{
  URL:        "https://example.com",
  RenderType: "html",
 }

 resp, err := client.DoPage(req)
 if err != nil {
  log.Fatalf("Error: %v", err)
 }

 fmt.Printf("Cost: %f credits\n", resp.Metadata.BillingCreditCost)
 fmt.Printf("Content: %s\n", resp.PageResponses[0].Content)
}
```

### Advanced Automation and Browser Scripting

Use the `OverseerScriptBuilder` to construct scripts that navigate, interact, and manipulate the DOM before returning the final rendered output.

```go
package main

import (
 "fmt"
 "log"
 "github.com/amafjarkasi/go-phantomjs"
)

func main() {
 client := phantomjscloud.NewClient("YOUR-API-KEY")

 // Build an automation script dynamically
 script := phantomjscloud.NewOverseerScriptBuilder().
  WaitForSelector("body").
  Type("input#search", "golang", 100).
  Click("button#submit").
  WaitForNavigation().
  Build()

 req := &phantomjscloud.PageRequest{
  URL:            "https://example.com",
  RenderType:     "png", // Take a screenshot after the script runs
  OutputAsJson:   true,
  OverseerScript: script,
  RequestSettings: phantomjscloud.RequestSettings{
   // Connect through a US residential proxy
   Proxy: phantomjscloud.ProxyGeoUS,
            // Skip image loading for faster performance if we just want data
   ResourceModifier: []phantomjscloud.ResourceModifier{
    {Type: "image", IsBlacklisted: true},
   },
  },
  RenderSettings: phantomjscloud.RenderSettings{
   Viewport: phantomjscloud.Viewport{Width: 1280, Height: 720},
  },
 }

 resp, err := client.DoPage(req)
 if err != nil {
  log.Fatalf("Error: %v", err)
 }

 // Assuming success, resp.PageResponses[0].Content contains the base64 encoded PNG.
 fmt.Printf("Captured screenshot length: %d bytes\n", len(resp.PageResponses[0].Content))
}
```

### Comprehensive Automation Scripting

`OverseerScriptBuilder` exposes the full Puppeteer-style surface area supported by PhantomJsCloud. Every method is chainable and generates the correct `await` expression automatically:

```go
 script := phantomjscloud.NewOverseerScriptBuilder().
  // ── Fingerprinting & identity
  UseProfile(useragents.ChromeWindowsProfile()). // UA + all matching headers in one call
  ApplyStealth().                // 14 evasions: navigator, WebGL, chrome APIs, iFrames
  ApplyViewport(viewport.FHD.Viewport). // Full viewport flags: isMobile, hasTouch, scale

  // ── Navigation
  Goto("https://example.com").
  WaitForNavigation().           // Wait for next load/redirect
  Reload().                      // Native browser refresh

  // ── DOM interaction
  AddScriptTag("https://example.com/utility.js").
  Evaluate("() => { console.log('Injected!'); }").
  WaitForSelector("body").
  WaitForXPath("//div[@class='results']"). // XPath support
  WaitForFunction("window.ready === true"). // Wait on any JS expression
  WaitForDelay(2000).            // Pause for UI transitions

  // ── Form automation
  Focus("input#search").
  ClearInput("input#search").    // Reliably empties the field via JS
  Type("input#search", "hello world", 100).
  Select("select#country", "US", "UK"). // Multi-select supported
  KeyboardPress("Enter", 1).     // Native keystroke
  ClickAndWaitForNavigation("button[type=submit]"). // Atomic: no race condition

  // ── Scrolling & mouse
  ScrollBy(0, 500).              // Relative pixel scroll
  ScrollToBottom().              // Jump to document.body.scrollHeight
  Hover("button#menu").          // Trigger CSS :hover states
  MouseMove(100, 200).           // OS-level cursor movement
  MouseClickPosition(300, 400).  // Coordinate click, bypasses DOM events

  // ── Headers, cookies & style
  SetUserAgent("MyAgent").       // Raw UA override (prefer UseProfile)
  SetExtraHTTPHeaders(map[string]string{"Authorization": "Bearer token"}).
  SetCookie("session", "abc", "example.com").
  DeleteCookie("session", "example.com").
  AddStyleTag("body { background: red; }").
  SetViewport(1920, 1080).       // Simple w/h override (prefer ApplyViewport)

  // ── Rendering & completion
  RenderContent().               // Capture HTML mid-script
  RenderScreenshot(true).        // Synchronous screenshot mid-execution
  ManualWait().                  // Disable auto-completion
  Done().                        // Signal manual completion
  Build()
```

### Stealth Evasions

Call `ApplyStealth()` before navigating to inject a comprehensive set of browser fingerprinting evasions derived from [`puppeteer-extra-plugin-stealth`](https://github.com/berstend/puppeteer-extra/tree/master/packages/puppeteer-extra-plugin-stealth). This spoofs the APIs that bot-detection scripts most commonly probe:

| Evasion | What it fixes |
|---|---|
| `navigator.webdriver` | Removes the automation flag |
| `chrome.app` / `chrome.csi` / `chrome.runtime` | Mocks missing Chrome extension APIs |
| `navigator.plugins` / `navigator.languages` / `navigator.vendor` | Supplies realistic plugin lists |
| `webgl.vendor` | Spoofs GPU vendor strings |
| `iframe.contentWindow` | Hides proxy artifacts in iframes |
| `media.codecs` | Reports realistic codec support |
| `window.outerdimensions` | Corrects outer window size |

```go
script := phantomjscloud.NewOverseerScriptBuilder().
  ApplyStealth().              // Inject all evasions before page load
  Goto("https://bot.sannysoft.com").
  WaitForDelay(2000).
  RenderScreenshot(true).
  Build()
```

> **Regenerating the payload** — the stealth JS lives in `ext/stealth/evasions.js` and is embedded at compile time via `//go:embed`. To update it after upgrading the npm package:
>
> ```bash
> npm update puppeteer-extra-plugin-stealth
> node scripts/gen_stealth.js
> ```

### Blocklist — Block Ads, Trackers & Dead Weight

The `ext/blocklist` package provides pre-built `ResourceModifier` slices for the most common scraping optimizations. Blocking ads and trackers alone can cut a typical page's request volume by 40–60% and reduce billing costs proportionally.

```go
import "github.com/amafjarkasi/go-phantomjs/ext/blocklist"

req := &phantomjscloud.PageRequest{
  URL: "https://edition.cnn.com",
  RequestSettings: phantomjscloud.RequestSettings{
    // Lightweight: block ads + trackers + fonts only
    ResourceModifier: blocklist.Lightweight(),

    // Or Full: also block images and video
    // ResourceModifier: blocklist.Full(),

    // Or compose your own:
    // ResourceModifier: append(blocklist.Ads(), blocklist.Fonts()...),
  },
}
```

| Preset | Blocks |
|---|---|
| `blocklist.Ads()` | 20 major ad networks |
| `blocklist.Trackers()` | 28 analytics/tracking beacons |
| `blocklist.Fonts()` | Google Fonts, Typekit, web font files |
| `blocklist.Media()` | All image and video assets |
| `blocklist.Lightweight()` | Ads + Trackers + Fonts _(recommended default)_ |
| `blocklist.Full()` | Ads + Trackers + Fonts + Media |

### User Agents — Realistic Browser Profiles

The `ext/useragents` package ships 15 current UA string constants covering every major browser and platform, plus `Profile` structs that bundle a UA with a complete matching header set (`Sec-CH-UA`, `Accept`, `Accept-Language`, `Sec-Fetch-*`) — because mismatched UA/header combinations are one of the most common bot signals.

**UA constants by platform:**

| Constant | Browser |
|---|---|
| `ChromeWin`, `ChromeWin11` | Chrome 122 on Windows 10 / 11 |
| `ChromeMac`, `ChromeLinux` | Chrome 122 on macOS / Linux |
| `FirefoxWin`, `FirefoxMac` | Firefox 123 on Windows / macOS |
| `SafariMac`, `SafariIPad`, `SafariIPhone` | Safari 17 on macOS / iOS |
| `EdgeWin` | Edge 122 on Windows |
| `ChromeAndroid`, `ChromeAndroidTablet` | Chrome 122 on Pixel 8 / Pixel Tablet |
| `Googlebot`, `GooglebotMobile`, `Bingbot` | Search crawler UAs |

**Profile constructors:** `ChromeWindowsProfile()`, `ChromeMacProfile()`, `FirefoxWindowsProfile()`

```go
import "github.com/amafjarkasi/go-phantomjs/ext/useragents"

// Simple: just the UA string in RequestSettings
req := &phantomjscloud.PageRequest{
  RequestSettings: phantomjscloud.RequestSettings{
    UserAgent: useragents.ChromeWin,
  },
}

// Better: full browser profile with matching headers in RequestSettings
profile := useragents.ChromeWindowsProfile()
req = &phantomjscloud.PageRequest{
  RequestSettings: phantomjscloud.RequestSettings{
    UserAgent:     profile.UserAgent,
    CustomHeaders: profile.Headers,
  },
}

// Best: set profile inside the automation script — UA and headers applied at evaluation time
script := phantomjscloud.NewOverseerScriptBuilder().
  UseProfile(useragents.ChromeWindowsProfile()). // setUserAgent + setExtraHTTPHeaders in one call
  Goto("https://example.com").
  Build()
```

### Viewport — Named Display Presets

The `ext/viewport` package provides named `Preset` variables for common screen configurations. Presets can be used in two ways:

**In `RenderSettings`** (affects the rendered output size and clip):

```go
import "github.com/amafjarkasi/go-phantomjs/ext/viewport"

req := &phantomjscloud.PageRequest{
  URL:            "https://example.com",
  RenderType:     "jpeg",
  RenderSettings: viewport.FHD.AsRenderSettings(),           // 1920×1080 desktop
}

req = &phantomjscloud.PageRequest{
  URL:            "https://example.com",
  RenderType:     "jpeg",
  RenderSettings: viewport.Thumbnail1200.AsRenderSettings(), // 1200×630 OG image, with clip
}
```

**In the automation script** (live mobile emulation via `page.setViewport`):

```go
script := phantomjscloud.NewOverseerScriptBuilder().
  ApplyViewport(viewport.MobilePortrait.Viewport). // isMobile:true, hasTouch:true, 390×844
  Goto("https://example.com").
  Build()
```

Available presets: `HD`, `FHD`, `QHD`, `UHD`, `Laptop`, `MobilePortrait`, `MobileLandscape`, `TabletPortrait`, `TabletLandscape`, `Thumbnail640`, `Thumbnail1200`, `Custom(w, h)`.

### Full Stealth Scrape Pattern

Combining all extensions gives you the strongest bot-evasion setup:

```go
import (
  phantomjscloud "github.com/amafjarkasi/go-phantomjs"
  "github.com/amafjarkasi/go-phantomjs/ext/blocklist"
  "github.com/amafjarkasi/go-phantomjs/ext/useragents"
  "github.com/amafjarkasi/go-phantomjs/ext/viewport"
)

profile := useragents.ChromeWindowsProfile()

script := phantomjscloud.NewOverseerScriptBuilder().
  UseProfile(profile).                          // Realistic UA + spoofed headers
  ApplyStealth().                               // 14 fingerprint evasions
  ApplyViewport(viewport.FHD.Viewport).         // 1920×1080 desktop emulation
  Goto("https://example.com").
  WaitForDelay(1500).
  Build()

// PageRequestBuilder composes the full request without struct literals
req := phantomjscloud.NewPageRequestBuilder("https://about:blank").
  WithRenderType("jpeg").
  WithProfile(profile).
  WithBlocklist(blocklist.Lightweight()).
  WithOverseerScript(script).
  Build()
```

### Advanced Automation Workflows

#### Auto-Login & Navigation

`ClickAndWaitForNavigation()` atomically pairs the click with the page-load wait — eliminating the race condition where a fast server redirects before a subsequent `WaitForNavigation()` call can register.

```go
 script := phantomjscloud.NewOverseerScriptBuilder().
  WaitForSelector("input#username").
  ClearInput("input#username").
  Type("input#username", "USER@EXAMPLE.COM", 50).
  ClearInput("input#password").
  Type("input#password", "PASSWORD", 50).
  ClickAndWaitForNavigation("button[type=submit]"). // Atomic: click + wait, no race condition
  Build()

 req := &phantomjscloud.PageRequest{
  URL:            "https://www.linkedin.com/uas/login",
  RenderType:     "jpeg",
  OverseerScript: script,
 }
```

#### Speeding up Long Requests (DOM Content Loaded)

If a page has heavy ad trackers or infinite lazy loading, `PhantomJsCloud` might timeout waiting for the network idle state. You can override this to finish rendering as soon as the DOM is available or use `DoneWhen` in `RequestSettings`.

```go
 // Method 1: Inject a manual wait and exit specifically on the domcontentloaded event
 script := phantomjscloud.NewOverseerScriptBuilder().
  WaitForNavigationEvent("domcontentloaded").
  Done().
  Build()

 // Method 2: Configure it declaratively in RequestSettings natively
 reqSettings := phantomjscloud.RequestSettings{
  DoneWhen: []phantomjscloud.DoneWhen{
   {Event: "domReady"},
  },
 }
```

### Advanced Features

#### Emulate Print Media for PDF Generation

Use the `EmulateMedia` parameter to generate a PDF exactly as it would look when printed.

```go
 req := &phantomjscloud.PageRequest{
  URL:        "https://example.com/invoice.html",
  RenderType: "pdf",
  RenderSettings: phantomjscloud.RenderSettings{
   EmulateMedia: "print", // Generate PDF using the @media:print CSS rules
  },
 }
```

#### Intercept and Modify Requests (Change URL & Blacklist)

You can use the `ResourceModifier` to change domains on the fly, or blacklist certain requests completely to save bandwidth (like CSS files).

```go
 req := &phantomjscloud.PageRequest{
  URL:        "https://www.highcharts.com",
  RenderType: "jpg",
  RequestSettings: phantomjscloud.RequestSettings{
   ClearCache: true, // Forces re-requesting css to be caught by the blacklist
   ResourceModifier: []phantomjscloud.ResourceModifier{
    {
     Regex:     ".*highcharts.com.*",
     ChangeUrl: "$$protocol:$$port//en.wikipedia.org/wiki$$path",
    },
    {
     Regex:         ".*css.*",
     IsBlacklisted: true,
    },
   },
  },
 }
```

#### Render Thumbnails and Zooming

Combine `Viewport`, `ClipRectangle`, and `ZoomFactor` to capture perfect thumbnails.

```go
 req := &phantomjscloud.PageRequest{
  URL:        "https://cnn.com",
  RenderType: "jpeg",
  RenderSettings: phantomjscloud.RenderSettings{
   ZoomFactor: 0.45,
   Viewport:      &phantomjscloud.Viewport{Width: 640, Height: 500},
   ClipRectangle: &phantomjscloud.ClipRectangle{Width: 640, Height: 500},
  },
 }
```

#### Uploading POST Data and JSONP

To submit POST data to a target URL natively, use `UrlSettings`:

```go
 req := &phantomjscloud.PageRequest{
  URL: "https://example.com/api",
  UrlSettings: &phantomjscloud.UrlSettings{
   Operation: "POST",
   Data:      `{"my_key":"my_value"}`,
  },
 }
```

#### HTTP Basic Auth & Reduced JSON Verbosity

Using `OutputAsJson: true` will return a massive payload. You can suppress fields using `SuppressJson`. Also, bypass HTTP Basic Auth natively using `Authentication`.

```go
 req := &phantomjscloud.PageRequest{
  URL:          "http://httpbin.org/basic-auth/user/pass",
  OutputAsJson: true,
  SuppressJson: []string{"pageResponses", "originalRequest"},
  RequestSettings: phantomjscloud.RequestSettings{
   Authentication: &phantomjscloud.Authentication{
    UserName: "user",
    Password: "pass",
   },
  },
 }
```

#### Full JSON Metadata Response and Cookies

To extract cookies, headers, and extensive metadata, use `OutputAsJson: true` and specify `Cookies` in the request settings.

```go
 req := &phantomjscloud.PageRequest{
  URL:          "http://example.com",
  RenderType:   "plainText",
  OutputAsJson: true,
  RequestSettings: phantomjscloud.RequestSettings{
   Cookies: []phantomjscloud.Cookie{
    {Domain: "example.com", Name: "myCookie1", Value: "value1"},
   },
  },
 }
```

#### Render Raw HTML Strings

`RenderRawHTML()` lets you upload a raw HTML string and render it through PhantomJsCloud as if it were a real page — useful for templated PDF invoices, emails, or dynamically generated reports without needing a web server.

```go
 html := `<html><body><h1>Invoice #1234</h1><p>Total: $99.00</p></body></html>`

 pdfBytes, err := client.RenderRawHTML(html, "pdf", nil)
 if err != nil {
  log.Fatal(err)
 }
 os.WriteFile("invoice.pdf", pdfBytes, 0o644)
```

## Repository Structure

```
go-phantomjs/
├── automation.go        # OverseerScriptBuilder — ApplyStealth, UseProfile, ApplyViewport, ClickAndWaitForNavigation
├── builder.go           # PageRequestBuilder — fluent PageRequest composition from ext/ presets
├── client.go            # API client, Do/DoPage, FetchWithAutomation, response metadata
├── types.go             # PhantomJsCloud JSON type mappings
├── client_test.go       # Unit tests
│
├── ext/                 # Optional sub-packages (import only what you need)
│   ├── stealth/         # Browser fingerprint evasions (from puppeteer-extra-plugin-stealth)
│   │   ├── stealth.go
│   │   └── evasions.js  # GENERATED — run: node scripts/gen_stealth.js
│   ├── blocklist/       # Pre-built ResourceModifier URL blacklists
│   │   ├── blocklist.go
│   │   └── blocklist_test.go
│   ├── useragents/      # Realistic browser UA strings and Profile bundles
│   │   ├── useragents.go
│   │   └── useragents_test.go
│   └── viewport/        # Named Viewport/ClipRectangle/ZoomFactor presets
│       ├── viewport.go
│       └── viewport_test.go
│
├── .github/
│   └── workflows/
│       └── ci.yml       # go vet + go test -race + go build on every push/PR
│
├── scripts/
│   └── gen_stealth.js   # Node.js bundler: pulls from puppeteer-extra-plugin-stealth
│
├── example/
│   └── main.go          # Stealth scrape, auto-login, OG thumbnail examples
│
├── .golangci.yml        # golangci-lint config (errcheck, staticcheck, govet, misspell)
├── CHANGELOG.md         # Keep a Changelog — full history from v0.1.0
└── package.json         # npm deps for code-gen only (private, not published)
```

**Adding extensions**: drop a new directory under `ext/` with its own Go `package`, import it where needed, and optionally wire a method into `OverseerScriptBuilder`.

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Before submitting, make sure all of these pass locally:

```bash
go vet ./...
go test -race -count=1 ./...
go build ./...
```

CI runs `go vet`, `go test -race`, `go build`, and `golangci-lint` automatically on every push and PR.

## License

[MIT](https://choosealicense.com/licenses/mit/)
