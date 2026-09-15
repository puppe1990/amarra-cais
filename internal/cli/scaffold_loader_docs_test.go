package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestScaffoldDocs_documentTemplateLoaderContract locks the loader contract
// documented for app authors (#65): nested pages are addressable, partials and
// components are flat, and unknown components fail at boot.
func TestScaffoldDocs_documentTemplateLoaderContract(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "docsapp")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "docsapp",
		ModulePath: "github.com/puppe1990/docsapp",
	}, false, false); err != nil {
		t.Fatal(err)
	}

	readme, err := os.ReadFile(filepath.Join(appDir, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	agents, err := os.ReadFile(filepath.Join(appDir, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		file string
		body string
	}{
		{"README.md", string(readme)},
		{"AGENTS.md", string(agents)},
	} {
		for _, needle := range []string{
			"pages/*/*.html",
			`view.Page{Name: "blog/post"}`,
			"partials/*.html",
		} {
			if !strings.Contains(tc.body, needle) {
				t.Errorf("%s must document the loader contract: missing %q", tc.file, needle)
			}
		}
	}
}
