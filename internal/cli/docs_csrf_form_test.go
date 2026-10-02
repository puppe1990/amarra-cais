package cli

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// #293: kit <.form> injects csrf_token from $.CSRFToken. README/website/AGENTS
// examples that also call csrfField teach a double hidden field.
func TestDocs_kitFormExamplesDoNotDuplicateCSRFField(t *testing.T) {
	var files []string
	for _, root := range []string{"../../README.md", "../../AGENTS.md", "../../website/src/content/docs"} {
		info, err := os.Stat(root)
		if err != nil {
			t.Fatal(err)
		}
		if !info.IsDir() {
			files = append(files, root)
			continue
		}
		err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() {
				return nil
			}
			switch filepath.Ext(path) {
			case ".md", ".mdx":
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if len(files) == 0 {
		t.Fatal("expected README, AGENTS.md, and website docs to scan")
	}
	for _, path := range files {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, fence := range fencedMarkdownBlocks(string(body)) {
			from := 0
			for {
				rel := strings.Index(fence[from:], "<.form")
				if rel < 0 {
					break
				}
				start := from + rel
				endRel := strings.Index(fence[start:], "</.form>")
				if endRel < 0 {
					break
				}
				end := start + endRel + len("</.form>")
				block := fence[start:end]
				if strings.Contains(block, "csrfField") {
					t.Errorf("%s: kit <.form> example duplicates csrfField (kit injects $.CSRFToken) (#293):\n%s", path, block)
				}
				from = end
			}
		}
	}
}

func fencedMarkdownBlocks(text string) []string {
	var blocks []string
	rest := text
	for {
		start := strings.Index(rest, "```")
		if start < 0 {
			return blocks
		}
		rest = rest[start+3:]
		end := strings.Index(rest, "```")
		if end < 0 {
			return blocks
		}
		block := rest[:end]
		if nl := strings.Index(block, "\n"); nl >= 0 {
			block = block[nl+1:]
		}
		blocks = append(blocks, block)
		rest = rest[end+3:]
	}
}
