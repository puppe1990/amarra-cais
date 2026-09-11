package view

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/puppe1990/amarra-cais/pkg/cais"
)

func TestKit_fileInputMultipartForm(t *testing.T) {
	fsys := fstest.MapFS{
		"layouts/app.html": &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
		"pages/home.html": &fstest.MapFile{Data: []byte(
			`{{ define "content" }}<.form action="/upload" method="post" enctype="multipart/form-data"><.input type="file" name="logo" label="Logo" accept="image/*" error="too large" /></.form>{{ end }}`,
		)},
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
		`enctype="multipart/form-data"`,
		`type="file"`,
		`name="logo"`,
		`accept="image/*"`,
		`aria-invalid="true"`,
		`name="csrf_token"`,
		"too large",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("file form missing %q in %s", want, body)
		}
	}
	if strings.Contains(body, `value="`) && strings.Contains(body, `type="file"`) {
		// file inputs must not copy a value attribute (browsers ignore it; it leaks paths)
		if strings.Contains(body, `type="file"`) && strings.Contains(body, `value="`) {
			t.Errorf("file input should omit value=: %s", body)
		}
	}
}
