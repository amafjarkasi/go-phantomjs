package proxy

import "testing"

func TestHealthRouter_PrefersHealthyProxyAfterFailure(t *testing.T) {
	r := NewHealthRouter("default").
		RouteHost("example.com", "p1", "p2")

	first := r.GetProxyForURL("https://example.com")
	if first != "p1" {
		t.Fatalf("expected initial proxy p1, got %v", first)
	}

	r.ReportFailure("https://example.com", "p1")

	second := r.GetProxyForURL("https://example.com")
	if second != "p2" {
		t.Fatalf("expected healthy proxy p2 after p1 failure, got %v", second)
	}
}

func TestHealthRouter_HealthIsDomainScoped(t *testing.T) {
	r := NewHealthRouter("default").
		RouteHost("example.com", "p1", "p2").
		RouteHost("other.com", "p1", "p2")

	r.ReportFailure("https://example.com", "p1")

	examplePick := r.GetProxyForURL("https://example.com")
	if examplePick != "p2" {
		t.Fatalf("expected example.com to avoid p1, got %v", examplePick)
	}

	otherPick := r.GetProxyForURL("https://other.com")
	if otherPick != "p1" {
		t.Fatalf("expected other.com to keep p1 healthy, got %v", otherPick)
	}
}

func TestHealthRouter_AttemptFallbackUsesHealthAwareOrder(t *testing.T) {
	r := NewHealthRouter("default").
		RouteHost("example.com", "p1", "p2", "p3")

	r.ReportFailure("https://example.com", "p1")
	r.ReportFailure("https://example.com", "p1")
	r.ReportFailure("https://example.com", "p2")

	if got := r.GetProxyForURLAttempt("https://example.com", 0); got != "p3" {
		t.Fatalf("expected healthiest proxy p3 for attempt 0, got %v", got)
	}
	if got := r.GetProxyForURLAttempt("https://example.com", 1); got != "p2" {
		t.Fatalf("expected second healthiest proxy p2 for attempt 1, got %v", got)
	}
	if got := r.GetProxyForURLAttempt("https://example.com", 2); got != "p1" {
		t.Fatalf("expected least healthy proxy p1 for attempt 2, got %v", got)
	}
}

func TestHealthRouter_HealthForURL_ReturnsScoresInHealthOrder(t *testing.T) {
	r := NewHealthRouter("d1").
		RouteHost("example.com", "p1", "p2", "p3")

	r.ReportFailure("https://example.com", "p1")
	r.ReportFailure("https://example.com", "p1")
	r.ReportFailure("https://example.com", "p2")

	health := r.HealthForURL("https://example.com")
	if len(health) != 3 {
		t.Fatalf("expected 3 health entries, got %d", len(health))
	}
	if health[0].Proxy != "p3" || health[0].Score != 0 {
		t.Fatalf("expected healthiest p3 score 0, got %#v", health[0])
	}
	if health[1].Proxy != "p2" || health[1].Score != 1 {
		t.Fatalf("expected next p2 score 1, got %#v", health[1])
	}
	if health[2].Proxy != "p1" || health[2].Score != 2 {
		t.Fatalf("expected worst p1 score 2, got %#v", health[2])
	}
}

func TestHealthRouter_ReportSuccess_DecreasesPenalty(t *testing.T) {
	r := NewHealthRouter().RouteHost("example.com", "p1", "p2")
	r.ReportFailure("https://example.com", "p1")
	r.ReportFailure("https://example.com", "p1")
	r.ReportSuccess("https://example.com", "p1")

	health := r.HealthForURL("https://example.com")
	if len(health) != 2 {
		t.Fatalf("expected 2 health entries, got %d", len(health))
	}

	var p1Score int
	for _, h := range health {
		if h.Proxy == "p1" {
			p1Score = h.Score
		}
	}
	if p1Score != 1 {
		t.Fatalf("expected p1 score reduced to 1, got %d", p1Score)
	}
}
