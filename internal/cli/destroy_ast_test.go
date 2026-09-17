package cli

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestUnpatchResourceRoutes_keepsUserLinesMentioningAdminPath(t *testing.T) {
	data := dataForResource("bookmark")
	content := `package app

import (
	"github.com/puppe1990/amarra-cais/pkg/cais"
	"github.com/puppe1990/amarra-cais/pkg/cais/middleware"

	"example.com/app/internal/handlers"
)

func registerRoutes(r *cais.Router, deps Deps, cfg cais.Config) {
	r.Get("/", home.ServeHTTP)
	bookmarks := handlers.NewBookmarksHandler(deps.Store, deps.Site, deps.Inertia)
	r.Get("/bookmarks", bookmarks.List)
	adminBookmarks := handlers.NewAdminBookmarksHandler(deps.Store, deps.Site, deps.Inertia)
	r.Group(middleware.RequireAuth("/login"), func(g *cais.Router) {
		g.Get("/admin/bookmarks", adminBookmarks.Index)
		g.Post("/admin/bookmarks/{id}", cais.IntParam("id", adminBookmarks.Update))
	})
	// TODO(#42): move /admin/bookmarks export behind a queue
	exporter := handlers.NewExportHandler(deps.Store)
	r.Get("/admin/bookmarks-export", exporter.ServeHTTP)
}
`
	out, err := unpatchResourceRoutes(content, data)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"/admin/bookmarks-export",
		"// TODO(#42): move /admin/bookmarks export behind a queue",
		"NewExportHandler",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("user code removed: %q missing from output:\n%s", want, out)
		}
	}
	for _, gone := range []string{"r.Group(", "adminBookmarks := ", "g.Get(\"/admin/bookmarks\",", `"bookmarks", bookmarks.List`} {
		if strings.Contains(out, gone) {
			t.Errorf("generated statement survived: %q in output:\n%s", gone, out)
		}
	}
	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "routes.go", out, parser.SkipObjectResolution); err != nil {
		t.Errorf("output does not parse: %v\n%s", err, out)
	}
}

func TestRemoveStoreResourceMethods_exactNamesOnly(t *testing.T) {
	data := dataForResource("bookmark")
	content := `package store

import "database/sql"

type Store interface {
	Ping() error
	Close() error
	InsertBookmark(b models.Bookmark) (int64, error)
	countBookmarks() (int, error)
}

type SQLiteStore struct{ db *sql.DB }

func (s *SQLiteStore) InsertBookmark(b models.Bookmark) (int64, error) { return 1, nil }

func (s *SQLiteStore) countBookmarks() (int, error) { return 0, nil }

func (s *SQLiteStore) countBookmarksByUser(userID int64) (int, error) { return 0, nil }
`
	out, err := removeStoreResourceMethods(content, data)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "countBookmarksByUser") {
		t.Error("lookalike user method countBookmarksByUser must survive:\n" + out)
	}
	if strings.Contains(out, "countBookmarks()") || strings.Contains(out, "InsertBookmark") {
		t.Errorf("generated methods must be removed:\n%s", out)
	}
	if !strings.Contains(out, "Ping() error") || !strings.Contains(out, "Close() error") {
		t.Error("unrelated Store methods must survive:\n" + out)
	}
	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "store.go", out, parser.SkipObjectResolution); err != nil {
		t.Errorf("output does not parse: %v\n%s", err, out)
	}
}

func TestRemoveMethodsFromStoreInterface_onlyTouchesStoreIface(t *testing.T) {
	in := `package store

type Other interface {
	InsertTag(models.Tag) (int64, error)
}

type Store interface {
	InsertTag(models.Tag) (int64, error)
	ListAllTags() ([]models.Tag, error)
	Ping() error
	Close() error
}
`
	out, err := removeInterfaceMethods(in, "Store", map[string]bool{"InsertTag": true, "ListAllTags": true})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "type Store interface {\n\tInsertTag") {
		t.Error("InsertTag should be removed from Store interface")
	}
	if strings.Contains(out, "type Store interface {\n\tListAllTags") {
		t.Error("ListAllTags should be removed from Store interface")
	}
	if !strings.Contains(out, "type Other interface {\n\tInsertTag") {
		t.Error("Other interface methods should remain")
	}
	if !strings.Contains(out, "Ping() error") {
		t.Error("unrelated Store methods should remain")
	}
}

// #174: ignored os.ReadDir/read errors made partial destroys report success.
// Missing paths stay tolerated; real I/O errors must surface.
