package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseTailwindColorKeys_readsExtendColors(t *testing.T) {
	cfg := `theme: { extend: { colors: { ink: "#081014", foam: "#f3ead8", copper: "#c9893a" } } }`
	got := parseTailwindColorKeys(cfg)
	for _, want := range []string{"ink", "foam", "copper"} {
		if !got[want] {
			t.Errorf("missing color key %q in %v", want, got)
		}
	}
}

func TestMissingPaletteTokens_reportsUnknownCustomColor(t *testing.T) {
	markup := `<p class="text-copper bg-ink border-foam text-slate-900 text-xs border-r">hi</p>`
	missing := missingPaletteTokens(markup, map[string]bool{"ink": true})
	joined := strings.Join(missing, ",")
	if !strings.Contains(joined, "copper") || !strings.Contains(joined, "foam") {
		t.Errorf("missing = %v, want copper and foam", missing)
	}
	for _, keep := range missing {
		if keep == "slate" || keep == "xs" || keep == "r" {
			t.Errorf("should ignore Tailwind default/utility %q", keep)
		}
	}
}

func TestCheckPalette_okWhenConfigDefinesTokens(t *testing.T) {
	dir := t.TempDir()
	writePaletteFixture(t, dir, `theme: { extend: { colors: { ink: "#000", foam: "#fff", copper: "#c90", tide: "#123" } } }`,
		`<p class="text-copper bg-ink">x</p>`)
	c := checkPalette(dir)
	if !c.OK {
		t.Fatalf("expected OK, got %+v", c)
	}
}

func TestCheckPalette_warnsWhenColorsBlockIsGone(t *testing.T) {
	dir := t.TempDir()
	writePaletteFixture(t, dir, `theme: { extend: { fontFamily: { sans: ["system-ui"] } } }`,
		`<p class="text-copper">x</p>`)
	c := checkPalette(dir)
	if c.OK {
		t.Fatal("expected warning when custom colour classes remain after colors block is removed")
	}
	if !strings.Contains(c.Detail, "copper") {
		t.Errorf("Detail = %q, want copper", c.Detail)
	}
}

func TestCheckPalette_warnsWhenKitColorsDroppedFromConfig(t *testing.T) {
	dir := t.TempDir()
	writePaletteFixture(t, dir, `theme: { extend: { colors: { brand: "#111" } } }`,
		`<button class="text-copper bg-ink">toggle</button>`)
	c := checkPalette(dir)
	if c.OK {
		t.Fatal("expected warning when copper/ink are used but missing from tailwind.config.js (#207)")
	}
	if !c.Optional {
		t.Error("palette mismatch should be a warn, not a FAIL")
	}
	if !strings.Contains(c.Detail, "copper") || !strings.Contains(c.Detail, "ink") {
		t.Errorf("Detail = %q, want copper and ink", c.Detail)
	}
}

func writePaletteFixture(t *testing.T, dir, config, html string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, "web/templates/pages"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "tailwind.config.js"), []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "web/templates/pages/home.html"), []byte(html), 0o644); err != nil {
		t.Fatal(err)
	}
}
