package cais

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	if err := os.WriteFile(filepath.Join(dir, "icon.png"), []byte("png"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := NewRouter()
	r.StaticForEnv("/static", dir, Config{Env: "production"})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/static/icon.png", nil)
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	cc := rr.Header().Get("Cache-Control")
	if !strings.Contains(cc, "public") {
		t.Errorf("Cache-Control = %q, want public", cc)
	}
	age := cacheControlMaxAge(cc)
	if age < 86400 {
		t.Errorf("Cache-Control = %q, want max-age >= 86400 (got %d)", cc, age)
	}
	if strings.Contains(cc, "no-store") {
		t.Error("production static should not force no-store")
	}
}

// #359: styles.css and amarra.js are stable names. A year-long TTL keeps the
// previous deploy's bytes in the HTTP cache after HTML (no-store) already
// points at the new files.
func TestStaticForEnv_production_cssJSMustRevalidate(t *testing.T) {
	dir := t.TempDir()
	cssDir := filepath.Join(dir, "css")
	jsDir := filepath.Join(dir, "js")
	if err := os.MkdirAll(cssDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(jsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cssDir, "styles.css"), []byte("body{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(jsDir, "amarra.js"), []byte("/* js */"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := NewRouter()
	r.StaticForEnv("/static", dir, Config{Env: "production"})

	for _, path := range []string{"/static/css/styles.css", "/static/js/amarra.js"} {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		r.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("%s status = %d", path, rr.Code)
		}
		cc := rr.Header().Get("Cache-Control")
		if !strings.Contains(cc, "must-revalidate") && !strings.Contains(cc, "max-age=0") {
			t.Errorf("%s Cache-Control = %q, want must-revalidate or max-age=0", path, cc)
		}
		if cacheControlMaxAge(cc) >= 86400 {
			t.Errorf("%s must not be cached for a year, Cache-Control = %q", path, cc)
		}
		if strings.Contains(cc, "immutable") {
			t.Errorf("%s must not be immutable, got %q", path, cc)
		}
	}
}

func TestStaticForEnv_production_swJSMustRevalidate(t *testing.T) {
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
	cc := rr.Header().Get("Cache-Control")
	if !strings.Contains(cc, "must-revalidate") && !strings.Contains(cc, "max-age=0") {
		t.Errorf("sw.js Cache-Control = %q, want must-revalidate or max-age=0", cc)
	}
	if strings.Contains(cc, "immutable") {
		t.Errorf("sw.js must not be immutable, got %q", cc)
	}
}

func TestStatic_emptyConfigGetsProductionTTL(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "icon.png"), []byte("png"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := NewRouter()
	r.Static("/static", dir)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/static/icon.png", nil)
	r.ServeHTTP(rr, req)

	cc := rr.Header().Get("Cache-Control")
	if cc == "no-store" {
		t.Fatal("Static with empty Config must not force no-store")
	}
	if !strings.Contains(cc, "public") || cacheControlMaxAge(cc) < 86400 {
		t.Errorf("Static empty Config Cache-Control = %q, want production TTL", cc)
	}
}

func TestStaticForEnv_production_404NotLongLived(t *testing.T) {
	dir := t.TempDir()
	r := NewRouter()
	r.StaticForEnv("/static", dir, Config{Env: "production"})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/static/missing.js", nil)
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rr.Code)
	}
	cc := rr.Header().Get("Cache-Control")
	if cacheControlMaxAge(cc) >= 86400 {
		t.Errorf("404 must not be cached long, Cache-Control = %q", cc)
	}
}

func cacheControlMaxAge(cc string) int {
	for _, part := range strings.Split(cc, ",") {
		part = strings.TrimSpace(part)
		if !strings.HasPrefix(part, "max-age=") {
			continue
		}
		n, err := strconv.Atoi(strings.TrimPrefix(part, "max-age="))
		if err != nil {
			return 0
		}
		return n
	}
	return 0
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
