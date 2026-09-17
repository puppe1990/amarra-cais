package migrate

import (
	"database/sql"
	"path/filepath"
	"sync"
	"testing"
	"testing/fstest"

	_ "modernc.org/sqlite"
)

// #123: two processes booting together (rolling restart, server + worker) both
// saw a migration as pending and applied it; the loser failed with
// `UNIQUE constraint failed: schema_migrations.version` and did not boot.
func TestApply_concurrentProcessesBootCleanly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "migrate.db")
	dsn := "file:" + path + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"

	open := func() *sql.DB {
		db, err := sql.Open("sqlite", dsn)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		return db
	}

	migrations := fstest.MapFS{
		"migrations/001_widget.sql": &fstest.MapFile{
			// Non-idempotent DDL plus a slow insert: widens the window where
			// another process still sees the migration as pending.
			Data: []byte(`CREATE TABLE widget (id INTEGER PRIMARY KEY, n INTEGER);
WITH RECURSIVE c(x) AS (SELECT 1 UNION ALL SELECT x + 1 FROM c WHERE x < 50000)
INSERT INTO widget(n) SELECT x FROM c;`),
		},
	}

	const workers = 8
	errCh := make(chan error, workers)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		db := open()
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			errCh <- Apply(db, migrations, "migrations")
		}()
	}
	close(start)
	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			t.Fatalf("concurrent Apply failed: %v", err)
		}
	}

	db := open()
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = '001_widget'").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("schema_migrations rows = %d, want 1", count)
	}
	var rows int
	if err := db.QueryRow("SELECT COUNT(*) FROM widget").Scan(&rows); err != nil {
		t.Fatal(err)
	}
	if rows != 50000 {
		t.Fatalf("widget rows = %d, want 50000 (migration must run once)", rows)
	}
}
