package cli

import (
	"os"
	"strings"
	"testing"
)

// #403: Drive applyHead copies a closed set of document meta after a morph.
// The amarra.js reference (EN + pt-BR) is the contract page; dropping a
// needle here means the docs drifted from the JS.
func TestDocs_driveHeadCopiesCanonicalOgJSONLD(t *testing.T) {
	needles := []string{
		"canonical",
		"og:",
		"twitter:",
		"application/ld+json",
		"hreflang",
	}
	for _, file := range []string{
		"../../website/src/content/docs/docs/reference/amarra-js.md",
		"../../website/src/content/docs/pt-br/docs/reference/amarra-js.md",
	} {
		body, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		text := string(body)
		for _, needle := range needles {
			if !strings.Contains(text, needle) {
				t.Errorf("%s missing Drive head contract %q (#403)", file, needle)
			}
		}
	}
}
