package phantomjscloud_test

import (
	"strings"
	"testing"

	phantomjscloud "github.com/amafjarkasi/go-phantomjs"
)

func TestOverseerScriptBuilder_Injections(t *testing.T) {
	malicious := "'); alert(1); //"

	tests := []struct {
		name string
		f    func(*phantomjscloud.OverseerScriptBuilder)
	}{
		{"AddScriptTag", func(b *phantomjscloud.OverseerScriptBuilder) { b.AddScriptTag(malicious) }},
		{"WaitForNavigationEvent", func(b *phantomjscloud.OverseerScriptBuilder) { b.WaitForNavigationEvent(malicious) }},
		{"WaitForSelector", func(b *phantomjscloud.OverseerScriptBuilder) { b.WaitForSelector(malicious) }},
		{"Click", func(b *phantomjscloud.OverseerScriptBuilder) { b.Click(malicious) }},
		{"ClickAndWaitForNavigation", func(b *phantomjscloud.OverseerScriptBuilder) { b.ClickAndWaitForNavigation(malicious) }},
		{"Type", func(b *phantomjscloud.OverseerScriptBuilder) { b.Type(malicious, malicious, 0) }},
		{"Goto", func(b *phantomjscloud.OverseerScriptBuilder) { b.Goto(malicious) }},
		{"GotoWithWaitUntil", func(b *phantomjscloud.OverseerScriptBuilder) { b.GotoWithWaitUntil(malicious, malicious) }},
		{"KeyboardPress", func(b *phantomjscloud.OverseerScriptBuilder) { b.KeyboardPress(malicious, 1) }},
		{"Hover", func(b *phantomjscloud.OverseerScriptBuilder) { b.Hover(malicious) }},
		{"Focus", func(b *phantomjscloud.OverseerScriptBuilder) { b.Focus(malicious) }},
		{"Select", func(b *phantomjscloud.OverseerScriptBuilder) { b.Select(malicious, malicious) }},
		{"ClearInput", func(b *phantomjscloud.OverseerScriptBuilder) { b.ClearInput(malicious) }},
		{"SetUserAgent", func(b *phantomjscloud.OverseerScriptBuilder) { b.SetUserAgent(malicious) }},
		{"SetExtraHTTPHeaders", func(b *phantomjscloud.OverseerScriptBuilder) { b.SetExtraHTTPHeaders(map[string]string{malicious: malicious}) }},
		{"SetCookie", func(b *phantomjscloud.OverseerScriptBuilder) { b.SetCookie(malicious, malicious, malicious) }},
		{"DeleteCookie", func(b *phantomjscloud.OverseerScriptBuilder) { b.DeleteCookie(malicious, malicious) }},
		{"DragAndDrop", func(b *phantomjscloud.OverseerScriptBuilder) { b.DragAndDrop(malicious, malicious) }},
		{"WaitForUrl", func(b *phantomjscloud.OverseerScriptBuilder) { b.WaitForUrl(malicious) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb := phantomjscloud.NewOverseerScriptBuilder()
			tt.f(sb)
			script := sb.Build()
			// The malicious string should be escaped, so it shouldn't appear in its raw form
			// accompanied by the script-breaking quotes.
			// Since we use json.Marshal, " becomes \", so '); alert(1); // becomes "'); alert(1); //"
			// The vulnerability was that it was wrapped in ' ', e.g. ' '); alert(1); // '
			// Now it is " \" '); alert(1); // \" " (wait, no, json.Marshal of '); alert(1); // is "'); alert(1); //")
			// The point is that if we had input: ' + alert(1) + '
			// Old: ' ' + alert(1) + ' '  => alert fires
			// New: "' + alert(1) + '"    => literal string, no alert

			// To detect if it's FIXED, we should check that the script DOES NOT contain the unescaped sequence
			// that would break out of a single-quoted string.
			// But since we now use double quotes from json.Marshal, we should check it doesn't break out of those.

			// A better test is to check for the absence of the SPECIFIC VULNERABLE PATTERN.
			// The vulnerable pattern was: 'input'
			if strings.Contains(script, "'"+malicious+"'") {
				t.Errorf("%s is vulnerable to injection (single quote)! Script: %s", tt.name, script)
			}
		})
	}

	t.Run("AddStyleTag", func(t *testing.T) {
		maliciousStyle := "`); alert(1); //"
		sb := phantomjscloud.NewOverseerScriptBuilder()
		sb.AddStyleTag(maliciousStyle)
		script := sb.Build()
		if strings.Contains(script, "`"+maliciousStyle+"`") {
			t.Errorf("AddStyleTag is vulnerable to injection (backtick)! Script: %s", script)
		}
	})

	t.Run("WaitForXPath", func(t *testing.T) {
		maliciousXPath := "\"); alert(1); //"
		sb := phantomjscloud.NewOverseerScriptBuilder()
		sb.WaitForXPath(maliciousXPath)
		script := sb.Build()
		if strings.Contains(script, "\""+maliciousXPath+"\"") {
			t.Errorf("WaitForXPath is vulnerable to injection (double quote)! Script: %s", script)
		}
	})
}
