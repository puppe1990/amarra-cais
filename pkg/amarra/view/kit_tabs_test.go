package view

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/puppe1990/amarra-cais/pkg/cais"
)

func TestKit_tabsUseExclusiveDetails(t *testing.T) {
	fsys := fstest.MapFS{
		"layouts/app.html": &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
		"pages/home.html": &fstest.MapFile{Data: []byte(`{{ define "content" }}
<.tabs>
  <.tab title="Detalhes" open>
    Body A
  </.tab>
  <.tab title="Histórico">
    Body B
  </.tab>
</.tabs>
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
		"<details",
		"<summary",
		"Detalhes",
		"Histórico",
		"Body A",
		"Body B",
		` open`,
		`name="`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("tabs missing %q in %s", want, body)
		}
	}
	if strings.Contains(body, `type="radio"`) || strings.Contains(body, `type="checkbox"`) {
		t.Errorf("must not use radio/checkbox hack: %s", body)
	}
}
