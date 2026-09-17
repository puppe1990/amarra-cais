package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func scaffoldLayoutApp(t *testing.T) string {
	t.Helper()
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "layoutapp")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "layoutapp",
		ModulePath: "github.com/puppe1990/layoutapp",
	}, false, false); err != nil {
		t.Fatal(err)
	}
	return appDir
}

func readScaffoldFile(t *testing.T, appDir, rel string) string {
	t.Helper()
	body, err := os.ReadFile(filepath.Join(appDir, rel))
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

// The scaffold helper must not hardcode the layout: an app with a second layout
// (a marketing "landing") otherwise renders public pages inside the app chrome
// with no compile error and no test failure (#66).
func TestScaffoldViewData_writeViewTakesLayout(t *testing.T) {
	appDir := scaffoldLayoutApp(t)
	helper := readScaffoldFile(t, appDir, "internal/handlers/viewdata.go")

	for _, needle := range []string{
		"layout, name string",
		"Layout: layout",
		"Status: status",
	} {
		if !strings.Contains(helper, needle) {
			t.Errorf("writeView must take a layout and keep Status flowing: missing %q\n%s", needle, helper)
		}
	}
}

// Every scaffolded handler that renders through the helper must name its
// layout, so the choice is visible at the call site (#66).
func TestScaffoldHandlers_passLayoutExplicitly(t *testing.T) {
	appDir := scaffoldLayoutApp(t)

	for _, rel := range []string{
		"internal/handlers/contact.go",
		"internal/handlers/dashboard.go",
		"internal/handlers/auth.go",
	} {
		body := readScaffoldFile(t, appDir, rel)
		if !strings.Contains(body, `writeView(w, r, h.views, h.cfg, "app",`) {
			t.Errorf("%s must pass the layout to writeView:\n%s", rel, body)
		}
	}
}

// #131: /health is public; the scaffold must let netutil omit lan_urls in
// production instead of always shipping the host's RFC1918 addresses.
func TestScaffoldApp_healthPassesEnv(t *testing.T) {
	if !strings.Contains(tplApp, "netutil.HealthPayload(status, cfg.Port, cfg.Env)") {
		t.Error("app.go healthHandler must pass cfg.Env to netutil.HealthPayload")
	}
}
