package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// #245: `destroy model contact` removed the model and the migration but left
// store.go, seeds.go, dashboard.go and store_test.go pointing at the deleted
// type — `go build` broke with no hint about the command that caused it.
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

func TestDestroyModel_refusesWhenHandlerStillCallsMethod(t *testing.T) {
	appDir := scaffoldDestroyModelProbe(t, "destguarded")
	if err := destroyHandler(appDir, "contact", false, false); err != nil {
		t.Fatal(err)
	}
	storePath := filepath.Join(appDir, "internal/store/store.go")
	before := readDestroyFixture(t, storePath)

	err := destroyModel(appDir, "contact", false, false)
	if err == nil {
		t.Fatal("destroy model removed a model the dashboard handler still uses")
	}
	msg := err.Error()
	for _, want := range []string{
		"internal/handlers/dashboard.go",
		"CountContacts",
		"amarra-cais destroy handler dashboard",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("error %q missing %q", msg, want)
		}
	}
	if _, statErr := os.Stat(filepath.Join(appDir, "internal/models/contact.go")); statErr != nil {
		t.Error("model removed despite the refusal")
	}
	if readDestroyFixture(t, storePath) != before {
		t.Error("store.go changed despite the refusal")
	}
}

func TestDestroyModel_clearsStoreSeedsAndTests(t *testing.T) {
	appDir := scaffoldDestroyModelProbe(t, "destclean")
	for _, handler := range []string{"contact", "dashboard"} {
		if err := destroyHandler(appDir, handler, false, false); err != nil {
			t.Fatal(err)
		}
	}
	if err := destroyModel(appDir, "contact", false, false); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(appDir, "internal/models/contact.go")); !os.IsNotExist(err) {
		t.Error("model file survived")
	}
	entries, err := os.ReadDir(filepath.Join(appDir, "internal/store/migrations"))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.Contains(e.Name(), "_contacts.sql") {
			t.Errorf("migration survived: %s", e.Name())
		}
	}

	store := readDestroyFixture(t, filepath.Join(appDir, "internal/store/store.go"))
	if strings.Contains(store, "Contact") || strings.Contains(store, "contacts") {
		t.Errorf("store.go still references the removed model:\n%s", store)
	}
	requireParses(t, "store.go", store)

	seeds := readDestroyFixture(t, filepath.Join(appDir, "internal/db/seeds.go"))
	if strings.Contains(seeds, "Contact") || strings.Contains(seeds, "/internal/models") {
		t.Errorf("seeds.go still references the removed model:\n%s", seeds)
	}
	requireParses(t, "seeds.go", seeds)

	storeTest := readDestroyFixture(t, filepath.Join(appDir, "internal/store/store_test.go"))
	if strings.Contains(storeTest, "Contact") || strings.Contains(storeTest, "contacts") || strings.Contains(storeTest, "/internal/models") {
		t.Errorf("store_test.go still references the removed model:\n%s", storeTest)
	}
	requireParses(t, "store_test.go", storeTest)
}

func TestDestroyModel_dryRunReportsUnpatchWithoutWriting(t *testing.T) {
	appDir := scaffoldDestroyModelProbe(t, "destmodeldryref")
	for _, handler := range []string{"contact", "dashboard"} {
		if err := destroyHandler(appDir, handler, false, false); err != nil {
			t.Fatal(err)
		}
	}
	storePath := filepath.Join(appDir, "internal/store/store.go")
	seedsPath := filepath.Join(appDir, "internal/db/seeds.go")
	storeBefore := readDestroyFixture(t, storePath)
	seedsBefore := readDestroyFixture(t, seedsPath)

	if err := destroyModel(appDir, "contact", true, false); err != nil {
		t.Fatal(err)
	}
	if readDestroyFixture(t, storePath) != storeBefore {
		t.Error("dry-run changed store.go")
	}
	if readDestroyFixture(t, seedsPath) != seedsBefore {
		t.Error("dry-run changed seeds.go")
	}
}
