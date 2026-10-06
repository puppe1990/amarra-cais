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
	missing := missingPaletteTokens(markup, tailwindThemeKeys{colors: map[string]bool{"ink": true}})
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

// #246: a fresh scaffold warned about center/linecap/linejoin/width because the
// scanner read SVG attributes (stroke-linecap="round") as colour tokens.
func TestCheckPalette_ignoresSvgAttributes(t *testing.T) {
	dir := t.TempDir()
	writePaletteFixture(t, dir, `theme: { extend: { colors: { ink: "#000" } } }`,
		`<svg class="text-ink" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" stroke-dasharray="4 4"></svg>`)
	if c := checkPalette(dir); !c.OK {
		t.Fatalf("SVG attributes must not count as colour tokens, got %+v", c)
	}
}

func TestCheckPalette_ignoresAlignmentAndGradientUtilities(t *testing.T) {
	dir := t.TempDir()
	writePaletteFixture(t, dir, `theme: { extend: { colors: { ink: "#000" } } }`,
		`<p class="text-center text-left text-right text-justify bg-gradient-to-r from-ink to-ink">x</p>`)
	if c := checkPalette(dir); !c.OK {
		t.Fatalf("alignment/gradient utilities are not palette colours, got %+v", c)
	}
}

func TestCheckPalette_ignoresTypographyScaleKeys(t *testing.T) {
	dir := t.TempDir()
	config := `theme: { extend: {
		colors: { ink: "#000" },
		fontSize: { "headline-xl": ["2rem", { lineHeight: "2.4rem" }], "code-sm": ["0.875rem", {}] },
		fontFamily: { "code-md": ["monospace"] },
		boxShadow: { "card-lg": "0 1px 2px" }
	} }`
	writePaletteFixture(t, dir, config,
		`<p class="text-headline-xl text-code-sm font-code-md shadow-card-lg text-ink">x</p>`)
	if c := checkPalette(dir); !c.OK {
		t.Fatalf("fontSize/fontFamily/boxShadow keys are not palette colours, got %+v", c)
	}
}

// #318: composed colour tokens (on-surface, border-subtle, compliance-atencao)
// were cut at the first hyphen, so the check reported tokens that exist.
func TestMissingPaletteTokens_acceptsComposedColorKeys(t *testing.T) {
	markup := `<p class="text-on-surface border-border-subtle bg-compliance-atencao">x</p>`
	theme := tailwindThemeKeys{colors: map[string]bool{
		"on-surface": true, "border-subtle": true, "compliance-atencao": true,
	}}
	if missing := missingPaletteTokens(markup, theme); len(missing) != 0 {
		t.Fatalf("composed tokens in theme must not warn, got %v", missing)
	}
}

func TestMissingPaletteTokens_stillReportsUnknownComposedToken(t *testing.T) {
	markup := `<p class="text-on-surface">x</p>`
	theme := tailwindThemeKeys{colors: map[string]bool{}}
	missing := missingPaletteTokens(markup, theme)
	if len(missing) == 0 {
		t.Fatal("unknown composed token must still warn")
	}
}
