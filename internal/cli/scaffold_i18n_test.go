package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// #211: the shipped language toggle must switch copy, so `new` wires every
// catalog into view.Load (per-locale page sets) and into LocaleMiddleware, and
// handler copy resolves through the request catalog instead of the boot one.
func TestScaffoldNewApp_wiresPerRequestCatalogs(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "i18nwire")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "i18nwire",
		ModulePath: "github.com/puppe1990/i18nwire",
	}, false, false); err != nil {
		t.Fatal(err)
	}

	mainSrc := scaffoldRead(t, appDir, "cmd/server/main.go")
	if !strings.Contains(mainSrc, `view.Load(tmplFS, catalog, catalogs["en"], catalogs["pt"])`) {
		t.Error("main.go must pass every catalog to view.Load so templates render per locale")
	}
	if !strings.Contains(mainSrc, "Catalogs:") {
		t.Error("main.go must hand the catalog map to app.Deps")
	}

	appSrc := scaffoldRead(t, appDir, "internal/app/app.go")
	if !strings.Contains(appSrc, "catalogs := deps.Catalogs") || !strings.Contains(appSrc, "i18n.LocaleMiddleware(catalogs") {
		t.Error("app.go must feed deps.Catalogs to LocaleMiddleware")
	}

	authSrc := scaffoldRead(t, appDir, "internal/handlers/auth.go")
	if strings.Contains(authSrc, "h.catalog.T(") {
		t.Error("auth.go must translate with the request catalog (h.t), not the boot catalog")
	}
	if !strings.Contains(authSrc, "i18n.CatalogOr(r, h.catalog)") {
		t.Error("auth.go needs a request-catalog helper")
	}

	homeSrc := scaffoldRead(t, appDir, "internal/handlers/home.go")
	if !strings.Contains(homeSrc, `h.t(r, "home.title")`) {
		t.Error("home.go title must resolve per request")
	}
}

func scaffoldRead(t *testing.T, dir, rel string) string {
	t.Helper()
	body, err := os.ReadFile(filepath.Join(dir, rel))
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

// #258: the sidebar theme toggle swaps its label, so both catalogs carry the
// light/dark copy.
func TestScaffoldNewApp_i18nIncludesThemeKeys(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "themei18n")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "themei18n",
		ModulePath: "github.com/puppe1990/themei18n",
	}, false, false); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		file string
		want []string
	}{
		{"internal/i18n/en.go", []string{`"layout.theme_light"`, `"layout.theme_dark"`, "Light mode", "Dark mode"}},
		{"internal/i18n/pt.go", []string{`"layout.theme_light"`, `"layout.theme_dark"`, "Modo claro", "Modo escuro"}},
	} {
		body := scaffoldRead(t, appDir, tc.file)
		for _, want := range tc.want {
			if !strings.Contains(body, want) {
				t.Errorf("%s missing %q", tc.file, want)
			}
		}
	}
}
