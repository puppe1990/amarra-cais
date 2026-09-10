package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGenerateResourceSmoke_compiles(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	caisDir := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	t.Setenv("CAIS_REPLACE", caisDir)

	appDir := filepath.Join(t.TempDir(), "smokeapp")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "smokeapp",
		ModulePath: "github.com/puppe1990/smokeapp",
	}, false, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "item", resourceOpts{
		Fields: "name:string",
		Seed:   false,
	}); err != nil {
		t.Fatal(err)
	}

	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = appDir
	if out, err := tidy.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy: %v\n%s", err, out)
	}

	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = appDir
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build ./... failed: %v\n%s", err, out)
	}
}

func TestSmokeScaffoldScript_generatesReferenceResource(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	script := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "scripts", "smoke-scaffold.sh"))
	body, err := os.ReadFile(script)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	for _, needle := range []string{
		"g resource category",
		"name:string",
		"g resource bookmark",
		"category_id:references",
		"--public",
		"--paginate",
	} {
		if !strings.Contains(text, needle) {
			t.Errorf("smoke-scaffold.sh missing %q (#34)", needle)
		}
	}
}

func TestGenerateResourceSmoke_referencesPublicPaginateCompiles(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	caisDir := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	t.Setenv("CAIS_REPLACE", caisDir)

	appDir := filepath.Join(t.TempDir(), "kitlib")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "kitlib",
		ModulePath: "github.com/puppe1990/kitlib",
	}, false, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "category", resourceOpts{
		Fields: "name:string",
		Seed:   true,
	}); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "bookmark", resourceOpts{
		Fields:   "title:string,url:url,category_id:references,read:bool",
		Public:   true,
		Paginate: true,
		Seed:     true,
	}); err != nil {
		t.Fatal(err)
	}

	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = appDir
	if out, err := tidy.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy: %v\n%s", err, out)
	}

	testCmd := exec.Command("go", "test", "./...", "-count=1")
	testCmd.Dir = appDir
	testCmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	if out, err := testCmd.CombinedOutput(); err != nil {
		t.Fatalf("go test ./... failed: %v\n%s", err, out)
	}

	build := exec.Command("go", "build", "-o", filepath.Join(t.TempDir(), "server"), "./cmd/server")
	build.Dir = appDir
	build.Env = append(os.Environ(), "CGO_ENABLED=0")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build ./cmd/server failed: %v\n%s", err, out)
	}
}
