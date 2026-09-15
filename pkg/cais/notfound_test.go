package cais

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func designedNotFound(called *bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		*called = true
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("designed 404"))
	}
}

func TestRouter_NotFound_servesUnknownRoutes(t *testing.T) {
	r := NewRouter()
	called := false
	r.NotFound(designedNotFound(&called))
	r.Get("/{$}", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/rota-inexistente", nil))

	if !called {
		t.Fatal("custom not-found handler was not called for an unknown route")
	}
	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "designed 404") {
		t.Errorf("body = %q", rr.Body.String())
	}
}

func TestRouter_NotFound_keepsKnownRoutesAndRoot(t *testing.T) {
	r := NewRouter()
	called := false
	r.NotFound(designedNotFound(&called))
	r.Get("/{$}", func(w http.ResponseWriter, req *http.Request) {
		_, _ = w.Write([]byte("home"))
	})
	r.Get("/blog/{slug}", StringParam("slug", func(w http.ResponseWriter, req *http.Request, slug string) {
		_, _ = w.Write([]byte("post " + slug))
	}))

	for path, want := range map[string]string{"/": "home", "/blog/ola": "post ola"} {
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, path, nil))
		if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), want) {
			t.Errorf("GET %s = %d %q, want %q", path, rr.Code, rr.Body.String(), want)
		}
	}
	if called {
		t.Error("not-found handler must not run for a matching route")
	}
}

func TestRouter_NotFound_usedByParamHelpers(t *testing.T) {
	r := NewRouter()
	called := false
	r.NotFound(designedNotFound(&called))
	r.Get("/items/{id}", IntParam("id", func(w http.ResponseWriter, req *http.Request, id int64) {
		t.Error("handler must not run for an invalid path param")
	}))

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/items/abc", nil))

	if !called {
		t.Fatal("param helper must use the router not-found handler")
	}
	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rr.Code)
	}
}

func TestRouter_NotFound_appliesToGroupedRoutes(t *testing.T) {
	r := NewRouter()
	called := false
	r.NotFound(designedNotFound(&called))
	pass := func(next http.Handler) http.Handler { return next }
	r.Group(pass, func(g *Router) {
		g.Get("/admin/items/{id}", IntParam("id", func(w http.ResponseWriter, req *http.Request, id int64) {
			t.Error("handler must not run for an invalid path param")
		}))
	})

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/admin/items/x", nil))

	if !called {
		t.Fatal("grouped routes must share the parent not-found handler")
	}
}

func TestRouter_NotFound_defaultsToPlain404(t *testing.T) {
	r := NewRouter()
	r.Get("/{$}", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/nope", nil))

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "404 page not found") {
		t.Errorf("default body = %q", rr.Body.String())
	}
}

// Apps that adopted the documented `/{path...}` catch-all keep working: it is a
// real route, so it wins over the router fallback (#62).
func TestRouter_NotFound_coexistsWithAppCatchAll(t *testing.T) {
	r := NewRouter()
	catchAll := false
	r.Get("/{caminho...}", func(w http.ResponseWriter, req *http.Request) {
		catchAll = true
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("app catch-all"))
	})
	called := false
	r.NotFound(designedNotFound(&called))

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/qualquer", nil))

	if !catchAll {
		t.Error("the app's own catch-all route must keep serving")
	}
	if called {
		t.Error("router fallback must not run when a route matched")
	}
	if !strings.Contains(rr.Body.String(), "app catch-all") {
		t.Errorf("body = %q", rr.Body.String())
	}
}
