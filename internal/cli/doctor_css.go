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

func checkKitClasses(dir string) doctorCheck {
	const name = "kit css classes"
	cfgPath := filepath.Join(dir, "tailwind.config.js")
	cfg, err := os.ReadFile(cfgPath)
	if err != nil {
		return doctorCheck{Name: name, OK: true, Detail: "skipped (no tailwind.config.js)"}
	}
	kitPath := filepath.Join(dir, "web/static/css/kit-classes.txt")
	if _, err := os.Stat(kitPath); err != nil {
		return doctorCheck{
			Name:     name,
			Optional: true,
			Detail:   "web/static/css/kit-classes.txt missing — Tailwind prunes kit utilities when go list -m cannot see the module (#391)",
			FixHint:  "copy kit-classes.txt from a fresh amarra-cais new (or amarra-cais upgrade) and rebuild CSS",
		}
	}
	if !strings.Contains(string(cfg), "kit-classes.txt") {
		return doctorCheck{
			Name:     name,
			Optional: true,
			Detail:   "tailwind.config.js does not scan kit-classes.txt — kit CSS can go missing on a PaaS build (#391)",
			FixHint:  "add ./web/static/css/kit-classes.txt to content in tailwind.config.js",
		}
	}
	return doctorCheck{Name: name, OK: true, Detail: "kit-classes.txt vendored for CSS builds without the Go module"}
}

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
	if stale, ref, since := stylesCSSStale(dir); stale {
		return doctorCheck{
			Name:     "tailwind css",
			Optional: true,
			Detail:   fmt.Sprintf("styles.css is older than %s (since %s) — run amarra-cais css", ref, since.UTC().Format(time.RFC3339)),
			FixHint:  "amarra-cais css",
		}
	}
	return doctorCheck{Name: "tailwind css", OK: true}
}

// stylesCSSFresh reports whether the built styles.css is at least as new as every
// Tailwind input, so server/install rebuild after a template edit (#189).
func stylesCSSFresh(dir string) bool {
	stale, _, _ := stylesCSSStale(dir)
	return !stale
}

// stylesCSSStale compares styles.css modtime against input.css and web/templates/.
// The gitignored artifact is not rebuilt by `git pull`, so templates newer than it
// mean missing classes in production even though doctor would report [ok] (#189).
func stylesCSSStale(dir string) (bool, string, time.Time) {
	info, err := os.Stat(filepath.Join(dir, cssOutput))
	if err != nil {
		return false, "", time.Time{}
	}
	newest, ref := newestTailwindInput(dir)
	if ref == "" || !newest.After(info.ModTime()) {
		return false, "", time.Time{}
	}
	return true, ref, newest
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
	consider(filepath.Join(dir, "web/static/css/kit-classes.txt"), "web/static/css/kit-classes.txt")
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
