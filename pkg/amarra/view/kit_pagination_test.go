package view

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/puppe1990/amarra-cais/pkg/cais"
)

func TestKit_paginationBaseWithoutQueryUsesQuestionMark(t *testing.T) {
	fsys := fstest.MapFS{
		"layouts/app.html": &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
		"pages/home.html":  &fstest.MapFile{Data: []byte(`{{ define "content" }}<.pagination base="/admin/items" />{{ end }}`)},
	}
	rec, err := Load(fsys, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	Write(rr, httptest.NewRequest(http.MethodGet, "/", nil), rec, Page{
		Layout: "app", Name: "home",
		Data: map[string]any{"Page": 1, "HasNext": true, "NextPage": 2},
	}, cais.Config{})
	body := rr.Body.String()
	if strings.Contains(body, `/admin/items&amp;page=`) || strings.Contains(body, `/admin/items&page=`) {
		t.Errorf("bare Base must not join page with &: %s", body)
	}
	if !strings.Contains(body, `href="/admin/items?page=2"`) {
		t.Errorf("bare Base should use ?page=: %s", body)
	}
}
