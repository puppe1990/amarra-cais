package view

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/puppe1990/amarra-cais/pkg/amarra"
	"github.com/puppe1990/amarra-cais/pkg/cais"
)

func testFS() fs.FS {
	return fstest.MapFS{
		"layouts/app.html": &fstest.MapFile{Data: []byte(`{{ define "app" }}<!doctype html><html><body><nav id="amarra-nav">n</nav><main id="amarra-main">{{ template "content" . }}</main></body></html>{{ end }}`)},
		// html/template rejects nested {{ define }} inside content; frames are sibling defines.
		"pages/home.html": &fstest.MapFile{Data: []byte(`{{ define "content" }}<h1>{{ .Title }}</h1>{{ end }}{{ define "frame:box" }}<section id="box">{{ .Title }}</section>{{ end }}`)},
	}
}

func TestWrite_fullPage(t *testing.T) {
	rec, err := Load(testFS(), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	Write(rr, req, rec, Page{Layout: "app", Name: "home", Data: map[string]string{"Title": "Hi"}}, cais.Config{})
	body := rr.Body.String()
	if rr.Code != 200 {
		t.Fatalf("code %d", rr.Code)
	}
	if !strings.Contains(body, `id="amarra-main"`) || !strings.Contains(body, "<h1>Hi</h1>") {
		t.Fatalf("body %q", body)
	}
}

func TestWrite_driveSelectsMain(t *testing.T) {
	rec, err := Load(testFS(), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(amarra.HeaderDrive, "true")
	Write(rr, req, rec, Page{Layout: "app", Name: "home", Data: map[string]string{"Title": "Hi"}}, cais.Config{})
	body := rr.Body.String()
	if !strings.Contains(body, `id="amarra-main"`) || !strings.Contains(body, "<h1>Hi</h1>") {
		t.Fatalf("body %q", body)
	}
}

func TestWrite_frame(t *testing.T) {
	rec, err := Load(testFS(), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(amarra.HeaderFrame, "box")
	Write(rr, req, rec, Page{Layout: "app", Name: "home", Frame: "box", Data: map[string]string{"Title": "Hi"}}, cais.Config{})
	body := rr.Body.String()
	if strings.Contains(body, `id="amarra-nav"`) {
		t.Fatalf("frame leaked layout: %q", body)
	}
	if !strings.Contains(body, `<section id="box">Hi</section>`) {
		t.Fatalf("body %q", body)
	}
}

func TestLoad_expandsComponents(t *testing.T) {
	fsys := fstest.MapFS{
		"layouts/app.html":       &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
		"pages/home.html":        &fstest.MapFile{Data: []byte(`{{ define "content" }}<.button type="submit">Go</.button>{{ end }}`)},
		"components/button.html": &fstest.MapFile{Data: []byte(`<button type="{{ .Type }}">{{ .Inner }}</button>`)},
	}
	rec, err := Load(fsys, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	Write(rr, httptest.NewRequest(http.MethodGet, "/", nil), rec, Page{Layout: "app", Name: "home", Data: map[string]string{}}, cais.Config{})
	if !strings.Contains(rr.Body.String(), `<button type="submit">Go</button>`) {
		t.Fatalf("body %q", rr.Body.String())
	}
}

func TestLoad_nestedPageName(t *testing.T) {
	fsys := fstest.MapFS{
		"layouts/app.html":       &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
		"pages/items/index.html": &fstest.MapFile{Data: []byte(`{{ define "content" }}<p>{{ .Title }}</p>{{ end }}`)},
	}
	rec, err := Load(fsys, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	Write(rr, httptest.NewRequest(http.MethodGet, "/", nil), rec, Page{Name: "items/index", Data: map[string]string{"Title": "List"}}, cais.Config{})
	if rr.Code != 200 {
		t.Fatalf("code %d body %q", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "<p>List</p>") {
		t.Fatalf("body %q", rr.Body.String())
	}
}

func TestLoad_unknownComponentFailsBoot(t *testing.T) {
	fsys := fstest.MapFS{
		"layouts/app.html": &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
		"pages/home.html":  &fstest.MapFile{Data: []byte(`{{ define "content" }}<.missing />{{ end }}`)},
	}
	if _, err := Load(fsys, nil); err == nil {
		t.Fatal("expected boot error")
	}
}
