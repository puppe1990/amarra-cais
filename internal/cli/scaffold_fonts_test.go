package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The default design carries the docs-site heading font, so a fresh scaffold
// links Fraunces and must allow the font hosts in the CSP — otherwise the
// browser silently falls back to the system serif and doctor warns (#57).
func TestScaffold_FrauncesIsLinkedAndAllowedByCSP(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "fonty")

	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "fonty",
		ModulePath: "github.com/puppe1990/fonty",
	}, false, false); err != nil {
		t.Fatal(err)
	}

	read := func(rel string) string {
		t.Helper()
		body, err := os.ReadFile(filepath.Join(appDir, rel))
		if err != nil {
			t.Fatal(err)
		}
		return string(body)
	}

	layout := read("web/templates/layouts/app.html")
	for _, want := range []string{
		`rel="preconnect" href="https://fonts.gstatic.com"`,
		"family=Fraunces:opsz,wght@9..144,500;9..144,620",
	} {
		if !strings.Contains(layout, want) {
			t.Errorf("layout missing %q", want)
		}
	}

	env := read(".env")
	for _, want := range []string{
		"CSP_STYLE_SRC=https://fonts.googleapis.com",
		"CSP_FONT_SRC=https://fonts.gstatic.com",
	} {
		if !strings.Contains(env, want) {
			t.Errorf(".env missing %q — the browser would block the font", want)
		}
	}

	if check := checkGoogleFonts(appDir); !check.OK {
		t.Errorf("fresh scaffold fails the CSP fonts check: %s (%s)", check.Detail, check.FixHint)
	}
}
