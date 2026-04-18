package proxy

import (
	"net/url"
	"strings"
	"sync"
)

// URLProxyProvider selects a proxy based on the request URL.
type URLProxyProvider interface {
	GetProxyForURL(rawURL string) interface{}
}

// URLProxyHealthReporter allows routers to learn from request outcomes.
type URLProxyHealthReporter interface {
	ReportFailure(rawURL string, proxy interface{})
	ReportSuccess(rawURL string, proxy interface{})
}

// URLProxyHealthSnapshotProvider exposes current proxy health for a URL.
type URLProxyHealthSnapshotProvider interface {
	HealthForURL(rawURL string) []ProxyHealth
}

// URLProxyFallbackProvider selects a proxy based on URL and attempt index.
// Attempt zero is the primary selection; higher attempts select fallbacks.
type URLProxyFallbackProvider interface {
	URLProxyProvider
	GetProxyForURLAttempt(rawURL string, attempt int) interface{}
}

// HostRouter routes proxies by hostname with optional default proxies.
// Matching is exact by hostname (case-insensitive). "www." variants are normalized.
type HostRouter struct {
	defaultProxies []interface{}
	byHost         map[string][]interface{}
	rrCounter      map[string]uint64
	mu             sync.Mutex
}

// NewHostRouter creates a new host router with optional default proxies.
func NewHostRouter(defaultProxies ...interface{}) *HostRouter {
	return &HostRouter{
		defaultProxies: append([]interface{}(nil), defaultProxies...),
		byHost:         make(map[string][]interface{}),
		rrCounter:      make(map[string]uint64),
	}
}

// SetDefault replaces default proxies used when no host route matches.
func (r *HostRouter) SetDefault(proxies ...interface{}) *HostRouter {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.defaultProxies = append([]interface{}(nil), proxies...)
	return r
}

// RouteHost sets the proxy pool for a hostname.
func (r *HostRouter) RouteHost(host string, proxies ...interface{}) *HostRouter {
	host = normalizeHost(host)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byHost[host] = append([]interface{}(nil), proxies...)
	return r
}

// GetProxyForURL selects a proxy for the URL using round-robin on the selected host pool.
func (r *HostRouter) GetProxyForURL(rawURL string) interface{} {
	return r.GetProxyForURLAttempt(rawURL, 0)
}

// GetProxyForURLAttempt selects a proxy for the URL and attempt.
// attempt=0 uses round-robin for load distribution.
// attempt>0 uses deterministic index selection to provide fallback ordering.
func (r *HostRouter) GetProxyForURLAttempt(rawURL string, attempt int) interface{} {
	if attempt < 0 {
		attempt = 0
	}
	host := extractHost(rawURL)

	r.mu.Lock()
	defer r.mu.Unlock()

	pool, key := r.poolForHost(host)
	if len(pool) == 0 {
		return nil
	}

	if attempt == 0 {
		idx := int(r.rrCounter[key] % uint64(len(pool)))
		r.rrCounter[key]++
		return pool[idx]
	}

	idx := attempt % len(pool)
	return pool[idx]
}

func (r *HostRouter) poolForHost(host string) ([]interface{}, string) {
	if host != "" {
		if p, ok := r.byHost[host]; ok {
			return p, host
		}
		trimmed := strings.TrimPrefix(host, "www.")
		if p, ok := r.byHost[trimmed]; ok {
			return p, trimmed
		}
	}
	return r.defaultProxies, "__default__"
}

func extractHost(rawURL string) string {
	if rawURL == "" {
		return ""
	}

	u, err := url.Parse(rawURL)
	if err == nil && u.Hostname() != "" {
		return normalizeHost(u.Hostname())
	}

	if !strings.Contains(rawURL, "://") {
		u2, err2 := url.Parse("https://" + rawURL)
		if err2 == nil && u2.Hostname() != "" {
			return normalizeHost(u2.Hostname())
		}
	}

	return normalizeHost(rawURL)
}

func normalizeHost(host string) string {
	host = strings.TrimSpace(strings.ToLower(host))
	host = strings.TrimPrefix(host, "https://")
	host = strings.TrimPrefix(host, "http://")
	host = strings.TrimPrefix(host, "www.")
	if i := strings.IndexByte(host, '/'); i >= 0 {
		host = host[:i]
	}
	if i := strings.IndexByte(host, ':'); i >= 0 {
		host = host[:i]
	}
	return host
}
