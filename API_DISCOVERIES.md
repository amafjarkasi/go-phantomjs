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
