package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTplPackageJSON_hasBuildScript(t *testing.T) {
	// #82 — deploy Go só compila o CSS via `npm run build`; sem ele o styles.css (gitignored) nunca chega ao servidor.
	if !strings.Contains(tplPackageJSON, `"build"`) {
		t.Error("package.json should define a build script for deploys")
	}
	if !strings.Contains(tplPackageJSON, "tailwindcss -i input.css -o web/static/css/styles.css --minify") {
		t.Error("build script should compile tailwind input.css into web/static/css/styles.css")
	}
}

func TestScaffoldTooling_CITriggersMain(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "ciapp")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "ciapp",
		ModulePath: "github.com/puppe1990/ciapp",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	ci, err := os.ReadFile(filepath.Join(appDir, ".github/workflows/ci.yml"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(ci)
	if !strings.Contains(body, "branches: [main]") {
		t.Errorf("ci.yml should trigger on main, got:\n%s", body)
	}
	if strings.Contains(body, "master") {
		t.Errorf("ci.yml should not trigger on master, got:\n%s", body)
	}
}

func TestScaffoldTooling_PrettierAllowsEmptyOrIgnoresTemplates(t *testing.T) {
	// #146 — committing only web/templates must not fail prettier hook
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "fmtapp")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "fmtapp",
		ModulePath: "github.com/puppe1990/fmtapp",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	pre, err := os.ReadFile(filepath.Join(appDir, ".pre-commit-config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(pre)
	hasEmptyOK := strings.Contains(body, "no-error-on-unmatched-pattern")
	hasExcludeTemplates := strings.Contains(body, "web/templates")
	if !hasEmptyOK && !hasExcludeTemplates {
		t.Error("pre-commit prettier must tolerate template-only commits (args or exclude)")
	}
}

func TestScaffoldTooling_PreCommitUsesGoimportsNotOnlyGofmt(t *testing.T) {
	// #148
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "impapp")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "impapp",
		ModulePath: "github.com/puppe1990/impapp",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	pre, err := os.ReadFile(filepath.Join(appDir, ".pre-commit-config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(pre)
	if strings.Contains(body, "entry: go fmt ./...") {
		t.Error("pre-commit must not use bare go fmt (misses goimports local-prefixes)")
	}
	if !strings.Contains(body, "goimports") {
		t.Error("pre-commit must run goimports to match CI golangci formatters")
	}
}

// #117: the scaffold tells users to put ADMIN_TOKEN/SMTP_PASSWORD in .env,
// but the generated .gitignore did not ignore it — the first `git add .`
// committed production secrets.
func TestScaffoldNewApp_gitignoreSkipsEnvSecrets(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "envapp")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "envapp",
		ModulePath: "github.com/puppe1990/envapp",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(appDir, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	lines := map[string]bool{}
	for _, line := range strings.Split(string(body), "\n") {
		lines[strings.TrimSpace(line)] = true
	}
	for _, want := range []string{".env", ".env.*", "!.env.example"} {
		if !lines[want] {
			t.Errorf(".gitignore missing %q:\n%s", want, body)
		}
	}
	if _, err := os.Stat(filepath.Join(appDir, ".env.example")); err != nil {
		t.Errorf(".env.example should still be scaffolded: %v", err)
	}
}

// #122: the scaffolded store must put the SQLite pragmas in the DSN, so every
// driver connection inherits foreign_keys/busy_timeout after a reconnect.
func TestScaffoldStore_opensWithDSNPragmas(t *testing.T) {
	for name, tpl := range map[string]string{"full": tplStore, "minimal": tplStoreMinimal} {
		if !strings.Contains(tpl, `sql.Open("sqlite", caissqlite.DSN(dsn))`) {
			t.Errorf("%s store template should build the DSN with caissqlite.DSN: %s", name, tpl[:min(len(tpl), 200)])
		}
	}
}
