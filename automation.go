package phantomjscloud

import (
	"strconv"
)

// Proxy Types
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

// Build returns the finalized script.
func (b *OverseerScriptBuilder) Build() string {
	return b.script
}
