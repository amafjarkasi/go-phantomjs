package proxy

import "testing"

func TestHostRouter_RoundRobinByHost(t *testing.T) {
	r := NewHostRouter("default").
		RouteHost("example.com", "p1", "p2")

	got1 := r.GetProxyForURL("https://www.example.com")
	got2 := r.GetProxyForURL("https://example.com/products")

	if got1 != "p1" {
		t.Fatalf("expected first round-robin proxy p1, got %v", got1)
	}
	if got2 != "p2" {
		t.Fatalf("expected second round-robin proxy p2, got %v", got2)
	}
}

func TestHostRouter_FallbackByAttempt(t *testing.T) {
	r := NewHostRouter("default").
		RouteHost("example.com", "p1", "p2", "p3")

	if got := r.GetProxyForURLAttempt("https://example.com", 1); got != "p2" {
		t.Fatalf("expected attempt 1 => p2, got %v", got)
	}
	if got := r.GetProxyForURLAttempt("https://example.com", 2); got != "p3" {
		t.Fatalf("expected attempt 2 => p3, got %v", got)
	}
	if got := r.GetProxyForURLAttempt("https://example.com", 3); got != "p1" {
		t.Fatalf("expected attempt 3 => p1, got %v", got)
	}
}

func TestHostRouter_DefaultPool(t *testing.T) {
	r := NewHostRouter("d1", "d2")

	if got := r.GetProxyForURL("https://unknown-host.test"); got != "d1" {
		t.Fatalf("expected default proxy d1, got %v", got)
	}
	if got := r.GetProxyForURL("https://unknown-host.test"); got != "d2" {
		t.Fatalf("expected default proxy d2, got %v", got)
	}
}
