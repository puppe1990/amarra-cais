package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestDefaultScaffoldCaisVersion_is0141(t *testing.T) {
	if defaultScaffoldCaisVersion != "0.18.1" {
		t.Errorf("defaultScaffoldCaisVersion = %q, want 0.18.1 so amarra-cais new pins the tagged release", defaultScaffoldCaisVersion)
	}
}

func TestREADME_goInstallPinsV0141(t *testing.T) {
	body, err := os.ReadFile("../../README.md")
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if !strings.Contains(text, "amarra-cais@v0.18.1") {
		t.Error("README go install should pin @v0.18.1")
	}
	if strings.Contains(text, "amarra-cais@v0.18.0") {
		t.Error("README still pins @v0.18.0")
	}
}

// TestScaffoldVersion_dropsVCSBuildMetadata guards the module version a
// scaffold pins: `go build` stamps a dirty dev tree as "<tag>+dirty", which is
// not a valid module version for go.mod.
func TestScaffoldVersion_dropsVCSBuildMetadata(t *testing.T) {
	for _, tc := range []struct {
		in, want string
	}{
		{"v0.18.1", "0.18.1"},
		{"0.18.1", "0.18.1"},
		{"v0.18.1+dirty", "0.18.1"},
		{"0.18.1+incompatible", "0.18.1"},
		{"", defaultScaffoldCaisVersion},
		{"dev", defaultScaffoldCaisVersion},
	} {
		if got := scaffoldVersion(tc.in); got != tc.want {
			t.Errorf("scaffoldVersion(%q) = %q, want %q", tc.in, got, tc.want)
		}
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
