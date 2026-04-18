# Headless Automation Feature Backlog (Top 25)

Date: 2026-04-18  
Project: `go-phantomjs`

## Method

- Prioritized for: `impact on real-world scrape success`, `implementation simplicity`, `fit with current architecture (ext/* + builder + scraper orchestration)`.
- Scores:
  - `Value`: 1-5 (higher is better)
  - `Effort`: S / M / L (smaller is easier)

## Top 25 Candidate Features

| Rank | Idea | Borrowed From | Value | Effort | Why It Matters | Our Unique Twist |
|---|---|---|---:|:---:|---|---|
| 1 | Proxy health + tiered failover by domain | Crawlee tiered proxies, session-aware proxy routing | 5 | S | Faster unblock and lower spend than blind retries | Add `ext/proxy/health` with latency/error/block score and domain-specific tier memory |
| 2 | Challenge policy state machine (signal-based) | Browserless CAPTCHA + stealth workflows, Crawlee session/block scoring | 5 | S | Converts random fallback into deterministic unblock path | Add weighted challenge signals (HTTP, DOM, token widgets, nav loops) and explicit policy transitions |
| 3 | Durable session store (cookies + local state) | Playwright `storageState`, Browserless session persistence/reconnect | 5 | S | Better continuity across paginated + authenticated flows | Persist session bundles by `domain persona` and age out with strict host/scheme guards |
| 4 | Persona packs v2 (domain-specific identity bundles) | Playwright device/emulation profiles, Crawlee fingerprints | 5 | S | Better match between UA/headers/viewport/proxy improves pass rate | Add “persona recipes” with AB testing and per-domain winner memory |
| 5 | Adaptive concurrency autoscaler | Crawlee `AutoscaledPool` | 5 | M | Avoids overload/ban spikes while maximizing throughput | Scale worker count on block-rate + p95 latency + cost per success |
| 6 | Request frontier with dedupe/reclaim | Crawlee `RequestQueue` | 4 | M | Prevents duplicate spend and supports resilient deep crawling | Add frontier priorities (`category`, `product`, `next-page`) with retry reclaim reason codes |
| 7 | HAR record/replay mode for deterministic tests | Playwright `routeFromHAR` | 4 | M | Stable CI/live regression harness without hitting targets every run | Include “sanitized HAR snapshots” and diff reports for changed endpoints |
| 8 | Cooperative interception rule engine | Puppeteer interception + cooperative priorities | 4 | M | Multiple rule packs can compose safely | Priority-based resolver with trace of which rule won per request |
| 9 | WebSocket interception hooks | Playwright `routeWebSocket` | 4 | M | Modern sites stream data via WS; missing it loses extraction coverage | Add WS message filters + schema extraction path for live inventory/price feeds |
| 10 | Trace bundle for each run | Playwright tracing/trace viewer | 4 | S | Speeds debugging and root-cause analysis | Export compact bundle: request graph, timing, blockers, policy path, final extraction |
| 11 | CAPTCHA provider abstraction | Browserless solve flow, puppeteer recaptcha plugin ecosystem | 4 | M | Swap solving providers without touching core logic | `ext/captcha` with interface + conditional trigger only when challenge selector present |
| 12 | Blocklist auto-tuning by host | Existing blocklist + adaptive policy concepts | 4 | S | Reduces breakage from over-blocking and cost from under-blocking | Learn per-host rule deltas from observed challenge/success outcomes |
| 13 | Navigation completion strategy planner | Puppeteer wait states, Playwright actionability model | 4 | S | Fewer false timeouts and incomplete pages | Per-site completion profile (`domReady`, selector quorum, network idle window) |
| 14 | Actionability waits and resilient selectors | Playwright locators/auto-wait | 4 | M | Reduces flaky automation scripts | Add selector fallback chain: role/text/css/xpath + visibility/actionability checks |
| 15 | Authentication bridge (API->browser state) | Playwright APIRequestContext + storage state reuse | 4 | M | Faster login and less brittle UI auth automation | Add helper to inject auth cookies/tokens into page requests safely |
| 16 | Domain circuit breaker and backoff | General crawler reliability patterns | 4 | S | Prevents burning proxies and quota during outages/blocks | Per-domain breaker keyed by block-rate + error-rate with cool-down probes |
| 17 | Cost-aware retry budgeter | Existing metadata billing headers + orchestration | 4 | S | Keeps spend bounded under hard anti-bot targets | Stop conditions based on `cost per successful page` and dynamic attempt caps |
| 18 | Fingerprint alignment tied to proxy identity | Crawlee fingerprint tied to proxy URL | 4 | M | Consistent identity lowers suspicion | Stable fingerprint seed per `proxy-session + domain`, rotate on block retire |
| 19 | Regional strategy planner | Proxy geo options + emulation options | 3 | S | Improves localized content availability and fewer geo blocks | Auto-select region via first-pass geo hints in response content |
| 20 | Humanized interaction generator | Browserless bot-detection guidance + action APIs | 3 | M | Helps on behavior-sensitive flows | Deterministic jitter profiles (typing cadence, scroll rhythm, micro-pauses) |
| 21 | Live intervention hook protocol | Browserless live debugger/hybrid intervention | 3 | M | Lets human salvage hard flows without restarting | Pause/resume checkpoints with serialized run context for takeover tooling |
| 22 | Extraction contract + validators | Robust scraping pipeline patterns | 3 | S | Catches silent data corruption | JSON schema validation + anomaly thresholds before accepting output |
| 23 | Response transformation pipeline | Playwright/Puppeteer request-response handling patterns | 3 | M | Enables lightweight in-flight content shaping | Add modular transforms (strip scripts, normalize prices, canonicalize links) |
| 24 | Cross-backend protocol adapter | Puppeteer WebDriver BiDi trajectory + Selenium BiDi direction | 3 | L | Future-proofs against backend/protocol drift | Keep high-level orchestration independent of transport/backend details |
| 25 | Session replay artifact view | Browserless session replay concept | 3 | M | Great debugging aid for failed runs | Produce lightweight replay JSON for timeline playback in internal tooling |

## Fastest High-Impact Implementation Set (Recommended First 10)

1. Proxy health + tiered failover by domain  
2. Challenge policy state machine (signal-based)  
3. Durable session store (cookies + local state)  
4. Persona packs v2 (domain-specific identity bundles)  
5. Adaptive concurrency autoscaler  
6. Request frontier with dedupe/reclaim  
7. HAR record/replay mode for deterministic tests  
8. Cooperative interception rule engine  
9. WebSocket interception hooks  
10. Trace bundle for each run

## Sources

- Puppeteer docs:
  - Request interception: https://pptr.dev/guides/network-interception
  - Network logging: https://pptr.dev/guides/network-logging
  - Cookies/context isolation: https://pptr.dev/guides/cookies
  - Browser contexts: https://pptr.dev/guides/browser-management
  - Tracing: https://pptr.dev/api/puppeteer.tracing
  - Network condition emulation: https://pptr.dev/api/puppeteer.page.emulatenetworkconditions
  - Offline mode: https://pptr.dev/api/puppeteer.page.setofflinemode
  - Timezone emulation: https://pptr.dev/api/puppeteer.page.emulatetimezone
  - Locators and interaction model: https://pptr.dev/next/guides/page-interactions
  - WebDriver BiDi support status: https://pptr.dev/webdriver-bidi
- Playwright docs:
  - BrowserContext API (`routeFromHAR`, `routeWebSocket`, `storageState`, tracing): https://playwright.dev/docs/api/class-browsercontext
  - Network and mocking: https://playwright.dev/docs/network
  - HAR mocking guide: https://playwright.dev/docs/mock
  - Locators: https://playwright.dev/docs/locators
  - Auto-wait/actionability: https://playwright.dev/docs/actionability
  - Trace viewer: https://playwright.dev/docs/trace-viewer
  - Auth state reuse: https://playwright.dev/docs/auth
  - Best practices: https://playwright.dev/docs/best-practices
- Crawlee docs:
  - AutoscaledPool: https://crawlee.dev/api/3.11/core/class/AutoscaledPool
  - SessionPool: https://crawlee.dev/api/3.11/core/class/SessionPool
  - Session options: https://crawlee.dev/js/api/3.2/core/interface/SessionOptions
  - ProxyConfiguration and session-bound proxy URLs: https://crawlee.dev/api/3.11/core/class/ProxyConfiguration
  - Proxy management guide: https://crawlee.dev/docs/3.4/guides/proxy-management
  - RequestQueue/reclaim: https://crawlee.dev/api/3.10/core/class/RequestQueue
  - Browser Pool + fingerprint behavior: https://crawlee.dev/js/api/3.4/browser-pool
  - Impit HTTP/TLS fingerprinting: https://crawlee.dev/js/docs/3.15/guides/impit-http-client
- Browserless docs:
  - Session reconnect: https://docs.browserless.io/browserql/session-management/reconnect-to-browserless
  - Persisting session state: https://docs.browserless.io/browserql/session-management/persisting-state
  - Standard sessions: https://docs.browserless.io/baas/session-management/standard-sessions
  - CAPTCHA solving: https://docs.browserless.io/baas/bot-detection/captchas
  - Bot detection overview (stealth/proxies/fingerprints): https://docs.browserless.io/browserql/bot-detection/overview
  - Session replay: https://docs.browserless.io/baas/features/session-replay
  - Live debugger: https://docs.browserless.io/baas/interactive-browser-sessions/live-debugger
- Puppeteer plugin ecosystem:
  - `puppeteer-extra` framework: https://github.com/berstend/puppeteer-extra
  - Stealth plugin: https://www.npmjs.com/package/puppeteer-extra-plugin-stealth
  - Adblocker plugin: https://www.npmjs.com/package/puppeteer-extra-plugin-adblocker
  - Block resources plugin: https://www.npmjs.com/package/puppeteer-extra-plugin-block-resources
  - Recaptcha plugin family entry: https://www.npmjs.com/package/@extra/recaptcha
  - Puppeteer cluster: https://github.com/thomasdondorf/puppeteer-cluster
- Selenium BiDi direction:
  - BiDi overview: https://www.selenium.dev/documentation/webdriver/bidi/w3c/
  - BiDi network features: https://www.selenium.dev/documentation/webdriver/bidi/network/

