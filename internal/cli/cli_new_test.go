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
		"web/static/favicon.svg",
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
	if !strings.Contains(layoutBody, `href="/static/favicon.svg"`) {
		t.Error("layouts/app.html should use the docs boat favicon")
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
