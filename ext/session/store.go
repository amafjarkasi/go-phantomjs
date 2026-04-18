package session

import (
	"net/url"
	"strings"
	"sync"
	"time"

	phantomjscloud "github.com/amafjarkasi/go-phantomjs"
)

// Store keeps cookies across requests while enforcing host/scheme filtering.
type Store struct {
	mu      sync.Mutex
	cookies []phantomjscloud.Cookie
	now     func() time.Time
}

// NewStore creates an empty cookie store.
func NewStore() *Store {
	return &Store{
		now: time.Now,
	}
}

// Upsert adds or replaces cookies by (name, domain, path).
func (s *Store) Upsert(cookies []phantomjscloud.Cookie) {
	if len(cookies) == 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, c := range cookies {
		if c.Name == "" {
			continue
		}
		idx := -1
		for i := range s.cookies {
			if sameCookieIdentity(s.cookies[i], c) {
				idx = i
				break
			}
		}
		if idx >= 0 {
			s.cookies[idx] = c
		} else {
			s.cookies = append(s.cookies, c)
		}
	}
}

// CaptureFromResponse stores all cookies returned by page responses.
func (s *Store) CaptureFromResponse(resp *phantomjscloud.UserResponseWithMeta) {
	if resp == nil || len(resp.PageResponses) == 0 {
		return
	}
	all := make([]phantomjscloud.Cookie, 0, 8)
	for i := range resp.PageResponses {
		all = append(all, resp.PageResponses[i].Cookies...)
	}
	s.Upsert(all)
}

// CookiesForURL returns cookies safe to send for the given URL.
func (s *Store) CookiesForURL(rawURL string) []phantomjscloud.Cookie {
	target, ok := parseTarget(rawURL)
	if !ok {
		return nil
	}

	nowUnix := float64(s.now().Unix())
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]phantomjscloud.Cookie, 0, len(s.cookies))
	for _, c := range s.cookies {
		if c.Expires > 0 && c.Expires <= nowUnix {
			continue
		}
		if c.Secure && target.scheme != "https" {
			continue
		}
		if !cookieMatchesTarget(c, target) {
			continue
		}
		out = append(out, c)
	}
	return out
}

type targetURL struct {
	scheme string
	host   string
}

func parseTarget(rawURL string) (targetURL, bool) {
	u, err := url.Parse(rawURL)
	if err != nil || u.Hostname() == "" {
		return targetURL{}, false
	}
	return targetURL{
		scheme: strings.ToLower(u.Scheme),
		host:   normalizeHost(u.Hostname()),
	}, true
}

func cookieMatchesTarget(c phantomjscloud.Cookie, target targetURL) bool {
	if c.URL != "" {
		u, err := url.Parse(c.URL)
		if err == nil && u.Hostname() != "" {
			return domainMatches(target.host, normalizeHost(u.Hostname()))
		}
	}

	if c.Domain == "" {
		return true
	}
	return domainMatches(target.host, normalizeDomain(c.Domain))
}

func domainMatches(host, domain string) bool {
	if domain == "" {
		return false
	}
	if host == domain {
		return true
	}
	return strings.HasSuffix(host, "."+domain)
}

func sameCookieIdentity(a, b phantomjscloud.Cookie) bool {
	return a.Name == b.Name &&
		normalizeDomain(a.Domain) == normalizeDomain(b.Domain) &&
		a.Path == b.Path
}

func normalizeHost(host string) string {
	host = strings.TrimSpace(strings.ToLower(host))
	host = strings.TrimPrefix(host, "www.")
	return host
}

func normalizeDomain(domain string) string {
	domain = strings.TrimSpace(strings.ToLower(domain))
	domain = strings.TrimPrefix(domain, ".")
	domain = strings.TrimPrefix(domain, "www.")
	return domain
}

