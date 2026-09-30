package cli

import (
	"os"
	"strings"
	"testing"
)

func TestDocs_splashLinksSQLiteCeiling(t *testing.T) {
	cases := []struct {
		path, href string
	}{
		{"../../website/src/content/docs/index.mdx", "/amarra-cais/docs/explanation/jobs-and-sqlite/"},
		{"../../website/src/content/docs/pt-br/index.mdx", "/amarra-cais/pt-br/docs/explanation/jobs-and-sqlite/"},
	}
	for _, tc := range cases {
		body, err := os.ReadFile(tc.path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(body), tc.href) {
			t.Errorf("%s missing splash link %q", tc.path, tc.href)
		}
	}
}

func TestDocs_gettingStartedNamesSQLiteAndLiveCeiling(t *testing.T) {
	cases := []struct {
		path    string
		needles []string
	}{
		{
			"../../website/src/content/docs/docs/getting-started/index.md",
			[]string{"MaxOpenConns(1)", "jobs work", "single-replica", "sqlc"},
		},
		{
			"../../website/src/content/docs/pt-br/docs/getting-started/index.md",
			[]string{"MaxOpenConns(1)", "jobs work", "réplica única", "sqlc"},
		},
	}
	for _, tc := range cases {
		body, err := os.ReadFile(tc.path)
		if err != nil {
			t.Fatal(err)
		}
		text := string(body)
		for _, needle := range tc.needles {
			if !strings.Contains(text, needle) {
				t.Errorf("%s missing %q", tc.path, needle)
			}
		}
	}
}
