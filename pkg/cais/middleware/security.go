package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/puppe1990/amarra-cais/pkg/cais"
)

// SecurityHeaders sets baseline headers. script-src is 'self' plus a
// per-request nonce so the layout FOUC snippet and SW inline scripts can run
// without 'unsafe-inline' (#263). Set CSP_SCRIPT_SRC='unsafe-inline' to restore
// the old policy (a nonce would make that keyword a no-op in CSP3).
func SecurityHeaders(cfg cais.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = applySecurityHeaders(w, r, cfg)
			next.ServeHTTP(w, r)
		})
	}
}

func applySecurityHeaders(w http.ResponseWriter, r *http.Request, cfg cais.Config) *http.Request {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
	policy := cfg.PermissionsPolicy
	if policy == "" {
		policy = "camera=(), microphone=(), geolocation=()"
	}
	w.Header().Set("Permissions-Policy", policy)

	scriptSrc, nonce := scriptSrcDirective(cfg)
	if nonce != "" {
		r = r.WithContext(cais.WithScriptNonce(r.Context(), nonce))
	}
	w.Header().Set("Content-Security-Policy", fmt.Sprintf(
		"default-src 'self'; script-src %s; style-src %s; img-src %s; font-src %s; connect-src %s; media-src %s; frame-ancestors 'none'; base-uri 'self'; form-action 'self'",
		scriptSrc, styleSrc(cfg), imgSrc(cfg), fontSrc(cfg), connectSrc(cfg), mediaSrc(cfg),
	))
	if cfg.Env == "production" {
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	}
	return r
}

func scriptSrcDirective(cfg cais.Config) (directive, nonce string) {
	directive = "'self'"
	if strings.Contains(cfg.CSPScriptSrc, "'unsafe-inline'") {
		return appendSrc(directive, cfg.CSPScriptSrc), ""
	}
	nonce, err := cais.NewScriptNonce()
	if err == nil && nonce != "" {
		directive += " 'nonce-" + nonce + "'"
	}
	return appendSrc(directive, cfg.CSPScriptSrc), nonce
}

func styleSrc(cfg cais.Config) string {
	return appendSrc("'self' 'unsafe-inline'", cfg.CSPStyleSrc)
}

func connectSrc(cfg cais.Config) string {
	return appendSrc("'self'", cfg.CSPConnectSrc)
}

func mediaSrc(cfg cais.Config) string {
	return appendSrc("'self'", cfg.CSPMediaSrc)
}

func imgSrc(cfg cais.Config) string {
	return appendSrc("'self' data:", cfg.CSPImgSrc)
}

func fontSrc(cfg cais.Config) string {
	// font-src defaults to 'self' data: so hosted webfonts need CSP_FONT_SRC
	// (e.g. https://fonts.gstatic.com) alongside CSP_STYLE_SRC for stylesheets.
	return appendSrc("'self' data:", cfg.CSPFontSrc)
}

func appendSrc(base, extra string) string {
	if extra == "" {
		return base
	}
	return base + " " + extra
}
