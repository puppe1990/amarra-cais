package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeJobsFixture(t *testing.T, dir, appGo, storeGo string) {
	t.Helper()
	for rel, body := range map[string]string{
		"internal/app/app.go":     appGo,
		"internal/store/store.go": storeGo,
	} {
		path := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// #194: the hint must mention that the Store interface needs DB() too, otherwise
// adding jobsui.Register does not compile.
func TestCheckJobsUI_missingRegisterHintsDB(t *testing.T) {
	dir := t.TempDir()
	writeJobsFixture(t, dir,
		"package app\n",
		"package store\n\ntype Store interface {\n\tSessions() int\n}\n",
	)
	c := checkJobsUI(dir)
	if c.OK {
		t.Fatal("expected warn when jobsui.Register is missing")
	}
	if !strings.Contains(c.FixHint, "jobsui.Register") {
		t.Errorf("FixHint = %q, want jobsui.Register", c.FixHint)
	}
	if !strings.Contains(c.FixHint, "Store interface") {
		t.Errorf("FixHint = %q, want the Store.DB() interface requirement", c.FixHint)
	}
}

// #194: Register present but the interface lacks DB() is a compile error, not [ok].
func TestCheckJobsUI_registerWithoutDBWarns(t *testing.T) {
	dir := t.TempDir()
	writeJobsFixture(t, dir,
		"package app\n\nfunc New() { jobsui.Register(r, deps.Store.DB()) }\n",
		"package store\n\ntype Store interface {\n\tSessions() int\n}\n",
	)
	c := checkJobsUI(dir)
	if c.OK {
		t.Fatal("Store without DB() must not be [ok]")
	}
	if !c.Optional {
		t.Error("missing DB() should warn, not fail doctor")
	}
	if !strings.Contains(c.Detail, "DB()") {
		t.Errorf("Detail = %q, want the missing DB() method", c.Detail)
	}
}

func TestCheckJobsUI_registerWithDBOK(t *testing.T) {
	dir := t.TempDir()
	writeJobsFixture(t, dir,
		"package app\n\nfunc New() { jobsui.Register(r, deps.Store.DB()) }\n",
		"package store\n\nimport \"database/sql\"\n\ntype Store interface {\n\tDB() *sql.DB\n}\n",
	)
	if c := checkJobsUI(dir); !c.OK {
		t.Fatalf("expected OK with Register + DB(), got %+v", c)
	}
}
