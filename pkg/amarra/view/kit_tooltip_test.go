package view

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/puppe1990/amarra-cais/pkg/cais"
)

func TestKit_tooltipTextIsInHTML(t *testing.T) {
	fsys := fstest.MapFS{
		"layouts/app.html": &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
		"pages/home.html": &fstest.MapFile{Data: []byte(`{{ define "content" }}
<.tooltip text="Copiado para a área de transferência">
  <button type="button">?</button>
</.tooltip>
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
		"Copiado para a área de transferência",
		`role="tooltip"`,
		"<button",
		"group-hover:visible",
		"group-focus-within:visible",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("tooltip missing %q in %s", want, body)
		}
	}
	if strings.Contains(body, `title="Copiado`) {
		t.Errorf("tooltip must not rely on title only: %s", body)
	}
}
