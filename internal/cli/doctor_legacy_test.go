package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDoctor_FailsOnHxAttrsInTemplates(t *testing.T) {
	unsetCIEnv(t)
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := t.TempDir()
	if err := scaffoldNewApp(dir, scaffoldData{
		AppName:    "hxapp",
		ModulePath: "github.com/puppe1990/hxapp",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	writeBuiltStylesCSS(t, dir)
	page := filepath.Join(dir, "web/templates/pages/home.html")
	body, err := os.ReadFile(page)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(page, append(body, []byte(`<button hx-post="/x">Go</button>`)...), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err = runDoctor(&buf, dir, doctorOptions{})
	if err == nil {
		t.Fatalf("runDoctor should fail on hx-post, output:\n%s", buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "[FAIL] HTMX/Inertia leftovers") {
		t.Errorf("expected FAIL HTMX/Inertia leftovers, got:\n%s", out)
	}
	if !strings.Contains(out, "hx-post") {
		t.Errorf("expected hx-post in detail, got:\n%s", out)
	}
}

func TestDoctor_FailsOnGonertiaInGoMod(t *testing.T) {
	unsetCIEnv(t)
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := t.TempDir()
	if err := scaffoldNewApp(dir, scaffoldData{
		AppName:    "inertiaapp",
		ModulePath: "github.com/puppe1990/inertiaapp",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	writeBuiltStylesCSS(t, dir)
	mod := filepath.Join(dir, "go.mod")
	body, err := os.ReadFile(mod)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(mod, append(body, []byte("\nrequire github.com/hotwire-go/gonertia v0.1.0\n")...), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err = runDoctor(&buf, dir, doctorOptions{})
	if err == nil {
		t.Fatalf("runDoctor should fail on gonertia, output:\n%s", buf.String())
	}
	if !strings.Contains(buf.String(), "gonertia") {
		t.Errorf("expected gonertia in doctor output, got:\n%s", buf.String())
	}
}
