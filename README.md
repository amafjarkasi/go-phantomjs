# go-phantomjs

[![go-phantomjs logo](assets/logo.svg)](https://github.com/amafjarkasi/go-phantomjs)

[![Go Reference](https://pkg.go.dev/badge/github.com/amafjarkasi/go-phantomjs.svg)](https://pkg.go.dev/github.com/amafjarkasi/go-phantomjs)
[![CI](https://github.com/amafjarkasi/go-phantomjs/actions/workflows/ci.yml/badge.svg)](https://github.com/amafjarkasi/go-phantomjs/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/go-1.21+-blue)](https://go.dev/dl/)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

`go-phantomjs` is a production-focused Go client for [PhantomJsCloud](https://phantomjscloud.com/): typed API models, fluent automation scripting, and composable scraping extensions (`ext/*`) for stealth, profiles, routing, retries, and session persistence.

## Table Of Contents

- [Why This Library](#why-this-library)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Core Concepts](#core-concepts)
- [Request Builder](#request-builder)
- [Automation Script Builder](#automation-script-builder)
- [Extensions](#extensions)
- [Reliability Patterns](#reliability-patterns)
- [Live A/B Harness](#live-ab-harness)
- [API Compatibility Notes](#api-compatibility-notes)
- [Repository Layout](#repository-layout)
- [Contributing](#contributing)
- [License](#license)

## Why This Library

- Typed request/response models for PhantomJsCloud API payloads.
- Fluent script builder (`OverseerScriptBuilder`) for browser actions without string-concatenation.
- Fluent request builder (`PageRequestBuilder`) to compose profiles, routing, block policies, and render settings.
- Composable reliability modules:
  - host-aware proxy routing
  - health-aware proxy failover
  - adaptive block-policy retries
  - challenge orchestration with persona + session persistence
- Test-first codebase with race-safe CI checks.

## Installation

```bash
go get github.com/amafjarkasi/go-phantomjs
```

## Quick Start

### Minimal Client + HTML Fetch

```go
package main

import (
	"fmt"
	"log"
	"os"

	phantomjscloud "github.com/amafjarkasi/go-phantomjs"
)

func main() {
	client := phantomjscloud.NewClient(os.Getenv("PHANTOMJSCLOUD_API_KEY"))

	req := phantomjscloud.NewPageRequestBuilder("https://example.com").
		WithRenderType("html").
		WithOutputAsJson(true).
		Build()

	resp, err := client.DoPage(req)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("status:", resp.PageResponses[0].StatusCode)
	fmt.Println("cost:", resp.Metadata.BillingCreditCost)
}
```

### Convenience Fetchers

```go
text, _ := client.FetchPlainText("https://example.com")
pdf, _ := client.FetchPDF("https://example.com", nil)
png, _ := client.FetchScreenshot("https://example.com", nil)
```

## Core Concepts

### Client Construction

```go
client := phantomjscloud.NewClient(
	os.Getenv("PHANTOMJSCLOUD_API_KEY"),
	phantomjscloud.WithTimeout(90*time.Second),
	// phantomjscloud.WithHTTPClient(customHTTPClient),
)
```

### Context-Aware Calls

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

resp, err := client.DoPageContext(ctx, req)
```

### Full `UserRequest` Batch Calls

```go
userReq := &phantomjscloud.UserRequest{
	Pages: []phantomjscloud.PageRequest{
		{URL: "https://example.com/1", RenderType: "html"},
		{URL: "https://example.com/2", RenderType: "html"},
	},
}
resp, err := client.Do(userReq)
```

## Request Builder

`PageRequestBuilder` is the main composition surface for page requests.

```go
req := phantomjscloud.NewPageRequestBuilder("https://example.com").
	WithRenderType("html").
	WithOutputAsJson(true).
	WithQuality(85).
	WithHeader("x-my-header", "value").
	Build()
```

Common methods:

- Rendering: `WithRenderType`, `WithQuality`, `WithViewport`, `WithClipRectangle`, `WithZoomFactor`, `WithPdfOptions`
- Request behavior: `WithWaitInterval`, `WithIgnoreImages`, `WithClearCache`, `WithDoneWhen`
- Identity: `WithUserAgent`, `WithProfile`, `WithProxy`, `WithProxyRouter`, `WithProxyRouterAttempt`
- Payload: `WithContent`, `WithUrlSettings`, `WithSuppressJson`
- Auth/session: `WithAuthentication`, `WithCookies`
- Scripting: `WithOverseerScript`, `WithOverseerScriptBuilder`

## Automation Script Builder

`OverseerScriptBuilder` builds Puppeteer-style scripts with chainable helpers.

```go
script := phantomjscloud.NewOverseerScriptBuilder().
	WaitForSelector("input#search").
	Type("input#search", "golang", 75).
	ClickAndWaitForNavigation("button[type=submit]").
	WaitForDelay(1000).
	RenderContent().
	Build()
```

Useful groups:

- Navigation: `Goto`, `WaitForNavigation`, `Reload`, `GoBack`, `GoForward`
- Interaction: `Click`, `Type`, `Select`, `Hover`, `Focus`, `KeyboardPress`, `ScrollBy`
- Conditions: `WaitForSelector`, `WaitForXPath`, `WaitForFunction`, `WaitForNavigationEvent`
- Identity: `UseProfile`, `ApplyStealth`, `ApplyViewport`, `SetUserAgent`, `SetExtraHTTPHeaders`
- Cookies: `SetCookie`, `DeleteCookie`
- Completion: `ManualWait`, `Done`, `RenderContent`, `RenderScreenshot`

## Extensions

### `ext/stealth`

Stealth evasions ported from `puppeteer-extra-plugin-stealth` and embedded via `go:embed`.

```go
script := phantomjscloud.NewOverseerScriptBuilder().
	ApplyStealth().
	Goto("https://bot.sannysoft.com").
	RenderScreenshot(true).
	Build()
```

### `ext/useragents`

Realistic UA constants and profile bundles with matching headers.

```go
profile := useragents.ChromeWindowsProfile()
req := phantomjscloud.NewPageRequestBuilder("https://example.com").
	WithProfile(profile).
	Build()
```

### `ext/viewport`

Named viewport presets for desktop, mobile, tablet, thumbnails.

```go
req := phantomjscloud.NewPageRequestBuilder("https://example.com").
	WithRenderType("jpeg").
	WithViewport(viewport.FHD.Viewport).
	Build()
```

### `ext/blocklist`

Prebuilt URL/resource blocklists for cost/performance tuning.

```go
req := phantomjscloud.NewPageRequestBuilder("https://example.com").
	WithBlocklist(blocklist.Lightweight()).
	Build()
```

### `ext/blockpolicy`

Policy levels for progressive relax-on-block retries:

- `LevelAggressive`
- `LevelBalanced`
- `LevelRelaxed`
- `LevelOff`

### `ext/proxy`

Proxy capabilities:

- `HostRouter`: host-aware round-robin + deterministic fallback by attempt.
- `HealthRouter`: host-aware routing with per-domain health scores and failover learning.

```go
router := proxy.NewHealthRouter(phantomjscloud.ProxyAnonUS).
	RouteHost("amazon.com", phantomjscloud.ProxyAnonUS, phantomjscloud.ProxyAnonCA)
```

### `ext/persona`

Domain-routed persona bundles (proxy + profile + viewport + blockers).

```go
engine := persona.NewEngine().
	Define("desktop-us", persona.Config{
		Proxy:    phantomjscloud.ProxyAnonUS,
		Profile:  useragents.ChromeWindowsProfile(),
		Viewport: viewport.FHD,
	}).
	RouteHost("amazon.com", "desktop-us")
```

### `ext/session`

Cookie store with host/scheme/expiry filtering for safer persistence.

```go
store := session.NewStore()
store.CaptureFromResponse(resp)
cookies := store.CookiesForURL("https://example.com")
```

### `ext/scraper`

Higher-level orchestration helpers:

- `DoPageWithAdaptiveBlockPolicy`
- `DoPageWithRoutingAndAdaptivePolicy`
- `DoPageWithChallengeOrchestration`
- `BuildChallengeDebugReport`

## Reliability Patterns

### Host Routing + Adaptive Policy

```go
router := proxy.NewHostRouter(phantomjscloud.ProxyAnonUS).
	RouteHost("example.com", phantomjscloud.ProxyAnonUS, phantomjscloud.ProxyAnonCA)

resp, attempts, err := scraper.DoPageWithRoutingAndAdaptivePolicy(
	context.Background(),
	client,
	baseReq,
	router,
	blockpolicy.LevelAggressive,
	3,
)
```

### Challenge Orchestration (Persona + Session + Health Routing)

```go
router := proxy.NewHealthRouter(phantomjscloud.ProxyAnonUS).
	RouteHost("example.com", phantomjscloud.ProxyAnonUS, phantomjscloud.ProxyAnonCA)

engine := persona.NewEngine().
	Define("desktop-us", persona.Config{
		Proxy:    phantomjscloud.ProxyAnonUS,
		Profile:  useragents.ChromeWindowsProfile(),
		Viewport: viewport.FHD,
	}).
	RouteHost("example.com", "desktop-us")

store := session.NewStore()

resp, trace, err := scraper.DoPageWithChallengeOrchestration(
	context.Background(),
	client,
	baseReq,
	scraper.ChallengeOrchestrationOptions{
		Persona:     engine,
		Router:      router,
		Session:     store,
		StartLevel:  blockpolicy.LevelAggressive,
		MaxAttempts: 3,
	},
)

report := scraper.BuildChallengeDebugReport(trace)
_ = resp
_ = report
_ = err
```

Notes:

- Health router penalties are domain-scoped.
- Transport/API errors are penalized more strongly than challenge-page blocks.
- `ChallengeAttempt` and `AdaptiveAttempt` include trace fields (`Proxy`, `Blocked`, health snapshots when available).

## Live A/B Harness

`cmd/abtest` compares baseline vs advanced orchestration on live retailer targets.

### Run

```bash
set PHANTOMJSCLOUD_API_KEY=YOUR_KEY
go run ./cmd/abtest
```

### Optional Exports

Export machine-readable and markdown reports:

```bash
set ABTEST_EXPORT_JSON=abtest-report.json
set ABTEST_EXPORT_MD=abtest-report.md
go run ./cmd/abtest
```

## API Compatibility Notes

- `Cookie.Expires` uses `float64` to support non-integer expiration values returned by PhantomJsCloud.
- `PageResponse.contentErrors` supports both:
  - `[]string`
  - object arrays (`[{message: "..."}]`)
- Page-level proxy inputs are normalized before API send:
  - `ProxyBuiltin{Location:"us"}` -> `"anon-us"`
  - `ProxyOptions{Geolocation:"us"}` -> `"geo-us"`
  - `ProxyOptions{Custom:{Host,Auth}}` -> `"custom-{host}:{auth}"`

## Repository Layout

```text
go-phantomjs/
├── automation.go
├── builder.go
├── client.go
├── types.go
├── proxy_compat.go
├── cmd/
│   ├── abtest/
│   └── pjsc/
├── ext/
│   ├── blocklist/
│   ├── blockpolicy/
│   ├── persona/
│   ├── proxy/
│   ├── scraper/
│   ├── session/
│   ├── stealth/
│   ├── useragents/
│   └── viewport/
└── example/
```

## Contributing

Before opening a PR, run:

```bash
go vet ./...
go test -race -count=1 ./...
go build ./...
```

## License

[MIT](LICENSE)

