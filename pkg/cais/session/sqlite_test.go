package session

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := EnsureSQLiteSchema(db); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestSQLiteStore_CreateGetDelete(t *testing.T) {
	store := NewSQLiteStore(testDB(t))

	token, err := store.Create(42)
	if err != nil {
		t.Fatal(err)
	}
	if token == "" {
		t.Fatal("expected token")
	}

	id, ok := store.Get(token)
	if !ok || id != 42 {
		t.Fatalf("Get = (%d, %v), want (42, true)", id, ok)
	}

	if err := store.Delete(token); err != nil {
		t.Fatal(err)
	}
	if _, ok := store.Get(token); ok {
		t.Fatal("session should be deleted")
	}
}

func TestSQLiteStore_Get_rejectsExpiredSession(t *testing.T) {
	db := testDB(t)
	store := NewSQLiteStore(db)

	token, err := store.Create(42)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec("UPDATE sessions SET expires_at = datetime('now', '-1 hour') WHERE token = ?", token)
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := store.Get(token); ok {
		t.Fatal("expected expired session to be rejected")
	}
}

func TestSQLiteStore_PruneExpired_removesOldRows(t *testing.T) {
	db := testDB(t)
	store := NewSQLiteStore(db)

	token, err := store.Create(1)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec("UPDATE sessions SET expires_at = datetime('now', '-1 day') WHERE token = ?", token)
	if err != nil {
		t.Fatal(err)
	}

	n, err := store.PruneExpired()
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("PruneExpired = %d, want 1", n)
	}
	if _, ok := store.Get(token); ok {
		t.Fatal("pruned session should not be found")
	}
}

func TestEnsureSQLiteSchema_PopulatedLegacySessions(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	db.SetMaxOpenConns(1)
	for _, query := range []string{
		"CREATE TABLE sessions (token TEXT PRIMARY KEY, user_id INTEGER, created_at TEXT)",
		"INSERT INTO sessions VALUES ('recent', 1, datetime('now')), ('old', 2, datetime('now', '-8 days')), ('invalid', 3, 'invalid')",
	} {
		if _, err := db.Exec(query); err != nil {
			t.Fatal(err)
		}
	}
	for range 2 {
		if err := EnsureSQLiteSchema(db); err != nil {
			t.Fatalf("legacy migration: %v", err)
		}
	}
	store := NewSQLiteStore(db)
	if id, ok := store.Get("recent"); !ok || id != 1 {
		t.Fatalf("recent = (%d, %v)", id, ok)
	}
	for _, token := range []string{"old", "invalid"} {
		if _, ok := store.Get(token); ok {
			t.Fatalf("%s must be expired", token)
		}
	}
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM sessions WHERE expires_at IS NOT NULL AND expires_at != ''").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Fatalf("backfilled sessions = %d, want 3", count)
	}
	if _, err := store.Create(4); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureSQLiteSchema_LegacyBackfillRollsBack(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	db.SetMaxOpenConns(1)
	for _, query := range []string{
		"CREATE TABLE sessions (token TEXT PRIMARY KEY, user_id INTEGER, created_at TEXT)",
		"INSERT INTO sessions VALUES ('recent', 1, datetime('now'))",
		"CREATE TRIGGER refuse_backfill BEFORE UPDATE ON sessions BEGIN SELECT RAISE(FAIL, 'backfill unavailable'); END",
	} {
		if _, err := db.Exec(query); err != nil {
			t.Fatal(err)
		}
	}
	if err := EnsureSQLiteSchema(db); err == nil {
		t.Fatal("expected backfill failure")
	}
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('sessions') WHERE name = 'expires_at'").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("failed backfill must roll back the column")
	}
	if _, err := db.Exec("DROP TRIGGER refuse_backfill"); err != nil {
		t.Fatal(err)
	}
	if err := EnsureSQLiteSchema(db); err != nil {
		t.Fatal(err)
	}
	if _, ok := NewSQLiteStore(db).Get("recent"); !ok {
		t.Fatal("retry must preserve recent session")
	}
}
