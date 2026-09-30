package cli

import (
	"os"
	"strings"
	"testing"
)

// Intro pages point at the full kit/hooks tables instead of a stale subset.
func TestDocs_introLinksShippedKitAndHooks(t *testing.T) {
	cases := []struct {
		path    string
		needles []string
	}{
		{
			"../../README.md",
			[]string{
				"docs/reference/views-and-kit/",
				"docs/reference/amarra-js/",
				"g component --list",
			},
		},
		{
			"../../website/src/content/docs/docs/getting-started/index.md",
			[]string{
				"/amarra-cais/docs/reference/views-and-kit/#shipped-kit",
				"/amarra-cais/docs/reference/amarra-js/",
				"g component --list",
			},
		},
		{
			"../../website/src/content/docs/pt-br/docs/getting-started/index.md",
			[]string{
				"/amarra-cais/pt-br/docs/reference/views-and-kit/#kit-incluso",
				"/amarra-cais/pt-br/docs/reference/amarra-js/",
				"g component --list",
			},
		},
		{
			"../../website/src/content/docs/docs/explanation/views-and-drive.md",
			[]string{
				"/amarra-cais/docs/reference/views-and-kit/#shipped-kit",
				"/amarra-cais/docs/reference/amarra-js/",
				"g component --list",
			},
		},
		{
			"../../website/src/content/docs/pt-br/docs/explanation/views-and-drive.md",
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
		text := strings.Join(strings.Fields(string(body)), " ")
		for _, needle := range tc.needles {
			if !strings.Contains(text, needle) {
				t.Errorf("%s missing kit/hooks pointer %q", tc.path, needle)
			}
		}
	}
}
