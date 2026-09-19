# `amarra-cais upgrade` Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `amarra-cais upgrade [version] [--dry-run]` to bump the framework in `go.mod`, re-install deps, run `doctor`, and print a curated migration checklist.

**Architecture:** A new `upgrade.go` orchestrates the command; a data-only `upgrade_manifest.go` holds curated `migrationStep`s and pure range selection; `commands.go` gains a shared `installAppDeps` helper reused by `install` and `upgrade`. All flow tests are headless via the existing `fakeToolchain` stub (no network).

**Tech Stack:** Go stdlib (`os`, `path/filepath`, `strings`, `fmt`, `io`), existing CLI helpers (`runCmd`, `runDoctor`, `parseSemverCore`, `extractCaisVersion`, `fakeToolchain`).

**Spec:** `docs/superpowers/specs/2026-09-19-amarra-cais-upgrade-design.md`

---

## File structure

| File                                              | Responsibility                                      |
| ------------------------------------------------- | --------------------------------------------------- |
| `internal/cli/upgrade_manifest.go` (create)       | Curated migration steps + `migrationsBetween`       |
| `internal/cli/upgrade.go` (create)                | Arg parsing, guards, command flow, checklist output |
| `internal/cli/commands.go` (modify)               | Extract shared `installAppDeps` from `cmdInstall`   |
| `internal/cli/cli.go` (modify)                    | Dispatch `upgrade` + help line                      |
| `internal/cli/upgrade_manifest_test.go` (create)  | `migrationsBetween` cases                           |
| `internal/cli/upgrade_test.go` (create)           | Parsing + flow/guards/dry-run                       |
| `internal/cli/commands_test.go` (modify)          | `installAppDeps` tests                              |
| `README.md`, `AGENTS.md`, `CHANGELOG.md` (modify) | Document the command                                |

---

### Task 1: Migration manifest + range selection

**Files:**

- Create: `internal/cli/upgrade_manifest.go`
- Test: `internal/cli/upgrade_manifest_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/cli/upgrade_manifest_test.go`:

```go
package cli

import "testing"

func TestMigrationsBetween_range(t *testing.T) {
	got := migrationsBetween(parseSemverCore("0.2.2"), parseSemverCore("0.10.0"))
	if len(got) == 0 {
		t.Fatal("expected steps between 0.2.2 and 0.10.0")
	}
	lo := parseSemverCore("0.2.2")
	hi := parseSemverCore("0.10.0")
	for _, s := range got {
		v := parseSemverCore(s.Version)
		if compareSemverCore(v, lo) <= 0 || compareSemverCore(v, hi) > 0 {
			t.Errorf("step %s outside (0.2.2, 0.10.0]", s.Version)
		}
	}
}

func TestMigrationsBetween_emptyWhenUpToDate(t *testing.T) {
	if got := migrationsBetween(parseSemverCore("0.10.0"), parseSemverCore("0.10.0")); len(got) != 0 {
		t.Fatalf("expected no steps, got %d", len(got))
	}
}

func TestMigrationsBetween_unknownFromReturnsAll(t *testing.T) {
	got := migrationsBetween(semverCore{}, parseSemverCore("0.10.0"))
	if len(got) != len(frameworkMigrations) {
		t.Fatalf("unknown from should return every step <= to, got %d want %d", len(got), len(frameworkMigrations))
	}
}

func TestMigrationsBetween_unknownToReturnsNewer(t *testing.T) {
	got := migrationsBetween(parseSemverCore("0.9.0"), semverCore{})
	if len(got) == 0 {
		t.Fatal("expected steps newer than 0.9.0")
	}
	newerThan := parseSemverCore("0.9.0")
	for _, s := range got {
		if compareSemverCore(parseSemverCore(s.Version), newerThan) <= 0 {
			t.Errorf("step %s should be newer than 0.9.0", s.Version)
		}
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/cli/ -run TestMigrationsBetween -count=1`
Expected: FAIL — `undefined: migrationsBetween` (and `frameworkMigrations`).

- [ ] **Step 3: Write the manifest + selection**

Create `internal/cli/upgrade_manifest.go`:

```go
package cli

// migrationStep is a curated breaking change an app owner must apply when moving
// between framework versions. Versions are pinned to CHANGELOG sections (#192).
type migrationStep struct {
	Version string
	Title   string
	Action  string
}

// frameworkMigrations is maintained per release, newest first.
var frameworkMigrations = []migrationStep{
	{Version: "0.10.0", Title: "netutil.HealthPayload gained an env argument", Action: "call netutil.HealthPayload(status, port, cfg.Env)"},
	{Version: "0.10.0", Title: "jobs dashboard needs Store.DB()", Action: "add DB() *sql.DB to the Store interface in internal/store/store.go"},
	{Version: "0.10.0", Title: "HTMX assets deprecated", Action: "migrate hx-* templates to Amarra Views + Drive"},
	{Version: "0.9.0", Title: ".cais-generated.json tracks generated files", Action: "regenerate resources or accept that destroy skips untracked files"},
	{Version: "0.6.1", Title: "view.Write defaults dynamic pages to Cache-Control: no-store", Action: "set Page.CacheControl on pages that are safe to cache"},
	{Version: "0.5.0", Title: "writeView argument order changed", Action: "call writeView(w, r, views, cfg, layout, name, data, status)"},
	{Version: "0.3.0", Title: "doctor FAILs on hx-* / gonertia", Action: "finish the HTML-first (Amarra) migration — see docs/migrate-inertia.md"},
}

// migrationsBetween returns the steps strictly newer than from and up to to.
// An unknown from (semverCore{}) returns every step up to to; an unknown to
// (e.g. `latest`) returns every step newer than from (#192).
func migrationsBetween(from, to semverCore) []migrationStep {
	var out []migrationStep
	for _, step := range frameworkMigrations {
		v := parseSemverCore(step.Version)
		if !v.OK {
			continue
		}
		if from.OK && compareSemverCore(v, from) <= 0 {
			continue
		}
		if to.OK && compareSemverCore(v, to) > 0 {
			continue
		}
		out = append(out, step)
	}
	return out
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/cli/ -run TestMigrationsBetween -count=1`
Expected: PASS (ok).

- [ ] **Step 5: Commit**

```bash
gofmt -l internal/cli/upgrade_manifest.go internal/cli/upgrade_manifest_test.go
git add internal/cli/upgrade_manifest.go internal/cli/upgrade_manifest_test.go
git commit -m "feat(upgrade): curated migration manifest + range selection (#192)"
```

---

### Task 2: Extract shared `installAppDeps` from `cmdInstall`

**Files:**

- Modify: `internal/cli/commands.go:37-67`
- Test: `internal/cli/commands_test.go`

- [ ] **Step 1: Write the failing tests**

Append to `internal/cli/commands_test.go`:

```go
func TestInstallAppDeps_npmAndTidy(t *testing.T) {
	dir := installFixture(t)
	logPath := fakeToolchain(t)
	if err := installAppDeps(&bytes.Buffer{}, dir); err != nil {
		t.Fatal(err)
	}
	calls, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("no tool calls recorded: %v", err)
	}
	got := string(calls)
	if !strings.Contains(got, "npm install --include=dev") {
		t.Errorf("missing npm install --include=dev, calls:\n%s", got)
	}
	if !strings.Contains(got, "go mod tidy") {
		t.Errorf("missing go mod tidy, calls:\n%s", got)
	}
}

func TestInstallAppDeps_skipsNpmWithoutPackageJSON(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module github.com/example/up\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	logPath := fakeToolchain(t)
	if err := installAppDeps(&bytes.Buffer{}, dir); err != nil {
		t.Fatal(err)
	}
	calls, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("no tool calls recorded: %v", err)
	}
	if strings.Contains(string(calls), "npm") {
		t.Errorf("no package.json should skip npm, calls:\n%s", calls)
	}
	if !strings.Contains(string(calls), "go mod tidy") {
		t.Errorf("expected go mod tidy, calls:\n%s", calls)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/cli/ -run TestInstallAppDeps -count=1`
Expected: FAIL — `undefined: installAppDeps`.

- [ ] **Step 3: Extract the helper**

In `internal/cli/commands.go`, replace the body of `cmdInstall` (currently lines 37-67) with:

```go
func (c *CLI) cmdInstall() error {
	dir, err := c.appDir()
	if err != nil {
		return err
	}

	if err := installAppDeps(c.Out, dir); err != nil {
		return err
	}

	// Build Tailwind so styles.css exists after clone/install (#141).
	if _, err := os.Stat(filepath.Join(dir, cssInput)); err == nil {
		_, _ = fmt.Fprintln(c.Out, "→ tailwind build (styles.css)")
		if err := runTailwindBuild(dir, false); err != nil {
			// Fail loudly: exiting 0 here left users on an unstyled app (#54).
			return fmt.Errorf("css build failed: %w — retry: amarra-cais css (NODE_ENV=production skips npm devDependencies)", err)
		}
	}

	_, _ = fmt.Fprintln(c.Out, "Done. Run: amarra-cais dev")
	return nil
}

// installAppDeps runs the npm + go.mod steps shared by `install` and `upgrade`.
func installAppDeps(w io.Writer, dir string) error {
	if _, err := os.Stat(filepath.Join(dir, "package.json")); err == nil {
		_, _ = fmt.Fprintln(w, "→ npm install")
		// --include=dev: NODE_ENV=production would skip tailwindcss/prettier (#54).
		if err := runCmd(dir, "npm", "install", "--include=dev"); err != nil {
			return fmt.Errorf("npm install: %w", err)
		}
	}
	_, _ = fmt.Fprintln(w, "→ go mod tidy")
	if err := runCmd(dir, "go", "mod", "tidy"); err != nil {
		return fmt.Errorf("go mod tidy: %w", err)
	}
	return nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/cli/ -run 'TestInstallAppDeps|TestCLI_Install' -count=1`
Expected: PASS — new tests plus the existing `TestCLI_Install_npmIncludesDevDependencies` and `TestCLI_Install_failsWhenStylesheetCannotBeBuilt`.

- [ ] **Step 5: Commit**

```bash
gofmt -l internal/cli/commands.go internal/cli/commands_test.go
git add internal/cli/commands.go internal/cli/commands_test.go
git commit -m "refactor(install): extract shared installAppDeps (#192)"
```

---

### Task 3: Arg parsing, target normalization, checklist output

**Files:**

- Create: `internal/cli/upgrade.go`
- Test: `internal/cli/upgrade_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/cli/upgrade_test.go`:

```go
package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseUpgradeArgs_defaultsToLatest(t *testing.T) {
	dry, target, err := parseUpgradeArgs(nil)
	if err != nil || dry || target != "latest" {
		t.Fatalf("parseUpgradeArgs(nil) = %v, %q, %v; want false, latest, nil", dry, target, err)
	}
}

func TestParseUpgradeArgs_dryRunAndVersion(t *testing.T) {
	dry, target, err := parseUpgradeArgs([]string{"--dry-run"})
	if err != nil || !dry || target != "latest" {
		t.Fatalf("dry-run parse = %v, %q, %v", dry, target, err)
	}
	dry, target, err = parseUpgradeArgs([]string{"0.11.0"})
	if err != nil || dry || target != "v0.11.0" {
		t.Fatalf("version parse = %v, %q, %v; want v0.11.0", dry, target, err)
	}
	if _, target, err = parseUpgradeArgs([]string{"v0.11.0"}); err != nil || target != "v0.11.0" {
		t.Fatalf("v-prefix parse = %q, %v", target, err)
	}
}

func TestParseUpgradeArgs_rejectsBadInput(t *testing.T) {
	if _, _, err := parseUpgradeArgs([]string{"banana"}); err == nil {
		t.Error("expected invalid version error")
	}
	if _, _, err := parseUpgradeArgs([]string{"--nope"}); err == nil {
		t.Error("expected unknown flag error")
	}
	if _, _, err := parseUpgradeArgs([]string{"0.1.0", "0.2.0"}); err == nil {
		t.Error("expected usage error for two versions")
	}
}

func TestPrintMigrationChecklist_listsSteps(t *testing.T) {
	var buf bytes.Buffer
	printMigrationChecklist(&buf, parseSemverCore("0.2.2"), parseSemverCore("0.10.0"), "0.2.2")
	out := buf.String()
	if !strings.Contains(out, "migration checklist") {
		t.Errorf("missing checklist header:\n%s", out)
	}
	if !strings.Contains(out, "[0.10.0]") {
		t.Errorf("missing 0.10.0 step:\n%s", out)
	}
}

func TestPrintMigrationChecklist_unknownFromWarns(t *testing.T) {
	var buf bytes.Buffer
	printMigrationChecklist(&buf, semverCore{}, parseSemverCore("0.10.0"), "?")
	if !strings.Contains(buf.String(), "could not detect") {
		t.Errorf("expected unknown-baseline warning:\n%s", buf.String())
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/cli/ -run 'TestParseUpgradeArgs|TestPrintMigrationChecklist' -count=1`
Expected: FAIL — `undefined: parseUpgradeArgs` / `printMigrationChecklist`.

- [ ] **Step 3: Add parsing + checklist helpers**

Create `internal/cli/upgrade.go`:

```go
package cli

import (
	"fmt"
	"io"
	"strings"
)

// parseUpgradeArgs separates flags from the optional target version.
func parseUpgradeArgs(args []string) (dryRun bool, target string, err error) {
	target = "latest"
	hasTarget := false
	for _, arg := range args {
		switch {
		case arg == "--dry-run":
			dryRun = true
		case strings.HasPrefix(arg, "-"):
			return false, "", fmt.Errorf("unknown flag %q (usage: amarra-cais upgrade [version] [--dry-run])", arg)
		case hasTarget:
			return false, "", fmt.Errorf("usage: amarra-cais upgrade [version] [--dry-run]")
		default:
			target = arg
			hasTarget = true
		}
	}
	target, err = normalizeUpgradeTarget(target)
	return dryRun, target, err
}

// normalizeUpgradeTarget keeps `latest` or canonicalizes a version to vX.Y.Z.
func normalizeUpgradeTarget(arg string) (string, error) {
	arg = strings.TrimSpace(arg)
	if arg == "" || arg == "latest" {
		return "latest", nil
	}
	v := parseSemverCore(arg)
	if !v.OK {
		return "", fmt.Errorf("invalid version %q — use vX.Y.Z or omit for latest", arg)
	}
	return "v" + formatSemver(v), nil
}

// printMigrationChecklist writes the curated steps for the from→to range.
func printMigrationChecklist(w io.Writer, from, to semverCore, fromRaw string) {
	steps := migrationsBetween(from, to)
	if len(steps) == 0 {
		_, _ = fmt.Fprintln(w, "  no known breaking changes in this range — see CHANGELOG.md")
		return
	}
	if !from.OK {
		_, _ = fmt.Fprintf(w, "  could not detect the current version (%s); showing every step up to the target:\n", fromRaw)
	}
	_, _ = fmt.Fprintln(w, "  migration checklist:")
	for _, s := range steps {
		_, _ = fmt.Fprintf(w, "  - [%s] %s: %s\n", s.Version, s.Title, s.Action)
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/cli/ -run 'TestParseUpgradeArgs|TestPrintMigrationChecklist' -count=1`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
gofmt -l internal/cli/upgrade.go internal/cli/upgrade_test.go
git add internal/cli/upgrade.go internal/cli/upgrade_test.go
git commit -m "feat(upgrade): arg parsing, target normalization, checklist output (#192)"
```

---

### Task 4: `cmdUpgrade` flow (guards, dry-run, go get, install, doctor)

**Files:**

- Modify: `internal/cli/upgrade.go`
- Test: `internal/cli/upgrade_test.go`

- [ ] **Step 1: Write the failing tests**

Update the import block in `internal/cli/upgrade_test.go` to add `os` and `path/filepath`:

```go
import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)
```

Then append to `internal/cli/upgrade_test.go`:

```go
func upgradeFixture(t *testing.T, requireVersion string) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"go.mod":       "module github.com/puppe1990/up\n\ngo 1.26\n\nrequire github.com/puppe1990/amarra-cais v" + requireVersion + "\n",
		"package.json": `{"devDependencies":{"tailwindcss":"^4.0.0"}}`,
	}
	for rel, body := range files {
		if err := os.WriteFile(filepath.Join(dir, rel), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestCLI_Upgrade_dryRunPrintsChecklistOnly(t *testing.T) {
	dir := upgradeFixture(t, "0.2.2")
	logPath := fakeToolchain(t)
	t.Chdir(dir)

	var buf bytes.Buffer
	if err := (&CLI{Out: &buf}).cmdUpgrade([]string{"--dry-run"}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "migration checklist") {
		t.Errorf("expected checklist, got:\n%s", out)
	}
	if !strings.Contains(out, "--dry-run: no files changed") {
		t.Errorf("expected dry-run notice, got:\n%s", out)
	}
	if _, err := os.Stat(logPath); err == nil {
		calls, _ := os.ReadFile(logPath)
		t.Errorf("dry-run must not invoke tools; calls:\n%s", calls)
	}
}

func TestCLI_Upgrade_runsGoGetThenTidy(t *testing.T) {
	dir := upgradeFixture(t, "0.2.2")
	logPath := fakeToolchain(t)
	t.Chdir(dir)

	if err := (&CLI{Out: &bytes.Buffer{}}).cmdUpgrade([]string{"0.10.0"}); err != nil {
		t.Fatal(err)
	}
	calls, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("no tool calls recorded: %v", err)
	}
	got := string(calls)
	if !strings.Contains(got, "go get github.com/puppe1990/amarra-cais@v0.10.0") {
		t.Errorf("missing go get, calls:\n%s", got)
	}
	if !strings.Contains(got, "go mod tidy") {
		t.Errorf("missing go mod tidy, calls:\n%s", got)
	}
	if !strings.Contains(got, "npm install --include=dev") {
		t.Errorf("missing npm install, calls:\n%s", got)
	}
}

func TestCLI_Upgrade_defaultsToLatest(t *testing.T) {
	dir := upgradeFixture(t, "0.2.2")
	logPath := fakeToolchain(t)
	t.Chdir(dir)

	if err := (&CLI{Out: &bytes.Buffer{}}).cmdUpgrade(nil); err != nil {
		t.Fatal(err)
	}
	calls, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("no tool calls recorded: %v", err)
	}
	if !strings.Contains(string(calls), "go get github.com/puppe1990/amarra-cais@latest") {
		t.Errorf("missing @latest go get, calls:\n%s", calls)
	}
}

func TestCLI_Upgrade_abortsOnLocalReplace(t *testing.T) {
	dir := upgradeFixture(t, "0.2.2")
	goMod := filepath.Join(dir, "go.mod")
	body, err := os.ReadFile(goMod)
	if err != nil {
		t.Fatal(err)
	}
	body = append(body, []byte("\nreplace github.com/puppe1990/amarra-cais => ../amarra-cais\n")...)
	if err := os.WriteFile(goMod, body, 0o644); err != nil {
		t.Fatal(err)
	}
	logPath := fakeToolchain(t)
	t.Chdir(dir)

	err = (&CLI{Out: &bytes.Buffer{}}).cmdUpgrade(nil)
	if err == nil || !strings.Contains(err.Error(), "link --unlink") {
		t.Fatalf("expected link --unlink hint, got %v", err)
	}
	if _, statErr := os.Stat(logPath); statErr == nil {
		t.Error("must not invoke tools after the replace guard")
	}
}

func TestCLI_Upgrade_rejectsOlderTarget(t *testing.T) {
	dir := upgradeFixture(t, "0.10.0")
	t.Chdir(dir)

	err := (&CLI{Out: &bytes.Buffer{}}).cmdUpgrade([]string{"0.2.2"})
	if err == nil || !strings.Contains(err.Error(), "older") {
		t.Fatalf("expected older-target error, got %v", err)
	}
}

func TestCLI_Upgrade_requiresCaisApp(t *testing.T) {
	if err := (&CLI{Out: &bytes.Buffer{}}).cmdUpgrade(nil); err == nil {
		t.Fatal("expected error outside a Cais app")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/cli/ -run 'TestCLI_Upgrade' -count=1`
Expected: FAIL — `undefined: (*CLI).cmdUpgrade`.

- [ ] **Step 3: Add the command flow**

Append to `internal/cli/upgrade.go` and extend the import block to include `os` and `path/filepath`:

```go
import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)
```

```go
func (c *CLI) cmdUpgrade(args []string) error {
	dir, err := c.appDir()
	if err != nil {
		return err
	}
	dryRun, target, err := parseUpgradeArgs(args)
	if err != nil {
		return err
	}

	body, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return err
	}
	content := string(body)
	if strings.Contains(content, "replace "+frameworkModule) {
		return fmt.Errorf("go.mod has a local replace for %s — run: amarra-cais link --unlink, then retry", frameworkModule)
	}

	fromRaw := extractCaisVersion(content)
	from := parseSemverCore(fromRaw)
	to := parseUpgradeTargetCore(target)
	if from.OK && to.OK && compareSemverCore(to, from) < 0 {
		return fmt.Errorf("target v%s is older than the current v%s", formatSemver(to), formatSemver(from))
	}

	_, _ = fmt.Fprintf(c.Out, "→ upgrade %s → %s\n", displayUpgradeVersion(fromRaw), target)
	printMigrationChecklist(c.Out, from, to, fromRaw)

	if dryRun {
		_, _ = fmt.Fprintln(c.Out, "  --dry-run: no files changed")
		return nil
	}

	_, _ = fmt.Fprintf(c.Out, "→ go get %s@%s\n", frameworkModule, target)
	if err := runCmd(dir, "go", "get", frameworkModule+"@"+target); err != nil {
		return fmt.Errorf("go get: %w", err)
	}
	if err := installAppDeps(c.Out, dir); err != nil {
		return err
	}
	if err := runDoctor(c.Out, dir, doctorOptions{}); err != nil {
		_, _ = fmt.Fprintf(c.Out, "⚠ doctor reported issues: %v — fix the checklist items above, then re-run doctor\n", err)
	}
	return nil
}

// parseUpgradeTargetCore returns semverCore{} for `latest` (unknown target).
func parseUpgradeTargetCore(target string) semverCore {
	if target == "latest" {
		return semverCore{}
	}
	return parseSemverCore(target)
}

func displayUpgradeVersion(raw string) string {
	if raw == "" || raw == "?" {
		return "unknown"
	}
	return "v" + strings.TrimPrefix(raw, "v")
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/cli/ -run 'TestCLI_Upgrade' -count=1`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
gofmt -l internal/cli/upgrade.go internal/cli/upgrade_test.go
git add internal/cli/upgrade.go internal/cli/upgrade_test.go
git commit -m "feat(upgrade): command flow with guards, dry-run and doctor (#192)"
```

---

### Task 5: Wire `upgrade` into the CLI + help

**Files:**

- Modify: `internal/cli/cli.go:30-70` (dispatch), `internal/cli/cli.go:~106` (help)
- Test: `internal/cli/commands_test.go` (`TestCLI_Help_IncludesAppCommands`)

- [ ] **Step 1: Update the help test**

In `internal/cli/commands_test.go`, add `"amarra-cais upgrade"` to the slice in `TestCLI_Help_IncludesAppCommands`:

```go
	for _, cmd := range []string{"amarra-cais install", "amarra-cais css", "amarra-cais dev", "amarra-cais build", "amarra-cais server", "amarra-cais db migrate", "amarra-cais db status", "amarra-cais db rollback", "amarra-cais db prune-sessions", "amarra-cais db seed", "amarra-cais routes", "amarra-cais version", "amarra-cais g [--dry-run] ci", "amarra-cais g [--dry-run] console", "amarra-cais destroy", "amarra-cais upgrade"} {
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/cli/ -run TestCLI_Help_IncludesAppCommands -count=1`
Expected: FAIL — `help missing "amarra-cais upgrade"`.

- [ ] **Step 3: Add dispatch + help line**

In `internal/cli/cli.go`, after the `case "link":` block, add:

```go
	case "upgrade":
		return c.cmdUpgrade(args[1:])
```

In the help string, add this line right after the `amarra-cais pwa ...` line:

```
  amarra-cais upgrade [version]     Bump the framework in go.mod, tidy, run doctor, print migration steps
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/cli/ -run 'TestCLI_Help_IncludesAppCommands|TestCLI_Upgrade' -count=1`
Expected: PASS.

- [ ] **Step 5: Add an end-to-end dispatch test**

Append to `internal/cli/upgrade_test.go`:

```go
func TestCLI_Upgrade_dispatches(t *testing.T) {
	dir := upgradeFixture(t, "0.2.2")
	fakeToolchain(t)
	t.Chdir(dir)

	var buf bytes.Buffer
	if err := (&CLI{Out: &buf}).Run([]string{"upgrade", "--dry-run"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "migration checklist") {
		t.Errorf("dispatch did not reach cmdUpgrade:\n%s", buf.String())
	}
}
```

Run: `go test ./internal/cli/ -run 'TestCLI_Upgrade_dispatches' -count=1`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
gofmt -l internal/cli/cli.go internal/cli/upgrade_test.go
git add internal/cli/cli.go internal/cli/commands_test.go internal/cli/upgrade_test.go
git commit -m "feat(cli): wire amarra-cais upgrade + help (#192)"
```

---

### Task 6: Docs

**Files:**

- Modify: `README.md` (command table), `AGENTS.md` (command lists), `CHANGELOG.md` (Unreleased)

- [ ] **Step 1: README command table**

Add a row after the `amarra-cais pwa ...` row:

```markdown
| `amarra-cais upgrade [version] [--dry-run]` | Bump framework, tidy, doctor, migration checklist |
```

- [ ] **Step 2: AGENTS.md command lists**

In the "CLI generators" list, after the `amarra-cais pwa [--bump] [--force]` line, add:

```bash
amarra-cais upgrade [version] [--dry-run]  # bump framework, run doctor, print migration steps
```

In the "App commands" list, after the `amarra-cais pwa ...` line, add:

```bash
amarra-cais upgrade [version]  # bump the framework in go.mod + migration checklist
```

- [ ] **Step 3: CHANGELOG Unreleased**

Under `## Unreleased`, add an `### Added` section above `### Fixed`:

```markdown
### Added

- `amarra-cais upgrade [version] [--dry-run]`: bumps `github.com/puppe1990/amarra-cais` in `go.mod` (default `latest`), re-runs `npm install`/`go mod tidy`, runs `doctor`, and prints a curated migration checklist for the version range (#192).
```

- [ ] **Step 4: Verify formatting**

Run: `npx prettier --check .`
Expected: `All matched files use Prettier code style!` (fix with `npx prettier --write README.md AGENTS.md CHANGELOG.md` if not).

- [ ] **Step 5: Commit**

```bash
git add README.md AGENTS.md CHANGELOG.md
git commit -m "docs(upgrade): document amarra-cais upgrade (#192)"
```

---

### Task 7: Final validation

**Files:** none (verification only)

- [ ] **Step 1: Full focused test run**

Run: `go test ./internal/cli/... -count=1`
Expected: `ok` for `internal/cli` and `internal/cli/patch`.

- [ ] **Step 2: Lint**

Run: `golangci-lint run ./internal/cli/...`
Expected: `0 issues.`

- [ ] **Step 3: God-file guard + whole module**

Run: `go test ./... -count=1`
Expected: all packages `ok`. (The `-race` gate runs in CI; local disk may be tight, so `go test ./...` without `-race` is acceptable locally, with CI covering the race run.)

- [ ] **Step 4: Open PR, watch CI, merge if green**

```bash
git push -u origin feat-192-upgrade
gh pr create --title "feat(cli): amarra-cais upgrade (#192)" --body "Closes #192. See docs/superpowers/specs/2026-09-19-amarra-cais-upgrade-design.md"
gh pr checks --watch
gh pr merge --squash --delete-branch
```

---

## Self-review notes

- **Spec coverage:** command surface (Tasks 3-5), flow/guards/dry-run (Task 4), manifest (Task 1), shared install helper (Task 2), docs (Task 6), tests headless via `fakeToolchain` (Tasks 1-5).
- **Placeholder scan:** no TBD/TODO; every code step has complete code.
- **Type consistency:** `migrationStep`, `migrationsBetween`, `parseUpgradeArgs`, `normalizeUpgradeTarget`, `parseUpgradeTargetCore`, `displayUpgradeVersion`, `printMigrationChecklist`, `installAppDeps` are used with the same signatures across tasks.
