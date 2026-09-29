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

// TestScaffoldDocs_documentFakedata keeps new apps aware that sample data ships
// with the framework, so seeds and fixtures call fakedata instead of inventing
// literals by hand.
func TestScaffoldDocs_documentFakedata(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "fakedataapp")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "fakedataapp",
		ModulePath: "github.com/puppe1990/fakedataapp",
	}, false, false); err != nil {
		t.Fatal(err)
	}

	for _, file := range []string{"README.md", "AGENTS.md"} {
		body, err := os.ReadFile(filepath.Join(appDir, file))
		if err != nil {
			t.Fatal(err)
		}
		for _, needle := range []string{"pkg/cais/fakedata", "fakedata.Seed"} {
			if !strings.Contains(string(body), needle) {
				t.Errorf("%s must document fakedata: missing %q", file, needle)
			}
		}
	}
}
