package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/puppe1990/amarra-cais/pkg/amarra/view"
)

const tplComponent = `<div class="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
  {{"{{"}} .Inner {{"}}"}}
</div>
`

// kitComponentStem returns the shipped kit stem matching name, or "".
// Kit stems keep hyphens (`locale-toggle`) while CLI names arrive as
// snake_case, so a hyphenated stem is also tried (#63).
func kitComponentStem(name string) string {
	shipped := view.ShippedComponents()
	snake := toSnake(name)
	for _, candidate := range []string{snake, strings.ReplaceAll(snake, "_", "-")} {
		if _, ok := shipped[candidate]; ok {
			return candidate
		}
	}
	return ""
}

// componentFileName keeps the shipped stem for kit overrides: a snake_case
// file never matches its `<.locale-toggle>` tag (#63).
func componentFileName(name, stem string) string {
	if stem != "" {
		return stem + ".html"
	}
	return toSnake(name) + ".html"
}

func scaffoldComponent(dir, name string, dryRun bool) error {
	stem := kitComponentStem(name)
	rel := filepath.Join("web/templates/components", componentFileName(name, stem))
	path := filepath.Join(dir, rel)
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%s already exists", rel)
	}
	if stem != "" {
		// Seed the shipped markup so the app restyles the real contract
		// (attributes, error slot, hooks) instead of recreating it (#63).
		// Written verbatim: it is a Go template for the app, not scaffold data.
		return writeScaffoldFile(path, []byte(view.ShippedComponents()[stem]), 0o644, rel, dryRun)
	}
	return writeScaffoldTemplate(path, tplComponent, dataForHandler(name), rel, dryRun)
}

// printShippedComponents lists what `g component <name>` can seed (#63).
func printShippedComponents(w io.Writer) {
	shipped := view.ShippedComponents()
	names := make([]string, 0, len(shipped))
	for name := range shipped {
		names = append(names, name)
	}
	sort.Strings(names)
	_, _ = fmt.Fprintf(w, "Shipped kit components (%d) — `amarra-cais g component <name>` seeds the markup to restyle:\n", len(names))
	for _, name := range names {
		_, _ = fmt.Fprintf(w, "  %s\n", name)
	}
}

func destroyComponent(dir, name string, dryRun, force bool) error {
	stem := kitComponentStem(name)
	files := []string{
		filepath.Join("web/templates/components", componentFileName(name, stem)),
	}
	// Generators before #63 named hyphenated kit stems with underscores.
	if alt := toSnake(name) + ".html"; alt != componentFileName(name, stem) {
		files = append(files, filepath.Join("web/templates/components", alt))
	}
	return removeGeneratedFiles(dir, files, dryRun, force)
}
