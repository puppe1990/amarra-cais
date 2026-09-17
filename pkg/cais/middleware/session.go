package middleware

import (
	"net/http"

	"github.com/puppe1990/amarra-cais/pkg/cais/session"
)

// LoadSession reads the session cookie and attaches the user ID to the request context.
func LoadSession(store session.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if token := session.TokenFromRequest(r); token != "" {
				if id, ok := store.Get(token); ok {
					r = session.WithUserID(r, id)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAuthFunc wraps a handler with RequireAuth.
func RequireAuthFunc(loginURL string, h http.HandlerFunc) http.HandlerFunc {
	auth := RequireAuth(loginURL)
	return func(w http.ResponseWriter, r *http.Request) {
		auth(http.HandlerFunc(h)).ServeHTTP(w, r)
	}
}

// RequireAuth blocks unauthenticated requests with a 303 to loginURL. Drive
// follows redirects and morphs the login page, so 303 is the single redirect
// behavior — the old HX-Redirect branch was dead contract (#94).
func RequireAuth(loginURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := session.UserID(r); !ok {
				http.Redirect(w, r, loginURL, http.StatusSeeOther)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
