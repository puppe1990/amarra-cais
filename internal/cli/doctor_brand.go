package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/puppe1990/amarra-cais/pkg/cais/pwa"
)

// checkBrandAssets warns while the app still ships the scaffold's placeholder
// icons and OG image: a shared link or an installed PWA would show placeholder
// art instead of the product's brand (#64).
func checkBrandAssets(dir string) doctorCheck {
	defaults := pwa.DefaultBrandAssets()
	if len(defaults) == 0 {
		return doctorCheck{Name: "brand assets", OK: true, Detail: "skipped (no defaults)"}
	}
	var stale []string
	found := 0
	for rel, want := range defaults {
		body, err := os.ReadFile(filepath.Join(dir, "web", "static", filepath.FromSlash(rel)))
		if err != nil {
			continue
		}
		found++
		if bytes.Equal(body, want) {
			stale = append(stale, rel)
		}
	}
	switch {
	case found == 0:
		return doctorCheck{Name: "brand assets", OK: true, Detail: "skipped (no icons)"}
	case len(stale) == 0:
		return doctorCheck{Name: "brand assets", OK: true, Detail: "icons + og.png replaced"}
	}
	sort.Strings(stale)
	return doctorCheck{
		Name:     "brand assets",
		Optional: true,
		Detail:   "still the scaffold placeholder: " + strings.Join(stale, ", "),
		FixHint:  "replace web/static/icons/*.png and web/static/og.png with the product brand",
	}
}
