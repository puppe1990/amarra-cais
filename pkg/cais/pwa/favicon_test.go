package pwa

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// New Amarra apps use the docs boat mark as the tab favicon so a fresh
// scaffold looks like Amarra instead of the gray PWA placeholder tile.
func TestInstallForAmarra_writesDocsFavicon(t *testing.T) {
	dir := t.TempDir()
	if err := InstallForAmarra(dir, "Demo"); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(dir, "web/static/favicon.svg"))
	if err != nil {
		t.Fatal(err)
	}
	svg := string(body)
	for _, mark := range []string{`fill="#0b2f3a"`, `fill="#2eb8c9"`, `fill="#e07a4a"`} {
		if !strings.Contains(svg, mark) {
			t.Errorf("favicon.svg missing docs mark %s", mark)
		}
	}
}

func TestHeadHTML_linksDocsFavicon(t *testing.T) {
	html := HeadHTML()
	if !strings.Contains(html, `href="/static/favicon.svg"`) {
		t.Errorf("HeadHTML should link the SVG favicon, got:\n%s", html)
	}
	if !strings.Contains(html, `type="image/svg+xml"`) {
		t.Errorf("HeadHTML favicon should declare image/svg+xml, got:\n%s", html)
	}
}
