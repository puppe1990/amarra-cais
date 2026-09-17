package jobs

import (
	"context"
	"database/sql"
	"errors"
	"path"
	"sync"
	"testing"
	"time"
)

// #119: RunScheduler decided from a stale ListRecurring snapshot and
// enqueueRecurring had no compare-and-set, so two workers listing before any
// commit both inserted the same cron run.
func TestEnqueueRecurring_secondClaimIsRejected(t *testing.T) {
	db := testDB(t)
	store := NewStore(db)
	ctx := context.Background()
	if err := store.UpsertRecurring(ctx, RecurringOptions{Kind: "Hourly", Cron: "0 * * * *"}); err != nil {
		t.Fatal(err)
	}
	tasks, err := store.ListRecurring(ctx)
	if err != nil || len(tasks) != 1 {
		t.Fatalf("tasks = %+v err=%v", tasks, err)
	}
	task := tasks[0] // both schedulers see the same LastRun (nil)
	now := time.Now().UTC().Truncate(time.Second)

	if err := store.enqueueRecurring(ctx, task, now); err != nil {
		t.Fatal(err)
	}
	err = store.enqueueRecurring(ctx, task, now)
	if !errors.Is(err, errRecurringClaimed) {
		t.Fatalf("second claim error = %v, want errRecurringClaimed", err)
	}

	jobs, err := store.List(ctx, ListFilter{Kind: "Hourly", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 {
		t.Fatalf("ready jobs = %d, want 1", len(jobs))
	}
}

func TestRunScheduler_concurrentWorkersEnqueueOnce(t *testing.T) {
	db, dsn := recurringFileDB(t)
	storeA := NewStore(db)
	second, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = second.Close() })
	storeB := NewStore(second)

	ctx := context.Background()
	if err := storeA.UpsertRecurring(ctx, RecurringOptions{Kind: "EveryMinute", Cron: "* * * * *"}); err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	stores := []*Store{storeA, storeB}
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(s *Store) {
			defer wg.Done()
			<-start
			_, _ = RunScheduler(ctx, s, now)
		}(stores[i%2])
	}
	close(start)
	wg.Wait()

	jobs, err := storeA.List(ctx, ListFilter{Kind: "EveryMinute", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 {
		t.Fatalf("ready jobs = %d, want exactly 1 (concurrent schedulers duplicated)", len(jobs))
	}
}

// recurringFileDB mirrors fileTestDB in the dispatch tests; named locally to
// avoid colliding while both PRs are in flight.
func recurringFileDB(t *testing.T) (*sql.DB, string) {
	t.Helper()
	path := path.Join(t.TempDir(), "recurring.db")
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
