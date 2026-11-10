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
// test file must stay under the cap. Skip .worktrees so a local linked
// checkout cannot fail the cap on another branch's files (#406).
func TestGoSources_stayUnderLineCap(t *testing.T) {
	const lineCap = 500
	err := filepath.WalkDir("../..", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "node_modules", "vendor", "assets", "bin", "tmp", "testdata", ".worktrees":
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

// #406: Drive/Hook JS has the same ~500-line cap as Go. Vendor (idiomorph)
// and *.test.mjs stay out; hook_tour / hook_select are already under the cap.
func TestJSSources_stayUnderLineCap(t *testing.T) {
	const lineCap = 500
	err := filepath.WalkDir("../..", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "node_modules", "vendor", "assets", "bin", "tmp", "testdata", ".worktrees":
				return filepath.SkipDir
			}
			return nil
		}
		name := d.Name()
		if !strings.HasSuffix(name, ".mjs") {
			return nil
		}
		if strings.HasSuffix(name, ".test.mjs") {
			return nil
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if lines := bytes.Count(body, []byte("\n")) + 1; lines > lineCap {
			t.Errorf("%s has %d lines (cap %d) — split by domain (#406)", path, lines, lineCap)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
