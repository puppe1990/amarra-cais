package cli

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// #95: AGENTS.md caps files at ~500 lines (target 200-300) so agents can read
// them whole. Generated/vendored assets are exempt; every other Go source and
// test file must stay under the cap.
func TestGoSources_stayUnderLineCap(t *testing.T) {
	const lineCap = 500
	err := filepath.WalkDir("../..", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "node_modules", "vendor", "assets", "bin", "tmp", "testdata":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".go") {
			return nil
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if lines := bytes.Count(body, []byte("\n")) + 1; lines > lineCap {
			t.Errorf("%s has %d lines (cap %d) — split by domain (#95)", path, lines, lineCap)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// #288: files already near the 500-line cap split before they become god files.
func TestGoSources_nearCapModulesStayUnderSoftCap(t *testing.T) {
	const softCap = 400
	files := []string{
		"../../pkg/cais/pwa/pwa.go",
		"cli.go",
		"resource.go",
		"tpl_scaffold_css.go",
		"destroy_unpatch.go",
		"resource_gen_admin.go",
	}
	for _, path := range files {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("%s: %v", path, err)
			continue
		}
		lines := bytes.Count(body, []byte("\n")) + 1
		if lines > softCap {
			t.Errorf("%s has %d lines (soft cap %d) — split by responsibility (#288)", path, lines, softCap)
		}
	}
}
