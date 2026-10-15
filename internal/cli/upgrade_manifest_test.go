package cli

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

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
	to := parseSemverCore("0.10.0")
	got := migrationsBetween(semverCore{}, to)
	want := 0
	for _, s := range frameworkMigrations {
		v := parseSemverCore(s.Version)
		if v.OK && compareSemverCore(v, to) <= 0 {
			want++
		}
	}
	if len(got) != want {
		t.Fatalf("unknown from should return every step <= to, got %d want %d", len(got), want)
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

func TestMigrationsBetween_0211to0220_listsSignOutAndAdmin(t *testing.T) {
	got := migrationsBetween(parseSemverCore("0.21.1"), parseSemverCore("0.22.0"))
	if len(got) == 0 {
		t.Fatal("expected steps between 0.21.1 and 0.22.0")
	}
	joined := migrationActions(got)
	if !strings.Contains(joined, "SignOut") {
		t.Errorf("0.21.1→0.22.0 must list SignOut error handling, got:\n%s", joined)
	}
	if !strings.Contains(joined, "is_admin") {
		t.Errorf("0.21.1→0.22.0 must list users.is_admin / RequireAdmin, got:\n%s", joined)
	}
	if !strings.Contains(joined, "amarra-flash") {
		t.Errorf("0.21.1→0.22.0 must list the 0.21.2 flash id, got:\n%s", joined)
	}
	for _, s := range got {
		v := parseSemverCore(s.Version)
		if compareSemverCore(v, parseSemverCore("0.21.1")) <= 0 || compareSemverCore(v, parseSemverCore("0.22.0")) > 0 {
			t.Errorf("step %s outside (0.21.1, 0.22.0]", s.Version)
		}
	}
}

func TestFrameworkMigrations_coverChangelogUpgradeSections(t *testing.T) {
	body, err := os.ReadFile("../../CHANGELOG.md")
	if err != nil {
		t.Fatal(err)
	}
	have := map[string]bool{}
	for _, s := range frameworkMigrations {
		have[s.Version] = true
	}
	for _, v := range changelogUpgradeVersions(string(body)) {
		if !have[v] {
			t.Errorf("CHANGELOG [%s] has ### Upgrade but frameworkMigrations has no %s step", v, v)
		}
	}
}

func migrationActions(steps []migrationStep) string {
	var b strings.Builder
	for _, s := range steps {
		b.WriteString(s.Version)
		b.WriteByte(' ')
		b.WriteString(s.Title)
		b.WriteByte(' ')
		b.WriteString(s.Action)
		b.WriteByte('\n')
	}
	return b.String()
}

var changelogSection = regexp.MustCompile(`(?m)^## \[(\d+\.\d+\.\d+)\]`)

func changelogUpgradeVersions(text string) []string {
	idxs := changelogSection.FindAllStringSubmatchIndex(text, -1)
	var out []string
	for i, loc := range idxs {
		version := text[loc[2]:loc[3]]
		end := len(text)
		if i+1 < len(idxs) {
			end = idxs[i+1][0]
		}
		if strings.Contains(text[loc[0]:end], "### Upgrade") {
			out = append(out, version)
		}
	}
	return out
}
