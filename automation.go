package phantomjscloud

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/amafjarkasi/go-phantomjs/ext/stealth"
	"github.com/amafjarkasi/go-phantomjs/ext/useragents"
)

// Proxy Builtin Locations (to be used with ProxyBuiltin.Location)
const (
	ProxyLocationAny = "any" // Worldwide (Global)
	ProxyLocationAU  = "au"  // Australia
	ProxyLocationBR  = "br"  // Brazil
	ProxyLocationCN  = "cn"  // China
	ProxyLocationDE  = "de"  // Germany
	ProxyLocationES  = "es"  // Spain
	ProxyLocationFR  = "fr"  // France
	ProxyLocationGB  = "gb"  // Great Britain
	ProxyLocationHK  = "hk"  // Hong Kong
	ProxyLocationID  = "id"  // Indonesia
	ProxyLocationIL  = "il"  // Israel
	ProxyLocationIN  = "in"  // India
	ProxyLocationIT  = "it"  // Italy
	ProxyLocationJP  = "jp"  // Japan
	ProxyLocationKR  = "kr"  // South Korea
	ProxyLocationMX  = "mx"  // Mexico
	ProxyLocationMY  = "my"  // Malaysia
	ProxyLocationNL  = "nl"  // Netherlands
	ProxyLocationRU  = "ru"  // Russia
	ProxyLocationSA  = "sa"  // Saudi Arabia
	ProxyLocationSG  = "sg"  // Singapore
	ProxyLocationTH  = "th"  // Thailand
	ProxyLocationTR  = "tr"  // Turkey
	ProxyLocationTW  = "tw"  // Taiwan
	ProxyLocationUS  = "us"  // United States
	ProxyLocationAE  = "ae"  // United Arab Emirates
)

// Legacy Proxy Types (for simple proxy string assignment)
const (
	ProxyAnonAny = "anon-any" // Random Anonymous proxy
	ProxyAnonUK  = "anon-uk"  // Anonymous proxy in United Kingdom
	ProxyAnonUS  = "anon-us"  // Anonymous proxy in United States
	ProxyAnonEU  = "anon-eu"  // Anonymous proxy in Europe
	ProxyGeoUS   = "geo-us"   // Static IP in United States
	ProxyGeoUK   = "geo-uk"   // Static IP in United Kingdom
	ProxyGeoRU   = "geo-ru"   // Static IP in Russia
	ProxyGeoCN   = "geo-cn"   // Static IP in China
)

// OverseerScriptBuilder helps construct complex Automation API scripts safely.
type OverseerScriptBuilder struct {
	script string
}

// NewOverseerScriptBuilder returns a builder that constructs a PhantomJsCloud
// overseerScript step by step. Call Build() to get the final script string,
// then assign it to PageRequest.OverseerScript or pass the builder directly
// to FetchWithAutomation or WithOverseerScriptBuilder.
func NewOverseerScriptBuilder() *OverseerScriptBuilder {
	return &OverseerScriptBuilder{
		script: "",
	}
}

// AddScriptTag injects an external script into the page.
func (b *OverseerScriptBuilder) AddScriptTag(url string) *OverseerScriptBuilder {
	b.script += "await page.addScriptTag({url: '" + url + "'});\n"
	return b
}

// Evaluate appends an evaluation block. Make sure `functionBody` is a valid JS function or string.
func (b *OverseerScriptBuilder) Evaluate(functionBody string) *OverseerScriptBuilder {
	b.script += "await page.evaluate(" + functionBody + ");\n"
	return b
}

// WaitForNavigation waits for a page load or redirect to finish.
func (b *OverseerScriptBuilder) WaitForNavigation() *OverseerScriptBuilder {
	b.script += "await page.waitForNavigation();\n"
	return b
}

// WaitForNavigationEvent waits for the page to reach a specific load event.
// Common values: "load", "domcontentloaded", "networkidle0", "networkidle2".
func (b *OverseerScriptBuilder) WaitForNavigationEvent(event string) *OverseerScriptBuilder {
	b.script += "await page.waitForNavigation({waitUntil: '" + event + "'});\n"
	return b
}

// WaitForNetworkIdle waits until there are no more than idleConnections open
// network connections for at least idleMs milliseconds.
// Pass idleConnections=0 for networkidle0 (fully idle), or 2 for networkidle2.
func (b *OverseerScriptBuilder) WaitForNetworkIdle(idleConnections, idleMs int) *OverseerScriptBuilder {
	waitUntil := "networkidle0"
	if idleConnections > 0 {
		waitUntil = "networkidle2"
	}
	b.script += "await page.waitForNavigation({waitUntil: '" + waitUntil + "', timeout: " + strconv.Itoa(idleMs*2+5000) + "});\n"
	return b
}

// WaitForSelector waits for an element to appear in the DOM.
func (b *OverseerScriptBuilder) WaitForSelector(selector string) *OverseerScriptBuilder {
	b.script += "await page.waitForSelector('" + selector + "');\n"
	return b
}

// Click simulates a mouse click on an element.
func (b *OverseerScriptBuilder) Click(selector string) *OverseerScriptBuilder {
	b.script += "await page.click('" + selector + "');\n"
	return b
}

// ClickAndWaitForNavigation clicks an element and waits for the resulting page
// navigation to complete before continuing. This prevents race conditions when
// clicking links or form submit buttons that trigger a full navigation.
func (b *OverseerScriptBuilder) ClickAndWaitForNavigation(selector string) *OverseerScriptBuilder {
	b.script += "await Promise.all([\n  page.waitForNavigation(),\n  page.click('" + selector + "')\n]);\n"
	return b
}

// Type simulates typing into an input field.
func (b *OverseerScriptBuilder) Type(selector, text string, delayMs int) *OverseerScriptBuilder {
	b.script += "await page.type('" + selector + "', '" + text + "',{delay:" + strconv.Itoa(delayMs) + "});\n"
	return b
}

// Raw adds raw javascript code to the overseer script.
func (b *OverseerScriptBuilder) Raw(code string) *OverseerScriptBuilder {
	b.script += code + "\n"
	return b
}

// Goto navigates to a URL and waits for the default load event.
func (b *OverseerScriptBuilder) Goto(url string) *OverseerScriptBuilder {
	b.script += "await page.goto('" + url + "');\n"
	return b
}

// GotoWithWaitUntil navigates to a URL and waits for a specific load event.
// Common values: "load", "domcontentloaded", "networkidle0", "networkidle2".
// Prefer this over Goto + WaitForNavigationEvent for SPAs that fire no traditional load events.
func (b *OverseerScriptBuilder) GotoWithWaitUntil(url, waitUntil string) *OverseerScriptBuilder {
	b.script += "await page.goto('" + url + "', {waitUntil: '" + waitUntil + "'});\n"
	return b
}

// KeyboardPress presses a specific key (e.g., 'Backspace', 'Enter') a certain number of times.
func (b *OverseerScriptBuilder) KeyboardPress(key string, times int) *OverseerScriptBuilder {
	if times <= 1 {
		b.script += "await page.keyboard.press('" + key + "');\n"
	} else {
		b.script += "await page.keyboard.press('" + key + "', {times: " + strconv.Itoa(times) + "});\n"
	}
	return b
}

// WaitForDelay pauses script execution for a specified number of milliseconds.
func (b *OverseerScriptBuilder) WaitForDelay(ms int) *OverseerScriptBuilder {
	b.script += "await page.waitForDelay(" + strconv.Itoa(ms) + ");\n"
	return b
}

// RenderContent tells PhantomJS to capture the HTML content of the page immediately.
func (b *OverseerScriptBuilder) RenderContent() *OverseerScriptBuilder {
	b.script += "page.render.content();\n"
	return b
}

// RenderScreenshot tells PhantomJS to capture a screenshot immediately. Wait triggers synchronous render.
func (b *OverseerScriptBuilder) RenderScreenshot(wait bool) *OverseerScriptBuilder {
	if wait {
		b.script += "await page.render.screenshot();\n"
	} else {
		b.script += "page.render.screenshot();\n"
	}
	return b
}

// ManualWait informs the page renderer that the script requires manual management, disabling automatic completion.
func (b *OverseerScriptBuilder) ManualWait() *OverseerScriptBuilder {
	b.script += "page.manualWait();\n"
	return b
}

// Done signals manual termination to the renderer. Must be paired with ManualWait.
func (b *OverseerScriptBuilder) Done() *OverseerScriptBuilder {
	b.script += "page.done();\n"
	return b
}

// Hover simulates resting the mouse over an element.
func (b *OverseerScriptBuilder) Hover(selector string) *OverseerScriptBuilder {
	b.script += "await page.hover('" + selector + "');\n"
	return b
}

// Focus focuses on an element.
func (b *OverseerScriptBuilder) Focus(selector string) *OverseerScriptBuilder {
	b.script += "await page.focus('" + selector + "');\n"
	return b
}

// Select selects options in a dropdown.
func (b *OverseerScriptBuilder) Select(selector string, values ...string) *OverseerScriptBuilder {
	valStr := ""
	for i, v := range values {
		if i > 0 {
			valStr += ", "
		}
		valStr += "'" + v + "'"
	}
	b.script += "await page.select('" + selector + "', " + valStr + ");\n"
	return b
}

// Reload refreshes the current page.
func (b *OverseerScriptBuilder) Reload() *OverseerScriptBuilder {
	b.script += "await page.reload();\n"
	return b
}

// ClearInput is a convenience method that manually clears a text field by evaluating Javascript.
func (b *OverseerScriptBuilder) ClearInput(selector string) *OverseerScriptBuilder {
	b.script += "await page.evaluate((sel) => { document.querySelector(sel).value = ''; }, '" + selector + "');\n"
	return b
}

// ScrollBy scrolls the page by a specific X and Y pixel offset.
func (b *OverseerScriptBuilder) ScrollBy(x, y int) *OverseerScriptBuilder {
	b.script += "await page.evaluate((x, y) => { window.scrollBy(x, y); }, " + strconv.Itoa(x) + ", " + strconv.Itoa(y) + ");\n"
	return b
}

// ScrollToBottom scrolls the entire page to the absolute bottom perfectly matching document limits. Ideal for infinite scrolling loaders.
func (b *OverseerScriptBuilder) ScrollToBottom() *OverseerScriptBuilder {
	b.script += "await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));\n"
	return b
}

// AddStyleTag injects custom CSS into the page.
func (b *OverseerScriptBuilder) AddStyleTag(cssContent string) *OverseerScriptBuilder {
	b.script += "await page.addStyleTag({content: `" + cssContent + "`});\n"
	return b
}

// SetViewport dynamically overrides the browser viewport mid-script.
func (b *OverseerScriptBuilder) SetViewport(width, height int) *OverseerScriptBuilder {
	b.script += "await page.setViewport({width: " + strconv.Itoa(width) + ", height: " + strconv.Itoa(height) + "});\n"
	return b
}

// SetUserAgent dynamically overrides the browser user agent natively mid-script.
func (b *OverseerScriptBuilder) SetUserAgent(userAgent string) *OverseerScriptBuilder {
	b.script += "await page.setUserAgent('" + userAgent + "');\n"
	return b
}

// SetExtraHTTPHeaders dynamically injects new global headers into the underlying browser mid-script.
// Input map is converted natively to a JSON object payload.
func (b *OverseerScriptBuilder) SetExtraHTTPHeaders(headers map[string]string) *OverseerScriptBuilder {
	b.script += "await page.setExtraHTTPHeaders({"
	first := true
	for k, v := range headers {
		if !first {
			b.script += ", "
		}
		b.script += "'" + k + "': '" + v + "'"
		first = false
	}
	b.script += "});\n"
	return b
}

// WaitForFunction pauses execution until the provided Javascript function returns truthy.
func (b *OverseerScriptBuilder) WaitForFunction(jsFunc string) *OverseerScriptBuilder {
	b.script += "await page.waitForFunction(" + jsFunc + ");\n"
	return b
}

// SetCookie adds a cookie directly into the browser context.
func (b *OverseerScriptBuilder) SetCookie(name, value, domain string) *OverseerScriptBuilder {
	b.script += "await page.setCookie({name: '" + name + "', value: '" + value + "', domain: '" + domain + "'});\n"
	return b
}

// DeleteCookie removes a specific cookie from the browser context.
func (b *OverseerScriptBuilder) DeleteCookie(name, url string) *OverseerScriptBuilder {
	b.script += "await page.deleteCookie({name: '" + name + "', url: '" + url + "'});\n"
	return b
}

// MouseMove simulates moving the mouse cursor to a specific absolute X,Y coordinate.
func (b *OverseerScriptBuilder) MouseMove(x, y int) *OverseerScriptBuilder {
	b.script += "await page.mouse.move(" + strconv.Itoa(x) + ", " + strconv.Itoa(y) + ");\n"
	return b
}

// MouseClickPosition simulates a native operating system level mouse click on a specific absolute X,Y coordinate rather than relying on DOM targeting.
func (b *OverseerScriptBuilder) MouseClickPosition(x, y int) *OverseerScriptBuilder {
	b.script += "await page.mouse.click(" + strconv.Itoa(x) + ", " + strconv.Itoa(y) + ");\n"
	return b
}

// WaitForXPath explicitly waits for a specific XPath block to render into the DOM.
func (b *OverseerScriptBuilder) WaitForXPath(xpath string) *OverseerScriptBuilder {
	b.script += "await page.waitForXPath(\"" + xpath + "\");\n"
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
	b.script += "await page.evaluateOnNewDocument(" + stealth.JS + ");\n"
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
	b.script += fmt.Sprintf("await page.setUserAgent(%q);\n", p.UserAgent)
	if len(p.Headers) > 0 {
		raw, _ := json.Marshal(p.Headers)
		b.script += "await page.setExtraHTTPHeaders(" + string(raw) + ");\n"
	}
	return b
}

// ApplyViewport applies a fully configured Viewport struct to the page — supporting
// DeviceScaleFactor, IsMobile, HasTouch, and IsLandscape flags that SetViewport(w,h)
// doesn't expose. Use a named preset from ext/viewport by passing its Viewport field:
//
//	builder.ApplyViewport(viewport.MobilePortrait.Viewport)
func (b *OverseerScriptBuilder) ApplyViewport(v Viewport) *OverseerScriptBuilder {
	b.script += fmt.Sprintf(
		"await page.setViewport({width:%d,height:%d,deviceScaleFactor:%g,isMobile:%t,hasTouch:%t,isLandscape:%t});\n",
		v.Width, v.Height, v.DeviceScaleFactor, v.IsMobile, v.HasTouch, v.IsLandscape,
	)
	return b
}

// Build returns the finalized script.
func (b *OverseerScriptBuilder) Build() string {
	return b.script
}
