package proxy

import (
	"sync"
	"sync/atomic"
)

// ProxyProvider is an interface for dynamic proxy selection.
type ProxyProvider interface {
	GetProxy() interface{}
}

// StaticProxyProvider returns a fixed proxy.
type StaticProxyProvider struct {
	Proxy interface{}
}

func (p *StaticProxyProvider) GetProxy() interface{} {
	return p.Proxy
}

// RotatingProxyProvider selects proxies from a pool using a Round-Robin strategy.
type RotatingProxyProvider struct {
	proxies []interface{}
	counter uint64
	mu      sync.RWMutex
}

// NewRotatingProxyProvider creates a new provider with the given proxy list.
func NewRotatingProxyProvider(proxies ...interface{}) *RotatingProxyProvider {
	return &RotatingProxyProvider{
		proxies: proxies,
	}
}

func (p *RotatingProxyProvider) GetProxy() interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.proxies) == 0 {
		return nil
	}

	idx := atomic.AddUint64(&p.counter, 1) % uint64(len(p.proxies))
	return p.proxies[idx]
}

// AddProxy adds a new proxy to the pool.
func (p *RotatingProxyProvider) AddProxy(proxy interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.proxies = append(p.proxies, proxy)
}
