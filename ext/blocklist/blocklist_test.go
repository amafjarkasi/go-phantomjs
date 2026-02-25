package blocklist_test

import (
	"testing"

	phantomjscloud "github.com/jbdt/go-phantomjs"
	"github.com/jbdt/go-phantomjs/ext/blocklist"
)

func assertAllBlocked(t *testing.T, name string, rules []phantomjscloud.ResourceModifier) {
	t.Helper()
	if len(rules) == 0 {
		t.Errorf("%s: expected non-empty slice, got 0 rules", name)
		return
	}
	for i, r := range rules {
		if !r.IsBlacklisted {
			t.Errorf("%s[%d]: IsBlacklisted is false for regex %q", name, i, r.Regex)
		}
		if r.Regex == "" {
			t.Errorf("%s[%d]: empty Regex", name, i)
		}
	}
}

func TestAds(t *testing.T) {
	assertAllBlocked(t, "Ads", blocklist.Ads())
}

func TestTrackers(t *testing.T) {
	assertAllBlocked(t, "Trackers", blocklist.Trackers())
}

func TestFonts(t *testing.T) {
	assertAllBlocked(t, "Fonts", blocklist.Fonts())
}

func TestMedia(t *testing.T) {
	assertAllBlocked(t, "Media", blocklist.Media())
}

func TestLightweight_IsSupersetOfAdsTrackersAndFonts(t *testing.T) {
	lw := blocklist.Lightweight()
	assertAllBlocked(t, "Lightweight", lw)

	ads := len(blocklist.Ads())
	trackers := len(blocklist.Trackers())
	fonts := len(blocklist.Fonts())
	want := ads + trackers + fonts

	if got := len(lw); got != want {
		t.Errorf("Lightweight: expected %d rules (Ads+Trackers+Fonts), got %d", want, got)
	}
}

func TestFull_IsSupersetOfLightweightAndMedia(t *testing.T) {
	full := blocklist.Full()
	assertAllBlocked(t, "Full", full)

	want := len(blocklist.Ads()) + len(blocklist.Trackers()) + len(blocklist.Fonts()) + len(blocklist.Media())
	if got := len(full); got != want {
		t.Errorf("Full: expected %d rules, got %d", want, got)
	}
}
