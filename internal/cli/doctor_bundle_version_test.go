package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeBundleFixture(t *testing.T, dir, gomod, bundle string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, "web/static/js"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(gomod), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "web/static/js/amarra.js"), []byte(bundle), 0o644); err != nil {
		t.Fatal(err)
	}
}

// #320: vendored amarra.js sem marcador de versão deixa o doctor cego a
// bundles defasados do go.mod.
func TestCheckBundleVersion_warnsWhenOutdated(t *testing.T) {
	dir := t.TempDir()
	writeBundleFixture(t, dir,
		"module app\n\nrequire github.com/puppe1990/amarra-cais v0.14.1\n",
		"/* amarra-cais v0.8.1 */\n(() => {})();\n")
	c := checkBundleVersion(dir)
	if c.OK {
		t.Fatal("bundle defasado deve avisar")
	}
	if !strings.Contains(c.FixHint, "pwa") {
		t.Errorf("FixHint = %q", c.FixHint)
	}
}

func TestCheckBundleVersion_okWhenMatching(t *testing.T) {
	dir := t.TempDir()
	writeBundleFixture(t, dir,
		"module app\n\nrequire github.com/puppe1990/amarra-cais v0.14.1\n",
		"/* amarra-cais v0.14.1 */\n(() => {})();\n")
	if c := checkBundleVersion(dir); !c.OK {
		t.Fatalf("bundle alinhado deve passar, got %+v", c)
	}
}

func TestCheckBundleVersion_warnsWhenMarkerMissing(t *testing.T) {
	dir := t.TempDir()
	writeBundleFixture(t, dir,
		"module app\n\nrequire github.com/puppe1990/amarra-cais v0.14.1\n",
		"(() => {})();\n")
	if c := checkBundleVersion(dir); c.OK {
		t.Fatal("bundle sem marcador deve avisar")
	}
}
