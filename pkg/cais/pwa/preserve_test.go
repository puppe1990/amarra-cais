package pwa

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// brandFixture overwrites every app-owned brand asset with sentinel content so a
// later install can prove it left them alone.
func brandFixture(t *testing.T, staticDir string) map[string]string {
	t.Helper()
	custom := map[string]string{
		"manifest.webmanifest":        `{"name":"Cifra","display":"standalone"}`,
		"offline.html":                "<html><body>custom offline</body></html>",
		"og.png":                      "custom og",
		"favicon.svg":                 "custom favicon",
		"icons/icon.png":              "custom icon",
		"icons/icon-192.png":          "custom 192",
		"icons/icon-512.png":          "custom 512",
		"icons/icon-512-maskable.png": "custom maskable",
	}
	for rel, body := range custom {
		if err := os.WriteFile(filepath.Join(staticDir, filepath.FromSlash(rel)), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return custom
}

func assertBrandPreserved(t *testing.T, staticDir string, custom map[string]string) {
	t.Helper()
	for rel, want := range custom {
		got, err := os.ReadFile(filepath.Join(staticDir, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		if string(got) != want {
			t.Errorf("%s was overwritten by pwa:\n got %q\nwant %q", rel, got, want)
		}
	}
}

// #186: refreshing the runtime must not clobber app-owned branding. This mirrors
// the reported aws-finops case where the manifest regressed to the directory name.
func TestWriteStaticAmarra_preservesExistingBrandAssets(t *testing.T) {
	dir := t.TempDir()
	if err := InstallForAmarra(dir, "Cifra"); err != nil {
		t.Fatal(err)
	}
	staticDir := filepath.Join(dir, "web", "static")
	custom := brandFixture(t, staticDir)

	// Stale runtime after a framework upgrade.
	amarra := filepath.Join(staticDir, "js", "amarra.js")
	if err := os.WriteFile(amarra, []byte("// stale runtime"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := InstallForAmarra(dir, "aws-finops"); err != nil {
		t.Fatal(err)
	}

	assertBrandPreserved(t, staticDir, custom)
	body, err := os.ReadFile(amarra)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) == "// stale runtime" {
		t.Error("framework runtime amarra.js was not refreshed")
	}
}

// A per-file gap (e.g. maskable added after the app was scaffolded) still lands:
// missing assets are written, present ones are preserved.
func TestWriteStaticAmarra_backfillsOnlyMissingAssets(t *testing.T) {
	dir := t.TempDir()
	if err := InstallForAmarra(dir, "Cifra"); err != nil {
		t.Fatal(err)
	}
	staticDir := filepath.Join(dir, "web", "static")
	custom := brandFixture(t, staticDir)

	maskable := filepath.Join(staticDir, "icons", "icon-512-maskable.png")
	if err := os.Remove(maskable); err != nil {
		t.Fatal(err)
	}
	delete(custom, "icons/icon-512-maskable.png")

	if err := InstallForAmarra(dir, "Cifra"); err != nil {
		t.Fatal(err)
	}
	assertBrandPreserved(t, staticDir, custom)
	if _, err := os.Stat(maskable); err != nil {
		t.Errorf("missing maskable icon was not backfilled: %v", err)
	}
}

// Config.Force is the explicit escape hatch to regenerate framework defaults.
func TestWriteStaticInertia_forceOverwritesBrandAssets(t *testing.T) {
	dir := t.TempDir()
	if err := WriteStaticInertia(dir, DefaultConfig("Default")); err != nil {
		t.Fatal(err)
	}
	staticDir := filepath.Join(dir, "web", "static")
	brandFixture(t, staticDir)

	cfg := DefaultConfig("Default")
	cfg.Force = true
	if err := WriteStaticInertia(dir, cfg); err != nil {
		t.Fatal(err)
	}

	manifest, err := os.ReadFile(filepath.Join(staticDir, "manifest.webmanifest"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(manifest), `"name": "Default"`) {
		t.Errorf("--force should rewrite the manifest, got: %s", manifest)
	}
	offline := readStatic(t, staticDir, "offline.html")
	if strings.Contains(string(offline), "custom offline") {
		t.Error("--force should rewrite offline.html")
	}
}
