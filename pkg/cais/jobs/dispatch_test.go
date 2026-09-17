package jobs

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// fileTestDB opens a WAL SQLite file so two connections can contend for the
// write lock (":memory:" would isolate them).
func fileTestDB(t *testing.T) (*sql.DB, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "jobs.db")
	dsn := "file:" + path + "?_pragma=busy_timeout(2000)&_pragma=journal_mode(WAL)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := EnsureSchema(db); err != nil {
		t.Fatal(err)
	}
	return db, dsn
}

// #120: DispatchDue opened a deferred transaction, read scheduled_jobs and
// only then wrote. Under WAL, promoting the read snapshot to a write fails
// with SQLITE_BUSY (busy_timeout does not cover the upgrade), so scheduled
// jobs were intermittently not promoted under concurrent writes.
func TestDispatchDue_waitsForConcurrentWriter(t *testing.T) {
	db, dsn := fileTestDB(t)
	store := NewStore(db)
	ctx := context.Background()

	if _, err := SetWait(ctx, store, 0, Options{Kind: "Due"}); err != nil {
		t.Fatal(err)
	}

	blocker, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = blocker.Close() }()
	conn, err := blocker.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()
	if _, err := conn.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.ExecContext(ctx,
		"INSERT INTO job_workers (id, heartbeat_at, started_at) VALUES ('blocker', datetime('now'), datetime('now'))"); err != nil {
		t.Fatal(err)
	}

	errCh := make(chan error, 1)
	go func() {
		_, err := store.DispatchDue(ctx, time.Now().UTC())
		errCh <- err
	}()

	time.Sleep(150 * time.Millisecond)
	select {
	case err := <-errCh:
		t.Fatalf("DispatchDue failed while a writer held the lock: %v", err)
	default:
	}

	if _, err := conn.ExecContext(ctx, "COMMIT"); err != nil {
		t.Fatal(err)
	}

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("DispatchDue after writer commit: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("DispatchDue did not finish after the writer committed")
	}

	job, err := store.Claim(ctx, DefaultQueue)
	if err != nil {
		t.Fatal(err)
	}
	if job == nil || job.Kind != "Due" {
		t.Fatalf("scheduled job was not promoted: %+v", job)
	}
}
