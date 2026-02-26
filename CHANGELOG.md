# Changelog

All notable changes to this project will be documented in this file.
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

---

## [0.1.0] — 2026-02-25

### Added

#### Core

- `PageRequestBuilder` — fluent builder for `PageRequest` that composes `ext/` presets without struct literal nesting. Key methods: `WithRenderType`, `WithProxy`, `WithProfile`, `WithBlocklist`, `WithRenderSettings`, `WithViewport`, `WithOverseerScriptBuilder`, and more.
- `OverseerScriptBuilder.ClickAndWaitForNavigation(selector)` — atomic `Promise.all([waitForNavigation, click])` pattern; eliminates race conditions on fast-redirecting form submissions.
- `OverseerScriptBuilder.UseProfile(useragents.Profile)` — sets `setUserAgent` + `setExtraHTTPHeaders` in a single chained call.
- `OverseerScriptBuilder.ApplyViewport(Viewport)` — emits `page.setViewport` with all six flags (width, height, deviceScaleFactor, isMobile, hasTouch, isLandscape).
- `OverseerScriptBuilder.ApplyStealth()` — injects a compiled IIFE of 14 browser fingerprinting evasions.
- `OverseerScriptBuilder.WaitForXPath(xpath)` — waits for an XPath expression to resolve in the DOM.

#### `ext/stealth`

- Embeds a compiled JavaScript IIFE (generated from `puppeteer-extra-plugin-stealth`) via `//go:embed`. Evasions cover: `navigator.webdriver`, WebGL vendor strings, Chrome runtime APIs (`chrome.app`, `chrome.csi`, `chrome.runtime`), `navigator.plugins`/`languages`/`vendor`, iframe `contentWindow`, media codec lists, and outer window dimensions.
- `scripts/gen_stealth.js` — Node.js code generator that bundles evasions from `node_modules` into `ext/stealth/evasions.js`.

#### `ext/blocklist`

- `Ads()` — 20 advertising network URL patterns.
- `Trackers()` — 28 analytics/tracking beacon patterns.
- `Fonts()` — Google Fonts, Typekit, and generic web font CDN patterns.
- `Media()` — image and video asset patterns.
- `Lightweight()` — Ads + Trackers + Fonts (recommended scraping default).
- `Full()` — Ads + Trackers + Fonts + Media.

#### `ext/useragents`

- 15 UA string constants: `ChromeWin`, `ChromeWin11`, `ChromeMac`, `ChromeLinux`, `FirefoxWin`, `FirefoxMac`, `SafariMac`, `SafariIPad`, `SafariIPhone`, `EdgeWin`, `ChromeAndroid`, `ChromeAndroidTablet`, `Googlebot`, `GooglebotMobile`, `Bingbot`.
- `Profile` type — bundles a UA with a matching header map (`Sec-CH-UA`, `Accept`, `Accept-Language`, `Sec-Fetch-*`).
- `ChromeWindowsProfile()`, `ChromeMacProfile()`, `FirefoxWindowsProfile()` constructors.

#### `ext/viewport`

- Named `Preset` variables: `HD`, `FHD`, `QHD`, `UHD`, `Laptop`, `MobilePortrait`, `MobileLandscape`, `TabletPortrait`, `TabletLandscape`, `Thumbnail640`, `Thumbnail1200`.
- `Custom(width, height int) Preset` — ad-hoc preset with no clip or zoom.
- `Preset.AsRenderSettings() RenderSettings` — maps preset into `PageRequest.RenderSettings`.

#### CI / Quality

- `.github/workflows/ci.yml` — `go vet`, `go test -race -count=1`, `go build`, and `golangci-lint` run on every push and pull request.
- `.golangci.yml` — linters: `errcheck`, `staticcheck`, `govet`, `godot`, `misspell`, `unconvert`, `unused`.

### Changed

- `example/main.go` — replaced placeholder with three complete runnable examples: full stealth scrape, auto-login with `ClickAndWaitForNavigation`, and OG image thumbnail via `viewport.Thumbnail1200`.
- `README.md` — comprehensive rewrite: expanded Features list, comprehensive builder method reference, UA constant table, viewport preset list, "Full Stealth Scrape" combined pattern, updated Repository Structure tree and Contributing section.

### Fixed

- `OverseerScriptBuilder.Click()` — was missing `await`; now correctly emits `await page.click(...)`.

---

[0.1.0]: https://github.com/amafjarkasi/go-phantomjs/releases/tag/v0.1.0
