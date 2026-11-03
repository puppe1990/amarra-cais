package session

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetCookie_AndTokenFromRequest(t *testing.T) {
	rr := httptest.NewRecorder()
	SetCookie(rr, "abc123", CookieOptions{Secure: false})

	res := rr.Result()
	defer func() { _ = res.Body.Close() }()

	cookies := res.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %d, want 1", len(cookies))
	}
	c := cookies[0]
	if c.Name != DefaultCookieName {
		t.Errorf("name = %q, want %q", c.Name, DefaultCookieName)
	}
	if c.Value != "abc123" {
		t.Errorf("value = %q, want abc123", c.Value)
	}
	if !c.HttpOnly {
		t.Error("cookie should be HttpOnly")
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(c)
	if got := TokenFromRequest(req); got != "abc123" {
		t.Errorf("TokenFromRequest() = %q, want abc123", got)
	}
}

func TestClearCookie(t *testing.T) {
	rr := httptest.NewRecorder()
	ClearCookie(rr, CookieOptions{})

	res := rr.Result()
	defer func() { _ = res.Body.Close() }()

	cookies := res.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %d, want 1", len(cookies))
	}
	if cookies[0].MaxAge != -1 {
		t.Errorf("MaxAge = %d, want -1", cookies[0].MaxAge)
	}
}

func TestClearCookie_SecureFlag(t *testing.T) {
	rr := httptest.NewRecorder()
	ClearCookie(rr, CookieOptions{Secure: true})

	res := rr.Result()
	defer func() { _ = res.Body.Close() }()

	for _, c := range res.Cookies() {
		if !c.Secure {
			t.Errorf("%s should preserve Secure flag", c.Name)
		}
	}
}

// #395: production session cookies use __Host- so a sibling subdomain cannot
// Set-Cookie Domain=.apps.gestaobem.com and compete with the host-only session.
func TestSetCookie_secureUsesHostPrefixedName(t *testing.T) {
	rr := httptest.NewRecorder()
	SetCookie(rr, "tok", CookieOptions{Secure: true})

	var host *http.Cookie
	for _, c := range rr.Result().Cookies() {
		if c.Name == DefaultCookieName {
			t.Error("plain cais_session must not be set when Secure")
		}
		if c.Name == "__Host-cais_session" {
			host = c
		}
	}
	if host == nil {
		t.Fatalf("missing __Host-cais_session cookie: %+v", rr.Result().Cookies())
	}
	if host.Value != "tok" {
		t.Errorf("value = %q, want tok", host.Value)
	}
	if !host.Secure || host.Path != "/" || host.Domain != "" || !host.HttpOnly {
		t.Errorf("__Host- cookie violates prefix rules: %+v", host)
	}
	if host.SameSite != http.SameSiteLaxMode {
		t.Errorf("SameSite = %v, want Lax", host.SameSite)
	}
}

func TestSetCookie_insecureKeepsPlainName(t *testing.T) {
	rr := httptest.NewRecorder()
	SetCookie(rr, "tok", CookieOptions{Secure: false})
	for _, c := range rr.Result().Cookies() {
		if c.Name == "__Host-cais_session" {
			t.Error("__Host- cookie must not be set without Secure")
		}
	}
}

func TestTokenFromRequest_prefersHostPrefixedCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: DefaultCookieName, Value: "legacy"})
	req.AddCookie(&http.Cookie{Name: "__Host-cais_session", Value: "host"})
	if got := TokenFromRequest(req); got != "host" {
		t.Errorf("TokenFromRequest = %q, want host", got)
	}
}

func TestTokenFromRequest_fallsBackToPlainName(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: DefaultCookieName, Value: "legacy"})
	if got := TokenFromRequest(req); got != "legacy" {
		t.Errorf("TokenFromRequest = %q, want legacy", got)
	}
}

func TestClearCookie_secureClearsHostAndPlainNames(t *testing.T) {
	rr := httptest.NewRecorder()
	ClearCookie(rr, CookieOptions{Secure: true})
	names := map[string]bool{}
	for _, c := range rr.Result().Cookies() {
		if c.MaxAge != -1 {
			t.Errorf("%s MaxAge = %d, want -1", c.Name, c.MaxAge)
		}
		if !c.Secure {
			t.Errorf("%s should be Secure", c.Name)
		}
		if c.Path != "/" || c.Domain != "" {
			t.Errorf("%s Path/Domain violate __Host- rules: %+v", c.Name, c)
		}
		names[c.Name] = true
	}
	if !names["__Host-cais_session"] || !names[DefaultCookieName] {
		t.Errorf("ClearCookie(Secure) names = %v, want both __Host-cais_session and %s", names, DefaultCookieName)
	}
}
