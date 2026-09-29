package cli

import (
	"path/filepath"
	"strings"
	"testing"
)

// #127: scaffolded handlers must not write raw store errors to the client.
func TestScaffoldHandlers_useSanitizedServerError(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "errorsanitize")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "errorsanitize",
		ModulePath: "github.com/puppe1990/errorsanitize",
	}, false, false); err != nil {
		t.Fatal(err)
	}

	for _, rel := range []string{
		"internal/handlers/contact.go",
		"internal/handlers/dashboard.go",
		"internal/handlers/auth.go",
	} {
		body := mustReadFile(t, filepath.Join(appDir, rel))
		if strings.Contains(body, "http.Error(w, err.Error(), http.StatusInternalServerError)") {
			t.Errorf("%s still writes the raw store error to the client", rel)
		}
		if !strings.Contains(body, "httpx.ServerError(w, err, h.cfg)") {
			t.Errorf("%s should route store errors through httpx.ServerError", rel)
		}
	}

	sitemap := buildSitemapHandler("github.com/puppe1990/errorsanitize")
	if strings.Contains(sitemap, "http.Error(w, err.Error()") {
		t.Error("sitemap handler still writes the raw store error to the client")
	}
	if !strings.Contains(sitemap, "httpx.ServerError(w, err, h.cfg)") {
		t.Error("sitemap handler should route store errors through httpx.ServerError")
	}
}
