package cli

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// #108: destroying a parent resource still referenced by a child removed
// models.Category/InsertCategory while SeedDemoBookmarks (and tests) kept
// calling InsertCategory — silent broken build after exit 0.
func scaffoldParentChildApp(t *testing.T, name string) string {
	t.Helper()
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), name)
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    name,
		ModulePath: "github.com/puppe1990/" + name,
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "category", resourceOpts{Fields: "name:string", Seed: false}); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "bookmark", resourceOpts{Fields: "title:string,category_id:references", Seed: false}); err != nil {
		t.Fatal(err)
	}
	return appDir
}

func TestDestroyResource_refusesParentStillReferenced(t *testing.T) {
	appDir := scaffoldParentChildApp(t, "destparent")

	err := destroyResource(appDir, "category", false, false)
	if err == nil {
		t.Fatal("destroy removed a parent still referenced by another resource")
	}
	if !strings.Contains(err.Error(), "bookmark") {
		t.Errorf("error should name the referencing resource, got: %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(appDir, "internal/models/category.go")); statErr != nil {
		t.Error("parent model removed despite the refusal")
	}
	routes := readDestroyFixture(t, filepath.Join(appDir, "internal/app/routes.go"))
	if !strings.Contains(routes, `"/admin/categories"`) {
		t.Error("parent routes removed despite the refusal")
	}
}

func TestDestroyResource_childFirstLeavesNoOrphans(t *testing.T) {
	appDir := scaffoldParentChildApp(t, "destchildfirst")

	if err := destroyResource(appDir, "bookmark", false, false); err != nil {
		t.Fatal(err)
	}
	store := readDestroyFixture(t, filepath.Join(appDir, "internal/store/store.go"))
	if strings.Contains(store, "ListCategoryOptions") {
		t.Error("orphan ListCategoryOptions survived destroying the only child")
	}
	if !strings.Contains(store, "InsertCategory") {
		t.Error("parent methods must survive a child destroy")
	}
	requireParses(t, "store.go", store)

	if err := destroyResource(appDir, "category", false, false); err != nil {
		t.Fatalf("destroying the parent after its child should work: %v", err)
	}
	store = readDestroyFixture(t, filepath.Join(appDir, "internal/store/store.go"))
	if strings.Contains(store, "Category") {
		t.Errorf("store.go still references Category:\n%s", store)
	}
	if strings.Contains(store, "models.") {
		t.Error("store.go should drop the unused models import")
	}
	requireParses(t, "store.go", store)

	storeTest := readDestroyFixture(t, filepath.Join(appDir, "internal/store/store_test.go"))
	if strings.Contains(storeTest, "InsertCategory") {
		t.Error("store_test.go still references InsertCategory")
	}
	requireParses(t, "store_test.go", storeTest)
}

// A child that is NOT being destroyed must keep its List<Ref>Options.
func TestDestroyResource_keepsOptionMethodsUsedByOtherChildren(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "desttwokids")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "desttwokids",
		ModulePath: "github.com/puppe1990/desttwokids",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "category", resourceOpts{Fields: "name:string", Seed: false}); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "bookmark", resourceOpts{Fields: "title:string,category_id:references", Seed: false}); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "note", resourceOpts{Fields: "body:string,category_id:references", Seed: false}); err != nil {
		t.Fatal(err)
	}

	if err := destroyResource(appDir, "bookmark", false, false); err != nil {
		t.Fatal(err)
	}
	store := readDestroyFixture(t, filepath.Join(appDir, "internal/store/store.go"))
	if !strings.Contains(store, "ListCategoryOptions") {
		t.Error("ListCategoryOptions still used by the note resource must survive")
	}
	requireParses(t, "store.go", store)
}

func readDestroyFixture(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(raw)
}

func requireParses(t *testing.T, name, src string) {
	t.Helper()
	if _, err := parser.ParseFile(token.NewFileSet(), name, src, parser.SkipObjectResolution); err != nil {
		t.Errorf("%s does not parse: %v", name, err)
	}
}
