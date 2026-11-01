package cli

import (
	"os"
	"strings"
	"testing"
)

// #387: product vocabulary has one glossary page (EN + pt-BR). Each term
// links to its contract page; that page must mention the term. Drift names
// the file and the term, like #295.

const (
	glossaryEN = "../../website/src/content/docs/docs/explanation/glossary.md"
	glossaryPT = "../../website/src/content/docs/pt-br/docs/explanation/glossary.md"
)

type glossaryTerm struct {
	name       string
	hrefsEN    []string
	hrefsPT    []string
	destFiles  []string
	destNeedle string
}

func glossaryProductTerms() []glossaryTerm {
	docs := "../../website/src/content/docs"
	return []glossaryTerm{
		{
			name:       "Cais vs Amarra",
			hrefsEN:    []string{"/amarra-cais/docs/explanation/cais-vs-amarra/"},
			hrefsPT:    []string{"/amarra-cais/pt-br/docs/explanation/cais-vs-amarra/"},
			destFiles:  []string{docs + "/docs/explanation/cais-vs-amarra.md", docs + "/pt-br/docs/explanation/cais-vs-amarra.md"},
			destNeedle: "Inertia",
		},
		{
			name:    "Drive",
			hrefsEN: []string{"/amarra-cais/docs/explanation/views-and-drive/", "/amarra-cais/docs/reference/amarra-js/"},
			hrefsPT: []string{"/amarra-cais/pt-br/docs/explanation/views-and-drive/", "/amarra-cais/pt-br/docs/reference/amarra-js/"},
			destFiles: []string{
				docs + "/docs/explanation/views-and-drive.md",
				docs + "/pt-br/docs/explanation/views-and-drive.md",
				docs + "/docs/reference/amarra-js.md",
				docs + "/pt-br/docs/reference/amarra-js.md",
			},
			destNeedle: "Drive",
		},
		{
			name:       "Frame",
			hrefsEN:    []string{"/amarra-cais/docs/reference/amarra-js/"},
			hrefsPT:    []string{"/amarra-cais/pt-br/docs/reference/amarra-js/"},
			destFiles:  []string{docs + "/docs/reference/amarra-js.md", docs + "/pt-br/docs/reference/amarra-js.md"},
			destNeedle: "amarra-frame",
		},
		{
			name:       "Stream",
			hrefsEN:    []string{"/amarra-cais/docs/how-to/streaming-chat/"},
			hrefsPT:    []string{"/amarra-cais/pt-br/docs/how-to/streaming-chat/"},
			destFiles:  []string{docs + "/docs/how-to/streaming-chat.md", docs + "/pt-br/docs/how-to/streaming-chat.md"},
			destNeedle: "Stream",
		},
		{
			name:       "Live",
			hrefsEN:    []string{"/amarra-cais/docs/how-to/live-updates/"},
			hrefsPT:    []string{"/amarra-cais/pt-br/docs/how-to/live-updates/"},
			destFiles:  []string{docs + "/docs/how-to/live-updates.md", docs + "/pt-br/docs/how-to/live-updates.md"},
			destNeedle: "Live",
		},
		{
			name:       "kit",
			hrefsEN:    []string{"/amarra-cais/docs/reference/views-and-kit/"},
			hrefsPT:    []string{"/amarra-cais/pt-br/docs/reference/views-and-kit/"},
			destFiles:  []string{docs + "/docs/reference/views-and-kit.md", docs + "/pt-br/docs/reference/views-and-kit.md"},
			destNeedle: "kit",
		},
		{
			name:       "morph",
			hrefsEN:    []string{"/amarra-cais/docs/explanation/views-and-drive/"},
			hrefsPT:    []string{"/amarra-cais/pt-br/docs/explanation/views-and-drive/"},
			destFiles:  []string{docs + "/docs/explanation/views-and-drive.md", docs + "/pt-br/docs/explanation/views-and-drive.md"},
			destNeedle: "morph",
		},
		{
			name:       "amarra-hook",
			hrefsEN:    []string{"/amarra-cais/docs/reference/amarra-js/"},
			hrefsPT:    []string{"/amarra-cais/pt-br/docs/reference/amarra-js/"},
			destFiles:  []string{docs + "/docs/reference/amarra-js.md", docs + "/pt-br/docs/reference/amarra-js.md"},
			destNeedle: "amarra-hook",
		},
		{
			name:       "data-amarra-shell",
			hrefsEN:    []string{"/amarra-cais/docs/reference/amarra-js/"},
			hrefsPT:    []string{"/amarra-cais/pt-br/docs/reference/amarra-js/"},
			destFiles:  []string{docs + "/docs/reference/amarra-js.md", docs + "/pt-br/docs/reference/amarra-js.md"},
			destNeedle: "data-amarra-shell",
		},
		{
			name:       "#amarra-main",
			hrefsEN:    []string{"/amarra-cais/docs/getting-started/project-layout/"},
			hrefsPT:    []string{"/amarra-cais/pt-br/docs/getting-started/project-layout/"},
			destFiles:  []string{docs + "/docs/getting-started/project-layout.md", docs + "/pt-br/docs/getting-started/project-layout.md"},
			destNeedle: "#amarra-main",
		},
		{
			name:       "g resource",
			hrefsEN:    []string{"/amarra-cais/docs/reference/generators/"},
			hrefsPT:    []string{"/amarra-cais/pt-br/docs/reference/generators/"},
			destFiles:  []string{docs + "/docs/reference/generators.md", docs + "/pt-br/docs/reference/generators.md"},
			destNeedle: "g resource",
		},
		{
			name:       "store",
			hrefsEN:    []string{"/amarra-cais/docs/how-to/database-and-migrations/"},
			hrefsPT:    []string{"/amarra-cais/pt-br/docs/how-to/database-and-migrations/"},
			destFiles:  []string{docs + "/docs/how-to/database-and-migrations.md", docs + "/pt-br/docs/how-to/database-and-migrations.md"},
			destNeedle: "store",
		},
		{
			name:       "doctor",
			hrefsEN:    []string{"/amarra-cais/docs/reference/cli/"},
			hrefsPT:    []string{"/amarra-cais/pt-br/docs/reference/cli/"},
			destFiles:  []string{docs + "/docs/reference/cli.md", docs + "/pt-br/docs/reference/cli.md"},
			destNeedle: "doctor",
		},
		{
			name:       "/jobs",
			hrefsEN:    []string{"/amarra-cais/docs/explanation/jobs-and-sqlite/"},
			hrefsPT:    []string{"/amarra-cais/pt-br/docs/explanation/jobs-and-sqlite/"},
			destFiles:  []string{docs + "/docs/explanation/jobs-and-sqlite.md", docs + "/pt-br/docs/explanation/jobs-and-sqlite.md"},
			destNeedle: "job",
		},
		{
			name:       "destroy",
			hrefsEN:    []string{"/amarra-cais/docs/reference/generators/"},
			hrefsPT:    []string{"/amarra-cais/pt-br/docs/reference/generators/"},
			destFiles:  []string{docs + "/docs/reference/generators.md", docs + "/pt-br/docs/reference/generators.md"},
			destNeedle: "destroy",
		},
		{
			name:       "patch marker",
			hrefsEN:    []string{"/amarra-cais/docs/getting-started/project-layout/"},
			hrefsPT:    []string{"/amarra-cais/pt-br/docs/getting-started/project-layout/"},
			destFiles:  []string{docs + "/docs/getting-started/project-layout.md", docs + "/pt-br/docs/getting-started/project-layout.md"},
			destNeedle: "registerRoutes",
		},
	}
}

func TestDocs_glossaryCoversProductTerms(t *testing.T) {
	pages := []struct {
		path  string
		hrefs func(glossaryTerm) []string
	}{
		{glossaryEN, func(term glossaryTerm) []string { return term.hrefsEN }},
		{glossaryPT, func(term glossaryTerm) []string { return term.hrefsPT }},
	}
	for _, page := range pages {
		body, err := os.ReadFile(page.path)
		if err != nil {
			t.Errorf("%s: missing glossary — add the page or remove it from the canonical source: %v", page.path, err)
			continue
		}
		text := string(body)
		for _, term := range glossaryProductTerms() {
			if !strings.Contains(text, term.name) {
				t.Errorf("%s: missing term %q — add a glossary row or remove it from the canonical source", page.path, term.name)
			}
			for _, href := range page.hrefs(term) {
				if !strings.Contains(text, href) {
					t.Errorf("%s: term %q missing link %q — add the href or retarget the canonical source", page.path, term.name, href)
				}
			}
		}
		for _, stale := range []string{"bounded context", "ubiquitous language", "saga", "CQRS", "aggregate root"} {
			if strings.Contains(text, stale) {
				t.Errorf("%s still mentions DDD jargon %q", page.path, stale)
			}
		}
	}
}

func TestDocs_glossaryDestinationsMentionTerms(t *testing.T) {
	for _, term := range glossaryProductTerms() {
		for _, dest := range term.destFiles {
			body, err := os.ReadFile(dest)
			if err != nil {
				t.Errorf("%s: missing destination for %q: %v", dest, term.name, err)
				continue
			}
			if !strings.Contains(string(body), term.destNeedle) {
				t.Errorf("%s: missing %s %q — add it to the destination or retarget the glossary", dest, term.name, term.destNeedle)
			}
		}
	}
}

func TestDocs_glossaryLinkedFromIndexes(t *testing.T) {
	cases := []struct {
		path, href string
	}{
		{"../../website/src/content/docs/docs/explanation/index.md", "/amarra-cais/docs/explanation/glossary/"},
		{"../../website/src/content/docs/pt-br/docs/explanation/index.md", "/amarra-cais/pt-br/docs/explanation/glossary/"},
		{"../../website/src/content/docs/docs/getting-started/index.md", "/amarra-cais/docs/explanation/glossary/"},
		{"../../website/src/content/docs/pt-br/docs/getting-started/index.md", "/amarra-cais/pt-br/docs/explanation/glossary/"},
	}
	for _, tc := range cases {
		body, err := os.ReadFile(tc.path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(body), tc.href) {
			t.Errorf("%s missing glossary link %q", tc.path, tc.href)
		}
	}
}
