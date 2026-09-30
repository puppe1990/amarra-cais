package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScaffoldResource_adminUpdateAndDeleteUseHTTPVerbs(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "restmark")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "restmark",
		ModulePath: "github.com/puppe1990/restmark",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "bookmark", resourceOpts{
		Fields: "title:string,url:url",
	}); err != nil {
		t.Fatal(err)
	}

	routes := mustReadFile(t, filepath.Join(appDir, "internal/app/routes.go"))
	if !strings.Contains(routes, `g.Put("/admin/bookmarks/{id}"`) {
		t.Error("admin update must be PUT (#269)")
	}
	if !strings.Contains(routes, `g.Delete("/admin/bookmarks/{id}"`) {
		t.Error("admin destroy must be DELETE on the resource URL (#269)")
	}
	if strings.Contains(routes, `g.Post("/admin/bookmarks/{id}"`) {
		t.Error("admin update must not stay POST (#269)")
	}
	if strings.Contains(routes, `/{id}/delete"`) {
		t.Error("admin destroy must not use a /delete suffix (#269)")
	}
	if !strings.Contains(routes, `g.Post("/admin/bookmarks"`) {
		t.Error("admin create stays POST")
	}
	if !strings.Contains(routes, `g.Post("/admin/bookmarks/bulk-delete"`) {
		t.Error("bulk-delete stays POST")
	}

	form := mustReadFile(t, filepath.Join(appDir, "web/templates/pages/admin_bookmark_form.html"))
	if !strings.Contains(form, `name="_method" value="put"`) {
		t.Error("edit form must send _method=put so Drive issues PUT (#269)")
	}

	index := mustReadFile(t, filepath.Join(appDir, "web/templates/pages/admin_bookmarks.html"))
	if !strings.Contains(index, `(dict "method" "delete")`) {
		t.Error("admin index delete link must use method delete (#269)")
	}
	if strings.Contains(index, `/admin/bookmarks/%d/delete`) {
		t.Error("admin index delete link must target the resource URL (#269)")
	}

	show, err := os.ReadFile(filepath.Join(appDir, "web/templates/pages/admin_bookmark_show.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(show), `(dict "method" "delete")`) {
		t.Error("admin show delete link must use method delete (#269)")
	}

	handlerTest := mustReadFile(t, filepath.Join(appDir, "internal/handlers/admin_bookmarks_test.go"))
	if !strings.Contains(handlerTest, `http.MethodDelete`) {
		t.Error("generated Delete test must call MethodDelete (#269)")
	}
	if strings.Contains(handlerTest, `/admin/bookmarks/1/delete`) {
		t.Error("generated Delete test must hit the resource URL (#269)")
	}
}
