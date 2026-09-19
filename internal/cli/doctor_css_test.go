package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// writeFileAt writes body and pins mtime so staleness comparison is deterministic.
func writeFileAt(t *testing.T, path, body string, mod time.Time) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, mod, mod); err != nil {
		t.Fatal(err)
	}
}

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

// #189: after a `git pull` deploy, the gitignored styles.css can be older than a
// template edit that added new Tailwind classes. doctor must not report [ok].
func TestCheckCSS_staleWhenTemplateIsNewer(t *testing.T) {
	dir := t.TempDir()
	built := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	edited := built.Add(time.Hour)
	writeFileAt(t, filepath.Join(dir, cssOutput), ".flex{display:flex}", built)
	writeFileAt(t, filepath.Join(dir, "web/templates/pages/home.html"), "<p>group-hover:block</p>", edited)

	c := checkCSS(dir)
	if c.OK {
		t.Fatal("stale styles.css must not be [ok]")
	}
	if !c.Optional {
		t.Error("stale styles.css should warn, not fail doctor")
	}
	if !strings.Contains(c.Detail, "older") {
		t.Errorf("Detail = %q, want staleness explanation", c.Detail)
	}
	if !strings.Contains(c.Detail, "since") {
		t.Errorf("Detail = %q, want the newer input timestamp", c.Detail)
	}
	if !strings.Contains(c.FixHint, "amarra-cais css") {
		t.Errorf("FixHint = %q, want amarra-cais css", c.FixHint)
	}
}

func TestCheckCSS_staleWhenInputCSSIsNewer(t *testing.T) {
	dir := t.TempDir()
	built := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	edited := built.Add(time.Hour)
	writeFileAt(t, filepath.Join(dir, cssOutput), ".flex{display:flex}", built)
	writeFileAt(t, filepath.Join(dir, cssInput), "@import \"tailwindcss\";", edited)

	if c := checkCSS(dir); c.OK {
		t.Fatal("input.css newer than styles.css must not be [ok]")
	}
}

func TestCheckCSS_freshAfterRebuild(t *testing.T) {
	dir := t.TempDir()
	template := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	rebuilt := template.Add(time.Hour)
	writeFileAt(t, filepath.Join(dir, "web/templates/pages/home.html"), "<p>group-hover:block</p>", template)
	writeFileAt(t, filepath.Join(dir, cssOutput), ".flex{display:flex}", rebuilt)

	if c := checkCSS(dir); !c.OK {
		t.Fatalf("fresh styles.css should be OK, got %+v", c)
	}
}
