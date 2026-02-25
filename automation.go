package phantomjscloud

import (
	"strconv"
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

// WaitForSelector waits for an element to appear in the DOM.
func (b *OverseerScriptBuilder) WaitForSelector(selector string) *OverseerScriptBuilder {
	b.script += "await page.waitForSelector('" + selector + "');\n"
	return b
}

// Click simulates a mouse click on an element.
func (b *OverseerScriptBuilder) Click(selector string) *OverseerScriptBuilder {
	b.script += "page.click('" + selector + "');\n"
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

// Goto navigates to a URL.
func (b *OverseerScriptBuilder) Goto(url string) *OverseerScriptBuilder {
	b.script += "await page.goto('" + url + "');\n"
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

// Build returns the finalized script.
func (b *OverseerScriptBuilder) Build() string {
	return b.script
}
