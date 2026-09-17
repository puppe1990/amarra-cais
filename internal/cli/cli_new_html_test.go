package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
				"Fullbleed",
				"page-owned chrome",
				"IIFE",
				"event delegation",
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
	if !strings.Contains(body, "signal.NotifyContext") {
		t.Error("main.go should wire signal.NotifyContext so air's restart signal shuts down gracefully (#77)")
	}
	if !strings.Contains(body, "RunContext") {
		t.Error("main.go should run the app with RunContext so SIGINT/SIGTERM releases the port and sqlite (#77)")
	}
	air, err := os.ReadFile(filepath.Join(appDir, ".air.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(air), `"html"`) {
		t.Error(".air.toml include_ext must include html so air rebuilds (and re-embeds) templates")
	}
	if !strings.Contains(string(air), "send_interrupt") {
		t.Error(".air.toml must set send_interrupt so air SIGINTs tmp/main before killing it (#77)")
	}
	if !strings.Contains(string(air), "kill_delay") {
		t.Error(".air.toml must set kill_delay so tmp/main can exit before SIGKILL (#77)")
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
