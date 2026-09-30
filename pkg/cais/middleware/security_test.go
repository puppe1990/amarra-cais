package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/puppe1990/amarra-cais/pkg/cais"
)

func TestSecurityHeaders_production(t *testing.T) {
	cfg := cais.Config{Env: "production", AppURL: "https://app.example.com"}
	h := SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	for _, key := range []string{
		"X-Content-Type-Options",
		"X-Frame-Options",
		"Referrer-Policy",
		"Permissions-Policy",
		"Content-Security-Policy",
		"Strict-Transport-Security",
	} {
		if rr.Header().Get(key) == "" {
			t.Errorf("missing header %s", key)
		}
	}
}

func TestSecurityHeaders_customPolicy(t *testing.T) {
	cfg := cais.Config{
		Env:               "development",
		PermissionsPolicy: "camera=(self), geolocation=(self)",
		CSPStyleSrc:       "https://fonts.googleapis.com",
		CSPFontSrc:        "https://fonts.gstatic.com",
	}
	h := SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	if got := rr.Header().Get("Permissions-Policy"); got != "camera=(self), geolocation=(self)" {
		t.Errorf("Permissions-Policy = %q", got)
	}
	csp := rr.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "https://fonts.googleapis.com") {
		t.Errorf("CSP missing style src extra: %q", csp)
	}
	if !strings.Contains(csp, "font-src 'self' data: https://fonts.gstatic.com") {
		t.Errorf("CSP missing font-src extra: %q", csp)
	}
}

func TestSecurityHeaders_defaultFontSrc(t *testing.T) {
	cfg := cais.Config{Env: "production"}
	h := SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	csp := rr.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "font-src 'self' data:") {
		t.Errorf("CSP missing default font-src: %q", csp)
	}
}

func TestSecurityHeaders_mediaSrc(t *testing.T) {
	cfg := cais.Config{
		Env:         "development",
		CSPMediaSrc: "blob:",
	}
	h := SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	csp := rr.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "media-src 'self' blob:") {
		t.Errorf("CSP missing media-src blob:, got %q", csp)
	}
}

func TestSecurityHeaders_development_defaultsAreNeutral(t *testing.T) {
	// #262: an empty Config (what Load() now yields in development) must not
	// open the camera or a third-party image CDN.
	cfg := cais.Config{Env: "development", CSPMediaSrc: "blob:"}
	h := SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	if got := rr.Header().Get("Permissions-Policy"); got != "camera=(), microphone=(), geolocation=()" {
		t.Errorf("Permissions-Policy = %q, want camera=()", got)
	}
	csp := rr.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "img-src 'self' data:") {
		t.Errorf("CSP missing img-src 'self' data:, got %q", csp)
	}
	if strings.Contains(csp, "openfoodfacts") {
		t.Errorf("CSP still allows Open Food Facts images: %q", csp)
	}
}

func TestSecurityHeaders_development_allowsCamera(t *testing.T) {
	cfg := cais.Config{Env: "development", PermissionsPolicy: "camera=(self), microphone=(), geolocation=()", CSPMediaSrc: "blob:"}
	h := SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	if got := rr.Header().Get("Permissions-Policy"); !strings.Contains(got, "camera=(self)") {
		t.Errorf("Permissions-Policy = %q, want camera=(self)", got)
	}
	csp := rr.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "media-src 'self' blob:") {
		t.Errorf("CSP = %q", csp)
	}
}

func TestSecurityHeaders_development_noHSTS(t *testing.T) {
	cfg := cais.Config{Env: "development"}
	h := SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	if rr.Header().Get("Strict-Transport-Security") != "" {
		t.Error("HSTS should not be set in development")
	}
}

func TestSecurityHeaders_scriptSrcOmitsUnsafeInline(t *testing.T) {
	// #263: the FOUC snippet and SW register/unregister ride a per-request
	// nonce so script-src can drop 'unsafe-inline'. style-src still allows it.
	cfg := cais.Config{Env: "development"}
	h := SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	csp := rr.Header().Get("Content-Security-Policy")
	script := cspDirective(csp, "script-src")
	if script == "" {
		t.Fatalf("missing script-src in %q", csp)
	}
	if strings.Contains(script, "'unsafe-inline'") {
		t.Errorf("script-src still allows unsafe-inline: %q", script)
	}
	if !strings.Contains(script, "'self'") {
		t.Errorf("script-src missing 'self': %q", script)
	}
	if !strings.Contains(script, "'nonce-") {
		t.Errorf("script-src missing nonce: %q", script)
	}
}

func TestSecurityHeaders_nonceDiffersPerRequest(t *testing.T) {
	cfg := cais.Config{Env: "development"}
	h := SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	rr1 := httptest.NewRecorder()
	h.ServeHTTP(rr1, httptest.NewRequest(http.MethodGet, "/", nil))
	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, httptest.NewRequest(http.MethodGet, "/", nil))

	n1 := cspDirective(rr1.Header().Get("Content-Security-Policy"), "script-src")
	n2 := cspDirective(rr2.Header().Get("Content-Security-Policy"), "script-src")
	if n1 == "" || n1 == n2 {
		t.Errorf("expected distinct script-src nonces, got %q and %q", n1, n2)
	}
}

func TestSecurityHeaders_unsafeInlineEscapeHatchSkipsNonce(t *testing.T) {
	// A nonce would make 'unsafe-inline' a no-op in CSP3, so the hatch
	// must omit the nonce entirely (#263).
	cfg := cais.Config{Env: "development", CSPScriptSrc: "'unsafe-inline'"}
	var got string
	h := SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = cais.ScriptNonceFromRequest(r)
	}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	script := cspDirective(rr.Header().Get("Content-Security-Policy"), "script-src")
	if !strings.Contains(script, "'unsafe-inline'") {
		t.Errorf("script-src missing unsafe-inline hatch: %q", script)
	}
	if strings.Contains(script, "'nonce-") {
		t.Errorf("script-src still issued a nonce: %q", script)
	}
	if got != "" {
		t.Errorf("request nonce = %q, want empty when hatch is set", got)
	}
}

func TestSecurityHeaders_cspScriptSrcAppendsHosts(t *testing.T) {
	cfg := cais.Config{Env: "development", CSPScriptSrc: "https://cdn.example.com"}
	h := SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	script := cspDirective(rr.Header().Get("Content-Security-Policy"), "script-src")
	if !strings.Contains(script, "https://cdn.example.com") {
		t.Errorf("script-src missing extra host: %q", script)
	}
	if strings.Contains(script, "'unsafe-inline'") {
		t.Errorf("script-src gained unsafe-inline: %q", script)
	}
	if !strings.Contains(script, "'nonce-") {
		t.Errorf("script-src missing nonce: %q", script)
	}
}

func TestSecurityHeaders_exposesNonceOnRequest(t *testing.T) {
	cfg := cais.Config{Env: "development"}
	var got string
	h := SecurityHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = cais.ScriptNonceFromRequest(r)
	}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	script := cspDirective(rr.Header().Get("Content-Security-Policy"), "script-src")
	want := "'nonce-" + got + "'"
	if got == "" || !strings.Contains(script, want) {
		t.Errorf("request nonce %q missing from script-src %q", got, script)
	}
}

func cspDirective(csp, name string) string {
	for _, part := range strings.Split(csp, ";") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, name+" ") || part == name {
			return part
		}
	}
	return ""
}
