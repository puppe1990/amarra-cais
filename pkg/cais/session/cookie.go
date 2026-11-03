package session

import (
	"net/http"
	"time"

	"github.com/puppe1990/amarra-cais/pkg/cais"
)

const DefaultCookieName = "cais_session"

// SecureCookieName carries the __Host- prefix in production (#395): browsers
// then reject any sibling-subdomain attempt to set it on the parent domain,
// blocking session cookie fixation. __Host- requires Secure, Path=/, no Domain.
const SecureCookieName = "__Host-cais_session"

const defaultMaxAge = 7 * 24 * 60 * 60 // 7 days

type CookieOptions struct {
	Secure bool
}

func CookieOptionsFromConfig(cfg cais.Config) CookieOptions {
	return CookieOptions{Secure: cfg.CookieSecure()}
}

func cookieName(opts CookieOptions) string {
	if opts.Secure {
		return SecureCookieName
	}
	return DefaultCookieName
}

func writeSessionCookie(w http.ResponseWriter, name, value string, maxAge int, opts CookieOptions) {
	c := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   opts.Secure,
	}
	if maxAge < 0 {
		c.Expires = time.Unix(0, 0)
	}
	http.SetCookie(w, c)
}

func SetCookie(w http.ResponseWriter, token string, opts CookieOptions) {
	writeSessionCookie(w, cookieName(opts), token, defaultMaxAge, opts)
}

func ClearCookie(w http.ResponseWriter, opts CookieOptions) {
	writeSessionCookie(w, cookieName(opts), "", -1, opts)
	if opts.Secure {
		// Drop a leftover plain cookie from a pre-__Host- deploy (#395).
		writeSessionCookie(w, DefaultCookieName, "", -1, opts)
	}
}

func TokenFromRequest(r *http.Request) string {
	if c, err := r.Cookie(SecureCookieName); err == nil && c.Value != "" {
		return c.Value
	}
	c, err := r.Cookie(DefaultCookieName)
	if err != nil {
		return ""
	}
	return c.Value
}
