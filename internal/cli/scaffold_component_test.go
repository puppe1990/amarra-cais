package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateComponent_writesFile(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := filepath.Join(t.TempDir(), "app")
	if err := scaffoldNewApp(dir, scaffoldData{AppName: "app", ModulePath: "example.com/app"}, true, false); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	var buf bytes.Buffer
	if err := (&CLI{Out: &buf}).Run([]string{"g", "component", "card"}); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(dir, "web/templates/components/card.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "{{ .Inner }}") {
		t.Errorf("component missing Inner slot: %s", body)
	}
}

func TestGenerateComponent_dryRunWritesNothing(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := filepath.Join(t.TempDir(), "app")
	if err := scaffoldNewApp(dir, scaffoldData{AppName: "app", ModulePath: "example.com/app"}, true, false); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	if err := (&CLI{Out: bytes.NewBuffer(nil)}).Run([]string{"g", "--dry-run", "component", "card"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "web/templates/components/card.html")); !os.IsNotExist(err) {
		t.Error("dry-run should not create card.html")
	}
}

func TestDestroyComponent_removesFile(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := filepath.Join(t.TempDir(), "app")
	if err := scaffoldNewApp(dir, scaffoldData{AppName: "app", ModulePath: "example.com/app"}, true, false); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	if err := (&CLI{Out: bytes.NewBuffer(nil)}).Run([]string{"g", "component", "card"}); err != nil {
		t.Fatal(err)
	}
	if err := (&CLI{Out: bytes.NewBuffer(nil)}).Run([]string{"destroy", "component", "card"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "web/templates/components/card.html")); !os.IsNotExist(err) {
		t.Error("expected card.html removed")
	}
}
