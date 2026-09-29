package view

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/puppe1990/amarra-cais/pkg/cais"
)

func TestKit_selectStaysNativeAndMarksUserInvalid(t *testing.T) {
	fsys := fstest.MapFS{
		"layouts/app.html": &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
		"pages/home.html": &fstest.MapFile{Data: []byte(`{{ define "content" }}
<.select name="role" label="Role" error="pick one"><option>admin</option></.select>
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
		"<select",
		`name="role"`,
		"user-invalid:",
		`aria-invalid="true"`,
		"pick one",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("select missing %q in %s", want, body)
		}
	}
	if strings.Contains(body, "role=listbox") || strings.Contains(body, "cais-select-search") {
		t.Errorf("select must stay a native control: %s", body)
	}
}

func TestKit_inputKeepsServerErrorWithUserInvalid(t *testing.T) {
	fsys := fstest.MapFS{
		"layouts/app.html": &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
		"pages/home.html":  &fstest.MapFile{Data: []byte(`{{ define "content" }}<.input name="title" label="Title" error="required" />{{ end }}`)},
	}
	rec, err := Load(fsys, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	Write(rr, httptest.NewRequest(http.MethodGet, "/", nil), rec, Page{Layout: "app", Name: "home", Data: map[string]any{}}, cais.Config{})
	body := rr.Body.String()
	if !strings.Contains(body, "user-invalid:") || !strings.Contains(body, "required") || !strings.Contains(body, `aria-invalid="true"`) {
		t.Errorf("input should keep Go errors and user-invalid: %s", body)
	}
}
