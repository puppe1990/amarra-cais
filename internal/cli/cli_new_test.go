package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLI_NewMinimalCreatesSlimApp(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "slim")

	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "slim",
		ModulePath: "github.com/puppe1990/slim",
	}, true, false); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		"internal/handlers/home.go",
		"go.mod",
	} {
		if _, err := os.Stat(filepath.Join(appDir, path)); err != nil {
			t.Errorf("missing %s: %v", path, err)
		}
	}

	for _, path := range []string{
		"internal/handlers/contact.go",
		"internal/handlers/dashboard.go",
		"internal/store/migrations/001_contacts.sql",
	} {
		if _, err := os.Stat(filepath.Join(appDir, path)); err == nil {
			t.Errorf("minimal app should not have %s", path)
		}
	}
}

func TestScaffoldNewApp_includesAgentsMD(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	for _, tc := range []struct {
		name           string
		minimal, blank bool
	}{
		{"full", false, false},
		{"minimal", true, false},
		{"blank", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			appDir := filepath.Join(t.TempDir(), tc.name)
			if err := scaffoldNewApp(appDir, scaffoldData{
				AppName:    tc.name,
				ModulePath: "github.com/puppe1990/" + tc.name,
			}, tc.minimal, tc.blank); err != nil {
				t.Fatal(err)
			}
			body, err := os.ReadFile(filepath.Join(appDir, "AGENTS.md"))
			if err != nil {
				t.Fatalf("missing AGENTS.md: %v", err)
			}
			text := string(body)
			for _, needle := range []string{
				"TDD",
				"Amarra",
				"view.Write",
				"flash.Set",
				"amarra-cais g",
				"internal/handlers",
				"web/templates/pages",
				`amarra-hook="password"`,
				`localStorage.getItem("amarra-theme")`,
				tc.name, // AppName rendered into title
			} {
				if !strings.Contains(text, needle) {
					t.Errorf("AGENTS.md missing %q", needle)
				}
			}
			for _, stale := range []string{"Inertia", "web/src/pages"} {
				if strings.Contains(text, stale) {
					t.Errorf("AGENTS.md still mentions %q", stale)
				}
			}
		})
	}
}

func TestScaffoldNewApp_includesLocaleToggle(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "localeapp")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "localeapp",
		ModulePath: "github.com/puppe1990/localeapp",
	}, false, false); err != nil {
		t.Fatal(err)
	}
	layout, err := os.ReadFile(filepath.Join(appDir, "web/templates/layouts/app.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(layout), `<.locale-toggle`) {
		t.Error("layouts/app.html should include <.locale-toggle />")
	}
	routes, err := os.ReadFile(filepath.Join(appDir, "internal/app/routes.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(routes), `r.Post("/locale"`) {
		t.Error("routes.go should register POST /locale")
	}
	if _, err := os.Stat(filepath.Join(appDir, "internal/handlers/locale.go")); err != nil {
		t.Errorf("missing locale handler: %v", err)
	}
}

func TestScaffoldNewApp_htmlFirstNoInertia(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "demo")
	if err := scaffoldNewApp(appDir, scaffoldData{AppName: "demo", ModulePath: "example.com/demo", CaisVersion: "0.1.0"}, false, false); err != nil {
		t.Fatal(err)
	}
	mustExist := []string{
		"web/templates/layouts/app.html",
		"web/templates/pages/home.html",
		"web/templates/pages/login.html",
		"web/templates/components/.gitkeep",
		"web/static/js/amarra.js",
		"internal/handlers/home.go",
	}
	mustNot := []string{
		"vite.config.js",
		"svelte.config.js",
		"web/src/pages/Home.svelte",
		"web/src/main.js",
		"internal/handlers/inertia_test.go",
	}
	for _, p := range mustExist {
		if _, err := os.Stat(filepath.Join(appDir, p)); err != nil {
			t.Errorf("missing %s", p)
		}
	}
	for _, p := range mustNot {
		if _, err := os.Stat(filepath.Join(appDir, p)); err == nil {
			t.Errorf("should not exist %s", p)
		}
	}
	gomod, _ := os.ReadFile(filepath.Join(appDir, "go.mod"))
	if strings.Contains(string(gomod), "gonertia") {
		t.Error("go.mod still has gonertia")
	}
	if !strings.Contains(string(gomod), "github.com/puppe1990/amarra-cais") {
		t.Error("go.mod missing amarra-cais")
	}
	home, _ := os.ReadFile(filepath.Join(appDir, "internal/handlers/home.go"))
	if strings.Contains(string(home), "inertia") {
		t.Error("home handler still uses inertia")
	}
	if !strings.Contains(string(home), "view.Write") {
		t.Error("home handler missing view.Write")
	}
}

func TestScaffoldNewApp_driveFlashLoginEmailHomeTest(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "drivefix")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "drivefix",
		ModulePath: "github.com/puppe1990/drivefix",
	}, false, false); err != nil {
		t.Fatal(err)
	}

	login, err := os.ReadFile(filepath.Join(appDir, "web/templates/pages/login.html"))
	if err != nil {
		t.Fatal(err)
	}
	loginBody := string(login)
	if strings.Contains(loginBody, `value="{{ if .Email }}`) {
		t.Error("login email value attr must not embed {{ if .Email }} (ExpandAll treats it as static)")
	}
	if !strings.Contains(loginBody, `value="{{ .Email }}"`) {
		t.Error(`login email should use value="{{ .Email }}"`)
	}

	auth, err := os.ReadFile(filepath.Join(appDir, "internal/handlers/auth.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(auth), `"Email":`) {
		t.Error("auth handler must set Email on login GET and 422")
	}
	if !strings.Contains(string(auth), "demo@example.com") {
		t.Error("login GET should prefill demo@example.com outside production")
	}

	layout, err := os.ReadFile(filepath.Join(appDir, "web/templates/layouts/app.html"))
	if err != nil {
		t.Fatal(err)
	}
	layoutBody := string(layout)
	mainIdx := strings.Index(layoutBody, `id="amarra-main"`)
	if mainIdx < 0 {
		t.Fatal("layouts/app.html missing #amarra-main")
	}
	flashIdx := strings.Index(layoutBody, `<.flash`)
	if flashIdx < 0 {
		t.Fatal("layouts/app.html missing <.flash>")
	}
	closeMain := strings.Index(layoutBody[mainIdx:], "</main>")
	if closeMain < 0 {
		t.Fatal("layouts/app.html missing </main>")
	}
	if flashIdx < mainIdx || flashIdx > mainIdx+closeMain {
		t.Error("<.flash /> must be inside #amarra-main so Drive morph keeps notices")
	}
	if strings.Contains(layoutBody, "}}-") {
		t.Error("layout has stray '-' after }} (define/end should use {{- … -}})")
	}

	homeTest, err := os.ReadFile(filepath.Join(appDir, "internal/handlers/home_test.go"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(homeTest), "You're on Cais!") {
		t.Error("home test must not assert unescaped You're (html/template emits &#39;)")
	}
	if !strings.Contains(string(homeTest), "made landfall") {
		t.Error("home test should still assert the heading")
	}
}

func TestCLI_NewCreatesApp(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "myapp")

	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "myapp",
		ModulePath: "github.com/puppe1990/myapp",
	}, false, false); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		"go.mod",
		"cmd/server/main.go",
		"internal/i18n/en.go",
		"internal/i18n/pt.go",
		".env.example",
		"internal/handlers/dashboard.go",
		"internal/handlers/viewdata.go",
		"web/templates/layouts/app.html",
		"web/templates/pages/home.html",
		"web/templates/pages/contact.html",
		"web/templates/pages/dashboard.html",
		"web/templates/pages/login.html",
		"web/static/js/amarra.js",
		"package.json",
		"web/static/manifest.webmanifest",
		"web/static/js/sw.js",
		"web/static/img/go-on-cais.jpg",
		"web/static/og.png",
		"web/static/icons/icon.png",
	} {
		if _, err := os.Stat(filepath.Join(appDir, path)); err != nil {
			t.Errorf("missing %s: %v", path, err)
		}
	}

	for _, path := range []string{
		"web/static/js/htmx.min.js",
		"web/static/js/cais.js",
		"vite.config.js",
		"svelte.config.js",
		"web/src/pages/Home.svelte",
		"internal/handlers/inertia_test.go",
	} {
		if _, err := os.Stat(filepath.Join(appDir, path)); err == nil {
			t.Errorf("Inertia/HTMX scaffold artifact should not exist: %s", path)
		}
	}

	appGo, err := os.ReadFile(filepath.Join(appDir, "internal/app/app.go"))
	if err != nil {
		t.Fatal(err)
	}
	appGoBody := string(appGo)
	if !strings.Contains(appGoBody, "Views     *view.Renderer") {
		t.Error("app.go should wire *view.Renderer in Deps")
	}
	if !strings.Contains(appGoBody, `r.Handle("GET /amarra/live", hub.Handler())`) {
		t.Error("app.New must register live hub as GET /amarra/live")
	}
	if !strings.Contains(appGoBody, "registerLiveViews") {
		t.Error("app.New must call registerLiveViews")
	}
	if !strings.Contains(appGoBody, "jobsui.Register") {
		t.Error("app.go should mount the localhost /jobs dashboard")
	}

	storeGo, err := os.ReadFile(filepath.Join(appDir, "internal/store/store.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(storeGo), "DB() *sql.DB") {
		t.Error("Store interface should expose DB() for jobsui.Register")
	}

	gomod, err := os.ReadFile(filepath.Join(appDir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	gomodBody := string(gomod)
	if strings.Contains(gomodBody, "gonertia") {
		t.Error("go.mod must not require gonertia")
	}
	if !strings.Contains(gomodBody, "github.com/puppe1990/amarra-cais v"+defaultScaffoldCaisVersion) &&
		!strings.Contains(gomodBody, "github.com/puppe1990/amarra-cais v") {
		t.Errorf("go.mod should require a current amarra-cais version, got:\n%s", gomodBody)
	}

	login, err := os.ReadFile(filepath.Join(appDir, "web/templates/pages/login.html"))
	if err != nil {
		t.Fatal(err)
	}
	loginBody := string(login)
	if !strings.Contains(loginBody, `action="/login"`) {
		t.Error("login.html should post to /login")
	}
	if !strings.Contains(loginBody, `<.form`) {
		t.Error("login.html should use kit <.form")
	}
	if strings.Contains(loginBody, "csrfField") {
		t.Error("login.html must not duplicate csrfField inside <.form> (#37)")
	}
	contact, err := os.ReadFile(filepath.Join(appDir, "web/templates/pages/contact.html"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(contact), "csrfField") {
		t.Error("contact.html must not duplicate csrfField inside <.form> (#37)")
	}
	if !strings.Contains(loginBody, `<.password name="password"`) {
		t.Error("login.html should use the <.password> kit with eye toggle (#29)")
	}

	auth, err := os.ReadFile(filepath.Join(appDir, "internal/handlers/auth.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(auth), "httpx.ParseFormOrJSON") {
		t.Error("auth handler should use httpx.ParseFormOrJSON")
	}
	if strings.Contains(string(auth), "r.ParseForm()") {
		t.Error("auth handler must not call r.ParseForm() alone (breaks JSON bodies)")
	}

	appHTML, err := os.ReadFile(filepath.Join(appDir, "web/templates/layouts/app.html"))
	if err != nil {
		t.Fatal(err)
	}
	layoutBody := string(appHTML)
	if !strings.Contains(layoutBody, "/static/js/amarra.js") {
		t.Error("layouts/app.html should load amarra.js")
	}
	if !strings.Contains(layoutBody, `id="amarra-main"`) {
		t.Error("layouts/app.html should include #amarra-main")
	}
	if strings.Contains(layoutBody, "hx-ext") || strings.Contains(layoutBody, "htmx") {
		t.Error("layouts/app.html must not load htmx")
	}
	if !strings.Contains(layoutBody, `localStorage.getItem("amarra-theme")`) {
		t.Error("layouts/app.html should include the theme FOUC snippet before CSS")
	}

	dash, err := os.ReadFile(filepath.Join(appDir, "web/templates/pages/dashboard.html"))
	if err != nil {
		t.Fatal(err)
	}
	dashBody := string(dash)
	if !strings.Contains(dashBody, `action="/logout"`) {
		t.Error("Dashboard logout should post to /logout")
	}
	if strings.Contains(dashBody, "csrfField") {
		t.Error("dashboard.html must not duplicate csrfField inside <.form> (#37)")
	}

	pkg, err := os.ReadFile(filepath.Join(appDir, "package.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(pkg), "@inertiajs/svelte") || strings.Contains(string(pkg), "vite") {
		t.Error("package.json should be Tailwind-only (no Vite/Svelte)")
	}

	css, err := os.ReadFile(filepath.Join(appDir, "input.css"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(css), "fonts.googleapis.com") {
		t.Error("input.css should not import Google Fonts (CSP blocked)")
	}
}

func TestScaffoldNewApp_i18nIncludesSignupKeys(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "i18napp")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "i18napp",
		ModulePath: "github.com/puppe1990/i18napp",
	}, false, false); err != nil {
		t.Fatal(err)
	}
	en, err := os.ReadFile(filepath.Join(appDir, "internal/i18n/en.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{
		`"auth.signup_prompt"`,
		`"auth.signup_title"`,
		`"auth.signup_submit"`,
		`"auth.login_prompt"`,
		`"auth.email_taken"`,
	} {
		if !strings.Contains(string(en), key) {
			t.Errorf("internal/i18n/en.go missing %s", key)
		}
	}
}

func TestScaffold_InputCSSIncludesHTMXStyles(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "styles")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "styles",
		ModulePath: "github.com/puppe1990/styles",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	css, err := os.ReadFile(filepath.Join(appDir, "input.css"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(css)
	for _, needle := range []string{
		".no-scrollbar",
		".cais-toast-enter", ".cais-skeleton",
		".cais-chat-scroll-down", ".cais-msg-time", ".cais-thinking-dots",
		".amarra-grain",
	} {
		if !strings.Contains(body, needle) {
			t.Errorf("input.css missing %q", needle)
		}
	}
	for _, leftover := range []string{
		".htmx-swapping", ".htmx-settling", ".htmx-indicator", ".htmx-request",
		"from-indigo", "text-indigo", "bg-indigo",
	} {
		if strings.Contains(body, leftover) {
			t.Errorf("input.css still has leftover %q", leftover)
		}
	}
	tailwind, err := os.ReadFile(filepath.Join(appDir, "tailwind.config.js"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(tailwind), "fonts.googleapis.com") {
		t.Error("tailwind.config.js should not reference Google Fonts")
	}
	if !strings.Contains(string(tailwind), "system-ui") {
		t.Error("tailwind.config.js should use system font stack")
	}
}

func TestScaffold_IncludesQualityTooling(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")

	for _, tc := range []struct {
		name           string
		minimal, blank bool
	}{
		{"full", false, false},
		{"minimal", true, false},
		{"blank", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			appDir := filepath.Join(t.TempDir(), tc.name)
			if err := scaffoldNewApp(appDir, scaffoldData{
				AppName:    tc.name,
				ModulePath: "github.com/puppe1990/" + tc.name,
			}, tc.minimal, tc.blank); err != nil {
				t.Fatal(err)
			}

			for _, path := range []string{
				".github/workflows/ci.yml",
				".pre-commit-config.yaml",
				".golangci.yml",
				".prettierrc.json",
				".prettierignore",
			} {
				if _, err := os.Stat(filepath.Join(appDir, path)); err != nil {
					t.Errorf("missing %s: %v", path, err)
				}
			}

			makefile, err := os.ReadFile(filepath.Join(appDir, "Makefile"))
			if err != nil {
				t.Fatal(err)
			}
			body := string(makefile)
			for _, target := range []string{"lint:", "format-check:", "pre-commit-install:", "ci:"} {
				if !strings.Contains(body, target) {
					t.Errorf("Makefile missing target %s", target)
				}
			}

			golangci, err := os.ReadFile(filepath.Join(appDir, ".golangci.yml"))
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(golangci), "github.com/puppe1990/"+tc.name) {
				t.Error(".golangci.yml missing module local-prefix")
			}

			ci, err := os.ReadFile(filepath.Join(appDir, ".github/workflows/ci.yml"))
			if err != nil {
				t.Fatal(err)
			}
			ciBody := string(ci)
			for _, needle := range []string{"go test", "golangci-lint", "prettier", "npm test"} {
				if !strings.Contains(ciBody, needle) {
					t.Errorf("ci.yml missing %q", needle)
				}
			}

			pkg, err := os.ReadFile(filepath.Join(appDir, "package.json"))
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(pkg), `"test"`) {
				t.Error("package.json missing test script")
			}
		})
	}
}

func TestScaffoldNewApp_ContactHandlerValidatesName(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "contactapp")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "contactapp",
		ModulePath: "github.com/puppe1990/contactapp",
	}, false, false); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(appDir, "internal/handlers/contact.go"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(body)
	if !strings.Contains(s, `errs.Add("name"`) {
		t.Errorf("contact handler missing name validation: %s", s)
	}
	if !strings.Contains(s, `contact.name_required`) {
		t.Errorf("contact handler missing name_required i18n key: %s", s)
	}
}

func TestScaffoldBlankApp_IncludesSecurityMiddleware(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "blankapp")
	if err := scaffoldNewApp(appDir, scaffoldData{AppName: "blankapp", ModulePath: "github.com/puppe1990/blankapp"}, false, true); err != nil {
		t.Fatal(err)
	}
	body, _ := os.ReadFile(filepath.Join(appDir, "internal/app/app.go"))
	s := string(body)
	for _, want := range []string{
		"middleware.Recover",
		"middleware.SecurityHeaders(cfg)",
		"ReadHeaderTimeout",
		"ReadTimeout",
		"r.Static",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("blank app missing %q in app.go", want)
		}
	}
}

func TestScaffoldBlankApp_IncludesSessionMiddleware(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "blankapp")
	if err := scaffoldNewApp(appDir, scaffoldData{AppName: "blankapp", ModulePath: "github.com/puppe1990/blankapp"}, false, true); err != nil {
		t.Fatal(err)
	}
	appGo, _ := os.ReadFile(filepath.Join(appDir, "internal/app/app.go"))
	s := string(appGo)
	for _, want := range []string{
		"middleware.LoadSession(deps.Store.Sessions())",
		"r.Use(middleware.Flash(cfg))",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("blank app missing %q in app.go", want)
		}
	}
	storeGo, _ := os.ReadFile(filepath.Join(appDir, "internal/store/store.go"))
	if !strings.Contains(string(storeGo), "Sessions() session.Store") {
		t.Error("blank store missing Sessions() on interface")
	}
}

func TestCLI_NewBlankCreatesEmptyApp(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "empty")

	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "empty",
		ModulePath: "github.com/puppe1990/empty",
	}, false, true); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		"go.mod",
		"cmd/server/main.go",
		"internal/app/app.go",
		"internal/app/routes.go",
	} {
		if _, err := os.Stat(filepath.Join(appDir, path)); err != nil {
			t.Errorf("missing %s: %v", path, err)
		}
	}

	for _, path := range []string{
		"internal/handlers/home.go",
		"web/templates/layouts/app.html",
		"web/templates/pages/home.html",
		"web/static/js/amarra.js",
	} {
		if _, err := os.Stat(filepath.Join(appDir, path)); err != nil {
			t.Errorf("blank app missing HTML file %s: %v", path, err)
		}
	}

	for _, path := range []string{
		"internal/handlers/contact.go",
		"internal/handlers/dashboard.go",
		"internal/models/contact.go",
		"internal/store/migrations/001_contacts.sql",
		"web/templates/pages/contact.html",
		"web/static/js/htmx.min.js",
		"vite.config.js",
	} {
		if _, err := os.Stat(filepath.Join(appDir, path)); err == nil {
			t.Errorf("blank app should not have %s", path)
		}
	}

	routesBody, err := os.ReadFile(filepath.Join(appDir, "internal/app/routes.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(routesBody), "home.ServeHTTP") {
		t.Error("blank app routes should register welcome home handler")
	}
}

func TestScaffoldNewApp_CustomModule(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "myapp")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "myapp",
		ModulePath: "github.com/acme/myapp",
	}, false, false); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(appDir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "module github.com/acme/myapp") {
		t.Errorf("go.mod missing custom module path: %s", body)
	}
}

func TestCLI_New_CustomModule(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	root := t.TempDir()
	appDir := filepath.Join(root, "myapp")

	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	if err := c.Run([]string{"new", "myapp", appDir, "--module", "github.com/acme/myapp"}); err != nil {
		t.Fatal(err)
	}

	body, err := os.ReadFile(filepath.Join(appDir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "module github.com/acme/myapp") {
		t.Errorf("go.mod missing custom module path: %s", body)
	}
}

func TestCLI_New_CustomModule_DefaultWhenOmitted(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	root := t.TempDir()
	appDir := filepath.Join(root, "cool-app")

	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	if err := c.Run([]string{"new", "cool-app", appDir}); err != nil {
		t.Fatal(err)
	}

	body, err := os.ReadFile(filepath.Join(appDir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "module github.com/puppe1990/coolapp") {
		t.Errorf("go.mod missing default module path: %s", body)
	}
}

func TestCLI_New_ModuleRequiresValue(t *testing.T) {
	c := &CLI{Out: os.Stdout}
	if err := c.Run([]string{"new", "myapp", "--module"}); err == nil {
		t.Fatal("expected error for --module without value")
	}
}

func TestCLI_New_unknownFlag_errors(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := t.TempDir()
	t.Chdir(dir)

	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	err := c.Run([]string{"new", "--seed"})
	if err == nil {
		t.Fatal("expected error for unknown flag --seed")
	}
	if !strings.Contains(err.Error(), "unknown flag --seed") {
		t.Errorf("error = %v, want unknown flag --seed", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "--seed")); err == nil {
		t.Fatal("unknown flag must not create a --seed directory")
	}
}

func TestParseNewArgs_bareHelpIsAppName(t *testing.T) {
	opts, err := parseNewArgs([]string{"help"})
	if err != nil {
		t.Fatal(err)
	}
	if opts.name != "help" || opts.dir != "help" {
		t.Errorf("got name=%q dir=%q, want help", opts.name, opts.dir)
	}
}

func TestParseNewArgs_knownFlags(t *testing.T) {
	opts, err := parseNewArgs([]string{"myapp", "outdir", "--minimal", "--blank", "--module", "github.com/acme/x"})
	if err != nil {
		t.Fatal(err)
	}
	if opts.name != "myapp" || opts.dir != "outdir" {
		t.Errorf("got name=%q dir=%q, want myapp/outdir", opts.name, opts.dir)
	}
	if !opts.minimal || !opts.blank || opts.module != "github.com/acme/x" {
		t.Errorf("got %+v, want minimal+blank with custom module", opts)
	}
}

func TestCLI_NewMainUsesTemplateHotReload(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "hotreload")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "hotreload",
		ModulePath: "github.com/puppe1990/hotreload",
	}, false, false); err != nil {
		t.Fatal(err)
	}
	mainGo, err := os.ReadFile(filepath.Join(appDir, "cmd/server/main.go"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(mainGo)
	if !strings.Contains(body, "view.Load") {
		t.Error("main.go should use view.Load for Amarra templates")
	}
	air, err := os.ReadFile(filepath.Join(appDir, ".air.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(air), `"html"`) {
		t.Error(".air.toml include_ext must include html so air rebuilds (and re-embeds) templates")
	}
}

func TestCLI_NewIncludesAmarraHTML(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "full")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "full",
		ModulePath: "github.com/puppe1990/full",
	}, false, false); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{
		"web/templates/layouts/app.html",
		"web/templates/pages/home.html",
		"web/static/js/amarra.js",
		".air.toml",
	} {
		if _, err := os.Stat(filepath.Join(appDir, path)); err != nil {
			t.Errorf("missing %s: %v", path, err)
		}
	}
}
