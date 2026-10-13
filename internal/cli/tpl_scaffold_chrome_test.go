package cli

import (
	"strings"
	"testing"
)

// Shell chrome shared by the anonymous header and the signed-in sidebar: the
// theme toggle, the brand version, the footer gate and the sign-out position.
// Split from tpl_scaffold_layout_test.go to stay under the file cap (#95).

// The default toggle shows the icon of the theme it switches to — sun while the
// app is dark, moon while it is light — next to the label. The swap is CSS-only
// (`html.light`), so the hook keeps owning just the label slot (#258).
func TestLayoutTemplates_themeToggleSwapsIcon(t *testing.T) {
	for _, name := range []string{"icon_sun", "icon_moon"} {
		if !strings.Contains(tplPartialIcons, `define "`+name+`"`) {
			t.Errorf("layout sprite missing %q", name)
		}
	}
	for _, token := range []string{
		`{{"{{"}} template "icon_sun" . {{"}}"}}`,
		`{{"{{"}} template "icon_moon" . {{"}}"}}`,
		`[html.light_&]:hidden`,
		`[html.light_&]:block`,
	} {
		if !strings.Contains(tplLayoutThemeToggle, token) {
			t.Errorf("theme toggle missing %q", token)
		}
	}
}

// The brand carries the installed framework version — the same string the boot
// banner and doctor report — so a running app can be matched to its go.mod.
func TestLayoutTemplates_brandShowsFrameworkVersion(t *testing.T) {
	for _, token := range []string{
		`Amarra v{{"{{"}} .FrameworkVersion {{"}}"}}`,
		`title="Cais v{{"{{"}} .FrameworkVersionFull {{"}}"}}"`,
	} {
		if !strings.Contains(tplLayoutBrand, token) {
			t.Errorf("brand missing %q", token)
		}
	}
	for _, token := range []string{
		"boot.CaisVersion()",
		`"FrameworkVersion"`,
		`"FrameworkVersionFull"`,
	} {
		if !strings.Contains(tplViewData, token) {
			t.Errorf("amarraData missing %q for the brand version", token)
		}
	}
}

// The sidebar footer ends with sign out: the locale and theme controls sit
// above it, so the last thing in the rail is not the destructive action.
func TestLayoutTemplates_signOutLastInSidebar(t *testing.T) {
	for name, tpl := range map[string]string{
		"full":    tplLayout,
		"minimal": tplLayoutMinimal,
		"blank":   tplLayoutBlank,
	} {
		aside := navAside(t, name, tpl)
		logout := strings.Index(aside, `action="/logout"`)
		if logout == -1 {
			t.Fatalf("%s layout sidebar has no logout form", name)
		}
		for _, token := range []string{`<.locale-toggle`, `amarra-hook="theme"`} {
			at := strings.Index(aside, token)
			if at == -1 || at > logout {
				t.Errorf("%s layout: %s must render above sign out (#258)", name, token)
			}
		}
	}
}

// The footer is visitor chrome: a signed-in page ends at the sidebar, so the
// app shell does not carry the marketing strip.
func TestLayoutTemplates_footerOnlyForAnonymousShell(t *testing.T) {
	marker := `{{"{{"}} if not .Site.LoggedIn {{"}}"}}`
	idx := strings.Index(tplLayoutBaseClose, "<footer")
	if idx == -1 {
		t.Fatal("layout has no footer")
	}
	open := strings.LastIndex(tplLayoutBaseClose[:idx], marker)
	if open == -1 {
		t.Fatal("footer must sit inside a `if not .Site.LoggedIn` gate")
	}
	if strings.Contains(tplLayoutBaseClose[open+len(marker):idx], `{{"{{"}} end {{"}}"}}`) {
		t.Error("footer gate closes before the footer — the footer still renders for signed-in pages")
	}
	if !strings.Contains(tplLayoutBaseClose[idx:], `{{"{{"}} end {{"}}"}}`) {
		t.Error("footer gate is never closed")
	}
}
