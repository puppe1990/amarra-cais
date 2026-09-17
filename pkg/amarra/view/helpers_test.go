package view

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/puppe1990/amarra-cais/pkg/cais"
)

func TestLinkTo_driveAnchor(t *testing.T) {
	// Drive intercepts every same-origin link by default (#31); data-amarra-drive
	// is a no-op the runtime never reads. The anchor must stay plain.
	got := string(LinkTo("/x", "Hi"))
	if !strings.Contains(got, `href="/x"`) || !strings.Contains(got, "Hi") {
		t.Fatalf("%q", got)
	}
	if strings.Contains(got, `data-amarra-drive`) {
		t.Fatalf("no-op drive attr leaked into anchor: %q", got)
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

func TestLinkTo_methodConfirmFrame(t *testing.T) {
	got := string(LinkTo("/x", "Del", map[string]any{
		"method":  "delete",
		"confirm": "Sure?",
		"frame":   "cart",
	}))
	for _, want := range []string{
		`href="/x"`,
		`data-amarra-method="delete"`,
		`data-amarra-confirm="Sure?"`,
		`data-amarra-frame="cart"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in %s", want, got)
		}
	}
}

// #116: LinkTo bypasses html/template's URL filter, so `{{ linkTo .UserURL }}`
// with javascript: emitted an executable href (Drive does not intercept it).
func TestLinkTo_neutralizesUnsafeSchemes(t *testing.T) {
	for _, href := range []string{
		"javascript:alert(1)",
		"JaVaScRiPt:alert(1)",
		" javascript:alert(1)",
		"java\tscript:alert(1)",
		"vbscript:msgbox(1)",
		"data:text/html,<script>alert(1)</script>",
	} {
		got := string(LinkTo(href, "click"))
		if strings.Contains(strings.ToLower(got), "javascript:") ||
			strings.Contains(strings.ToLower(got), "vbscript:") ||
			strings.Contains(strings.ToLower(got), "data:text/html") {
			t.Errorf("unsafe href survived (%q): %q", href, got)
		}
		if !strings.Contains(got, `href="#"`) {
			t.Errorf("unsafe href %q should become #, got %q", href, got)
		}
		if !strings.Contains(got, "click") {
			t.Errorf("label dropped for %q: %q", href, got)
		}
	}
}

func TestLinkTo_keepsAllowedSchemesAndRelativePaths(t *testing.T) {
	for _, href := range []string{
		"/items/1",
		"items/1",
		"#anchor",
		"?page=2",
		"https://example.com/x",
		"http://example.com",
		"mailto:hi@example.com",
		"tel:+5511999999999",
	} {
		got := string(LinkTo(href, "x"))
		if !strings.Contains(got, `href="`+template.HTMLEscapeString(href)+`"`) {
			t.Errorf("safe href %q was rewritten: %q", href, got)
		}
	}
}

func TestLoad_linkToDictOpts(t *testing.T) {
	fsys := fstest.MapFS{
		"layouts/app.html": &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
		"pages/home.html":  &fstest.MapFile{Data: []byte(`{{ define "content" }}{{ linkTo "/x" "Del" (dict "method" "delete" "confirm" "Sure?") }}{{ end }}`)},
	}
	rec, err := Load(fsys, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	Write(rr, httptest.NewRequest(http.MethodGet, "/", nil), rec, Page{Layout: "app", Name: "home", Data: map[string]any{}}, cais.Config{})
	body := rr.Body.String()
	if !strings.Contains(body, `data-amarra-method="delete"`) || !strings.Contains(body, `data-amarra-confirm="Sure?"`) {
		t.Fatalf("%q", body)
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
	if !strings.Contains(body, `href="/x"`) || !strings.Contains(body, "Hi") || strings.Contains(body, `data-amarra-drive`) {
		t.Fatalf("%q", body)
	}
}
