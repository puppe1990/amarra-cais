package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLI_PWA_MigratesLegacyServiceWorker(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := t.TempDir()
	if err := scaffoldNewApp(dir, scaffoldData{
		AppName:    "pwa",
		ModulePath: "github.com/puppe1990/pwa",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	// Simulate pre-#127 SW (cache-first for all /static/).
	legacy := `const CACHE_VERSION = 3;
if (url.pathname.startsWith("/static/")) {
  caches.match(request);
}
`
	swPath := filepath.Join(dir, "web/static/js/sw.js")
	if err := os.WriteFile(swPath, []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	if err := c.cmdPWA(nil); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(swPath)
	if err != nil {
		t.Fatal(err)
	}
	s := string(body)
	if !strings.Contains(s, "/static/js/amarra.js") || !strings.Contains(s, "networkFirst") {
		t.Errorf("sw.js not migrated to network-first amarra.js:\n%s", s)
	}
	if strings.Contains(s, "/static/build/") {
		t.Errorf("sw.js should not treat /static/build/ as the SPA bundle path:\n%s", s)
	}
	if !strings.Contains(s, "CACHE_VERSION = 3") {
		t.Errorf("should preserve CACHE_VERSION=3, got:\n%s", s)
	}
	if !strings.Contains(buf.String(), "sw.js updated") {
		t.Errorf("expected migration message, got:\n%s", buf.String())
	}
}

func TestCLI_PWA_HTMLAppInstallsAmarraNotHTMX(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := t.TempDir()
	if err := scaffoldNewApp(dir, scaffoldData{
		AppName:    "pwa",
		ModulePath: "github.com/puppe1990/pwa",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(dir, "web/static/js/amarra.js")); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	if err := c.cmdPWA(nil); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "web/static/js/amarra.js")); err != nil {
		t.Fatalf("expected amarra.js after pwa: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "web/static/js/htmx.min.js")); err == nil {
		t.Fatal("HTML app PWA must not install htmx")
	}
	for _, rel := range []string{
		"web/static/js/sse-ext.min.js",
		"web/static/js/cais-core.js",
		"web/static/js/cais.js",
	} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err == nil {
			t.Errorf("HTML app PWA must not install %s", rel)
		}
	}
}

func TestCLI_PWA_Bump(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := t.TempDir()
	if err := scaffoldNewApp(dir, scaffoldData{
		AppName:    "pwa",
		ModulePath: "github.com/puppe1990/pwa",
	}, true, false); err != nil {
		t.Fatal(err)
	}

	c := &CLI{Out: &bytes.Buffer{}}
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	if err := c.cmdPWA([]string{"--bump"}); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(dir, "web/static/js/sw.js"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "CACHE_VERSION = 2") {
		t.Errorf("expected bumped CACHE_VERSION, got:\n%s", body)
	}
}
