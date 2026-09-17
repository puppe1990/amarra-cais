package cli

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// #104: patch guards used substring matching, so "InsertPostComment" made the
// generator skip patching store.go for a later "post" resource (exit 0, app
// that does not compile).
func TestScaffoldResource_prefixCollisionPatchesEverything(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "prefixapp")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "prefixapp",
		ModulePath: "github.com/puppe1990/prefixapp",
	}, true, false); err != nil {
		t.Fatal(err)
	}

	if err := scaffoldResource(appDir, "post_comment", resourceOpts{Fields: "body:string", Seed: false}); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "post", resourceOpts{Fields: "title:string", Seed: false}); err != nil {
		t.Fatal(err)
	}

	store := readFileString(t, filepath.Join(appDir, "internal/store/store.go"))
	for _, want := range []string{
		"func (s *SQLiteStore) InsertPost(",
		"InsertPost(models.Post)",
		"func (s *SQLiteStore) ListAllPosts(",
	} {
		if !strings.Contains(store, want) {
			t.Errorf("store.go missing %q after prefix-colliding resource", want)
		}
	}
	assertParses(t, "store.go", store)

	storeTest := readFileString(t, filepath.Join(appDir, "internal/store/store_test.go"))
	if !strings.Contains(storeTest, "func TestStore_InsertPost(") {
		t.Error("store_test.go missing TestStore_InsertPost")
	}
	assertParses(t, "store_test.go", storeTest)

	routes := readFileString(t, filepath.Join(appDir, "internal/app/routes.go"))
	if !strings.Contains(routes, `"/admin/posts"`) {
		t.Error("routes.go missing the post admin routes")
	}
	assertParses(t, "routes.go", routes)
}

// Same collision class on plural names: generate the longer name first, then
// "tag" — its guards see "/admin/tags"/"SeedDemoTags" inside the archive paths
// and skip the routes/seeds/main patches.
func TestScaffoldResource_pluralPrefixCollisionPatchesRoutesAndSeeds(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "pluralprefix")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "pluralprefix",
		ModulePath: "github.com/puppe1990/pluralprefix",
	}, true, false); err != nil {
		t.Fatal(err)
	}

	if err := scaffoldResource(appDir, "tags_archive", resourceOpts{Fields: "title:string", Seed: true}); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "tag", resourceOpts{Fields: "name:string", Seed: true}); err != nil {
		t.Fatal(err)
	}

	routes := readFileString(t, filepath.Join(appDir, "internal/app/routes.go"))
	for _, want := range []string{`"/admin/tags"`, `"/admin/tags_archives"`, "adminTags := "} {
		if !strings.Contains(routes, want) {
			t.Errorf("routes.go missing %q", want)
		}
	}
	assertParses(t, "routes.go", routes)

	seeds := readFileString(t, filepath.Join(appDir, "internal/db/seeds.go"))
	if !strings.Contains(seeds, "s.SeedDemoTags()") {
		t.Error("seeds.go missing the tag seed call")
	}
	main := readFileString(t, filepath.Join(appDir, "cmd/server/main.go"))
	if !strings.Contains(main, "s.SeedDemoTags()") {
		t.Error("main.go missing the tag seed call")
	}
}

func readFileString(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(raw)
}

func assertParses(t *testing.T, name, src string) {
	t.Helper()
	if _, err := parser.ParseFile(token.NewFileSet(), name, src, parser.SkipObjectResolution); err != nil {
		t.Errorf("%s does not parse: %v", name, err)
	}
}
