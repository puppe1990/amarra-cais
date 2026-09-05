package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLI_Help(t *testing.T) {
	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	if err := c.Run([]string{"help"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "amarra-cais new") {
		t.Error("help missing amarra-cais new")
	}
}

func TestCLI_Help_UsesAmarraCais(t *testing.T) {
	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	if err := c.Run([]string{"help"}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{
		"amarra-cais new",
		"amarra-cais version",
		"amarra-cais g",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("help missing %q", want)
		}
	}
	if strings.Contains(out, "  cais new") {
		t.Error("help still documents cais new")
	}
	if !strings.Contains(out, "../amarra-cais") {
		t.Error("help missing sibling ../amarra-cais")
	}
}

func TestCLI_Help_IncludesResource(t *testing.T) {
	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	if err := c.Run([]string{"help"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "[--dry-run] resource") {
		t.Error("help missing g resource")
	}
}

func TestCLI_Help_IncludesLive(t *testing.T) {
	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	if err := c.Run([]string{"help"}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "g [--dry-run] live") {
		t.Error("help missing g live")
	}
	if !strings.Contains(out, "stream chat") {
		t.Error("help missing g stream chat")
	}
}

func TestCLI_Help_IncludesComponent(t *testing.T) {
	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	if err := c.Run([]string{"help"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "component") {
		t.Error("help missing g component")
	}
}

func TestCLI_Help_DevIsAirAndTailwind(t *testing.T) {
	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	if err := c.Run([]string{"help"}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	var devLine string
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "amarra-cais dev") {
			devLine = line
			break
		}
	}
	if devLine == "" {
		t.Fatal("help missing amarra-cais dev")
	}
	if !strings.Contains(devLine, "air") || !strings.Contains(devLine, "tailwind") {
		t.Errorf("dev help should be air + tailwind, got %q", devLine)
	}
	if strings.Contains(strings.ToLower(devLine), "vite") {
		t.Errorf("dev help must not mention vite: %q", devLine)
	}
}

func TestCLI_Help_IncludesModuleFlag(t *testing.T) {
	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	if err := c.Run([]string{"help"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "--module") {
		t.Error("help missing --module flag")
	}
}

func TestCLI_NewHelp_doesNotCreateDir(t *testing.T) {
	assertNewHelpDoesNotScaffold(t, []string{"new", "--help"}, "--help")
}

func TestCLI_NewHelp_shortFlag(t *testing.T) {
	assertNewHelpDoesNotScaffold(t, []string{"new", "-h"}, "-h")
}

func TestCLI_New_helpWithName_doesNotScaffold(t *testing.T) {
	assertNewHelpDoesNotScaffold(t, []string{"new", "myapp", "--help"}, "myapp", "--help")
}

func assertNewHelpDoesNotScaffold(t *testing.T, args []string, forbiddenDirs ...string) {
	t.Helper()
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := t.TempDir()
	t.Chdir(dir)

	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	if err := c.Run(args); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "amarra-cais new") {
		t.Errorf("help output missing amarra-cais new usage: %q", out)
	}
	if strings.Contains(out, "Created app") {
		t.Errorf("help must not scaffold: %q", out)
	}
	for _, name := range forbiddenDirs {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			t.Errorf("cais %v must not create directory %q", args, name)
		}
	}
}
