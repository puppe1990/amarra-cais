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
	fileAt := strings.Index(body, `type="file"`)
	if fileAt < 0 {
		t.Fatal("missing type=file")
	}
	start := strings.LastIndex(body[:fileAt], "<input")
	endRel := strings.Index(body[fileAt:], ">")
	if start < 0 || endRel < 0 {
		t.Fatalf("could not isolate file input tag: %s", body)
	}
	tag := body[start : fileAt+endRel+1]
	if strings.Contains(tag, "value=") {
		t.Errorf("file input should omit value=: %s", tag)
	}
}

func TestKit_formInsideRangeDoesNotReadEnctypeFromRow(t *testing.T) {
	type row struct{ ID int64 }
	fsys := fstest.MapFS{
		"layouts/app.html": &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
		"pages/home.html": &fstest.MapFile{Data: []byte(
			`{{ define "content" }}{{ range .Items }}<.form action="/toggle" method="post"><button type="submit">Go</button></.form>{{ end }}{{ end }}`,
		)},
	}
	rec, err := Load(fsys, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	Write(rr, httptest.NewRequest(http.MethodGet, "/", nil), rec, Page{
		Layout: "app", Name: "home",
		Data: map[string]any{"Items": []row{{ID: 1}}, "CSRFToken": "tok"},
	}, cais.Config{})
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), `action="/toggle"`) {
		t.Errorf("form missing: %s", rr.Body.String())
	}
}
