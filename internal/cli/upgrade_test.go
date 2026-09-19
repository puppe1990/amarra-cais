package cli

import (
	"bytes"
	"os"
	"path/filepath"
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
