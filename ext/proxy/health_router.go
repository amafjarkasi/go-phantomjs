package proxy

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// HealthRouter routes proxies by host and tracks per-host proxy health.
// Health scores are domain-scoped and used to bias selection toward healthier proxies.
type HealthRouter struct {
	defaultProxies []interface{}
	byHost         map[string][]interface{}
	rrCounter      map[string]uint64
	healthByHost   map[string]map[string]int
	mu             sync.Mutex
}

// ProxyHealth is a point-in-time proxy health snapshot entry.
type ProxyHealth struct {
	Proxy interface{}
	Score int
}

// NewHealthRouter creates a new health-aware host router.
func NewHealthRouter(defaultProxies ...interface{}) *HealthRouter {
	return &HealthRouter{
		defaultProxies: append([]interface{}(nil), defaultProxies...),
		byHost:         make(map[string][]interface{}),
		rrCounter:      make(map[string]uint64),
		healthByHost:   make(map[string]map[string]int),
	}
}

// SetDefault replaces default proxies used when no host route matches.
func (r *HealthRouter) SetDefault(proxies ...interface{}) *HealthRouter {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.defaultProxies = append([]interface{}(nil), proxies...)
	return r
}

// RouteHost sets the proxy pool for a hostname.
func (r *HealthRouter) RouteHost(host string, proxies ...interface{}) *HealthRouter {
	host = normalizeHost(host)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byHost[host] = append([]interface{}(nil), proxies...)
	return r
}

// GetProxyForURL selects a healthy proxy for the URL.
func (r *HealthRouter) GetProxyForURL(rawURL string) interface{} {
	return r.GetProxyForURLAttempt(rawURL, 0)
}

// GetProxyForURLAttempt selects a proxy for URL and attempt index.
// attempt=0 chooses among the healthiest proxies and round-robins them.
// attempt>0 traverses a deterministic health-ranked fallback order.
func (r *HealthRouter) GetProxyForURLAttempt(rawURL string, attempt int) interface{} {
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

	scores := r.healthForHost(key)
	if attempt == 0 {
		minScore := scoreForProxy(scores, pool[0])
		best := make([]int, 0, len(pool))
		for i := range pool {
			s := scoreForProxy(scores, pool[i])
			if s < minScore {
				minScore = s
				best = best[:0]
				best = append(best, i)
				continue
			}
			if s == minScore {
				best = append(best, i)
			}
		}
		idx := int(r.rrCounter[key] % uint64(len(best)))
		r.rrCounter[key]++
		return pool[best[idx]]
	}

	order := rankPoolByHealth(pool, scores)
	idx := attempt % len(order)
	return pool[order[idx]]
}

// ReportFailure increases health penalty for a host+proxy pair.
func (r *HealthRouter) ReportFailure(rawURL string, proxy interface{}) {
	host := extractHost(rawURL)

	r.mu.Lock()
	defer r.mu.Unlock()

	_, key := r.poolForHost(host)
	m := r.healthForHost(key)
	pk := proxyKey(proxy)
	m[pk] = m[pk] + 1
}

// ReportSuccess reduces health penalty for a host+proxy pair.
func (r *HealthRouter) ReportSuccess(rawURL string, proxy interface{}) {
	host := extractHost(rawURL)

	r.mu.Lock()
	defer r.mu.Unlock()

	_, key := r.poolForHost(host)
	m := r.healthForHost(key)
	pk := proxyKey(proxy)
	if m[pk] <= 1 {
		delete(m, pk)
		return
	}
	m[pk] = m[pk] - 1
}

// HealthForURL returns current proxy health for the URL's pool,
// sorted from healthiest (lowest score) to least healthy.
func (r *HealthRouter) HealthForURL(rawURL string) []ProxyHealth {
	host := extractHost(rawURL)

	r.mu.Lock()
	defer r.mu.Unlock()

	pool, key := r.poolForHost(host)
	if len(pool) == 0 {
		return nil
	}
	scores := r.healthForHost(key)
	order := rankPoolByHealth(pool, scores)

	out := make([]ProxyHealth, 0, len(pool))
	for _, idx := range order {
		p := pool[idx]
		out = append(out, ProxyHealth{
			Proxy: p,
			Score: scoreForProxy(scores, p),
		})
	}
	return out
}

func (r *HealthRouter) poolForHost(host string) ([]interface{}, string) {
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

func (r *HealthRouter) healthForHost(hostKey string) map[string]int {
	m, ok := r.healthByHost[hostKey]
	if ok {
		return m
	}
	m = make(map[string]int)
	r.healthByHost[hostKey] = m
	return m
}

func rankPoolByHealth(pool []interface{}, scores map[string]int) []int {
	type item struct {
		idx   int
		score int
	}
	items := make([]item, 0, len(pool))
	for i := range pool {
		items = append(items, item{
			idx:   i,
			score: scoreForProxy(scores, pool[i]),
		})
	}
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].score < items[j].score
	})

	out := make([]int, 0, len(items))
	for i := range items {
		out = append(out, items[i].idx)
	}
	return out
}

func scoreForProxy(scores map[string]int, proxy interface{}) int {
	return scores[proxyKey(proxy)]
}

func proxyKey(proxy interface{}) string {
	return fmt.Sprintf("%#v", proxy)
}
