package phantomjscloud

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/amafjarkasi/go-phantomjs/ext/stealth"
	"github.com/amafjarkasi/go-phantomjs/ext/useragents"
)

// Proxy Builtin Locations (to be used with ProxyBuiltin.Location)
const (
	ProxyLocationUS = "us"
	ProxyLocationUK = "uk"
	ProxyLocationDE = "de"
	ProxyLocationFR = "fr"
	ProxyLocationCA = "ca"
	ProxyLocationJP = "jp"
	ProxyLocationAU = "au"
)

// ProxyBuiltin options let you use a standard rotating residential proxy without
// managing your own pool. Just specify the country.
var (
	ProxyAnonUS = ProxyBuiltin{Location: ProxyLocationUS}
	ProxyAnonUK = ProxyBuiltin{Location: ProxyLocationUK}
	ProxyAnonDE = ProxyBuiltin{Location: ProxyLocationDE}
	ProxyAnonFR = ProxyBuiltin{Location: ProxyLocationFR}
	ProxyAnonCA = ProxyBuiltin{Location: ProxyLocationCA}
	ProxyAnonJP = ProxyBuiltin{Location: ProxyLocationJP}
	ProxyAnonAU = ProxyBuiltin{Location: ProxyLocationAU}
)

// OverseerScriptBuilder helps construct complex Automation API scripts safely.
type OverseerScriptBuilder struct {
	script strings.Builder
}

// NewOverseerScriptBuilder returns a builder that constructs a PhantomJsCloud
// overseerScript step by step. Call Build() to get the final script string,
// or pass the builder directly to FetchWithAutomation or WithOverseerScriptBuilder.
func NewOverseerScriptBuilder() *OverseerScriptBuilder {
	return &OverseerScriptBuilder{}
}

func (b *OverseerScriptBuilder) writeJSString(s string) {
	raw, _ := json.Marshal(s)
	b.script.Write(raw)
}

// AddScriptTag injects an external script into the page.
func (b *OverseerScriptBuilder) AddScriptTag(url string) *OverseerScriptBuilder {
	b.script.WriteString("await page.addScriptTag({url: ")
	b.writeJSString(url)
	b.script.WriteString("});\n")
	return b
}

// Evaluate appends an evaluation block. Make sure functionBody is a valid JS function or string.
func (b *OverseerScriptBuilder) Evaluate(functionBody string) *OverseerScriptBuilder {
	b.script.WriteString("await page.evaluate(")
	b.script.WriteString(functionBody)
	b.script.WriteString(");\n")
	return b
}

// WaitForNavigation waits for a navigation event to complete (default: load).
func (b *OverseerScriptBuilder) WaitForNavigation() *OverseerScriptBuilder {
	b.script.WriteString("await page.waitForNavigation();\n")
	return b
}

// WaitForNavigationEvent waits for a specific navigation event (load, domcontentloaded, networkidle0, networkidle2).
func (b *OverseerScriptBuilder) WaitForNavigationEvent(event string) *OverseerScriptBuilder {
	b.script.WriteString("await page.waitForNavigation({waitUntil: ")
	b.writeJSString(event)
	b.script.WriteString("});\n")
	return b
}

// WaitForNetworkIdle is a convenience wrapper that waits for network inactivity.
func (b *OverseerScriptBuilder) WaitForNetworkIdle(idleConnections, idleMs int) *OverseerScriptBuilder {
	// Note: Puppeteer usually uses 'networkidle0' or 'networkidle2' via waitForNavigation.
	// This helper constructs a custom wait logic or uses the built-in string if standard.
	// For standard Puppeteer 'networkidle0' (0 connections for 500ms):
	if idleConnections == 0 && idleMs == 500 {
		return b.WaitForNavigationEvent("networkidle0")
	}
	if idleConnections == 2 && idleMs == 500 {
		return b.WaitForNavigationEvent("networkidle2")
	}
	// Fallback or custom logic could be implemented here, but standard API usually suffices.
	return b.WaitForNavigationEvent("networkidle0")
}

// WaitForSelector waits for an element to appear in the DOM.
func (b *OverseerScriptBuilder) WaitForSelector(selector string) *OverseerScriptBuilder {
	b.script.WriteString("await page.waitForSelector(")
	b.writeJSString(selector)
	b.script.WriteString(");\n")
	return b
}

// Click clicks on an element matching the selector.
func (b *OverseerScriptBuilder) Click(selector string) *OverseerScriptBuilder {
	b.script.WriteString("await page.click(")
	b.writeJSString(selector)
	b.script.WriteString(");\n")
	return b
}

// ClickAndWaitForNavigation clicks an element and simultaneously waits for navigation to complete.
// This prevents race conditions where the navigation happens before the wait starts.
func (b *OverseerScriptBuilder) ClickAndWaitForNavigation(selector string) *OverseerScriptBuilder {
	b.script.WriteString("await Promise.all([\n")
	b.script.WriteString("  page.waitForNavigation(),\n")
	b.script.WriteString("  page.click(")
	b.writeJSString(selector)
	b.script.WriteString(")\n")
	b.script.WriteString("]);\n")
	return b
}

// Type types text into an element.
func (b *OverseerScriptBuilder) Type(selector, text string, delayMs int) *OverseerScriptBuilder {
	b.script.WriteString("await page.type(")
	b.writeJSString(selector)
	b.script.WriteString(", ")
	b.writeJSString(text)
	if delayMs > 0 {
		b.script.WriteString(",{delay:")
		b.script.WriteString(strconv.Itoa(delayMs))
		b.script.WriteString("});\n")
	} else {
		b.script.WriteString(");\n")
	}
	return b
}

// Raw appends a raw Javascript code block directly.
func (b *OverseerScriptBuilder) Raw(code string) *OverseerScriptBuilder {
	b.script.WriteString(code)
	b.script.WriteString("\n")
	return b
}

// Goto navigates to a URL and waits for the default load event.
func (b *OverseerScriptBuilder) Goto(url string) *OverseerScriptBuilder {
	b.script.WriteString("await page.goto(")
	b.writeJSString(url)
	b.script.WriteString(");\n")
	return b
}

// GotoWithWaitUntil navigates to a URL and waits for a specific load event.
// Common values: "load", "domcontentloaded", "networkidle0", "networkidle2".
// Prefer this over Goto + WaitForNavigationEvent for SPAs that fire no traditional load events.
func (b *OverseerScriptBuilder) GotoWithWaitUntil(url, waitUntil string) *OverseerScriptBuilder {
	b.script.WriteString("await page.goto(")
	b.writeJSString(url)
	b.script.WriteString(", {waitUntil: ")
	b.writeJSString(waitUntil)
	b.script.WriteString("});\n")
	return b
}

// KeyboardPress presses a specific key (e.g., 'Backspace', 'Enter') a certain number of times.
func (b *OverseerScriptBuilder) KeyboardPress(key string, times int) *OverseerScriptBuilder {
	if times <= 1 {
		b.script.WriteString("await page.keyboard.press(")
		b.writeJSString(key)
		b.script.WriteString(");\n")
	} else {
		b.script.WriteString("await page.keyboard.press(")
		b.writeJSString(key)
		b.script.WriteString(", {times: ")
		b.script.WriteString(strconv.Itoa(times))
		b.script.WriteString("});\n")
	}
	return b
}

// WaitForDelay pauses script execution for a specified number of milliseconds.
func (b *OverseerScriptBuilder) WaitForDelay(ms int) *OverseerScriptBuilder {
	b.script.WriteString("await page.waitForDelay(")
	b.script.WriteString(strconv.Itoa(ms))
	b.script.WriteString(");\n")
	return b
}

// RenderContent tells PhantomJS to capture the HTML content of the page immediately.
func (b *OverseerScriptBuilder) RenderContent() *OverseerScriptBuilder {
	b.script.WriteString("page.render.content();\n")
	return b
}

// RenderScreenshot tells PhantomJS to capture a screenshot immediately. Wait triggers synchronous render.
func (b *OverseerScriptBuilder) RenderScreenshot(wait bool) *OverseerScriptBuilder {
	if wait {
		b.script.WriteString("await page.render.screenshot();\n")
	} else {
		b.script.WriteString("page.render.screenshot();\n")
	}
	return b
}

// ManualWait informs the page renderer that the script requires manual management, disabling automatic completion.
func (b *OverseerScriptBuilder) ManualWait() *OverseerScriptBuilder {
	b.script.WriteString("page.manualWait();\n")
	return b
}

// Done signals manual termination to the renderer. Must be paired with ManualWait.
func (b *OverseerScriptBuilder) Done() *OverseerScriptBuilder {
	b.script.WriteString("page.done();\n")
	return b
}

// Hover simulates resting the mouse over an element.
func (b *OverseerScriptBuilder) Hover(selector string) *OverseerScriptBuilder {
	b.script.WriteString("await page.hover(")
	b.writeJSString(selector)
	b.script.WriteString(");\n")
	return b
}

// Focus focuses on an element.
func (b *OverseerScriptBuilder) Focus(selector string) *OverseerScriptBuilder {
	b.script.WriteString("await page.focus(")
	b.writeJSString(selector)
	b.script.WriteString(");\n")
	return b
}

// Select selects options in a dropdown.
func (b *OverseerScriptBuilder) Select(selector string, values ...string) *OverseerScriptBuilder {
	b.script.WriteString("await page.select(")
	b.writeJSString(selector)
	b.script.WriteString(", ")
	raw, _ := json.Marshal(values)
	// Remove brackets if we want a comma separated list of arguments,
	// but page.select in Puppeteer takes (selector, ...values)
	// wait, page.select(selector, ...values) means values are passed as separate arguments.
	// The original code was: await page.select('selector', 'v1', 'v2');
	// json.Marshal(values) gives ["v1", "v2"].
	// We need to strip the brackets to get "v1", "v2"
	valStr := string(raw)
	if len(valStr) >= 2 {
		b.script.WriteString(valStr[1 : len(valStr)-1])
	}
	b.script.WriteString(");\n")
	return b
}

// Reload refreshes the current page.
func (b *OverseerScriptBuilder) Reload() *OverseerScriptBuilder {
	b.script.WriteString("await page.reload();\n")
	return b
}

// ClearInput is a convenience method that manually clears a text field by evaluating Javascript.
func (b *OverseerScriptBuilder) ClearInput(selector string) *OverseerScriptBuilder {
	b.script.WriteString("await page.evaluate((sel) => { document.querySelector(sel).value = ''; }, ")
	b.writeJSString(selector)
	b.script.WriteString(");\n")
	return b
}

// ScrollBy scrolls the page by a specific X and Y pixel offset.
func (b *OverseerScriptBuilder) ScrollBy(x, y int) *OverseerScriptBuilder {
	b.script.WriteString("await page.evaluate((x, y) => { window.scrollBy(x, y); }, ")
	b.script.WriteString(strconv.Itoa(x))
	b.script.WriteString(", ")
	b.script.WriteString(strconv.Itoa(y))
	b.script.WriteString(");\n")
	return b
}

// ScrollToBottom scrolls the entire page to the absolute bottom perfectly matching document limits. Ideal for infinite scrolling loaders.
func (b *OverseerScriptBuilder) ScrollToBottom() *OverseerScriptBuilder {
	b.script.WriteString("await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));\n")
	return b
}

// AddStyleTag injects custom CSS into the page.
func (b *OverseerScriptBuilder) AddStyleTag(cssContent string) *OverseerScriptBuilder {
	b.script.WriteString("await page.addStyleTag({content: ")
	b.writeJSString(cssContent)
	b.script.WriteString("});\n")
	return b
}

// SetViewport dynamically overrides the browser viewport mid-script.
func (b *OverseerScriptBuilder) SetViewport(width, height int) *OverseerScriptBuilder {
	b.script.WriteString("await page.setViewport({width: ")
	b.script.WriteString(strconv.Itoa(width))
	b.script.WriteString(", height: ")
	b.script.WriteString(strconv.Itoa(height))
	b.script.WriteString("});\n")
	return b
}

// SetUserAgent dynamically overrides the browser user agent natively mid-script.
func (b *OverseerScriptBuilder) SetUserAgent(userAgent string) *OverseerScriptBuilder {
	b.script.WriteString("await page.setUserAgent(")
	b.writeJSString(userAgent)
	b.script.WriteString(");\n")
	return b
}

// SetExtraHTTPHeaders dynamically injects new global headers into the underlying browser mid-script.
// Input map is converted natively to a JSON object payload.
func (b *OverseerScriptBuilder) SetExtraHTTPHeaders(headers map[string]string) *OverseerScriptBuilder {
	b.script.WriteString("await page.setExtraHTTPHeaders(")
	raw, _ := json.Marshal(headers)
	b.script.Write(raw)
	b.script.WriteString(");\n")
	return b
}

// WaitForFunction pauses execution until the provided Javascript function returns truthy.
func (b *OverseerScriptBuilder) WaitForFunction(jsFunc string) *OverseerScriptBuilder {
	b.script.WriteString("await page.waitForFunction(")
	b.script.WriteString(jsFunc)
	b.script.WriteString(");\n")
	return b
}

// SetCookie adds a cookie directly into the browser context.
func (b *OverseerScriptBuilder) SetCookie(name, value, domain string) *OverseerScriptBuilder {
	b.script.WriteString("await page.setCookie({name: ")
	b.writeJSString(name)
	b.script.WriteString(", value: ")
	b.writeJSString(value)
	b.script.WriteString(", domain: ")
	b.writeJSString(domain)
	b.script.WriteString("});\n")
	return b
}

// DeleteCookie removes a specific cookie from the browser context.
func (b *OverseerScriptBuilder) DeleteCookie(name, url string) *OverseerScriptBuilder {
	b.script.WriteString("await page.deleteCookie({name: ")
	b.writeJSString(name)
	b.script.WriteString(", url: ")
	b.writeJSString(url)
	b.script.WriteString("});\n")
	return b
}

// MouseMove simulates moving the mouse cursor to a specific absolute X,Y coordinate.
func (b *OverseerScriptBuilder) MouseMove(x, y int) *OverseerScriptBuilder {
	b.script.WriteString("await page.mouse.move(")
	b.script.WriteString(strconv.Itoa(x))
	b.script.WriteString(", ")
	b.script.WriteString(strconv.Itoa(y))
	b.script.WriteString(");\n")
	return b
}

// MouseClickPosition simulates a native operating system level mouse click on a specific absolute X,Y coordinate rather than relying on DOM targeting.
func (b *OverseerScriptBuilder) MouseClickPosition(x, y int) *OverseerScriptBuilder {
	b.script.WriteString("await page.mouse.click(")
	b.script.WriteString(strconv.Itoa(x))
	b.script.WriteString(", ")
	b.script.WriteString(strconv.Itoa(y))
	b.script.WriteString(");\n")
	return b
}

// WaitForXPath explicitly waits for a specific XPath block to render into the DOM.
func (b *OverseerScriptBuilder) WaitForXPath(xpath string) *OverseerScriptBuilder {
	b.script.WriteString("await page.waitForXPath(")
	b.writeJSString(xpath)
	b.script.WriteString(");\n")
	return b
}

// ApplyStealth injects a comprehensive suite of browser fingerprinting evasions
// derived from puppeteer-extra-plugin-stealth. Spoofs navigator, WebGL, chrome,
// iframe, media codec, and many other APIs that bot-detection scripts probe.
//
// Call this early in your script, ideally before Goto, so the evasions are
// registered before any page content is loaded.
//
// The JS payload lives in ext/stealth/evasions.js and can be regenerated with:
//
//	node scripts/gen_stealth.js
func (b *OverseerScriptBuilder) ApplyStealth() *OverseerScriptBuilder {
	b.script.WriteString("await page.evaluateOnNewDocument(")
	b.script.WriteString(stealth.JS)
	b.script.WriteString(");\n")
	return b
}

// UseProfile sets both the user agent string and the accompanying request headers
// (Accept, Accept-Language, Sec-CH-UA, Sec-Fetch-* etc.) in a single call.
// This is the recommended way to spoof a specific browser fingerprint because
// mismatched UA/header combinations are a common bot signal.
//
// Use the profile constructors in ext/useragents:
//
//	builder.UseProfile(useragents.ChromeWindowsProfile())
func (b *OverseerScriptBuilder) UseProfile(p useragents.Profile) *OverseerScriptBuilder {
	fmt.Fprintf(&b.script, "await page.setUserAgent(%q);\n", p.UserAgent)
	if len(p.Headers) > 0 {
		raw, _ := json.Marshal(p.Headers)
		b.script.WriteString("await page.setExtraHTTPHeaders(")
		b.script.Write(raw)
		b.script.WriteString(");\n")
	}
	return b
}

// ApplyViewport applies a fully configured Viewport struct to the page — supporting
// DeviceScaleFactor, IsMobile, HasTouch, and IsLandscape flags that SetViewport(w,h)
// doesn't expose. Use a named preset from ext/viewport by passing its Viewport field:
//
//	builder.ApplyViewport(viewport.MobilePortrait.Viewport)
func (b *OverseerScriptBuilder) ApplyViewport(v Viewport) *OverseerScriptBuilder {
	fmt.Fprintf(&b.script,
		"await page.setViewport({width:%d,height:%d,deviceScaleFactor:%g,isMobile:%t,hasTouch:%t,isLandscape:%t});\n",
		v.Width, v.Height, v.DeviceScaleFactor, v.IsMobile, v.HasTouch, v.IsLandscape,
	)
	return b
}

// DragAndDrop simulates dragging an element from one selector to another.
func (b *OverseerScriptBuilder) DragAndDrop(sourceSelector, targetSelector string) *OverseerScriptBuilder {
	b.script.WriteString("await page.dragAndDrop(")
	b.writeJSString(sourceSelector)
	b.script.WriteString(", ")
	b.writeJSString(targetSelector)
	b.script.WriteString(");\n")
	return b
}

// WaitForUrl waits until the page URL contains the specified string.
func (b *OverseerScriptBuilder) WaitForUrl(urlFragment string) *OverseerScriptBuilder {
	b.script.WriteString("await page.waitForFunction((url) => window.location.href.includes(url), {}, ")
	b.writeJSString(urlFragment)
	b.script.WriteString(");\n")
	return b
}

// GoBack navigates to the previous page in history.
func (b *OverseerScriptBuilder) GoBack() *OverseerScriptBuilder {
	b.script.WriteString("await page.goBack();\n")
	return b
}

// GoForward navigates to the next page in history.
func (b *OverseerScriptBuilder) GoForward() *OverseerScriptBuilder {
	b.script.WriteString("await page.goForward();\n")
	return b
}

// WaitUntilVisible waits for an element to be visible in the viewport.
func (b *OverseerScriptBuilder) WaitUntilVisible(selector string) *OverseerScriptBuilder {
	fmt.Fprintf(&b.script, "await page.waitForFunction((s) => {\n"+
		"  const el = document.querySelector(s);\n"+
		"  if (!el) return false;\n"+
		"  const style = window.getComputedStyle(el);\n"+
		"  return style && style.display !== 'none' && style.visibility !== 'hidden' && style.opacity !== '0';\n"+
		"}, {}, %q);\n", selector)
	return b
}

// WaitUntilHidden waits for an element to be removed from the DOM or hidden via CSS.
func (b *OverseerScriptBuilder) WaitUntilHidden(selector string) *OverseerScriptBuilder {
	fmt.Fprintf(&b.script, "await page.waitForFunction((s) => {\n"+
		"  const el = document.querySelector(s);\n"+
		"  if (!el) return true;\n"+
		"  const style = window.getComputedStyle(el);\n"+
		"  return !style || style.display === 'none' || style.visibility === 'hidden' || style.opacity === '0';\n"+
		"}, {}, %q);\n", selector)
	return b
}

// ClickByText clicks the first element that contains the specified text.
func (b *OverseerScriptBuilder) ClickByText(text string) *OverseerScriptBuilder {
	fmt.Fprintf(&b.script, "await page.evaluate((t) => {\n"+
		"  const xpath = `//*[contains(text(),'${t}')]`;\n"+
		"  const matchingElement = document.evaluate(xpath, document, null, XPathResult.FIRST_ORDERED_NODE_TYPE, null).singleNodeValue;\n"+
		"  if (matchingElement) matchingElement.click();\n"+
		"}, %q);\n", text)
	return b
}

// ScrollToElement scrolls the page until the specified element is in view.
func (b *OverseerScriptBuilder) ScrollToElement(selector string) *OverseerScriptBuilder {
	fmt.Fprintf(&b.script, "await page.evaluate((s) => {\n"+
		"  const el = document.querySelector(s);\n"+
		"  if (el) el.scrollIntoView({ behavior: 'smooth', block: 'center' });\n"+
		"}, %q);\n", selector)
	return b
}

// HighlightElement draws a red border around an element — useful for debugging screenshots.
func (b *OverseerScriptBuilder) HighlightElement(selector string) *OverseerScriptBuilder {
	fmt.Fprintf(&b.script, "await page.evaluate((s) => {\n"+
		"  const el = document.querySelector(s);\n"+
		"  if (el) el.style.border = '5px solid red';\n"+
		"}, %q);\n", selector)
	return b
}

// SelectByLabel selects a dropdown option based on its visible label text.
func (b *OverseerScriptBuilder) SelectByLabel(selector, label string) *OverseerScriptBuilder {
	fmt.Fprintf(&b.script, "await page.evaluate((s, l) => {\n"+
		"  const select = document.querySelector(s);\n"+
		"  if (!select) return;\n"+
		"  const option = Array.from(select.options).find(o => o.text === l);\n"+
		"  if (option) { select.value = option.value; select.dispatchEvent(new Event('change')); }\n"+
		"}, %q, %q);\n", selector, label)
	return b
}

// Build returns the finalized script.
func (b *OverseerScriptBuilder) Build() string {
	s := b.script.String()
	if strings.Contains(s, "__pjsc_result") {
		s += "window.__pjsc_result;\n"
	}
	return s
}
