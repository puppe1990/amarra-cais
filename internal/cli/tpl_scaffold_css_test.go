package cli

import (
	"strings"
	"testing"
)

func TestTplInputCSS_passwordToggleOverlaysInput(t *testing.T) {
	for _, want := range []string{
		".cais-password-wrap {",
		"@apply relative",
		"padding-right: 2.5rem",
		".cais-password-toggle {",
		"@apply absolute right-0 top-0",
		"border: none",
		"text-foam/70",
	} {
		if !strings.Contains(tplInputCSS, want) {
			t.Errorf("input.css missing password overlay %q", want)
		}
	}
	if strings.Contains(tplInputCSS, "text-slate-400 hover:text-slate-600") {
		t.Error("password toggle still uses slate instead of foam")
	}
}

func TestTplInputCSS_documentsBaseSelectSupports(t *testing.T) {
	for _, want := range []string{
		"@supports (appearance: base-select)",
		"appearance: base-select",
		":user-invalid",
	} {
		if !strings.Contains(tplInputCSS, want) {
			t.Errorf("input.css missing %q", want)
		}
	}
}

// #258: the palette lives in CSS variables so html.light (sidebar toggle) flips
// every kit class — bg-ink/text-foam/border-copper/bg-tide — not a few overrides.
func TestTplInputCSS_paletteVariablesSwitchTheme(t *testing.T) {
	dark := cssBlock(t, tplInputCSS, ":root {")
	light := cssBlock(t, tplInputCSS, "html.light {")
	for _, v := range []string{"--amarra-ink", "--amarra-foam", "--amarra-copper", "--amarra-tide"} {
		darkVal, ok := cssVar(dark, v)
		if !ok {
			t.Errorf(":root missing %s (#258)", v)
			continue
		}
		lightVal, ok := cssVar(light, v)
		if !ok {
			t.Errorf("html.light missing %s (#258)", v)
			continue
		}
		if darkVal == lightVal {
			t.Errorf("html.light does not change %s (%q) — the toggle would be a no-op", v, darkVal)
		}
	}
}

func TestTplTailwind_mapsPaletteToCSSVariables(t *testing.T) {
	for _, key := range []string{"ink", "foam", "copper", "tide"} {
		want := key + `: "rgb(var(--amarra-` + key + `) / <alpha-value>)"`
		if !strings.Contains(tplTailwind, want) {
			t.Errorf("tailwind.config.js missing %s so html.light can flip it (#258)", want)
		}
	}
}

func cssBlock(t *testing.T, css, selector string) string {
	t.Helper()
	start := strings.Index(css, selector)
	if start == -1 {
		t.Fatalf("input.css missing %q block", selector)
	}
	end := strings.Index(css[start:], "}")
	if end == -1 {
		t.Fatalf("input.css %q block is not closed", selector)
	}
	return css[start+len(selector) : start+end]
}

func cssVar(block, name string) (string, bool) {
	for _, line := range strings.Split(block, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, name+":") {
			continue
		}
		return strings.TrimSuffix(strings.TrimSpace(strings.TrimPrefix(line, name+":")), ";"), true
	}
	return "", false
}
