package middleware

import (
	"net/http"
	"strings"

	"github.com/puppe1990/amarra-cais/pkg/cais"
	"github.com/puppe1990/amarra-cais/pkg/cais/csrf"
	"github.com/puppe1990/amarra-cais/pkg/cais/httpx"
)

// CSRF protects state-changing requests with a double-submit cookie token.
func CSRF(cfg cais.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return csrfHandler(cfg, next)
	}
}

func csrfHandler(cfg cais.Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if skipCSRF(r) {
			next.ServeHTTP(w, r)
			return
		}

		if isSafeMethod(r.Method) {
			token, err := csrf.EnsureToken(w, r, cfg.CookieSecure())
			if err != nil {
				http.Error(w, "csrf token error", http.StatusInternalServerError)
				return
			}
			r = r.WithContext(csrf.WithToken(r.Context(), token))
			next.ServeHTTP(w, r)
			return
		}

		// ParseForm does not read multipart bodies, so the csrf_token form field
		// of a classic multipart upload would never reach FormValue and every
		// such request died with 403 (#26). ParseFormOrJSON also covers JSON
		// posts that carry the token in the body instead of the header.
		//
		// Cap the total body first (#221): ParseMultipartForm's memory threshold
		// bounds only the in-memory fraction, so an anonymous client could spill
		// arbitrarily large uploads to disk before the token was ever checked.
		httpx.LimitBody(w, r, cfg.BodyLimit())
		if err := httpx.ParseFormOrJSON(r); err != nil {
			httpx.CleanupMultipart(r)
			if httpx.BodyTooLarge(err) {
				http.Error(w, "request entity too large", http.StatusRequestEntityTooLarge)
				return
			}
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if !csrf.Valid(r) {
			http.Error(w, "invalid csrf token", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isSafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}

// skipCSRF reports whether a request bypasses the double-submit check: health
// and static probes, the /api/ machine-client convention, and requests carrying
// an Authorization: Bearer header (#331) — a machine client has no CSRF cookie
// or form field, so the 403 used to run before the handler authenticated it.
//
// A classic HTML form post keeps its CSRF protection even when a stray Bearer
// header is present: cross-site forms cannot attach Authorization, so a Bearer
// on form content there is browser/extension noise, not a machine client.
func skipCSRF(r *http.Request) bool {
	path := r.URL.Path
	if path == "/health" ||
		strings.HasPrefix(path, "/static/") ||
		strings.HasPrefix(path, "/api/") {
		return true
	}
	return hasBearerAuthorization(r) && !isFormContentType(r)
}

func hasBearerAuthorization(r *http.Request) bool {
	return strings.HasPrefix(strings.ToLower(r.Header.Get("Authorization")), "bearer ")
}

func isFormContentType(r *http.Request) bool {
	ct := r.Header.Get("Content-Type")
	if i := strings.IndexByte(ct, ';'); i >= 0 {
		ct = ct[:i]
	}
	switch strings.ToLower(strings.TrimSpace(ct)) {
	case "application/x-www-form-urlencoded", "multipart/form-data", "text/plain":
		return true
	default:
		return false
	}
}
