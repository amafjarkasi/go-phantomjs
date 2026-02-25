# PhantomJsCloud API Discoveries

This document serves to log advanced functionalities and nuances discovered within the [PhantomJsCloud API Documentation](https://phantomjscloud.com/docs/http-api/index.html) that have been mapped directly into this Golang client library (`go-phantomjs`).

These discoveries were made during the deep dive into their raw TS interfaces, advanced endpoints, and script-injection systems.

## 1. The `Automation API` vs `Scripts API`

The `Scripts` API (injecting JS into `domReady`, `loadFinished` phases via array injections) is considered **deprecated** or vastly inferior to the new **Automation API**, driven by the `overseerScript` parameter.

* **Actionable Implementation**: We built `OverseerScriptBuilder` to allow Golang developers to dynamically chain asynchronous Puppeteer/Playwright-like commands natively instead of parsing array strings for injecting scripts. The builder handles `AddScriptTag`, `Evaluate`, `WaitForSelector`, `Type`, `Click`, and injecting arbitrary execution blocks safely.

## 2. Paginating / Isolating the Output (`automationResult`, `scriptOutput`)

A major hurdle when executing client-side scripts is capturing the specific JSON object extracted by the script itself rather than returning the raw HTML structure.

* **Discovery**: The API has native properties for this on the `PageResponse` JSON structure directly: `automationResult` and `scriptOutput`.
* **Actionable Implementation**: We ensured that `PageResponse.AutomationResult` is strictly mapped as a generic `interface{}` parameter. This means whenever users execute `overseerScript` payloads containing logic like `page.meta.store.set("myScrapedData", myObject)`, it automatically natively populates back into the top-level API payload instead of the user needing to parse HTML logs manually.

## 3. Network Lifecycle / Event Properties (`eventPhase`, `doneDetail`)

Debugging failures on headless Chromium environments usually involves heavy manual network interception. PhantomJsCloud wraps this lifecycle automatically.

* **Discovery**: The interfaces document detailed `eventPhase`, `pageExecLastWaitedOn`, and stringified `doneDetail` values on Chrome-based endpoints whenever executions fail or time out.
* **Actionable Implementation**: Added mapping for `EventPhase` (string), `Resources` (`[]interface{}` array containing load time analysis), and expanded `Errors` arrays to the `PageResponse` struct.

## 4. Query JSON Payload Compression

When fetching APIs or extracting pure JSON blobs using PhantomJsCloud, the response is typically bloated with surrounding HTML `<pre>` tags or full `UserResponse` structures.

* **Discovery**: The parameter `outputAsJson: true` enforces that the core `content` field is returned. More importantly, using `queryJson` allows developers to parse response output via JSONPath directly *before* the output hits the wire, vastly reducing egress costs.
* **Actionable Implementation**: We added `QueryJson` directly to `UserResponse` because it modifies the structure natively on the top level output.

## 5. Security Context Overrides (`RequestSettings`)

Extracting data from highly-secure single-page applications (SPAs) often requires bypassing standard headless browser locks.

* **Discovery**: The `RequestSettings` interface provides advanced overrides: `disableSecureHeaders: true` (disables CSP), `webSecurityEnabled: false` (turns off CORS restriction checks), and `xssAuditingEnabled`.
* **Actionable Implementation**: Explicitly mapped these overrides as boolean native properties on the `RequestSettings` payload for use in extracting heavily obfuscated SPAs.

## 6. Real-Time Resource Modification

Fetching heavy external scripts, ads, and fonts during processing uses up quota credits faster than needed.

* **Discovery**: The API supports injecting regex string arrays against network traffic allowing you to drop it via `ResourceModifier`. What's cooler is that that modifier supports the `changeUrl` property which redirects matched network resources elsewhere.
* **Actionable Implementation**: Mapped `ResourceModifier` handling `Regex`, `Type`, `IsBlacklisted`, `SetHeader` and most importantly `ChangeUrl` to easily override ad networks or intercept trackers dynamically before they parse.

## 7. Metadata and Quota Headers (`pjsc-*`)

Most APIs throw 429 warnings on quota usage directly in the API payload payload. PhantomJsCloud uses custom HTTP Response Headers so that metrics exist *even on basic HTML returns*.

* **Discovery**: We identified `pjsc-billing-credit-cost`, `pjsc-content-status-code`, and `pjsc-content-done-when`.
* **Actionable Implementation**: Built `parseMetadata()` manually mapping `http.Header` injections into a natively bound struct `ResponseMetadata` attached whenever the struct is deserialized.

## 8. Batch Loading via UserRequest

While `PageRequest` is the fundamental building block for loading a page, PhantomJsCloud explicitly supports bulk processes internally via the `UserRequest` array structure.

* **Discovery**: Submitting a JSON array of `pages` to the root API will execute all of them internally but *only* render the last successfully loaded page. This is surprisingly useful for executing multi-stage logins where each stage triggers a top-level page redirection.
* **Actionable Implementation**: We fully fleshed out the `UserRequest` struct, giving users the explicit design choice between submitting independent `PageRequest` definitions or a batched `UserRequest`. Furthermore, `UserRequest` allows configuring global overrides like `BackendDiscrete` and `OutputAsJson`.

## 9. Specialized Proxy Contexts (IProxyOptions)

Bypassing region locks usually involves setting simple strings, but PhantomJsCloud provides much more granular routing options that aren't obvious at first glance.

* **Discovery**: The `Proxy` parameter is a generic union type: `string | IProxyOptions`. It wasn't just a basic URL override string. Using the `IProxyOptions` object format, it allows targeting internal service proxies like anonymizing traffic specifically through German IPs `(ProxyBuiltin{Location: "de"})`, routing through datacenter IPs vs residential, or securely managing proxy HTTP basic auth headers directly without appending them to the connection string manually.
* **Actionable Implementation**: Mapped the union property natively utilizing Go's `interface{}` parameter mechanism. Also added structured classes for `ProxyOptions`, `ProxyBuiltin`, and `ProxyCustom` allowing strongly typed programmatic access instead of throwing string concatenation at the wall.

## 10. Granular PDF & Screenshot Options

Generating PDFs correctly from a browser engine requires intricate tweaks to trick the page into formatting for print layout.

* **Discovery**: The `PdfOptions` interface contains specific nuances beyond just margins. Properties like `omitBackground` (forcing PDFs to render transparent/removed background graphics), `preferCSSPageSize` (forcing the PDF to adhere to the webpage's `@page` CSS rule if defined), and `onepageFudgeFactor` (fixing standard bugs when converting entire continuous scroll pages into single-page PDFs).
* **Actionable Implementation**: Updated the struct to strictly map `OmitBackground`, `PreferCSSPageSize`, `OnepageFudgeFactor`, and explicitly supported integer timeouts overrides just for the PDF rendering phase itself.

## 11. Automating Bot-Evasion (Puppeteer Stealth Concepts)

When navigating restricted platforms (like LinkedIn or Cloudflare-protected sites), basic headless configurations often trigger captchas.

* **Discovery**: Unlike pure Node.js environments where you can inject `puppeteer-extra-plugin-stealth` *before* the browser launches, PhantomJsCloud handles the browser context natively. However, by leveraging our `OverseerScriptBuilder` functions like `SetUserAgent`, `SetExtraHTTPHeaders`, and `DeleteCookie`, we can effectively mimic heavy evasion techniques dynamically inside the cloud container. Also, injecting `Raw` evaluation scripts modifying `navigator.webdriver` assists in masking the headless state.
* **Actionable Implementation**: We appended heavy-duty interaction features like `MouseMove`, `MouseClickPosition` (bypassing DOM event listeners and triggering OS-level coordinates), `ScrollToBottom`, and `SetUserAgent`. These natively mimic human interactions commonly used by stealth scraping frameworks directly via PhantomJSCloud's Chrome Debug protocol translation.

## 12. Complete Proxy Datacenter Geolocation List

Standard documentation hinted at a handful of built-in proxies, but reading raw interface dumps (`/examples/helpers/proxy-builtin-locations`) revealed a massive global network.

* **Discovery**: PhantomJsCloud inherently supports a massive list of built-in datacenter exits spanning nearly two dozen countries implicitly formatted as two-letter country codes in their `ProxyBuiltin` objects (e.g., `"sg"` for Singapore, `"ie"` for Ireland, etc.).
* **Actionable Implementation**: Scraped the definitive list of backend keys and generated native Golang constants (`ProxyLocationUS`, `ProxyLocationDE`, `ProxyLocationAE`, etc.) mapping to every single built-in geo-routing target. This saves developers from guessing API string formatting or referencing external charts.
