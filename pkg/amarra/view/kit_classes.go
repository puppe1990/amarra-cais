package view

import (
	_ "embed"
	"regexp"
	"sort"
	"strings"
)

//go:embed kit-classes.txt
var kitClassesFile string

// KitClassesFile is the vendored Tailwind content scan (#391). Cleat runs
// `npm run build` before `go: downloading`, so shippedKitContent() cannot
// see the kit HTML; this list keeps kit utilities in styles.css anyway.
func KitClassesFile() string {
	return kitClassesFile
}

var htmlClassAttr = regexp.MustCompile(`class="([^"]*)"`)
var templateAction = regexp.MustCompile(`\{\{[\s\S]*?\}\}`)

func extractShippedKitClassTokens() []string {
	seen := map[string]bool{}
	for _, html := range ShippedComponents() {
		stripped := templateAction.ReplaceAllString(html, " ")
		for _, m := range htmlClassAttr.FindAllStringSubmatch(stripped, -1) {
			for _, tok := range strings.Fields(m[1]) {
				if tok == "" {
					continue
				}
				seen[tok] = true
			}
		}
	}
	out := make([]string, 0, len(seen))
	for tok := range seen {
		out = append(out, tok)
	}
	sort.Strings(out)
	return out
}
