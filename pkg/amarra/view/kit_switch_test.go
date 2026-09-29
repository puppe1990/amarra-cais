package view

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/puppe1990/amarra-cais/pkg/cais"
)

func TestKit_switchIsLabeledCheckbox(t *testing.T) {
	fsys := fstest.MapFS{
		"layouts/app.html": &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
		"pages/home.html": &fstest.MapFile{Data: []byte(`{{ define "content" }}
<.form action="/items" method="post">
  <.switch name="published" label="Publicado" checked />
</.form>
{{ end }}`)},
	}
	rec, err := Load(fsys, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	Write(rr, httptest.NewRequest(http.MethodGet, "/", nil), rec, Page{
		Layout: "app", Name: "home",
		Data: map[string]any{"CSRFToken": "tok"},
	}, cais.Config{})
	body := rr.Body.String()
	for _, want := range []string{
		`type="checkbox"`,
		`name="published"`,
		`id="published"`,
		`for="published"`,
		"Publicado",
		"checked",
		`name="csrf_token"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("switch missing %q in %s", want, body)
		}
	}
}
