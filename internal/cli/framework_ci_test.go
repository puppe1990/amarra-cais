package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// #129: amarra.js is committed and shipped by `amarra-cais new`/`pwa`, but CI
// only ran js:test — a source edit without rebuild passed CI and shipped a
// stale runtime. The framework CI must rebuild and fail on diff; the step lives
// in the Frontend workflow, so every workflow file is scanned.
func TestFrameworkCI_verifiesCommittedJSBundles(t *testing.T) {
	entries, err := os.ReadDir("../../.github/workflows")
	if err != nil {
		t.Fatal(err)
	}
	var ci strings.Builder
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yml") {
			continue
		}
		body, err := os.ReadFile(filepath.Join("../../.github/workflows", entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		ci.Write(body)
	}
	text := ci.String()
	for _, want := range []string{
		"npm run js:build",
		"git diff --exit-code",
		"pkg/cais/pwa/assets/amarra.js",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("framework CI missing %q (committed bundles must be rebuild-checked)", want)
		}
	}
}

// #396: smoke-production used to boot /login 200 without running Tailwind, so
// #391 (kit classes missing when go list -m fails) was invisible in CI.
func TestSmokeProduction_buildsCSSWithoutGoModule(t *testing.T) {
	body, err := os.ReadFile("../../scripts/smoke-production.sh")
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	for _, want := range []string{
		"kit-classes.txt",
		"bg-copper",
		"amarra-btn",
		"tailwindcss",
		"/health",
		"/login",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("smoke-production.sh missing %q", want)
		}
	}
	if !strings.Contains(text, "GOMODCACHE") {
		t.Error("CSS gate must empty GOMODCACHE so shippedKitContent fails like Cleat")
	}
	// GH runners keep `go` in /usr/bin; hiding it from PATH is not enough —
	// stub the binary so `go list -m` fails the same way as a missing tool.
	if !strings.Contains(text, "hide-go") && !strings.Contains(text, "exit 127") {
		t.Error("CSS gate must stub go so /usr/bin/go cannot resolve the module")
	}
	wf, err := os.ReadFile("../../.github/workflows/ci.yml")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(wf), "setup-node") {
		t.Error("ci.yml smoke-production needs Node to run Tailwind")
	}
}

// #407: ci.yml paths-ignore website/**, so a docs-only PR never ran TestDocs_*.
// docs-site.yml is the job that still fires; pin the Go contract there without
// the full smoke/scaffold suite (#295).
func TestDocsSiteCI_runsTestDocs(t *testing.T) {
	body, err := os.ReadFile("../../.github/workflows/docs-site.yml")
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	for _, want := range []string{
		"actions/setup-go",
		"go test ./internal/cli/",
		"-run TestDocs",
		"#295",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("docs-site.yml missing %q (website PRs must run TestDocs)", want)
		}
	}
	for _, refuse := range []string{
		"make ci",
		"smoke-production",
		"amarra-cais new",
	} {
		if strings.Contains(text, refuse) {
			t.Errorf("docs-site.yml must not run %q on docs-only PRs", refuse)
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
