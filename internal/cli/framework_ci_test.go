package cli

import (
	"os"
	"strings"
	"testing"
)

// #129: amarra.js is committed and shipped by `amarra-cais new`/`pwa`, but CI
// only ran js:test — a source edit without rebuild passed CI and shipped a
// stale runtime. The framework CI must rebuild and fail on diff.
func TestFrameworkCI_verifiesCommittedJSBundles(t *testing.T) {
	body, err := os.ReadFile("../../.github/workflows/ci.yml")
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	for _, want := range []string{
		"npm run js:build",
		"git diff --exit-code",
		"pkg/cais/pwa/assets/amarra.js",
		"pkg/cais/pwa/assets/cais-chat-logic.mjs",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("ci.yml missing %q (committed bundles must be rebuild-checked)", want)
		}
	}
}

func TestMakefile_hasBundleCheck(t *testing.T) {
	body, err := os.ReadFile("../../Makefile")
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if !strings.Contains(text, "js-bundle-check:") {
		t.Error("Makefile should define js-bundle-check for local parity with CI")
	}
	if !strings.Contains(text, "ci: test js-test lint format-check js-bundle-check") {
		t.Error("make ci should include js-bundle-check")
	}
}
