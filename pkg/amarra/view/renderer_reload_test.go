package view

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/puppe1990/amarra-cais/pkg/cais"
	"github.com/puppe1990/amarra-cais/pkg/cais/i18n"
)

// writeReloadFixture drops a loadable template tree whose page body carries the
// marker, so a test can tell which revision the renderer is serving.
func writeReloadFixture(t *testing.T, dir, pageBody string) {
	t.Helper()
	write := func(rel, content string) {
		path := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("layouts/app.html", `{{ define "app" }}<html lang="{{ htmlLang }}"><body>{{ template "content" . }}{{ t "greet" }}</body></html>{{ end }}`)
	write("pages/home.html", fmt.Sprintf(`{{ define "content" }}{{ t "greet" }} %s{{ end }}`, pageBody))
}

func renderLocale(t *testing.T, rec *Renderer, locale string) string {
	t.Helper()
	en, pt := i18nTestCatalogs()
	catalogs := map[string]*i18n.Catalog{"en": en, "pt": pt}
	h := i18n.LocaleMiddleware(catalogs, "en")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Write(w, r, rec, Page{Layout: "app", Name: "home", Data: map[string]any{}}, cais.Config{})
	}))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if locale != "" {
		req.AddCookie(&http.Cookie{Name: i18n.CookieName, Value: locale})
	}
	h.ServeHTTP(rr, req)
	return rr.Body.String()
}

// #340: `amarra-cais dev` edits a .html and the running process serves it without a
// rebuild — ReloadFromDir re-parses the same catalog sets and swaps them in place.
func TestRenderer_ReloadFromDir_swapsPagesAndLocales(t *testing.T) {
	dir := t.TempDir()
	writeReloadFixture(t, dir, "home v1")

	en, pt := i18nTestCatalogs()
	rec, err := Load(os.DirFS(dir), en, pt)
	if err != nil {
		t.Fatal(err)
	}
	if body := renderLocale(t, rec, ""); !strings.Contains(body, "home v1") {
		t.Fatalf("boot body = %q", body)
	}

	writeReloadFixture(t, dir, "home v2")
	if err := rec.ReloadFromDir(dir); err != nil {
		t.Fatal(err)
	}

	if body := renderLocale(t, rec, ""); !strings.Contains(body, "home v2") || !strings.Contains(body, "Hello") {
		t.Fatalf("reloaded default body = %q", body)
	}
	// The extra catalogs are re-parsed too, or the pt page keeps the old markup (#211).
	body := renderLocale(t, rec, "pt")
	if !strings.Contains(body, "home v2") || !strings.Contains(body, "Olá") {
		t.Fatalf("reloaded pt body = %q", body)
	}
}

func TestRenderer_ReloadFromDir_keepsPreviousOnParseError(t *testing.T) {
	dir := t.TempDir()
	writeReloadFixture(t, dir, "home v1")

	rec, err := Load(os.DirFS(dir), nil)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "pages", "home.html"), []byte(`{{ define "content" }}{{ if }}broken`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := rec.ReloadFromDir(dir); err == nil {
		t.Fatal("a broken template must fail the reload")
	}
	if body := renderLocale(t, rec, ""); !strings.Contains(body, "home v1") {
		t.Fatalf("renderer must keep serving the last good parse, body = %q", body)
	}
}

// A page added while the server runs is addressable after the reload — the whole
// page set is re-globbed, not only the files that existed at boot.
func TestRenderer_ReloadFromDir_picksUpNewPages(t *testing.T) {
	dir := t.TempDir()
	writeReloadFixture(t, dir, "home v1")

	rec, err := Load(os.DirFS(dir), nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "pages", "fresh.html"), []byte(`{{ define "content" }}fresh{{ end }}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := rec.ReloadFromDir(dir); err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	Write(rr, httptest.NewRequest(http.MethodGet, "/fresh", nil), rec, Page{Layout: "app", Name: "fresh"}, cais.Config{})
	if body := rr.Body.String(); !strings.Contains(body, "fresh") {
		t.Fatalf("new page body = %q", body)
	}
}
