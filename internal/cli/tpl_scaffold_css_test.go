package cli

import (
	"os"
	"os/exec"
	"path/filepath"
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

// #372: the kit HTML lives inside the Go module, so a scaffold that scans only
// web/templates/** never emits the kit utilities (bg-copper, px-3, …) and
// input/button/flash render unstyled.
func TestTplTailwind_scansShippedKit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tailwind.config.js")
	if err := writeTemplate(path, tplTailwind, scaffoldData{}); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"shippedKitContent()",
		"go list -m -f {{.Dir}} github.com/puppe1990/amarra-cais",
		"/pkg/amarra/view/components/**/*.html",
	} {
		if !strings.Contains(string(body), want) {
			t.Errorf("tailwind.config.js missing %q:\n%s", want, body)
		}
	}

	out, err := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/puppe1990/amarra-cais").Output()
	if err != nil {
		t.Fatalf("go list module dir: %v", err)
	}
	matches, err := filepath.Glob(filepath.Join(strings.TrimSpace(string(out)), "pkg/amarra/view/components/*.html"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) < 20 {
		t.Errorf("kit glob resolved %d components, want the shipped kit", len(matches))
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

// Default design = the published docs-site palette (harbor night + cream paper,
// sea-teal accent). The keys stay ink/foam/copper/tide (kit + doctor contract);
// only the values are the documentation look.
func TestTplInputCSS_defaultPaletteMatchesDocsSite(t *testing.T) {
	dark := cssBlock(t, tplInputCSS, ":root {")
	light := cssBlock(t, tplInputCSS, "html.light {")
	for name, pair := range map[string][2]string{
		"--amarra-ink":    {"8 19 24", "247 241 230"},
		"--amarra-foam":   {"244 234 216", "16 33 38"},
		"--amarra-copper": {"31 139 163", "15 107 128"},
		"--amarra-tide":   {"18 36 40", "228 217 196"},
	} {
		if got, ok := cssVar(dark, name); !ok || got != pair[0] {
			t.Errorf(":root %s = %q, want docs palette %q", name, got, pair[0])
		}
		if got, ok := cssVar(light, name); !ok || got != pair[1] {
			t.Errorf("html.light %s = %q, want docs palette %q", name, got, pair[1])
		}
	}
}

func TestTplInputCSS_shipsCoralAccent(t *testing.T) {
	for _, selector := range []string{":root {", "html.light {"} {
		block := cssBlock(t, tplInputCSS, selector)
		got, ok := cssVar(block, "--amarra-coral")
		if !ok || got != "224 122 74" {
			t.Errorf("%s --amarra-coral = %q, want docs coral %q", selector, got, "224 122 74")
		}
	}
}

func TestTplTailwind_shipsCoralAndFraunces(t *testing.T) {
	for _, want := range []string{
		`coral: "rgb(var(--amarra-coral) / <alpha-value>)"`,
		`"Fraunces"`,
	} {
		if !strings.Contains(tplTailwind, want) {
			t.Errorf("tailwind.config.js missing %q", want)
		}
	}
}

// The docs palette is harbor/teal: no leftover copper literals in the scaffold
// markup or CSS, otherwise the accent reads amber next to the teal classes.
func TestTplScaffold_hasNoCopperAccentLeftovers(t *testing.T) {
	for name, blob := range map[string]string{
		"home":   tplPageHome,
		"dash":   tplPageDashboard,
		"login":  tplPageLogin,
		"layout": tplLayout,
		"css":    tplInputCSS,
	} {
		for _, leftover := range []string{"201,137,58", "201 137 58", "#c9893a", "#081014", "#f6f1e7"} {
			if strings.Contains(blob, leftover) {
				t.Errorf("%s still carries the old copper palette value %q", name, leftover)
			}
		}
	}
	if !strings.Contains(tplLayout, `content="#081318"`) {
		t.Error("layout theme-color meta must carry the harbor surface #081318")
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
