package view

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/puppe1990/amarra-cais/pkg/cais"
)

func TestLinkTo_driveAnchor(t *testing.T) {
	got := string(LinkTo("/x", "Hi"))
	if !strings.Contains(got, `href="/x"`) || !strings.Contains(got, "Hi") {
		t.Fatalf("%q", got)
	}
	if !strings.Contains(got, `data-amarra-drive="true"`) {
		t.Fatalf("missing drive attr: %q", got)
	}
}

func TestLinkTo_escapesHrefAndLabel(t *testing.T) {
	got := string(LinkTo(`/" onclick="x`, `<script>alert(1)</script>`))
	if strings.Contains(got, `" onclick="`) || strings.Contains(got, "<script>") {
		t.Fatalf("unescaped: %q", got)
	}
	if !strings.Contains(got, "alert(1)") {
		t.Fatalf("label dropped: %q", got)
	}
}

func TestLoad_linkToAvailableInTemplates(t *testing.T) {
	fsys := fstest.MapFS{
		"layouts/app.html": &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
		"pages/home.html":  &fstest.MapFile{Data: []byte(`{{ define "content" }}{{ linkTo "/x" "Hi" }}{{ end }}`)},
	}
	rec, err := Load(fsys, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	Write(rr, httptest.NewRequest(http.MethodGet, "/", nil), rec, Page{Layout: "app", Name: "home", Data: map[string]any{}}, cais.Config{})
	body := rr.Body.String()
	if !strings.Contains(body, `href="/x"`) || !strings.Contains(body, "Hi") || !strings.Contains(body, `data-amarra-drive="true"`) {
		t.Fatalf("%q", body)
	}
}
