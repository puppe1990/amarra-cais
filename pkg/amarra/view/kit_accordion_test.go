package view

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/puppe1990/amarra-cais/pkg/cais"
)

func TestKit_accordionCollapseUsesDetails(t *testing.T) {
	fsys := fstest.MapFS{
		"layouts/app.html": &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
		"pages/home.html": &fstest.MapFile{Data: []byte(`{{ define "content" }}
<.accordion>
  <.collapse title="Dados fiscais" open>
    CNPJ
  </.collapse>
  <.collapse title="Endereço" name="addr">
    Rua
  </.collapse>
</.accordion>
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
		"Dados fiscais",
		"CNPJ",
		"Endereço",
		"Rua",
		` open`,
		`name="addr"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("accordion/collapse missing %q in %s", want, body)
		}
	}
	if strings.Contains(body, `type="checkbox"`) {
		t.Errorf("must not use checkbox hack: %s", body)
	}
}
