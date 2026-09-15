package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDoctor_MobileOKOnFreshScaffold(t *testing.T) {
	unsetCIEnv(t)
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := t.TempDir()
	if err := scaffoldNewApp(dir, scaffoldData{
		AppName:    "mobile",
		ModulePath: "github.com/puppe1990/mobile",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	writeBuiltStylesCSS(t, dir)

	var buf bytes.Buffer
	if err := runDoctor(&buf, dir, doctorOptions{Mobile: true}); err != nil {
		t.Fatalf("doctor --mobile failed: %v\n%s", err, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "[ok] flash template") {
		t.Errorf("expected flash template ok, got:\n%s", out)
	}
	if !strings.Contains(out, "[ok] CSP fonts") {
		t.Errorf("expected CSP fonts ok, got:\n%s", out)
	}
	if !strings.Contains(out, "amarra.js") {
		t.Errorf("mobile doctor should check amarra.js, got:\n%s", out)
	}
	if strings.Contains(out, "cais.js") || strings.Contains(out, "htmx.min.js") {
		t.Errorf("mobile doctor must not require cais.js/htmx, got:\n%s", out)
	}
}

func TestDoctor_MobileWarnsGoogleFonts(t *testing.T) {
	unsetCIEnv(t)
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := t.TempDir()
	if err := scaffoldNewApp(dir, scaffoldData{
		AppName:    "mobile",
		ModulePath: "github.com/puppe1990/mobile",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	cssPath := filepath.Join(dir, "input.css")
	body, err := os.ReadFile(cssPath)
	if err != nil {
		t.Fatal(err)
	}
	patched := "@import url('https://fonts.googleapis.com/css');\n" + string(body)
	if err := os.WriteFile(cssPath, []byte(patched), 0o644); err != nil {
		t.Fatal(err)
	}
	writeBuiltStylesCSS(t, dir)

	var buf bytes.Buffer
	if err := runDoctor(&buf, dir, doctorOptions{Mobile: true}); err != nil {
		t.Fatalf("doctor should pass with warning: %v\n%s", err, buf.String())
	}
	if !strings.Contains(buf.String(), "[warn] CSP fonts") {
		t.Errorf("expected CSP fonts warning, got:\n%s", buf.String())
	}
}

// linkGoogleFontsInLayout patches the generated layout with an external
// stylesheet link — the path used when porting an existing design (#57).
func linkGoogleFontsInLayout(t *testing.T, dir string) {
	t.Helper()
	layout := filepath.Join(dir, "web/templates/layouts/app.html")
	body, err := os.ReadFile(layout)
	if err != nil {
		t.Fatal(err)
	}
	link := "<link rel=\"stylesheet\" href=\"https://fonts.googleapis.com/css2?family=Inter\">\n" +
		"<link rel=\"preconnect\" href=\"https://fonts.gstatic.com\" crossorigin>\n"
	if err := os.WriteFile(layout, append([]byte(link), body...), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDoctor_MobileWarnsGoogleFontsInTemplate(t *testing.T) {
	unsetCIEnv(t)
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := t.TempDir()
	if err := scaffoldNewApp(dir, scaffoldData{
		AppName:    "mobile",
		ModulePath: "github.com/puppe1990/mobile",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	linkGoogleFontsInLayout(t, dir)
	writeBuiltStylesCSS(t, dir)

	var buf bytes.Buffer
	if err := runDoctor(&buf, dir, doctorOptions{Mobile: true}); err != nil {
		t.Fatalf("doctor should pass with warning: %v\n%s", err, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "[warn] CSP fonts") {
		t.Errorf("expected CSP fonts warning for a template <link>, got:\n%s", out)
	}
	if !strings.Contains(out, "CSP_STYLE_SRC") {
		t.Errorf("hint should name the CSP escape hatch, got:\n%s", out)
	}
}
