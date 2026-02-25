package phantomjscloud_test

import (
	"testing"

	phantomjscloud "github.com/amafjarkasi/go-phantomjs"
)

func TestOverseerScriptBuilder_Raw(t *testing.T) {
	sb := phantomjscloud.NewOverseerScriptBuilder()
	sb.Raw("console.log('test');")

	expected := "console.log('test');\n"
	if sb.Build() != expected {
		t.Errorf("expected %q, got %q", expected, sb.Build())
	}

	sb.Raw("alert('hello');")
	expected = "console.log('test');\nalert('hello');\n"
	if sb.Build() != expected {
		t.Errorf("expected %q, got %q", expected, sb.Build())
	}
}

func TestOverseerScriptBuilder_Chaining(t *testing.T) {
	sb := phantomjscloud.NewOverseerScriptBuilder().
		Goto("https://example.com").
		WaitForSelector(".main").
		Raw("console.log('done');")

	expected := "await page.goto('https://example.com');\nawait page.waitForSelector('.main');\nconsole.log('done');\n"
	if sb.Build() != expected {
		t.Errorf("expected %q, got %q", expected, sb.Build())
	}
}
