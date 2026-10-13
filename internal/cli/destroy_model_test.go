package cli

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// #245: `destroy model` removed the model and the migration but left store.go,
// seeds.go, the handlers and store_test.go pointing at the deleted type —
// `go build` broke with no hint about the command that caused it.
func scaffoldDestroyModelProbe(t *testing.T, name string) string {
	t.Helper()
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), name)
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    name,
		ModulePath: "example.com/" + name,
	}, false, false); err != nil {
		t.Fatal(err)
	}
	return appDir
}

// destroyModelProbe adds a generated resource: the scaffold ships no example
// model, so the model under destroy is one the app created with `g resource`.
func destroyModelProbe(t *testing.T, name string) string {
	t.Helper()
	appDir := scaffoldDestroyModelProbe(t, name)
	t.Chdir(appDir)
	if err := (&CLI{Out: io.Discard}).Run([]string{"g", "resource", "bookmark", "--fields", "title:string"}); err != nil {
		t.Fatal(err)
	}
	return appDir
}

func TestDestroyModel_refusesWhenHandlerStillCallsMethod(t *testing.T) {
	appDir := destroyModelProbe(t, "destguarded")
	storePath := filepath.Join(appDir, "internal/store/store.go")
	before := readDestroyFixture(t, storePath)

	err := destroyModel(appDir, "bookmark", false, false)
	if err == nil {
		t.Fatal("destroy model removed a model the admin handler still uses")
	}
	msg := err.Error()
	for _, want := range []string{
		"internal/handlers/admin_bookmarks.go",
		"Bookmark",
		"pass --force",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("error %q missing %q", msg, want)
		}
	}
	if _, statErr := os.Stat(filepath.Join(appDir, "internal/models/bookmark.go")); statErr != nil {
		t.Error("model removed despite the refusal")
	}
	if readDestroyFixture(t, storePath) != before {
		t.Error("store.go changed despite the refusal")
	}
}

func TestDestroyModel_clearsStoreSeedsAndTests(t *testing.T) {
	appDir := destroyModelProbe(t, "destclean")
	// --force: the generated handlers still call the methods (that guard has its
	// own test above); this one covers the store/seeds/test cleanup.
	if err := destroyModel(appDir, "bookmark", false, true); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(appDir, "internal/models/bookmark.go")); !os.IsNotExist(err) {
		t.Error("model file survived")
	}
	entries, err := os.ReadDir(filepath.Join(appDir, "internal/store/migrations"))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.Contains(e.Name(), "_bookmarks.sql") {
			t.Errorf("migration survived: %s", e.Name())
		}
	}

	store := readDestroyFixture(t, filepath.Join(appDir, "internal/store/store.go"))
	if strings.Contains(store, "Bookmark") || strings.Contains(store, "bookmarks") {
		t.Errorf("store.go still references the removed model:\n%s", store)
	}
	requireParses(t, "store.go", store)

	seeds := readDestroyFixture(t, filepath.Join(appDir, "internal/db/seeds.go"))
	if strings.Contains(seeds, "Bookmark") || strings.Contains(seeds, "/internal/models") {
		t.Errorf("seeds.go still references the removed model:\n%s", seeds)
	}
	requireParses(t, "seeds.go", seeds)

	storeTest := readDestroyFixture(t, filepath.Join(appDir, "internal/store/store_test.go"))
	if strings.Contains(storeTest, "Bookmark") || strings.Contains(storeTest, "bookmarks") || strings.Contains(storeTest, "/internal/models") {
		t.Errorf("store_test.go still references the removed model:\n%s", storeTest)
	}
	requireParses(t, "store_test.go", storeTest)
}

func TestDestroyModel_dryRunReportsUnpatchWithoutWriting(t *testing.T) {
	appDir := destroyModelProbe(t, "destmodeldryref")
	storePath := filepath.Join(appDir, "internal/store/store.go")
	seedsPath := filepath.Join(appDir, "internal/db/seeds.go")
	storeBefore := readDestroyFixture(t, storePath)
	seedsBefore := readDestroyFixture(t, seedsPath)

	if err := destroyModel(appDir, "bookmark", true, true); err != nil {
		t.Fatal(err)
	}
	if readDestroyFixture(t, storePath) != storeBefore {
		t.Error("dry-run changed store.go")
	}
	if readDestroyFixture(t, seedsPath) != seedsBefore {
		t.Error("dry-run changed seeds.go")
	}
}
