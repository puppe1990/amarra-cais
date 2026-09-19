package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCheckCSS_missing(t *testing.T) {
	dir := t.TempDir()
	c := checkCSS(dir)
	if c.OK {
		t.Fatal("expected FAIL when styles.css missing")
	}
	if !strings.Contains(c.FixHint, "amarra-cais css") {
		t.Errorf("FixHint = %q, want amarra-cais css", c.FixHint)
	}
}

func TestCheckCSS_placeholderOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "web/static/css")
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	// Scaffold ships a comment-only stub; treating it as OK leaves apps unstyled (#141).
	if err := os.WriteFile(filepath.Join(path, "styles.css"), []byte("/* Run: cais css */\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	c := checkCSS(dir)
	if c.OK {
		t.Fatal("expected FAIL for placeholder styles.css")
	}
	if !strings.Contains(strings.ToLower(c.Detail), "empty") && !strings.Contains(strings.ToLower(c.Detail), "not built") {
		t.Errorf("Detail = %q, want empty/not built", c.Detail)
	}
}

func TestCheckCSS_emptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "web/static/css")
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "styles.css"), []byte("   \n"), 0o644); err != nil {
		t.Fatal(err)
	}
	c := checkCSS(dir)
	if c.OK {
		t.Fatal("expected FAIL for empty styles.css")
	}
}

func TestCheckCSS_builtOK(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "web/static/css")
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	// Minified Tailwind output is large and has real selectors.
	body := "*,:after,:before{box-sizing:border-box}.text-stone-900{color:#1c1917}"
	if err := os.WriteFile(filepath.Join(path, "styles.css"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	c := checkCSS(dir)
	if !c.OK {
		t.Fatalf("expected OK for built CSS, got %+v", c)
	}
}

func TestStylesCSSReady(t *testing.T) {
	dir := t.TempDir()
	if stylesCSSReady(dir) {
		t.Fatal("missing file should not be ready")
	}
	path := filepath.Join(dir, cssOutput)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("/* Run: cais css */"), 0o644); err != nil {
		t.Fatal(err)
	}
	if stylesCSSReady(dir) {
		t.Fatal("placeholder should not be ready")
	}
	if err := os.WriteFile(path, []byte(".flex{display:flex}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !stylesCSSReady(dir) {
		t.Fatal("built CSS should be ready")
	}
}

func writeBuiltCSSAt(t *testing.T, dir string, mtime time.Time) {
	t.Helper()
	path := filepath.Join(dir, cssOutput)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(".flex{display:flex}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, mtime, mtime); err != nil {
		t.Fatal(err)
	}
}

func TestStylesCSSStale_inputNewer(t *testing.T) {
	dir := t.TempDir()
	old := time.Now().Add(-2 * time.Hour)
	newer := time.Now().Add(-time.Minute)
	writeBuiltCSSAt(t, dir, old)
	if err := os.WriteFile(filepath.Join(dir, cssInput), []byte("@import \"tailwindcss\";\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(filepath.Join(dir, cssInput), newer, newer); err != nil {
		t.Fatal(err)
	}
	stale, src, _ := stylesCSSStale(dir)
	if !stale {
		t.Fatal("expected stale when input.css is newer")
	}
	if src != cssInput {
		t.Errorf("src = %q, want %s", src, cssInput)
	}
}

func TestStylesCSSStale_templateNewer(t *testing.T) {
	dir := t.TempDir()
	old := time.Now().Add(-2 * time.Hour)
	newer := time.Now().Add(-time.Minute)
	writeBuiltCSSAt(t, dir, old)
	tpl := filepath.Join(dir, cssTemplates, "layout.html")
	if err := os.MkdirAll(filepath.Dir(tpl), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tpl, []byte(`<div class="group-hover:block"></div>`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(tpl, newer, newer); err != nil {
		t.Fatal(err)
	}
	stale, src, _ := stylesCSSStale(dir)
	if !stale {
		t.Fatal("expected stale when a template is newer")
	}
	if !strings.Contains(src, "layout.html") {
		t.Errorf("src = %q, want template path", src)
	}
}

func TestStylesCSSStale_cssNewer(t *testing.T) {
	dir := t.TempDir()
	old := time.Now().Add(-2 * time.Hour)
	newer := time.Now().Add(-time.Minute)
	if err := os.WriteFile(filepath.Join(dir, cssInput), []byte("@import \"tailwindcss\";\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(filepath.Join(dir, cssInput), old, old); err != nil {
		t.Fatal(err)
	}
	writeBuiltCSSAt(t, dir, newer)
	stale, src, _ := stylesCSSStale(dir)
	if stale {
		t.Fatalf("expected not stale when styles.css is newest, src=%s", src)
	}
}

func TestCheckCSS_staleWarns(t *testing.T) {
	dir := t.TempDir()
	old := time.Now().Add(-2 * time.Hour)
	newer := time.Now().Add(-time.Minute)
	writeBuiltCSSAt(t, dir, old)
	if err := os.WriteFile(filepath.Join(dir, cssInput), []byte("@import \"tailwindcss\";\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(filepath.Join(dir, cssInput), newer, newer); err != nil {
		t.Fatal(err)
	}
	c := checkCSS(dir)
	if c.OK {
		t.Fatal("expected warn (not OK) for stale styles.css")
	}
	if !c.Optional {
		t.Fatal("stale CSS should warn, not FAIL doctor")
	}
	if !strings.Contains(c.Detail, "stale styles.css") {
		t.Errorf("Detail = %q, want stale styles.css", c.Detail)
	}
	if !strings.Contains(c.FixHint, "amarra-cais css") {
		t.Errorf("FixHint = %q", c.FixHint)
	}
}
