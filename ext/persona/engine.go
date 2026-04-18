package persona

import (
	"net/url"
	"strings"
	"sync"

	phantomjscloud "github.com/amafjarkasi/go-phantomjs"
	"github.com/amafjarkasi/go-phantomjs/ext/useragents"
	"github.com/amafjarkasi/go-phantomjs/ext/viewport"
)

// Config defines how a persona mutates a page request.
type Config struct {
	Proxy    interface{}
	Profile  useragents.Profile
	Viewport viewport.Preset
	Blockers []phantomjscloud.ResourceModifier
}

// URLPersonaProvider applies a persona for a request URL and attempt number.
type URLPersonaProvider interface {
	ApplyForURLAttempt(req *phantomjscloud.PageRequest, attempt int) string
}

// Engine routes personas by host with deterministic fallback by attempt.
type Engine struct {
	defaultPersonas []string
	byHost          map[string][]string
	personas        map[string]Config
	rrCounter       map[string]uint64
	mu              sync.Mutex
}

// NewEngine creates an empty persona engine.
func NewEngine() *Engine {
	return &Engine{
		byHost:    make(map[string][]string),
		personas:  make(map[string]Config),
		rrCounter: make(map[string]uint64),
	}
}

// Define registers or replaces a named persona.
func (e *Engine) Define(name string, cfg Config) *Engine {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.personas[name] = cfg
	return e
}

// RouteHost sets persona pool for a host.
func (e *Engine) RouteHost(host string, personas ...string) *Engine {
	host = normalizeHost(host)
	e.mu.Lock()
	defer e.mu.Unlock()
	e.byHost[host] = append([]string(nil), personas...)
	return e
}

// SetDefault sets fallback personas for unmatched hosts.
func (e *Engine) SetDefault(personas ...string) *Engine {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.defaultPersonas = append([]string(nil), personas...)
	return e
}

// ApplyForURLAttempt picks and applies a persona based on request URL and attempt index.
func (e *Engine) ApplyForURLAttempt(req *phantomjscloud.PageRequest, attempt int) string {
	if req == nil {
		return ""
	}
	if attempt < 0 {
		attempt = 0
	}
	host := extractHost(req.URL)

	e.mu.Lock()
	defer e.mu.Unlock()

	pool, key := e.poolForHost(host)
	if len(pool) == 0 {
		return ""
	}

	var idx int
	if attempt == 0 {
		idx = int(e.rrCounter[key] % uint64(len(pool)))
		e.rrCounter[key]++
	} else {
		idx = attempt % len(pool)
	}

	name := pool[idx]
	cfg, ok := e.personas[name]
	if !ok {
		return ""
	}
	applyConfig(req, cfg)
	return name
}

func (e *Engine) poolForHost(host string) ([]string, string) {
	if host != "" {
		if p, ok := e.byHost[host]; ok {
			return p, host
		}
		trimmed := strings.TrimPrefix(host, "www.")
		if p, ok := e.byHost[trimmed]; ok {
			return p, trimmed
		}
	}
	return e.defaultPersonas, "__default__"
}

func applyConfig(req *phantomjscloud.PageRequest, cfg Config) {
	if cfg.Proxy != nil {
		req.Proxy = cfg.Proxy
	}
	if cfg.Profile.UserAgent != "" {
		req.RequestSettings.UserAgent = cfg.Profile.UserAgent
	}
	if len(cfg.Profile.Headers) > 0 {
		req.RequestSettings.CustomHeaders = copyMap(cfg.Profile.Headers)
	}
	if cfg.Viewport.Viewport.Width > 0 && cfg.Viewport.Viewport.Height > 0 {
		rs := cfg.Viewport.AsRenderSettings()
		req.RenderSettings.Viewport = rs.Viewport
		req.RenderSettings.ClipRectangle = rs.ClipRectangle
		req.RenderSettings.ZoomFactor = rs.ZoomFactor
	}
	if len(cfg.Blockers) > 0 {
		req.RequestSettings.ResourceModifier = append([]phantomjscloud.ResourceModifier(nil), cfg.Blockers...)
	}
}

func copyMap(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
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

