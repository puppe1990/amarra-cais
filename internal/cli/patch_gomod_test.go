package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindCaisSiblingFrom_amarraCaisDir(t *testing.T) {
	root := t.TempDir()
	fw := filepath.Join(root, "amarra-cais")
	app := filepath.Join(root, "demo")
	for _, d := range []string{fw, app} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(fw, "go.mod"), []byte("module github.com/puppe1990/amarra-cais\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := findCaisSiblingFrom(app)
	if got != fw {
		t.Fatalf("findCaisSiblingFrom = %q, want %q", got, fw)
	}
}

func writeFakeCaisCheckout(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module github.com/puppe1990/amarra-cais\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readGoMod(t *testing.T, appDir string) string {
	t.Helper()
	body, err := os.ReadFile(filepath.Join(appDir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

// A sibling checkout next to the app must not be linked implicitly: the
// replace breaks clone and CI on every machine without that path (#56).
func TestScaffoldNewApp_doesNotAutoLinkSiblingCais(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	t.Setenv("CAIS_REPLACE", "")
	root := t.TempDir()
	caisDir := filepath.Join(root, "Cais")
	appsDir := filepath.Join(root, "Cais-apps", "demo")
	if err := os.MkdirAll(filepath.Dir(appsDir), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFakeCaisCheckout(t, caisDir)

	if err := scaffoldNewApp(appsDir, scaffoldData{
		AppName:    "demo",
		ModulePath: "github.com/puppe1990/demo",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if mod := readGoMod(t, appsDir); strings.Contains(mod, "replace github.com/puppe1990/amarra-cais") {
		t.Errorf("new must not write a machine-local replace (#56):\n%s", mod)
	}
}

// Same for a sibling reachable only from the current directory: the replace
// used to depend on where the user happened to run `new` (#56).
func TestScaffoldNewApp_doesNotAutoLinkFromCwd(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	t.Setenv("CAIS_REPLACE", "")
	root := t.TempDir()
	caisDir := filepath.Join(root, "Cais")
	appsDir := filepath.Join(root, "Cais-apps")
	appDir := filepath.Join(root, "remote", "testapp")
	for _, d := range []string{appsDir, filepath.Dir(appDir)} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	writeFakeCaisCheckout(t, caisDir)
	t.Chdir(appsDir)

	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "testapp",
		ModulePath: "github.com/puppe1990/testapp",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if mod := readGoMod(t, appDir); strings.Contains(mod, "replace github.com/puppe1990/amarra-cais") {
		t.Errorf("new must not link a cwd sibling implicitly (#56):\n%s", mod)
	}
}

func TestScaffoldNewApp_linksWhenCAISReplaceIsSet(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	t.Setenv("CAIS_REPLACE", "../Cais")
	appDir := filepath.Join(t.TempDir(), "demo")

	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "demo",
		ModulePath: "github.com/puppe1990/demo",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if mod := readGoMod(t, appDir); !strings.Contains(mod, "replace github.com/puppe1990/amarra-cais => ../Cais") {
		t.Errorf("CAIS_REPLACE must be written to go.mod:\n%s", mod)
	}
}

// `amarra-cais link` is the explicit path and keeps discovering a sibling
// checkout, unlike `new` (#56).
func TestCLI_Link_discoversSiblingCais(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	t.Setenv("CAIS_REPLACE", "")
	fakeToolchain(t)
	root := t.TempDir()
	caisDir := filepath.Join(root, "Cais")
	appDir := filepath.Join(root, "demo")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFakeCaisCheckout(t, caisDir)
	mod := "module github.com/puppe1990/demo\n\ngo 1.26\n\nrequire github.com/puppe1990/amarra-cais v0.3.0\n"
	if err := os.WriteFile(filepath.Join(appDir, "go.mod"), []byte(mod), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(appDir)

	var buf bytes.Buffer
	if err := (&CLI{Out: &buf}).Run([]string{"link"}); err != nil {
		t.Fatalf("link should find the sibling checkout: %v\n%s", err, buf.String())
	}
	if got := readGoMod(t, appDir); !strings.Contains(got, "replace github.com/puppe1990/amarra-cais => ../Cais") {
		t.Errorf("link must write the sibling replace:\n%s", got)
	}
}

func TestCLI_New_warnsWhenCAISReplaceIsSet(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	t.Setenv("CAIS_REPLACE", "../Cais")
	appDir := filepath.Join(t.TempDir(), "demo")

	var buf bytes.Buffer
	if err := (&CLI{Out: &buf}).Run([]string{"new", "demo", appDir}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "do not commit this replace") {
		t.Errorf("new must warn that the replace is machine-local:\n%s", out)
	}
	if !strings.Contains(out, "amarra-cais link --unlink") {
		t.Errorf("new must show the unlink hint:\n%s", out)
	}
}
