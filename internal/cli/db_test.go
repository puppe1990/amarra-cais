package cli

import (
	"bytes"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/puppe1990/amarra-cais/pkg/cais/session"

	_ "modernc.org/sqlite"
)

func TestCLI_DBStatus_listsMigrations(t *testing.T) {
	dir := t.TempDir()
	writeMinimalApp(t, dir)

	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	if err := c.Run([]string{"db", "status"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "001_contacts") {
		t.Errorf("status output missing migration: %q", buf.String())
	}
}

// #374: without a -- down section a rollback would keep the schema and drop
// only the record, so the next apply can fail. Refuse and point at the flag.
func TestCLI_DBRollback_requiresDownSection(t *testing.T) {
	dir := t.TempDir()
	writeMinimalApp(t, dir)

	c := &CLI{Out: &bytes.Buffer{}}
	if err := c.Run([]string{"db", "migrate"}); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err := (&CLI{Out: &buf}).Run([]string{"db", "rollback"})
	if err == nil {
		t.Fatal("expected rollback to refuse a migration without -- down")
	}
	for _, want := range []string{"001_contacts", "no -- down section", "--allow-record-only"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q should mention %q", err, want)
		}
	}
	if strings.Contains(buf.String(), "=> Rolled back") {
		t.Errorf("refused rollback must not report success: %q", buf.String())
	}
}

func TestCLI_DBRollback_allowRecordOnly(t *testing.T) {
	dir := t.TempDir()
	writeMinimalApp(t, dir)

	c := &CLI{Out: &bytes.Buffer{}}
	if err := c.Run([]string{"db", "migrate"}); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	c2 := &CLI{Out: &buf}
	if err := c2.Run([]string{"db", "rollback", "--allow-record-only"}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "001_contacts") {
		t.Errorf("rollback output missing version: %q", out)
	}
	if !strings.Contains(out, "=> Rolled back") {
		t.Errorf("rollback output missing success message: %q", out)
	}
	if !strings.Contains(out, "no -- down section") {
		t.Errorf("rollback output missing down-section warning: %q", out)
	}
}

func TestCLI_DBPruneSessions_removesExpired(t *testing.T) {
	dir := t.TempDir()
	writeMinimalApp(t, dir)

	dbPath := filepath.Join(dir, "data", "app.db")
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := session.EnsureSQLiteSchema(db); err != nil {
		t.Fatal(err)
	}
	store := session.NewSQLiteStore(db)
	token, err := store.Create(1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(
		"UPDATE sessions SET expires_at = datetime('now', '-1 day') WHERE token = ?",
		token,
	); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	if err := c.Run([]string{"db", "prune-sessions"}); err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); !strings.Contains(got, "=> Pruned 1 expired session(s)") {
		t.Errorf("prune output = %q, want pruned count message", got)
	}
}

func TestCLI_DBSeed_missingSeedsFile(t *testing.T) {
	dir := t.TempDir()
	writeMinimalApp(t, dir)

	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	if err := c.Run([]string{"db", "seed"}); err == nil {
		t.Fatal("expected error without seeds.go")
	}
}

func TestCLI_DBMigrate_isIdempotent(t *testing.T) {
	dir := t.TempDir()
	writeMinimalApp(t, dir)

	c := &CLI{Out: &bytes.Buffer{}}
	if err := c.Run([]string{"db", "migrate"}); err != nil {
		t.Fatal(err)
	}
	if err := c.Run([]string{"db", "migrate"}); err != nil {
		t.Fatalf("second migrate failed: %v", err)
	}
}

func writeMinimalApp(t *testing.T, dir string) {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	files := map[string]string{
		"go.mod": `module testapp

require github.com/puppe1990/amarra-cais v0.3.0
`,
		"internal/store/migrations/001_contacts.sql": `CREATE TABLE IF NOT EXISTS contacts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT NOT NULL
);`,
	}
	for path, content := range files {
		full := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
