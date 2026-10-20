package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// #369: Idiomorph reuses a node with the same id. The scaffold contract must
// teach delegation/teardown instead of claiming the listener stays on an
// orphaned node, and must warn that data-* attributes written by JS (including
// dataset.ready) are erased by the morph.
func TestScaffoldAgents_documentMorphNodeReuse(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "morphdocs")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "morphdocs",
		ModulePath: "example.com/morphdocs",
	}, false, false); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(appDir, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	for _, needle := range []string{
		"same `id`",
		"same DOM node",
		"`data-*`",
		"`dataset.ready`",
		"`addEventListener`",
		"`amarra:morphed`",
		"`removeEventListener`",
	} {
		if !strings.Contains(text, needle) {
			t.Errorf("AGENTS.md missing morph reuse contract %q", needle)
		}
	}
	if strings.Contains(text, "orphaned node") {
		t.Error("AGENTS.md still claims the listener stays on an orphaned node; a stable id reuses the node (#369)")
	}
}

func TestWebsiteDocs_documentMorphNodeReuse(t *testing.T) {
	for _, tc := range []struct {
		file     string
		sameNode string
	}{
		{"../../website/src/content/docs/docs/explanation/views-and-drive.md", "same DOM node"},
		{"../../website/src/content/docs/pt-br/docs/explanation/views-and-drive.md", "mesmo nó do DOM"},
		{"../../website/src/content/docs/docs/reference/amarra-js.md", "same DOM node"},
		{"../../website/src/content/docs/pt-br/docs/reference/amarra-js.md", "mesmo nó do DOM"},
	} {
		body, err := os.ReadFile(tc.file)
		if err != nil {
			t.Fatal(err)
		}
		text := string(body)
		for _, needle := range []string{
			tc.sameNode,
			"`data-*`",
			"`dataset.ready`",
			"`addEventListener`",
			"`amarra:morphed`",
			"`removeEventListener`",
		} {
			if !strings.Contains(text, needle) {
				t.Errorf("%s missing morph reuse contract %q", tc.file, needle)
			}
		}
	}
}
