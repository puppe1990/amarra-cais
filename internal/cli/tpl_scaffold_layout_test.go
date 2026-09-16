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
		if !strings.Contains(tpl, `data-amarra-layout="app"`) {
			t.Errorf("%s layout missing data-amarra-layout marker", name)
		}
	}
}

func TestScaffoldPartials_iconsRenderNonEmpty(t *testing.T) {
	dir := t.TempDir()
	data := scaffoldData{AppName: "demo", ModulePath: "github.com/acme/demo"}
	for path, tpl := range map[string]string{
		"web/templates/partials/icons.html": tplPartialIcons,
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

func TestLayoutTemplates_dashboardUsesStat(t *testing.T) {
	if !strings.Contains(tplPageDashboard, `<.stat`) {
		t.Error("dashboard should use kit <.stat> for KPIs (#42)")
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

func TestLayoutTemplates_sidebarShell(t *testing.T) {
	for name, tpl := range map[string]string{
		"full":    tplLayout,
		"minimal": tplLayoutMinimal,
		"blank":   tplLayoutBlank,
	} {
		for _, token := range []string{
			`<aside id="amarra-nav"`,
			`fixed left-0 top-[57px] bottom-0`,
			`w-60`,
			`lg:ml-60`,
			`amarra-sidebar-toggle`,
			`peer-checked:translate-x-0`,
			`amarra-hook="nav"`,
			`data-amarra-nav-on`,
			`data-amarra-nav-off`,
			`href="/dashboard"`,
			`<.form action="/logout"`,
			`<.flash />`,
			`amarra-toast-host`,
			`template "icon_chart_nav"`,
			`<!-- cais:nav -->`,
		} {
			if !strings.Contains(tpl, token) {
				t.Errorf("%s layout missing sidebar token %q", name, token)
			}
		}
		for _, gone := range []string{
			`<nav id="amarra-nav"`,
			`template "nav_links"`,
			`href="/contact"`,
		} {
			if strings.Contains(tpl, gone) {
				t.Errorf("%s layout should not contain %q", name, gone)
			}
		}
		if strings.Contains(tpl, ">Home<") {
			t.Errorf("%s layout sidebar should not contain a Home item", name)
		}
		if !strings.Contains(tpl, `id="amarra-main" class="flex-grow px-4 sm:px-6 lg:px-8 py-5 lg:ml-60"`) {
			t.Errorf("%s layout should tie lg:ml-60 to #amarra-main", name)
		}
		if !strings.Contains(tpl, "<aside") {
			t.Errorf("%s layout missing <aside sidebar", name)
			continue
		}
		if strings.Index(tpl, "<aside") > strings.Index(tpl, "<!-- cais:nav -->") {
			t.Errorf("%s layout should render <!-- cais:nav --> inside the sidebar", name)
		}
		if strings.Index(tpl, "<!-- cais:nav -->") >= strings.Index(tpl, "</aside>") {
			t.Errorf("%s layout should render <!-- cais:nav --> before </aside>", name)
		}
		side := tpl[strings.Index(tpl, "<aside"):strings.Index(tpl, "</aside>")]
		for _, token := range []string{
			`amarra-hook="nav"`,
			`data-amarra-nav-on`,
			`data-amarra-nav-off`,
		} {
			if !strings.Contains(side, token) {
				t.Errorf("%s layout sidebar should contain %q", name, token)
			}
		}
	}
}
