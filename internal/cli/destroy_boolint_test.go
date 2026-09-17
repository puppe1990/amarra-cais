package cli

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// #105: the boolInt regex stopped at the first closing brace (the inner if),
// leaving an orphan `return 0` in store.go after destroy/--force.
func TestRemoveStoreResourceMethods_removesWholeBoolIntHelper(t *testing.T) {
	data := dataForResource("flag")
	content := `package store

import "app/internal/models"

func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func (s *SQLiteStore) InsertFlag(f models.Flag) (int64, error) { return boolInt(f.Active), nil }
`
	out, err := removeStoreResourceMethods(content, data)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "boolInt") {
		t.Errorf("boolInt helper survived by name:\n%s", out)
	}
	if strings.Contains(out, "return 0") {
		t.Errorf("boolInt body survived as orphan statements:\n%s", out)
	}
	if _, err := parser.ParseFile(token.NewFileSet(), "store.go", out, parser.SkipObjectResolution); err != nil {
		t.Errorf("output does not parse: %v\n%s", err, out)
	}
}

func TestDestroyResource_boolOnlyKeepsStoreParseable(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "boolflag")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "boolflag",
		ModulePath: "github.com/puppe1990/boolflag",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "flag", resourceOpts{Fields: "title:string,active:bool", Seed: false}); err != nil {
		t.Fatal(err)
	}

	storePath := filepath.Join(appDir, "internal/store/store.go")
	patched, err := os.ReadFile(storePath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(patched), "boolInt") {
		t.Fatal("scaffold should have added the boolInt helper for a bool field")
	}

	if err := destroyResource(appDir, "flag", false, false); err != nil {
		t.Fatal(err)
	}

	after, err := os.ReadFile(storePath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := parser.ParseFile(token.NewFileSet(), "store.go", after, parser.SkipObjectResolution); err != nil {
		t.Errorf("store.go does not parse after destroy: %v\n%s", err, after)
	}
}
