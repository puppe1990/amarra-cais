package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// #107: references without the parent resource generated a broken app (seeds
// and tests calling s.InsertCategory, FK to a table that never exists) with
// exit 0. The generator must fail before writing anything.
func TestScaffoldResource_missingReferenceParentFailsEarly(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "refparent")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "refparent",
		ModulePath: "github.com/puppe1990/refparent",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	storeBefore := mustReadFile(t, filepath.Join(appDir, "internal/store/store.go"))

	err := scaffoldResource(appDir, "bookmark", resourceOpts{
		Fields: "title:string,category_id:references",
		Seed:   false,
	})
	if err == nil {
		t.Fatal("scaffoldResource accepted a reference without its parent resource")
	}
	if !strings.Contains(err.Error(), "amarra-cais g resource category") {
		t.Errorf("error should point at the parent generator command, got: %v", err)
	}

	if _, statErr := os.Stat(filepath.Join(appDir, "internal/models/bookmark.go")); !os.IsNotExist(statErr) {
		t.Error("generator wrote bookmark.go before validating the parent")
	}
	if got := mustReadFile(t, filepath.Join(appDir, "internal/store/store.go")); got != storeBefore {
		t.Error("generator patched store.go before validating the parent")
	}
	entries, readErr := os.ReadDir(filepath.Join(appDir, "internal/store/migrations"))
	if readErr != nil {
		t.Fatal(readErr)
	}
	for _, e := range entries {
		if strings.Contains(e.Name(), "_bookmarks.sql") {
			t.Errorf("generator wrote migration %s before validating the parent", e.Name())
		}
	}
}

func TestScaffoldResource_belongsToMissingParentFailsEarly(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "refbelongs")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "refbelongs",
		ModulePath: "github.com/puppe1990/refbelongs",
	}, true, false); err != nil {
		t.Fatal(err)
	}

	err := scaffoldResource(appDir, "bookmark", resourceOpts{
		Fields: "title:string,category:belongs_to",
		Seed:   false,
	})
	if err == nil {
		t.Fatal("scaffoldResource accepted belongs_to without its parent resource")
	}
	if !strings.Contains(err.Error(), "amarra-cais g resource category") {
		t.Errorf("error should point at the parent generator command, got: %v", err)
	}
}

func TestScaffoldResource_referenceParentPresentSucceeds(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "refok")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "refok",
		ModulePath: "github.com/puppe1990/refok",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "category", resourceOpts{Fields: "name:string", Seed: false}); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "bookmark", resourceOpts{
		Fields: "title:string,category_id:references",
		Seed:   false,
	}); err != nil {
		t.Fatalf("scaffoldResource with parent present: %v", err)
	}
}

func mustReadFile(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(raw)
}
