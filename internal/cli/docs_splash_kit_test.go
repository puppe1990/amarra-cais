package cli

import (
	"os"
	"strings"
	"testing"
)

// #272: splash kit+hooks card points at the full tables instead of a stale subset.
func TestDocs_splashLinksShippedKitAndHooks(t *testing.T) {
	cases := []struct {
		path    string
		needles []string
	}{
		{
			"../../website/src/content/docs/index.mdx",
			[]string{
				"/amarra-cais/docs/reference/views-and-kit/#shipped-kit",
				"/amarra-cais/docs/reference/amarra-js/",
				"g component --list",
			},
		},
		{
			"../../website/src/content/docs/pt-br/index.mdx",
			[]string{
				"/amarra-cais/pt-br/docs/reference/views-and-kit/#kit-incluso",
				"/amarra-cais/pt-br/docs/reference/amarra-js/",
				"g component --list",
			},
		},
	}
	for _, tc := range cases {
		body, err := os.ReadFile(tc.path)
		if err != nil {
			t.Fatal(err)
		}
		// Prettier may wrap `g component --list` across lines in the pt-BR card.
		text := strings.Join(strings.Fields(string(body)), " ")
		for _, needle := range tc.needles {
			if !strings.Contains(text, needle) {
				t.Errorf("%s missing splash kit/hooks pointer %q (#272)", tc.path, needle)
			}
		}
	}
}
