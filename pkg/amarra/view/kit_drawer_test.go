package view

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/puppe1990/amarra-cais/pkg/cais"
)

func TestKit_drawerIsSideDialog(t *testing.T) {
	fsys := fstest.MapFS{
		"layouts/app.html": &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
		"pages/home.html": &fstest.MapFile{Data: []byte(`{{ define "content" }}
<.drawer id="filters-drawer" title="Filtros">
  <.filters action="/admin/items" frame="items-table">q</.filters>
</.drawer>
{{ end }}`)},
	}
	rec, err := Load(fsys, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	Write(rr, httptest.NewRequest(http.MethodGet, "/", nil), rec, Page{Layout: "app", Name: "home", Data: map[string]any{}}, cais.Config{})
	body := rr.Body.String()
	for _, want := range []string{
		"<dialog",
		`data-amarra-dialog-target`,
		`id="filters-drawer"`,
		"ml-auto",
		"h-full",
		"max-w-sm",
		"Filtros",
		`action="/admin/items"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("drawer missing %q in %s", want, body)
		}
	}
}
