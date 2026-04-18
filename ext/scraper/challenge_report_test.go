package scraper

import (
	"testing"

	"github.com/amafjarkasi/go-phantomjs/ext/proxy"
)

func TestBuildChallengeDebugReport_ComputesHealthDeltaPerAttempt(t *testing.T) {
	attempts := []ChallengeAttempt{
		{
			Proxy:   "anon-us",
			Blocked: true,
			Health: []proxy.ProxyHealth{
				{Proxy: "anon-us", Score: 0},
				{Proxy: "anon-ca", Score: 0},
			},
		},
		{
			Proxy:   "anon-ca",
			Blocked: false,
			Health: []proxy.ProxyHealth{
				{Proxy: "anon-us", Score: 1},
				{Proxy: "anon-ca", Score: 0},
			},
		},
	}

	report := BuildChallengeDebugReport(attempts)
	if len(report.Attempts) != 2 {
		t.Fatalf("expected 2 attempts, got %d", len(report.Attempts))
	}
	if report.Attempts[0].Attempt != 1 || report.Attempts[1].Attempt != 2 {
		t.Fatalf("unexpected attempt numbers: %#v", report.Attempts)
	}
	if !report.Attempts[0].Blocked || report.Attempts[1].Blocked {
		t.Fatalf("blocked flags mismatch: %#v", report.Attempts)
	}
	if report.Attempts[0].SelectedProxy != "anon-us" {
		t.Fatalf("expected attempt1 selected proxy anon-us, got %q", report.Attempts[0].SelectedProxy)
	}

	// Health delta uses previous attempt snapshot as baseline.
	if got := report.Attempts[1].HealthDelta["anon-us"]; got != 1 {
		t.Fatalf("expected anon-us delta +1 on attempt2, got %d", got)
	}
	if got := report.Attempts[1].HealthDelta["anon-ca"]; got != 0 {
		t.Fatalf("expected anon-ca delta 0 on attempt2, got %d", got)
	}
}

func TestBuildChallengeDebugReport_HandlesEmpty(t *testing.T) {
	report := BuildChallengeDebugReport(nil)
	if len(report.Attempts) != 0 {
		t.Fatalf("expected empty report, got %#v", report)
	}
}

