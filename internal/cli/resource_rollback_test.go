package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newRollbackApp(t *testing.T, name string) string {
	t.Helper()
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), name)
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    name,
		ModulePath: "github.com/puppe1990/" + name,
	}, true, false); err != nil {
		t.Fatal(err)
	}
	return appDir
}

// #130: the store interface insert failed loudly on a missing marker, but the
// implementation insert was a silent strings.Replace — the generator reported
// success with an interface no struct implements.
func TestScaffoldResource_missingStoreImplMarkerFailsWithoutMutation(t *testing.T) {
	appDir := newRollbackApp(t, "implmarker")
	storePath := filepath.Join(appDir, "internal/store/store.go")
	body := mustReadFile(t, storePath)
	broken := strings.Replace(body, "func (s *SQLiteStore) Close() error {", "func (s *SQLiteStore) closeDisabled() error {", 1)
	if broken == body {
		t.Fatal("test setup: Close() implementation not found")
	}
	if err := os.WriteFile(storePath, []byte(broken), 0o644); err != nil {
		t.Fatal(err)
	}

	err := scaffoldResource(appDir, "widget", resourceOpts{Fields: "title:string", Seed: false})
	if err == nil {
		t.Fatal("generator reported success without patching the store implementation")
	}
	if got := mustReadFile(t, storePath); got != broken {
		t.Error("store.go was mutated despite the failed patch")
	}
	if _, statErr := os.Stat(filepath.Join(appDir, "internal/models/widget.go")); !os.IsNotExist(statErr) {
		t.Error("generated files should be rolled back on failure")
	}
}

// Same class: a marker missing from a shared file must fail before any file is
// written, not leave a half-patched tree.
func TestScaffoldResource_missingRoutesMarkerFailsBeforeWriting(t *testing.T) {
	appDir := newRollbackApp(t, "routemarker")
	routesPath := filepath.Join(appDir, "internal/app/routes.go")
	body := mustReadFile(t, routesPath)
	broken := strings.Replace(body, "func registerRoutes", "func registerRoutesBroken", 1)
	if broken == body {
		t.Fatal("test setup: registerRoutes not found")
	}
	if err := os.WriteFile(routesPath, []byte(broken), 0o644); err != nil {
		t.Fatal(err)
	}
	storeBefore := mustReadFile(t, filepath.Join(appDir, "internal/store/store.go"))

	err := scaffoldResource(appDir, "widget", resourceOpts{Fields: "title:string", Seed: false})
	if err == nil {
		t.Fatal("generator accepted a routes.go without registerRoutes")
	}
	if _, statErr := os.Stat(filepath.Join(appDir, "internal/models/widget.go")); !os.IsNotExist(statErr) {
		t.Error("model written before patch preconditions were validated")
	}
	if got := mustReadFile(t, filepath.Join(appDir, "internal/store/store.go")); got != storeBefore {
		t.Error("store.go patched before patch preconditions were validated")
	}
}
