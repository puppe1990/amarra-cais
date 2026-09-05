package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLI_Help_IncludesConsole(t *testing.T) {
	var buf strings.Builder
	c := &CLI{Out: &buf}
	if err := c.Run([]string{"help"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "amarra-cais console") {
		t.Error("help missing amarra-cais console")
	}
}

func TestCLI_Console_missingMainHintsAmarraCais(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module app\n\nrequire github.com/puppe1990/amarra-cais v0.1.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	c := &CLI{Out: os.Stdout}
	err := c.Run([]string{"console"})
	if err == nil {
		t.Fatal("expected error when cmd/console/main.go is missing")
	}
	msg := err.Error()
	if !strings.Contains(msg, "amarra-cais g console") {
		t.Errorf("missing generate hint: %q", msg)
	}
	if strings.Contains(msg, "run: cais ") {
		t.Errorf("still documents old CLI: %q", msg)
	}
}

func TestCLI_Console_requiresCaisApp(t *testing.T) {
	c := &CLI{Out: os.Stdout}
	if err := c.Run([]string{"console"}); err == nil {
		t.Fatal("expected error outside cais app")
	}
}
