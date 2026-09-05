package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLayoutTemplates_containNavMarker(t *testing.T) {
	for name, tpl := range map[string]string{
		"full":    tplLayout,
		"minimal": tplLayoutMinimal,
		"blank":   tplLayoutBlank,
	} {
		if !strings.Contains(tpl, "<!-- cais:nav -->") {
			t.Errorf("%s layout missing <!-- cais:nav --> marker", name)
		}
	}
}

func TestLayoutTemplates_fullHasDefaultNavLinks(t *testing.T) {
	if !strings.Contains(tplLayout, `template "nav_links"`) {
		t.Error("full layout should render nav_links partial")
	}
	for _, link := range []string{`href="/contact"`, `href="/dashboard"`, `data-amarra-drive="true"`} {
		if !strings.Contains(tplPartialNavLinks, link) {
			t.Errorf("nav_links partial missing %s", link)
		}
	}
	if !strings.Contains(tplLayout, "amarra-toast-host") {
		t.Error("full layout missing amarra-toast-host")
	}
}

func TestLayoutTemplates_minimalAndBlankMatch(t *testing.T) {
	if tplLayoutMinimal != tplLayoutBlank {
		t.Error("minimal and blank base layouts should be identical")
	}
}

func TestLayoutTemplates_useSharedHead(t *testing.T) {
	for name, tpl := range map[string]string{
		"full":    tplLayout,
		"minimal": tplLayoutMinimal,
		"blank":   tplLayoutBlank,
	} {
		if !strings.Contains(tpl, "/static/js/amarra.js") || !strings.Contains(tpl, `define "app"`) {
			t.Errorf("%s layout missing shared head/shell fragments", name)
		}
		if strings.Contains(tpl, "htmx.min.js") || strings.Contains(tpl, `hx-ext`) {
			t.Errorf("%s layout must not load htmx", name)
		}
	}
}

func TestLayoutTemplates_hasDriveShell(t *testing.T) {
	for name, tpl := range map[string]string{
		"full":    tplLayout,
		"minimal": tplLayoutMinimal,
		"blank":   tplLayoutBlank,
	} {
		if !strings.Contains(tpl, `id="amarra-nav"`) {
			t.Errorf("%s layout missing amarra-nav id", name)
		}
		if !strings.Contains(tpl, `id="amarra-main"`) {
			t.Errorf("%s layout missing amarra-main shell", name)
		}
	}
}

func TestLayoutTemplates_navTabsHaveIcons(t *testing.T) {
	for _, icon := range []string{`icon_home_nav`, `icon_message_nav`, `icon_chart_nav`} {
		if !strings.Contains(tplPartialNavLinks, icon) {
			t.Errorf("nav partial should include %s", icon)
		}
	}
}

func TestScaffoldPartials_iconsRenderNonEmpty(t *testing.T) {
	dir := t.TempDir()
	data := scaffoldData{AppName: "demo", ModulePath: "github.com/acme/demo"}
	for path, tpl := range map[string]string{
		"web/templates/partials/icons.html":     tplPartialIcons,
		"web/templates/partials/nav_links.html": tplPartialNavLinks,
	} {
		if err := writeTemplate(filepath.Join(dir, path), tpl, data); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		body, err := os.ReadFile(filepath.Join(dir, path))
		if err != nil {
			t.Fatal(err)
		}
		if len(body) == 0 {
			t.Fatalf("%s rendered empty", path)
		}
		want := `define "icon_sparkles_md"`
		if path == "web/templates/partials/nav_links.html" {
			want = `define "nav_links"`
		}
		if !strings.Contains(string(body), want) {
			t.Fatalf("%s missing %s", path, want)
		}
	}
}

func TestLayoutTemplates_contactFormUsesKitForm(t *testing.T) {
	if !strings.Contains(tplPageContact, `<.form action="/contact"`) {
		t.Error("contact form should use <.form> kit tag")
	}
}

func TestLayoutTemplates_dashboardUsesLogoutForm(t *testing.T) {
	if !strings.Contains(tplPageDashboard, `action="/logout"`) {
		t.Error("dashboard page should post logout")
	}
}

func TestPageHome_greetingHarborLayout(t *testing.T) {
	for _, token := range []string{
		`data-testid="amarra-ready"`,
		"amarra-grain",
		"home.rails_heading",
		"home.cta_board",
		"home.berth_label",
		"home.manifest_label",
		"home.tide_mark",
	} {
		if !strings.Contains(tplPageHome, token) {
			t.Errorf("home page missing %q", token)
		}
	}
	if !strings.Contains(tplI18nEn, `"home.cta_board"`) || !strings.Contains(tplI18nEn, "Come aboard") {
		t.Error("en catalog missing home.cta_board")
	}
	if !strings.Contains(tplI18nPt, "Suba a bordo") {
		t.Error("pt catalog missing boarding CTA")
	}
	if !strings.Contains(tplLayoutBaseOpen, `href="/login"`) {
		t.Error("shell header should link to /login")
	}
	if !strings.Contains(tplHomeHandler, `"ActiveNav": "home"`) {
		t.Error("home handler should set ActiveNav")
	}
}

func TestScaffoldPages_dropIndigoAndHTMX(t *testing.T) {
	blobs := map[string]string{
		"login":     tplPageLogin,
		"signup":    tplPageSignup,
		"forgot":    tplPageForgotPassword,
		"reset":     tplPageResetPassword,
		"contact":   tplPageContact,
		"dashboard": tplPageDashboard,
		"home":      tplPageHome,
		"css":       tplInputCSS,
		"nav":       tplPartialNavLinks,
	}
	for name, blob := range blobs {
		for _, leftover := range []string{"indigo", "hx-ext", "htmx.min.js", "font-display"} {
			if strings.Contains(blob, leftover) {
				t.Errorf("%s still contains leftover %q", name, leftover)
			}
		}
	}
	for _, blob := range []string{tplPageLogin, tplPageContact, tplPageDashboard} {
		if !strings.Contains(blob, "border-copper") && !strings.Contains(blob, "text-copper") {
			t.Error("inner pages should use copper tokens on the harbor shell")
		}
	}
}

func TestLayoutTemplates_shellDesignTokens(t *testing.T) {
	for name, tpl := range map[string]string{
		"full":    tplLayout,
		"minimal": tplLayoutMinimal,
		"blank":   tplLayoutBlank,
	} {
		for _, token := range []string{
			"bg-ink",
			"text-foam",
			"text-copper",
			"no-scrollbar",
			"sticky top-0",
		} {
			if !strings.Contains(tpl, token) {
				t.Errorf("%s layout missing design token %q", name, token)
			}
		}
	}
}
