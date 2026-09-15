package view

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/puppe1990/amarra-cais/pkg/cais"
)

// TestLoad_contract_nestedPages locks the documented loader contract (#65):
// pages/*/*.html is addressable as "blog/post".
func TestLoad_contract_nestedPages(t *testing.T) {
	fsys := fstest.MapFS{
		"layouts/app.html":     &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
		"pages/home.html":      &fstest.MapFile{Data: []byte(`{{ define "content" }}home{{ end }}`)},
		"pages/blog/post.html": &fstest.MapFile{Data: []byte(`{{ define "content" }}post {{ template "cta" . }}{{ end }}`)},
		"partials/cta.html":    &fstest.MapFile{Data: []byte(`{{ define "cta" }}cta{{ end }}`)},
	}
	rec, err := Load(fsys, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	Write(rr, httptest.NewRequest(http.MethodGet, "/blog/post", nil), rec, Page{Layout: "app", Name: "blog/post"}, cais.Config{})
	if rr.Code != 200 {
		t.Fatalf("code %d body %q", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "post cta") {
		t.Fatalf("body %q", rr.Body.String())
	}
}

// TestLoad_contract_partialsAreFlat pins the trap: partials/*.html does not
// match nested files, so the failure surfaces when the page renders (#65).
func TestLoad_contract_partialsAreFlat(t *testing.T) {
	fsys := fstest.MapFS{
		"layouts/app.html":         &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
		"pages/home.html":          &fstest.MapFile{Data: []byte(`{{ define "content" }}{{ template "card" . }}{{ end }}`)},
		"partials/posts/card.html": &fstest.MapFile{Data: []byte(`{{ define "card" }}card{{ end }}`)},
	}
	rec, err := Load(fsys, nil)
	if err != nil {
		t.Fatalf("nested partials are ignored at load, not rejected: %v", err)
	}

	rr := httptest.NewRecorder()
	Write(rr, httptest.NewRequest(http.MethodGet, "/", nil), rec, Page{Layout: "app", Name: "home"}, cais.Config{})
	if rr.Code == 200 {
		t.Fatalf("nested partial must not resolve; body %q", rr.Body.String())
	}
}
