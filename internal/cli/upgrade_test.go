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
