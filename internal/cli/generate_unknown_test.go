package cli

import (
	"io"
	"path/filepath"
	"strings"
	"testing"
)

// #404: the unknown-generator error listed a phantom `app` and omitted live/job.
func TestCLI_UnknownGenerator_matchesUsage(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "unkgen")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "unkgen",
		ModulePath: "github.com/puppe1990/unkgen",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	t.Chdir(appDir)
	err := (&CLI{Out: io.Discard}).Run([]string{"g", "foo"})
	if err == nil {
		t.Fatal("expected unknown generator error")
	}
	msg := err.Error()
	if strings.Contains(msg, "app,") || strings.Contains(msg, ", app") {
		t.Errorf("phantom app generator still listed: %s", msg)
	}
	for _, kind := range []string{
		"handler", "page", "component", "migration", "resource", "model",
		"stream", "live", "job", "console", "auth", "ci", "sitemap",
	} {
		if !strings.Contains(msg, kind) {
			t.Errorf("unknown generator error missing %q: %s", kind, msg)
		}
	}
}
