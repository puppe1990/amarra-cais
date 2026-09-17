package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// #99: the public contract is amarra.js. Generators must never emit the HTMX
// legacy again, and the legacy packages must stay marked deprecated until the
// planned removal.
func TestScaffoldTemplates_neverEmitHTMX(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasPrefix(name, "tpl_") ||
			!strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		body, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		for i, line := range strings.Split(string(body), "\n") {
			if idx := strings.Index(line, "//"); idx >= 0 {
				line = line[:idx]
			}
			for _, needle := range []string{"hx-", "hx_", "IsHTMX", "htmxattrs"} {
				if strings.Contains(line, needle) {
					t.Errorf("%s:%d emits legacy HTMX (%q): %s", name, i+1, needle, line)
				}
			}
		}
	}
}

func TestHTMXLegacy_isMarkedDeprecated(t *testing.T) {
	for _, rel := range []string{
		"../../pkg/cais/htmx.go",
		"../../pkg/cais/htmxattrs/htmxattrs.go",
	} {
		body, err := os.ReadFile(filepath.Clean(rel))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(body), "Deprecated:") {
			t.Errorf("%s should carry a Deprecated: note with the removal plan", rel)
		}
	}
}
