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
	if !strings.Contains(out, "[ok] self-hosted fonts") {
		t.Errorf("fresh scaffold must not warn on missing/small fonts, got:\n%s", out)
	}
}

// dropGoogleFontsCSP removes the scaffold's font allowance so the warn path is
// exercised — a fresh scaffold ships the docs-site font hosts already allowed (#57).
func dropGoogleFontsCSP(t *testing.T, dir string) {
	t.Helper()
	path := filepath.Join(dir, ".env")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var kept []string
	for _, line := range strings.Split(string(body), "\n") {
		if strings.HasPrefix(line, "CSP_STYLE_SRC=") || strings.HasPrefix(line, "CSP_FONT_SRC=") {
			continue
		}
		kept = append(kept, line)
	}
	if err := os.WriteFile(path, []byte(strings.Join(kept, "\n")), 0o644); err != nil {
		t.Fatal(err)
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
	dropGoogleFontsCSP(t, dir)
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
	dropGoogleFontsCSP(t, dir)
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

func TestDoctor_MobileOKWhenCSPAllowsTemplateFonts(t *testing.T) {
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
	env := "CSP_STYLE_SRC=https://fonts.googleapis.com\nCSP_FONT_SRC=https://fonts.gstatic.com\n"
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(env), 0o644); err != nil {
		t.Fatal(err)
	}
	writeBuiltStylesCSS(t, dir)

	var buf bytes.Buffer
	if err := runDoctor(&buf, dir, doctorOptions{Mobile: true}); err != nil {
		t.Fatalf("doctor --mobile failed: %v\n%s", err, buf.String())
	}
	if !strings.Contains(buf.String(), "[ok] CSP fonts") {
		t.Errorf("expected CSP fonts ok when .env allows both hosts, got:\n%s", buf.String())
	}
}

func TestDoctor_MobileWarnsWhenOnlyStyleSrcAllowsFonts(t *testing.T) {
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
	env := "CSP_STYLE_SRC=https://fonts.googleapis.com\n"
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(env), 0o644); err != nil {
		t.Fatal(err)
	}
	writeBuiltStylesCSS(t, dir)

	var buf bytes.Buffer
	if err := runDoctor(&buf, dir, doctorOptions{Mobile: true}); err != nil {
		t.Fatalf("doctor should pass with warning: %v\n%s", err, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "[warn] CSP fonts") {
		t.Errorf("expected CSP fonts warning while font-src is missing, got:\n%s", out)
	}
	if !strings.Contains(out, "CSP_FONT_SRC") {
		t.Errorf("hint should name the missing directive, got:\n%s", out)
	}
}

func TestCheckSelfHostedFonts_warnsWhenWoff2Exceeds200KiB(t *testing.T) {
	dir := t.TempDir()
	fonts := filepath.Join(dir, "web", "static", "fonts")
	if err := os.MkdirAll(fonts, 0o755); err != nil {
		t.Fatal(err)
	}
	name := "material-symbols-outlined.woff2"
	if err := os.WriteFile(filepath.Join(fonts, name), make([]byte, 201*1024), 0o644); err != nil {
		t.Fatal(err)
	}

	c := checkSelfHostedFonts(dir)
	if c.OK || !c.Optional {
		t.Fatalf("huge woff2 must be an optional warning, got %+v", c)
	}
	if !strings.Contains(c.Detail, name) {
		t.Errorf("detail must name the file, got %q", c.Detail)
	}
	if !strings.Contains(c.Detail, "201 KiB") && !strings.Contains(c.Detail, "201KiB") {
		t.Errorf("detail must include size, got %q", c.Detail)
	}
	if c.FixHint == "" {
		t.Error("expected a subset FixHint")
	}
}

func TestCheckSelfHostedFonts_okWhenSmallOrMissing(t *testing.T) {
	dir := t.TempDir()
	c := checkSelfHostedFonts(dir)
	if !c.OK {
		t.Fatalf("missing fonts dir must skip, got %+v", c)
	}

	fonts := filepath.Join(dir, "web", "static", "fonts")
	if err := os.MkdirAll(fonts, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fonts, "inter.woff2"), make([]byte, 100*1024), 0o644); err != nil {
		t.Fatal(err)
	}
	c = checkSelfHostedFonts(dir)
	if !c.OK {
		t.Fatalf("100 KiB woff2 must pass, got %+v", c)
	}
}

func TestDoctor_MobileWarnsHugeSelfHostedWoff2(t *testing.T) {
	unsetCIEnv(t)
	dir := t.TempDir()
	fonts := filepath.Join(dir, "web", "static", "fonts")
	if err := os.MkdirAll(fonts, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fonts, "material-symbols-outlined.woff2"), make([]byte, 201*1024), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module app\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	_ = runDoctor(&buf, dir, doctorOptions{Mobile: true})
	out := buf.String()
	if !strings.Contains(out, "[warn] self-hosted fonts") {
		t.Errorf("expected optional warning for huge woff2, got:\n%s", out)
	}
	if !strings.Contains(out, "material-symbols-outlined.woff2") {
		t.Errorf("warning must name the file, got:\n%s", out)
	}
}
