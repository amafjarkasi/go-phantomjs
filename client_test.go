package phantomjscloud

import (
	"net/http"
	"testing"
)

func TestClient_Do(t *testing.T) {
	// ... test implementation
}

func TestClient_DoPage(t *testing.T) {
	// TODO: implement a test for DoPage using a mock server
	// overriding the base endpoint URL.
}

func TestParseMetadata(t *testing.T) {
	headers := http.Header{}
	headers.Set("pjsc-billing-credit-cost", "2.25")
	headers.Set("pjsc-content-status-code", "201")
	headers.Set("pjsc-content-done-when", "load")

	meta := parseMetadata(headers)

	if meta.BillingCreditCost != 2.25 {
		t.Errorf("Expected 2.25, got %f", meta.BillingCreditCost)
	}
	if meta.ContentStatusCode != 201 {
		t.Errorf("Expected 201, got %d", meta.ContentStatusCode)
	}
	if meta.ContentDoneWhen != "load" {
		t.Errorf("Expected load, got %s", meta.ContentDoneWhen)
	}
}

func TestOverseerScriptBuilder(t *testing.T) {
	b := NewOverseerScriptBuilder()
	script := b.AddScriptTag("http://example.com/script.js").
		Evaluate("() => { return 'done'; }").
		WaitForSelector("body").
		Type("input#name", "test user", 100).
		Build()

	expected := "await page.addScriptTag({url: 'http://example.com/script.js'});\n" +
		"await page.evaluate(() => { return 'done'; });\n" +
		"await page.waitForSelector('body');\n" +
		"await page.type('input#name', 'test user',{delay:100});\n"

	if script != expected {
		t.Errorf("Builder output mismatch.\nExpected:\n%s\nGot:\n%s", expected, script)
	}
}
