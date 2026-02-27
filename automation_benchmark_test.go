package phantomjscloud

import (
	"testing"
)

func BenchmarkOverseerScriptBuilder_AddScriptTag(b *testing.B) {
	for i := 0; i < b.N; i++ {
		builder := NewOverseerScriptBuilder()
		for j := 0; j < 1000; j++ {
			builder.AddScriptTag("https://example.com/script.js")
		}
		_ = builder.Build()
	}
}

func BenchmarkOverseerScriptBuilder_GeneralUsage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		builder := NewOverseerScriptBuilder()
		for j := 0; j < 100; j++ {
			builder.Goto("https://example.com").
				WaitForSelector(".main").
				Raw("console.log('done');").
				AddScriptTag("https://example.com/script.js")
		}
		_ = builder.Build()
	}
}
