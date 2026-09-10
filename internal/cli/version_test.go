package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestDefaultScaffoldCaisVersion_is022(t *testing.T) {
	if defaultScaffoldCaisVersion != "0.2.2" {
		t.Errorf("defaultScaffoldCaisVersion = %q, want 0.2.2 so amarra-cais new pins the tagged release", defaultScaffoldCaisVersion)
	}
}

func TestREADME_goInstallPinsV022(t *testing.T) {
	body, err := os.ReadFile("../../README.md")
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if !strings.Contains(text, "amarra-cais@v0.2.2") {
		t.Error("README go install should pin @v0.2.2")
	}
	if strings.Contains(text, "amarra-cais@v0.2.1") {
		t.Error("README still pins @v0.2.1")
	}
}

func TestCLI_Version(t *testing.T) {
	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	if err := c.Run([]string{"version"}); err != nil {
		t.Fatal(err)
	}
	out := strings.TrimSpace(buf.String())
	if out == "" {
		t.Fatal("expected non-empty version output")
	}
}

func TestCLI_Help_IncludesVersion(t *testing.T) {
	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	if err := c.Run([]string{"help"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "amarra-cais version") {
		t.Error("help missing amarra-cais version")
	}
}
