package sqlite

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

// #122: Configure applies PRAGMAs with Exec on whichever pooled connection it
// happens to get; a driver reconnect (ErrBadConn) hands out a fresh connection
// without foreign_keys/busy_timeout. Pragmas in the DSN are inherited by every
// new connection.
func TestDSN_appliesPragmasOnFreshConnections(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dsn.db")
	db, err := sql.Open("sqlite", DSN(path))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	// Drop idle connections so every query opens a brand-new connection.
	db.SetMaxIdleConns(0)

	for i := 0; i < 3; i++ {
		var busy, foreignKeys int
		if err := db.QueryRow("PRAGMA busy_timeout").Scan(&busy); err != nil {
			t.Fatal(err)
		}
		if err := db.QueryRow("PRAGMA foreign_keys").Scan(&foreignKeys); err != nil {
			t.Fatal(err)
		}
		if busy != 5000 {
			t.Errorf("conn %d: busy_timeout = %d, want 5000", i, busy)
		}
		if foreignKeys != 1 {
			t.Errorf("conn %d: foreign_keys = %d, want 1", i, foreignKeys)
		}
	}
}

func TestDSN_memoryAndAlreadyParametrized(t *testing.T) {
	if dsn := DSN(":memory:"); !strings.Contains(dsn, "_pragma=busy_timeout(5000)") {
		t.Errorf("DSN(:memory:) = %q, want pragmas", dsn)
	}
	custom := "file:x.db?_pragma=busy_timeout(10)"
	if got := DSN(custom); got != custom {
		t.Errorf("DSN should not duplicate existing _pragma params: %q", got)
	}
	withQuery := "file:x.db?cache=shared"
	if got := DSN(withQuery); !strings.HasPrefix(got, withQuery+"&") {
		t.Errorf("DSN should append with & when a query exists: %q", got)
	}
}
