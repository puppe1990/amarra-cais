package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// #24: generated resources must use the shipped kit (<.filters>, <.table>,
// <.pagination>, <.empty>, <.checkbox>) instead of one-off Tailwind copies.
func TestScaffoldResource_AdminIndexUsesKit(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "kitmark")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "kitmark",
		ModulePath: "github.com/puppe1990/kitmark",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "bookmark", resourceOpts{
		Fields:   "title:string,url:url,read:bool",
		Paginate: true,
	}); err != nil {
		t.Fatal(err)
	}

	partial, err := os.ReadFile(filepath.Join(appDir, "web/templates/partials/admin_bookmarks_index.html"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(partial)
	for _, tag := range []string{`<.filters`, `<.table`, `<.empty`, `<.pagination`, `frame="admin-bookmarks"`} {
		if !strings.Contains(body, tag) {
			t.Errorf("admin index missing %s (#24)", tag)
		}
	}
	for _, banned := range []string{"bg-slate-50", `<table`} {
		if strings.Contains(body, banned) {
			t.Errorf("admin index should not hand-roll %q anymore (#24)", banned)
		}
	}
	if !strings.Contains(body, `name="q"`) {
		t.Error("admin index filters should search q")
	}
	if !strings.Contains(body, `(dict "method" "post" "confirm"`) {
		t.Error("admin index delete should be a confirmed Drive link")
	}

	form, err := os.ReadFile(filepath.Join(appDir, "web/templates/pages/admin_bookmark_form.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(form), `<.checkbox`) {
		t.Errorf("admin form should use the <.checkbox> kit (#24): %s", form)
	}
	if strings.Contains(string(form), `type="checkbox"`) {
		t.Error("admin form should not hand-roll checkbox markup")
	}

	handler, err := os.ReadFile(filepath.Join(appDir, "internal/handlers/admin_bookmarks.go"))
	if err != nil {
		t.Fatal(err)
	}
	handlerBody := string(handler)
	for _, needle := range []string{`Query().Get("q")`, `Query().Get("sort")`, `Query().Get("dir")`, `Cols`, `"Base":`} {
		if !strings.Contains(handlerBody, needle) {
			t.Errorf("admin handler missing %s for kit index (#24)", needle)
		}
	}
}

func TestScaffoldResource_PublicIndexUsesKit(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "kitpub")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "kitpub",
		ModulePath: "github.com/puppe1990/kitpub",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "bookmark", resourceOpts{
		Fields:   "title:string,url:url",
		Public:   true,
		Paginate: true,
	}); err != nil {
		t.Fatal(err)
	}
	partial, err := os.ReadFile(filepath.Join(appDir, "web/templates/partials/bookmarks_list.html"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(partial)
	for _, tag := range []string{`<.filters`, `<.empty`, `<.pagination`} {
		if !strings.Contains(body, tag) {
			t.Errorf("public index missing %s (#24)", tag)
		}
	}
	handler, err := os.ReadFile(filepath.Join(appDir, "internal/handlers/bookmarks.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(handler), `Query().Get("q")`) {
		t.Error("public handler should read q for the filters form (#24)")
	}
}
