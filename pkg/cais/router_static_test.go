package cais

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestStaticForEnv_development_noCache(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "app.js"), []byte("v1"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := NewRouter()
	r.StaticForEnv("/static", dir, Config{Env: "development"})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/static/app.js", nil)
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	cc := rr.Header().Get("Cache-Control")
	if cc != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", cc)
	}
}

func TestStaticForEnv_production_allowsCache(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "app.js"), []byte("v1"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := NewRouter()
	r.StaticForEnv("/static", dir, Config{Env: "production"})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/static/app.js", nil)
	r.ServeHTTP(rr, req)

	if rr.Header().Get("Cache-Control") == "no-store" {
		t.Error("production static should not force no-store")
	}
}

// Default SW scope is the script directory (/static/js/). Pages live at /, so
// register({scope:"/"}) needs this header or Chrome rejects the wider scope (#294).
func TestStaticForEnv_swJSAllowsRootScope(t *testing.T) {
	dir := t.TempDir()
	jsDir := filepath.Join(dir, "js")
	if err := os.MkdirAll(jsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(jsDir, "sw.js"), []byte("/* sw */"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := NewRouter()
	r.StaticForEnv("/static", dir, Config{Env: "production"})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/static/js/sw.js", nil)
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	got := rr.Header().Get("Service-Worker-Allowed")
	if got != "/" {
		t.Errorf("Service-Worker-Allowed = %q, want /", got)
	}

	rrJS := httptest.NewRecorder()
	reqJS := httptest.NewRequest(http.MethodGet, "/static/js/amarra.js", nil)
	r.ServeHTTP(rrJS, reqJS)
	if rrJS.Header().Get("Service-Worker-Allowed") != "" {
		t.Error("non-SW scripts must not advertise Service-Worker-Allowed")
	}
}
