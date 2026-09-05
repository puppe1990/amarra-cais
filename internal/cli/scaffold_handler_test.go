package cli

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateHandler_writesHTMLNotSvelte(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := filepath.Join(t.TempDir(), "app")
	if err := scaffoldNewApp(dir, scaffoldData{AppName: "app", ModulePath: "example.com/app"}, true, false); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	if err := (&CLI{Out: io.Discard}).Run([]string{"g", "handler", "settings"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "web/templates/pages/settings.html")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "web/src/pages/Settings.svelte")); err == nil {
		t.Fatal("svelte page should not exist")
	}
	src, _ := os.ReadFile(filepath.Join(dir, "internal/handlers/settings.go"))
	if strings.Contains(string(src), "inertia") {
		t.Fatal("handler still references inertia")
	}
	if !strings.Contains(string(src), "view.Write") {
		t.Fatal("handler missing view.Write")
	}
}

func TestScaffoldHandler_AfterResourceRoutesCompile(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "menu")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "menu",
		ModulePath: "github.com/puppe1990/menu",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "dish", resourceOpts{Public: true}); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldHandler(appDir, "about", false); err != nil {
		t.Fatal(err)
	}

	routes, err := os.ReadFile(filepath.Join(appDir, "internal/app/routes.go"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(routes)
	if strings.Contains(body, "})about") || strings.Contains(body, "})\tabout") {
		t.Errorf("handler route insert must start on new line after resource group: %s", body)
	}
	if !strings.Contains(body, `r.Get("/about", about.ServeHTTP)`) {
		t.Error("missing about route")
	}
}
