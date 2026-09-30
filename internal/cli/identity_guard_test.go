package cli

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// #266: live code, tests and docs must use generic examples
// (demo@example.com, https://app.example.com). CHANGELOG history may still
// name the previous dogfood host.
func TestSources_omitLegacyProductHost(t *testing.T) {
	needle := "pulse" + "fit"
	root := filepath.Clean(filepath.Join("..", ".."))
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "node_modules", "vendor", "assets", "bin", "tmp", "testdata", "dist":
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() == "CHANGELOG.md" {
			return nil
		}
		switch filepath.Ext(d.Name()) {
		case ".go", ".md", ".mdx", ".html", ".mjs", ".js", ".yml", ".yaml", ".txt":
		default:
			return nil
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(strings.ToLower(string(body)), needle) {
			rel, _ := filepath.Rel(root, path)
			t.Errorf("%s still names the old dogfood host (#266)", rel)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
