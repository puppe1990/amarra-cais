package cli

import (
	"os"
	"strings"
	"testing"
)

func TestRoadmapRemaining_omitsShippedWork(t *testing.T) {
	plan, err := os.ReadFile("../../docs/superpowers/plans/2026-07-01-framework-roadmap.md")
	if err != nil {
		t.Fatal(err)
	}
	text := string(plan)
	idx := strings.Index(text, "## Remaining")
	if idx < 0 {
		t.Fatal("missing ## Remaining")
	}
	remaining := text[idx:]
	for _, stale := range []string{
		"- [ ] External docs site",
		"- [ ] esbuild / JS bundling",
		"- [ ] REST PUT/DELETE in generated admin",
		"- [ ] Resource show page in generator",
		"- [ ] Nonce-based CSP",
	} {
		if strings.Contains(remaining, stale) {
			t.Errorf("Remaining still lists shipped work as open: %q (#270)", stale)
		}
	}
}

func TestRoadmapSpec_isNotDraftAndNonceShipped(t *testing.T) {
	body, err := os.ReadFile("../../docs/superpowers/specs/2026-07-01-framework-roadmap-design.md")
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if strings.Contains(text, "**Status:** Draft") {
		t.Error("roadmap spec must not stay Draft (#270)")
	}
	if strings.Contains(text, "nonce-based CSP deferred") {
		t.Error("CSP nonce shipped in #263; spec must not say it is deferred (#270)")
	}
}
