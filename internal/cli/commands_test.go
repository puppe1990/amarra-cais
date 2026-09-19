package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCLI_Help_IncludesAppCommands(t *testing.T) {
	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	if err := c.Run([]string{"help"}); err != nil {
		t.Fatal(err)
	}
	for _, cmd := range []string{"amarra-cais install", "amarra-cais css", "amarra-cais dev", "amarra-cais build", "amarra-cais server", "amarra-cais db migrate", "amarra-cais db status", "amarra-cais db rollback", "amarra-cais db prune-sessions", "amarra-cais db seed", "amarra-cais routes", "amarra-cais version", "amarra-cais g [--dry-run] ci", "amarra-cais g [--dry-run] console", "amarra-cais destroy"} {
		if !strings.Contains(buf.String(), cmd) {
			t.Errorf("help missing %q", cmd)
		}
	}
}

func TestCLI_Install_requiresCaisApp(t *testing.T) {
	c := &CLI{Out: os.Stdout}
	if err := c.Run([]string{"install"}); err == nil {
		t.Fatal("expected error outside cais app")
	}
}

func TestCLI_CSS_requiresCaisApp(t *testing.T) {
	c := &CLI{Out: os.Stdout}
	if err := c.Run([]string{"css"}); err == nil {
		t.Fatal("expected error outside cais app")
	}
}

func TestCLI_Dev_requiresCaisApp(t *testing.T) {
	c := &CLI{Out: os.Stdout}
	if err := c.Run([]string{"dev"}); err == nil {
		t.Fatal("expected error outside cais app")
	}
}

func TestStartDevAssetWatchers_skipsViteWithoutNodeModules(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "vite.config.js"), []byte("export default {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"scripts":{"build":"vite build","css":"tailwindcss"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	stop, err := startDevAssetWatchers(dir)
	if err != nil {
		t.Fatalf("dev must not require Vite node_modules: %v", err)
	}
	if stop != nil {
		stop()
	}
}

func TestCLI_Build_requiresCaisApp(t *testing.T) {
	c := &CLI{Out: os.Stdout}
	if err := c.Run([]string{"build"}); err == nil {
		t.Fatal("expected error outside cais app")
	}
}

func TestFindAir(t *testing.T) {
	// always returns empty or a path — must not panic
	_ = findAir()
}

func TestRunTailwindBuild_missingInput(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\nrequire github.com/puppe1990/amarra-cais v0.3.0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := runTailwindBuild(dir, false)
	if err == nil {
		t.Fatal("expected error without input.css")
	}
}

func TestEnsureStylesCSS_skipsWhenReady(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, cssOutput)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(".p-4{padding:1rem}"), 0o644); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := ensureStylesCSS(&buf, dir); err != nil {
		t.Fatalf("ready CSS should not error: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no build log when ready, got %q", buf.String())
	}
}

// #189: a stale styles.css must trigger a Tailwind rebuild on server/install.
func TestEnsureStylesCSS_rebuildsWhenStale(t *testing.T) {
	dir := t.TempDir()
	built := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	edited := built.Add(time.Hour)
	writeFileAt(t, filepath.Join(dir, cssInput), "@import \"tailwindcss\";", edited)
	writeFileAt(t, filepath.Join(dir, cssOutput), ".flex{display:flex}", built)
	writeFileAt(t, filepath.Join(dir, "web/templates/pages/home.html"), "<p>group-hover:block</p>", edited)
	logPath := fakeToolchain(t)

	var buf bytes.Buffer
	_ = ensureStylesCSS(&buf, dir)
	if !strings.Contains(buf.String(), "tailwind build") {
		t.Errorf("stale CSS should trigger a rebuild, log: %q", buf.String())
	}
	calls, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("no tool calls recorded: %v", err)
	}
	if !strings.Contains(string(calls), "tailwindcss") {
		t.Errorf("expected npx tailwindcss, calls:\n%s", calls)
	}
}

func TestEnsureStylesCSS_errorsWithoutInput(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer
	err := ensureStylesCSS(&buf, dir)
	if err == nil {
		t.Fatal("expected error when styles missing and no input.css")
	}
	if !strings.Contains(err.Error(), "unstyled") && !strings.Contains(err.Error(), cssOutput) {
		t.Errorf("error = %v", err)
	}
}

// fakeToolchain puts stub npm/npx/go first on PATH and returns the call log.
// The npm stub encodes the real contract (#54): npm skips devDependencies when
// NODE_ENV=production unless --include=dev is passed.
func fakeToolchain(t *testing.T) string {
	t.Helper()
	binDir := t.TempDir()
	logPath := filepath.Join(binDir, "calls.log")
	scripts := map[string]string{
		"npm": "#!/bin/sh\n" +
			"echo \"npm $*\" >> \"$FAKE_CALLS\"\n" +
			"if [ \"$NODE_ENV\" = \"production\" ]; then\n" +
			"  case \" $* \" in *\" --include=dev \"*) ;; *) echo \"npm error could not determine executable to run\" >&2; exit 1 ;; esac\n" +
			"fi\n" +
			"exit 0\n",
		"npx": "#!/bin/sh\necho \"npx $*\" >> \"$FAKE_CALLS\"\nexit ${FAKE_NPX_EXIT:-0}\n",
		"go":  "#!/bin/sh\necho \"go $*\" >> \"$FAKE_CALLS\"\nexit 0\n",
	}
	for name, body := range scripts {
		if err := os.WriteFile(filepath.Join(binDir, name), []byte(body), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("FAKE_CALLS", logPath)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return logPath
}

func installFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"go.mod":       "module github.com/example/probe\n\ngo 1.26\n\nrequire github.com/puppe1990/amarra-cais v0.3.0\n",
		"package.json": `{"devDependencies":{"tailwindcss":"^4.0.0","prettier":"^3.5.3"}}`,
		"input.css":    "@import \"tailwindcss\";\n",
	}
	for rel, body := range files {
		if err := os.WriteFile(filepath.Join(dir, rel), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestCLI_Install_npmIncludesDevDependencies(t *testing.T) {
	dir := installFixture(t)
	logPath := fakeToolchain(t)
	t.Setenv("NODE_ENV", "production")
	t.Chdir(dir)

	var buf bytes.Buffer
	if err := (&CLI{Out: &buf}).Run([]string{"install"}); err != nil {
		t.Fatalf("install must install the build toolchain with NODE_ENV=production: %v\n%s", err, buf.String())
	}
	calls, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("no tool calls recorded: %v", err)
	}
	if !strings.Contains(string(calls), "--include=dev") {
		t.Errorf("npm install must pass --include=dev, calls:\n%s", calls)
	}
}

func TestCLI_Install_failsWhenStylesheetCannotBeBuilt(t *testing.T) {
	dir := installFixture(t)
	fakeToolchain(t)
	t.Setenv("FAKE_NPX_EXIT", "1")
	t.Chdir(dir)

	var buf bytes.Buffer
	err := (&CLI{Out: &buf}).Run([]string{"install"})
	if err == nil {
		t.Fatalf("install must not exit 0 leaving an unbuilt stylesheet\n%s", buf.String())
	}
	if !strings.Contains(err.Error(), "amarra-cais css") {
		t.Errorf("error should point at amarra-cais css, got: %v", err)
	}
	if strings.Contains(buf.String(), "Done. Run: amarra-cais dev") {
		t.Errorf("install must not report success after a failed css build:\n%s", buf.String())
	}
}
