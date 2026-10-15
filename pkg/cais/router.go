package cais

import (
	"net/http"
	"strconv"
	"strings"
)

type Middleware func(http.Handler) http.Handler

type IntHandler func(http.ResponseWriter, *http.Request, int64)
type StringHandler func(http.ResponseWriter, *http.Request, string)
type StringParamsHandler func(http.ResponseWriter, *http.Request, string, string)
type IntStringParamsHandler func(http.ResponseWriter, *http.Request, int64, string)

type Router struct {
	mux         *http.ServeMux
	middlewares []Middleware
	// notFound is shared with groups: the parent registration applies to routes
	// declared on child routers (#62).
	notFound *notFoundSlot
}

func NewRouter() *Router {
	return &Router{
		mux:      http.NewServeMux(),
		notFound: &notFoundSlot{},
	}
}

func (r *Router) Use(mw Middleware) {
	r.middlewares = append(r.middlewares, mw)
}

// Group registers routes with extra middleware (e.g. admin auth).
// Child routers share the parent ServeMux — only middleware differs, not URL namespace.
func (r *Router) Group(mw Middleware, fn func(*Router)) {
	child := &Router{
		mux:         r.mux,
		middlewares: append(append([]Middleware{}, r.middlewares...), mw),
		notFound:    r.notFound,
	}
	fn(child)
}

func (r *Router) Get(pattern string, handler http.HandlerFunc) {
	r.register("GET", pattern, handler)
}

func (r *Router) Post(pattern string, handler http.HandlerFunc) {
	r.register("POST", pattern, handler)
}

func (r *Router) Put(pattern string, handler http.HandlerFunc) {
	r.register("PUT", pattern, handler)
}

func (r *Router) Patch(pattern string, handler http.HandlerFunc) {
	r.register("PATCH", pattern, handler)
}

func (r *Router) Delete(pattern string, handler http.HandlerFunc) {
	r.register("DELETE", pattern, handler)
}

func (r *Router) Handle(pattern string, handler http.Handler) {
	r.handlePattern(pattern, r.wrap(handler))
}

func (r *Router) Static(prefix, dir string) {
	r.StaticForEnv(prefix, dir, Config{})
}

const (
	// Year-long browser cache for fingerprinted-or-SW-bumped assets. URLs are
	// stable (/static/css/styles.css); skip immutable so a deploy + CACHE_VERSION
	// bump can still replace them. Set only on 200 so a missing file is not
	// cached for a year (#355).
	staticCacheTTL = "public, max-age=31536000"
	// sw.js must revalidate: a long-lived SW would pin an old CACHE_VERSION.
	staticSWCacheControl = "public, max-age=0, must-revalidate"
)

// StaticForEnv serves files from disk. Development is no-store so JS/CSS edits
// apply without a rebuild. Production sets a long TTL on 200s and
// must-revalidate on sw.js. Static is registered via handlePattern, not wrap(),
// so session/flash middleware never runs on CSS and fonts (#355).
func (r *Router) StaticForEnv(prefix, dir string, cfg Config) {
	fs := http.FileServer(http.Dir(dir))
	handler := allowRootServiceWorker(http.StripPrefix(prefix, fs))
	if cfg.Env == "development" {
		handler = noCacheStatic(handler)
	} else {
		handler = cacheStatic(handler)
	}
	r.handlePattern("GET "+prefix+"/", handler)
}

func noCacheStatic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func cacheStatic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(&staticCacheWriter{
			ResponseWriter: w,
			sw:             strings.HasSuffix(r.URL.Path, "/sw.js"),
		}, r)
	})
}

type staticCacheWriter struct {
	http.ResponseWriter
	sw    bool
	wrote bool
}

func (w *staticCacheWriter) WriteHeader(code int) {
	if !w.wrote {
		w.wrote = true
		if code == http.StatusOK {
			cc := staticCacheTTL
			if w.sw {
				cc = staticSWCacheControl
			}
			w.Header().Set("Cache-Control", cc)
		}
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *staticCacheWriter) Write(b []byte) (int, error) {
	if !w.wrote {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

// Default SW scope is the script directory. HTML lives at /, so sw.js must
// advertise a wider max scope or register({scope:"/"}) is rejected (#294).
func allowRootServiceWorker(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/sw.js") {
			w.Header().Set("Service-Worker-Allowed", "/")
		}
		next.ServeHTTP(w, r)
	})
}

func (r *Router) register(method, pattern string, handler http.HandlerFunc) {
	r.handlePattern(method+" "+pattern, r.wrap(handler))
}

// handlePattern registers on ServeMux and rewrites conflict panics with a cais-specific hint (#142).
func (r *Router) handlePattern(pattern string, handler http.Handler) {
	pattern = exactRoot(pattern)
	handler = r.withNotFoundSlot(handler)
	defer func() {
		if rec := recover(); rec != nil {
			panic(formatRouteConflict(pattern, rec))
		}
	}()
	r.mux.Handle(pattern, handler)
}

// exactRoot rewrites root "/" patterns to the end-of-path wildcard "/{$}" (#32).
// On Go 1.22+ ServeMux a trailing slash is a subtree pattern, so "GET /" matches
// every unmatched GET path (/.env, /nope → 200 from the home handler). "/{$}"
// matches only the exact root. Real subtree patterns ("/static/") are kept.
func exactRoot(pattern string) string {
	if pattern == "/" {
		return "/{$}"
	}
	method, path, found := strings.Cut(pattern, " ")
	if found && path == "/" {
		return method + " /{$}"
	}
	return pattern
}

func (r *Router) wrap(handler http.Handler) http.Handler {
	h := handler
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		h = r.middlewares[i](h)
	}
	return h
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if _, pattern := r.mux.Handler(req); pattern == "" {
		r.serveNoPattern(w, req)
		return
	}
	r.mux.ServeHTTP(w, req)
}

// IntParam wraps a handler that receives a parsed int64 path parameter.
func IntParam(name string, fn IntHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue(name), 10, 64)
		if err != nil || id <= 0 {
			serveNotFound(w, r)
			return
		}
		fn(w, r, id)
	}
}

// IntStringParams wraps a handler with an int64 and string path parameter.
func IntStringParams(intName, stringName string, fn IntStringParamsHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue(intName), 10, 64)
		if err != nil || id <= 0 {
			serveNotFound(w, r)
			return
		}
		s := r.PathValue(stringName)
		if s == "" {
			serveNotFound(w, r)
			return
		}
		fn(w, r, id, s)
	}
}

// StringParams wraps a handler with two string path parameters (avoids nested StringParam callbacks).
func StringParams(nameA, nameB string, fn StringParamsHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a := r.PathValue(nameA)
		b := r.PathValue(nameB)
		if a == "" || b == "" {
			serveNotFound(w, r)
			return
		}
		fn(w, r, a, b)
	}
}

// StringParam wraps a handler that receives a string path parameter.
func StringParam(name string, fn StringHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		v := r.PathValue(name)
		if v == "" {
			serveNotFound(w, r)
			return
		}
		fn(w, r, v)
	}
}
