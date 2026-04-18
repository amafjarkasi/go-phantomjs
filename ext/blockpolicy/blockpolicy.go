package blockpolicy

import (
	"strings"

	phantomjscloud "github.com/amafjarkasi/go-phantomjs"
	"github.com/amafjarkasi/go-phantomjs/ext/blocklist"
)

// Level controls how aggressively resources are blocked.
type Level string

const (
	LevelAggressive Level = "aggressive"
	LevelBalanced   Level = "balanced"
	LevelRelaxed    Level = "relaxed"
	LevelOff        Level = "off"
)

// Rules returns blocklist rules for a policy level.
func Rules(level Level) []phantomjscloud.ResourceModifier {
	switch level {
	case LevelAggressive:
		return blocklist.Full()
	case LevelBalanced:
		return blocklist.Lightweight()
	case LevelRelaxed:
		rules := make([]phantomjscloud.ResourceModifier, 0, 48)
		rules = append(rules, blocklist.Ads()...)
		rules = append(rules, blocklist.Trackers()...)
		return rules
	case LevelOff:
		return nil
	default:
		return blocklist.Lightweight()
	}
}

// NextLevel returns the next less restrictive level.
func NextLevel(level Level) Level {
	switch level {
	case LevelAggressive:
		return LevelBalanced
	case LevelBalanced:
		return LevelRelaxed
	case LevelRelaxed:
		return LevelOff
	default:
		return LevelOff
	}
}

// Apply clones and sets resource modifiers for the requested level.
func Apply(req *phantomjscloud.PageRequest, level Level) {
	rules := Rules(level)
	if len(rules) == 0 {
		req.RequestSettings.ResourceModifier = nil
		return
	}
	req.RequestSettings.ResourceModifier = append([]phantomjscloud.ResourceModifier(nil), rules...)
}

// LooksBlocked checks whether response content looks like a challenge/block page.
func LooksBlocked(resp *phantomjscloud.UserResponseWithMeta) bool {
	if resp == nil || len(resp.PageResponses) == 0 {
		return true
	}

	code := resp.Metadata.ContentStatusCode
	if code == 0 {
		code = resp.PageResponses[0].StatusCode
	}
	if code == 403 || code == 429 || code == 503 {
		return true
	}

	content := resp.PageResponses[0].Content
	if content == "" && resp.PageResponses[0].FrameData != nil {
		content = resp.PageResponses[0].FrameData.Content
	}
	lc := strings.ToLower(content)
	signals := []string{
		"robot or human",
		"captcha",
		"continue shopping",
		"verify you are human",
		"access denied",
		"automated access",
	}
	for _, s := range signals {
		if strings.Contains(lc, s) {
			return true
		}
	}
	return false
}
