package view

import (
	"strings"
	"testing"
)

// #391: Tailwind on Cleat runs before `go list -m` can resolve the module, so
// the kit HTML is invisible and utilities like bg-copper are pruned. The
// vendored class list must keep those tokens even with an empty GOMODCACHE.
func TestKitClassesFile_includesPinnedUtilitiesAndKitTokens(t *testing.T) {
	body := KitClassesFile()
	for _, want := range []string{
		"bg-copper",
		"hover:bg-foam",
		"border-copper/40",
		"text-ink",
		"amarra-btn",
		"peer-checked:bg-copper",
	} {
		if !tokenPresent(body, want) {
			t.Errorf("kit-classes.txt missing %q", want)
		}
	}
	for _, tok := range extractShippedKitClassTokens() {
		if !tokenPresent(body, tok) {
			t.Errorf("kit-classes.txt missing kit token %q", tok)
		}
	}
}

func tokenPresent(body, tok string) bool {
	for _, line := range strings.Split(body, "\n") {
		if strings.TrimSpace(line) == tok {
			return true
		}
	}
	return false
}
