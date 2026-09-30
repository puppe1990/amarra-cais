package cli

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// #271: the load-at-call-time admin middleware is gone. AdminAuth(cfg) is
// the remaining API. CHANGELOG history and the upgrade manifest may still
// name the old function.
func TestSources_omitDeprecatedLoadAtCallAdminAuth(t *testing.T) {
	needle := "Token" + "Auth"
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
		switch d.Name() {
		case "CHANGELOG.md", "upgrade_manifest.go":
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
		if strings.Contains(string(body), needle) {
			rel, _ := filepath.Rel(root, path)
			t.Errorf("%s still names the deprecated load-at-call-time admin middleware (#271)", rel)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
