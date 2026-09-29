package view

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/puppe1990/amarra-cais/pkg/cais"
	"github.com/puppe1990/amarra-cais/pkg/cais/i18n"
)

func i18nPageFS() fstest.MapFS {
	return fstest.MapFS{
		"layouts/app.html": &fstest.MapFile{Data: []byte(`{{ define "app" }}<html lang="{{ htmlLang }}"><body>{{ template "content" . }}{{ t "greet" }}</body></html>{{ end }}`)},
		"pages/home.html":  &fstest.MapFile{Data: []byte(`{{ define "content" }}{{ t "greet" }}{{ end }}`)},
	}
}

func i18nTestCatalogs() (en, pt *i18n.Catalog) {
	locales := map[string]map[string]string{
		"en": {"greet": "Hello"},
		"pt": {"greet": "Olá"},
	}
	return i18n.NewCatalogFrom("en", locales), i18n.NewCatalogFrom("pt", locales)
}

func TestWrite_usesRequestCatalog(t *testing.T) {
	en, pt := i18nTestCatalogs()
	rec, err := Load(i18nPageFS(), en, pt)
	if err != nil {
		t.Fatal(err)
	}
	catalogs := map[string]*i18n.Catalog{"en": en, "pt": pt}
	h := i18n.LocaleMiddleware(catalogs, "en")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Write(w, r, rec, Page{Layout: "app", Name: "home", Data: map[string]any{}}, cais.Config{})
	}))

	render := func(t *testing.T, locale string) string {
		t.Helper()
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		if locale != "" {
			req.AddCookie(&http.Cookie{Name: i18n.CookieName, Value: locale})
		}
		h.ServeHTTP(rr, req)
		return rr.Body.String()
	}

	if body := render(t, ""); !strings.Contains(body, "Hello") || !strings.Contains(body, `lang="en"`) {
		t.Fatalf("default locale body = %q", body)
	}
	body := render(t, "pt")
	if !strings.Contains(body, "Olá") || !strings.Contains(body, `lang="pt-BR"`) {
		t.Fatalf("pt body = %q", body)
	}
	if strings.Contains(body, "Hello") {
		t.Fatalf("boot catalog leaked into pt render: %q", body)
	}
	// Interleaving back to the boot locale must still work (no execute-once clone).
	if body := render(t, ""); !strings.Contains(body, "Hello") {
		t.Fatalf("back to default body = %q", body)
	}
}

func TestWrite_noRequestCatalogUsesBootCatalog(t *testing.T) {
	en, pt := i18nTestCatalogs()
	rec, err := Load(i18nPageFS(), en, pt)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	Write(rr, httptest.NewRequest(http.MethodGet, "/", nil), rec, Page{Layout: "app", Name: "home"}, cais.Config{})
	if !strings.Contains(rr.Body.String(), "Hello") {
		t.Fatalf("body = %q", rr.Body.String())
	}
}
