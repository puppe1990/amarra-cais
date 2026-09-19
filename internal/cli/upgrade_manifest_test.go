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
