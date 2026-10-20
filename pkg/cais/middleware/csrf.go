package middleware

import (
	"net/http"
	"strings"

	"github.com/puppe1990/amarra-cais/pkg/cais"
	"github.com/puppe1990/amarra-cais/pkg/cais/csrf"
	"github.com/puppe1990/amarra-cais/pkg/cais/httpx"
)

// CSRF protects state-changing requests with a double-submit cookie token.
// Options add explicit bearer-only API prefixes; no path is exempt by naming
// convention alone (#364).
func CSRF(cfg cais.Config, opts ...CSRFOption) func(http.Handler) http.Handler {
	policy := newCSRFPolicy(cfg, opts...)
	return func(next http.Handler) http.Handler {
		return csrfHandler(cfg, policy, next)
	}
}

// CSRFOption configures a bearer-only API exemption.
type CSRFOption func(*csrfPolicy)

// BearerVerifier reports whether the request carries a Bearer token accepted by
// an app-specific API.
type BearerVerifier func(*http.Request) bool

// BearerAPI declares prefix as an API authenticated exclusively by a Bearer
// token accepted by verify. Requests there skip the double-submit check only
// when they carry a Bearer token and verify accepts it; form-encoded machine
// clients are fine because a browser form cannot attach Authorization. Never
// register a prefix that also serves session routes: the session fallback
// would turn a stray header into a CSRF bypass (#364).
func BearerAPI(prefix string, verify BearerVerifier) CSRFOption {
	return func(policy *csrfPolicy) {
		if prefix == "" || verify == nil {
			return
		}
		policy.bearerAPIs = append(policy.bearerAPIs, csrfBearerAPI{prefix: prefix, verify: verify})
	}
}

type csrfPolicy struct {
	adminToken string
	bearerAPIs []csrfBearerAPI
}

type csrfBearerAPI struct {
	prefix string
	verify BearerVerifier
}

func newCSRFPolicy(cfg cais.Config, opts ...CSRFOption) csrfPolicy {
	policy := csrfPolicy{adminToken: cfg.AdminToken}
	for _, opt := range opts {
		if opt != nil {
			opt(&policy)
		}
	}
	return policy
}

func csrfHandler(cfg cais.Config, policy csrfPolicy, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if skipCSRF(policy, r) {
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

// skipCSRF reports whether a request bypasses the double-submit check:
// health/static probes, crawler files that carry no form, and machine clients
// that authenticate with a Bearer token the route's auth gate would accept.
//
// #364: /api/ is a naming convention, not proof that the route is Bearer-only.
// A session route under it must keep CSRF; otherwise a cross-site form (or a
// stray Authorization header) could mutate with the victim's cookie.
func skipCSRF(policy csrfPolicy, r *http.Request) bool {
	path := r.URL.Path
	if path == "/health" || isCrawlerFile(path) || strings.HasPrefix(path, "/static/") {
		return true
	}
	return policy.skipsBearer(r)
}

// skipsBearer accepts a machine client only when the token is one AdminAuth
// would accept (or an explicitly registered BearerAPI verifier). The token is
// validated before any content-type check: curl -d sends a form body, and a
// cross-site form cannot attach Authorization at all. An unvalidated header on
// a form therefore still falls through to the double-submit check (#364).
func (p csrfPolicy) skipsBearer(r *http.Request) bool {
	if !hasBearerAuthorization(r) {
		return false
	}
	if bearerMatches(r, p.adminToken) {
		return true
	}
	for _, api := range p.bearerAPIs {
		if strings.HasPrefix(r.URL.Path, api.prefix) && api.verify(r) {
			return true
		}
	}
	return false
}

// isCrawlerFile covers the public documents a crawler fetches: the framework's
// /health sibling and the sitemap/robots/llms trio `g sitemap` generates. A
// session cookie on them is noise — Google Search Console reads a sitemap that
// sets cookies as a non-public resource and refuses to fetch it (#342).
//
// Closed set on purpose: an app's own *.xml route (feed, export) may be
// authenticated, so this must never become an extension match.
func isCrawlerFile(path string) bool {
	switch path {
	case "/sitemap.xml", "/robots.txt", "/llms.txt":
		return true
	default:
		return false
	}
}

func hasBearerAuthorization(r *http.Request) bool {
	return strings.HasPrefix(strings.ToLower(r.Header.Get("Authorization")), "bearer ")
}
