package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/puppe1990/amarra-cais/pkg/cais/pwa"
)

// A shared link or an installed PWA must not show the framework's placeholder
// art without the app noticing (#64).
func TestDoctor_WarnsWhileBrandAssetsAreScaffoldDefaults(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := scaffoldDoctorApp(t)

	out := runDoctorOutput(t, dir)
	if !strings.Contains(out, "[warn] brand assets") {
		t.Errorf("expected brand assets warning, got:\n%s", out)
	}
	if !strings.Contains(out, "icons/icon.png") || !strings.Contains(out, "og.png") {
		t.Errorf("warning should name the placeholder files, got:\n%s", out)
	}
}

func TestDoctor_OKAfterBrandAssetsReplaced(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := scaffoldDoctorApp(t)

	defaults := pwa.DefaultBrandAssets()
	if len(defaults) == 0 {
		t.Fatal("no default brand assets to replace")
	}
	for rel := range defaults {
		path := filepath.Join(dir, "web", "static", filepath.FromSlash(rel))
		if err := os.WriteFile(path, []byte("brand"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	out := runDoctorOutput(t, dir)
	if !strings.Contains(out, "[ok] brand assets") {
		t.Errorf("expected brand assets ok after replacing, got:\n%s", out)
	}
}
