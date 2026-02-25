package stealth_test

import (
	"strings"
	"testing"

	"github.com/amafjarkasi/go-phantomjs/ext/stealth"
)

// skipIfNoPayload skips the test when evasions.js has not been generated yet.
// The file is deliberately excluded from version control (it is built by running
// `node scripts/gen_stealth.js`) so CI will not have it unless that step runs.
func skipIfNoPayload(t *testing.T) {
	t.Helper()
	if len(stealth.JS) == 0 {
		t.Skip("stealth payload not present — run: node scripts/gen_stealth.js")
	}
}

func TestJS_NonEmpty(t *testing.T) {
	skipIfNoPayload(t)
}

func TestJS_ContainsEvasionMarkers(t *testing.T) {
	skipIfNoPayload(t)
	markers := []string{
		"navigator.webdriver",
		"WebGLRenderingContext",
		"chrome",
		"navigator.plugins",
		"navigator.languages",
	}
	for _, marker := range markers {
		if !strings.Contains(stealth.JS, marker) {
			t.Errorf("stealth.JS missing expected evasion marker %q", marker)
		}
	}
}

func TestJS_IsFunctionExpression(t *testing.T) {
	skipIfNoPayload(t)
	js := strings.TrimSpace(stealth.JS)
	// The payload must be a function expression (not bare statements) so that
	// page.evaluateOnNewDocument(stealth.JS) is valid — it expects a function arg.
	if !strings.Contains(js, "function") {
		t.Error("stealth.JS does not appear to contain a function expression")
	}
}
