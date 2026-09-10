package view

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/puppe1990/amarra-cais/pkg/cais"
)

func TestKit_tableSortPreservesQueryResetsPage(t *testing.T) {
	fsys := fstest.MapFS{
		"layouts/app.html": &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
		"pages/home.html": &fstest.MapFile{Data: []byte(
			`{{ define "content" }}<.table cols="{{ .Cols }}" sort="name" dir="asc" base="/admin/items?q=copper&page=2"></.table>{{ end }}`,
		)},
	}
	rec, err := Load(fsys, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	Write(rr, httptest.NewRequest(http.MethodGet, "/", nil), rec, Page{
		Layout: "app", Name: "home", Data: map[string]any{"Cols": tableCols()},
	}, cais.Config{})
	body := rr.Body.String()
	if !strings.Contains(body, "q=copper") {
		t.Errorf("sort link must keep the search query: %s", body)
	}
	if strings.Contains(body, "page=") {
		t.Errorf("sort must reset page (omit page=): %s", body)
	}
	if !strings.Contains(body, "/admin/items?") {
		t.Errorf("sort link must keep the index path: %s", body)
	}
}
