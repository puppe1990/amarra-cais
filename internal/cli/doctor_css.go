package cli

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func checkCSS(dir string) doctorCheck {
	path := filepath.Join(dir, cssOutput)
	if _, err := os.Stat(path); err != nil {
		return doctorCheck{Name: "tailwind css", Detail: "styles.css missing", FixHint: "amarra-cais css"}
	}
	if !stylesCSSReady(dir) {
		return doctorCheck{
			Name:    "tailwind css",
			Detail:  "styles.css empty or not built (app will look unstyled)",
			FixHint: "amarra-cais css",
		}
	}
	if stale, ref := stylesCSSStale(dir); stale {
		return doctorCheck{
			Name:    "tailwind css",
			Detail:  fmt.Sprintf("styles.css is older than %s — run amarra-cais css", ref),
			FixHint: "amarra-cais css",
		}
	}
	return doctorCheck{Name: "tailwind css", OK: true}
}

// stylesCSSFresh reports whether the built styles.css is at least as new as every
// Tailwind input, so server/install rebuild after a template edit (#189).
func stylesCSSFresh(dir string) bool {
	stale, _ := stylesCSSStale(dir)
	return !stale
}

// stylesCSSStale compares styles.css modtime against input.css and web/templates/.
// The gitignored artifact is not rebuilt by `git pull`, so templates newer than it
// mean missing classes in production even though doctor would report [ok] (#189).
func stylesCSSStale(dir string) (bool, string) {
	info, err := os.Stat(filepath.Join(dir, cssOutput))
	if err != nil {
		return false, ""
	}
	newest, ref := newestTailwindInput(dir)
	if ref == "" {
		return false, ""
	}
	return newest.After(info.ModTime()), ref
}

func newestTailwindInput(dir string) (time.Time, string) {
	var newest time.Time
	ref := ""
	consider := func(path, rel string) {
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			return
		}
		if info.ModTime().After(newest) {
			newest = info.ModTime()
			ref = rel
		}
	}
	consider(filepath.Join(dir, cssInput), cssInput)
	templates := filepath.Join(dir, "web", "templates")
	_ = filepath.WalkDir(templates, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		rel, rerr := filepath.Rel(dir, path)
		if rerr != nil {
			rel = path
		}
		consider(path, rel)
		return nil
	})
	return newest, ref
}

// stylesCSSReady reports whether web/static/css/styles.css looks like a real Tailwind build.
// Scaffold writes a comment-only stub; git clones omit the gitignored file entirely (#141).
func stylesCSSReady(dir string) bool {
	body, err := os.ReadFile(filepath.Join(dir, cssOutput))
	if err != nil {
		return false
	}
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return false
	}
	// Comment-only / placeholder stubs ship with cais new until `cais css` runs.
	withoutComments := regexp.MustCompile(`(?s)/\*.*?\*/`).ReplaceAllString(trimmed, "")
	withoutComments = strings.TrimSpace(withoutComments)
	if withoutComments == "" {
		return false
	}
	// Real Tailwind output always contains a CSS rule (selector + block).
	return strings.Contains(withoutComments, "{")
}
