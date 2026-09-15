package cais

import (
	"context"
	"net/http"
)

// notFoundKey carries the router's not-found slot through the request so the
// package-level param helpers (IntParam, StringParam, …) can render the app's
// page instead of the plain `404 page not found` (#62).
type notFoundKey struct{}

type notFoundSlot struct {
	handler http.HandlerFunc
}

// NotFound registers the handler for unmatched routes and for path parameters
// that fail to parse. The handler owns the status: pass
// view.Write(..., view.Page{Status: http.StatusNotFound}) to ship a designed page.
func (r *Router) NotFound(handler http.HandlerFunc) {
	r.notFound.handler = handler
}

// serveNotFound runs the router handler when the request came through the
// router; helpers called directly (tests, custom muxes) keep the plain 404.
func serveNotFound(w http.ResponseWriter, r *http.Request) {
	if slot, ok := r.Context().Value(notFoundKey{}).(*notFoundSlot); ok && slot.handler != nil {
		slot.handler(w, r)
		return
	}
	http.NotFound(w, r)
}

func (r *Router) withNotFoundSlot(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		next.ServeHTTP(w, req.WithContext(context.WithValue(req.Context(), notFoundKey{}, r.notFound)))
	})
}

// serveNoPattern handles a request no pattern matched, keeping the router
// middlewares (session/flash/CSRF) around the designed page. Detected via
// mux.Handler instead of registering "/", which would conflict with an app's
// own `/{path...}` catch-all.
func (r *Router) serveNoPattern(w http.ResponseWriter, req *http.Request) {
	handler := http.Handler(r.notFound.handler)
	if r.notFound.handler == nil {
		handler = http.NotFoundHandler()
	}
	r.wrap(handler).ServeHTTP(w, req)
}
