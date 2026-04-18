package session

import (
	"testing"
	"time"

	phantomjscloud "github.com/amafjarkasi/go-phantomjs"
)

func TestStore_CookiesForURL_FiltersByHostSchemeAndExpiry(t *testing.T) {
	now := time.Date(2026, 4, 18, 12, 0, 0, 0, time.UTC)
	s := NewStore()
	s.now = func() time.Time { return now }

	s.Upsert([]phantomjscloud.Cookie{
		{Name: "sid", Value: "a", Domain: ".example.com", Expires: float64(now.Add(10 * time.Minute).Unix())},
		{Name: "secure", Value: "b", Domain: ".example.com", Secure: true, Expires: float64(now.Add(10 * time.Minute).Unix())},
		{Name: "expired", Value: "x", Domain: ".example.com", Expires: float64(now.Add(-10 * time.Minute).Unix())},
		{Name: "other", Value: "y", Domain: ".other.com", Expires: float64(now.Add(10 * time.Minute).Unix())},
	})

	httpsCookies := s.CookiesForURL("https://shop.example.com/items")
	if len(httpsCookies) != 2 {
		t.Fatalf("expected 2 cookies over https, got %d", len(httpsCookies))
	}

	httpCookies := s.CookiesForURL("http://shop.example.com/items")
	if len(httpCookies) != 1 {
		t.Fatalf("expected 1 cookie over http (secure excluded), got %d", len(httpCookies))
	}
	if httpCookies[0].Name != "sid" {
		t.Fatalf("expected sid cookie, got %q", httpCookies[0].Name)
	}
}

func TestStore_CaptureFromResponse(t *testing.T) {
	s := NewStore()
	resp := &phantomjscloud.UserResponseWithMeta{
		UserResponse: phantomjscloud.UserResponse{
			PageResponses: []phantomjscloud.PageResponse{
				{
					Cookies: []phantomjscloud.Cookie{
						{Name: "sess", Value: "123", Domain: ".example.com"},
					},
				},
			},
		},
	}

	s.CaptureFromResponse(resp)
	got := s.CookiesForURL("https://www.example.com")
	if len(got) != 1 || got[0].Name != "sess" {
		t.Fatalf("expected captured sess cookie, got %#v", got)
	}
}

