package scraper

import (
	"fmt"

	"github.com/amafjarkasi/go-phantomjs/ext/proxy"
)

// ChallengeDebugReport is a structured summary of orchestration attempts.
type ChallengeDebugReport struct {
	Attempts []ChallengeDebugAttempt `json:"attempts"`
}

// ChallengeDebugAttempt is a normalized attempt record for logs/telemetry.
type ChallengeDebugAttempt struct {
	Attempt       int            `json:"attempt"`
	SelectedProxy string         `json:"selectedProxy,omitempty"`
	Blocked       bool           `json:"blocked"`
	HasError      bool           `json:"hasError"`
	Health        map[string]int `json:"health,omitempty"`
	HealthDelta   map[string]int `json:"healthDelta,omitempty"`
}

// BuildChallengeDebugReport converts orchestration attempts into stable debug data.
func BuildChallengeDebugReport(attempts []ChallengeAttempt) ChallengeDebugReport {
	report := ChallengeDebugReport{
		Attempts: make([]ChallengeDebugAttempt, 0, len(attempts)),
	}

	var prev map[string]int
	for i := range attempts {
		cur := snapshotHealth(attempts[i].Health)

		item := ChallengeDebugAttempt{
			Attempt:       i + 1,
			SelectedProxy: fmt.Sprintf("%v", attempts[i].Proxy),
			Blocked:       attempts[i].Blocked,
			HasError:      attempts[i].Err != nil,
			Health:        cur,
			HealthDelta:   diffHealth(prev, cur),
		}
		report.Attempts = append(report.Attempts, item)
		prev = cur
	}

	return report
}

func snapshotHealth(in []proxy.ProxyHealth) map[string]int {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]int, len(in))
	for i := range in {
		out[fmt.Sprintf("%v", in[i].Proxy)] = in[i].Score
	}
	return out
}

func diffHealth(prev, cur map[string]int) map[string]int {
	if len(cur) == 0 {
		return nil
	}
	out := make(map[string]int, len(cur))
	for k, v := range cur {
		out[k] = v - prev[k]
	}
	return out
}

