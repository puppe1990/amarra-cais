package jobs

import (
	"context"
	"testing"
	"time"
)

func TestTouchWorker_listsAsLive(t *testing.T) {
	store := NewStore(testDB(t))
	ctx := context.Background()

	if err := store.TouchWorker(ctx, WorkerPulse{
		ID: "w1", Queues: "default", Concurrency: 2,
	}); err != nil {
		t.Fatal(err)
	}

	live, err := store.ListLiveWorkers(ctx, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if len(live) != 1 || live[0].ID != "w1" || live[0].Concurrency != 2 {
		t.Fatalf("live = %+v", live)
	}
}

func TestListLiveWorkers_excludesStale(t *testing.T) {
	db := testDB(t)
	store := NewStore(db)
	ctx := context.Background()
	if err := store.TouchWorker(ctx, WorkerPulse{ID: "old"}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE job_workers SET heartbeat_at = datetime('now', '-2 minutes')`); err != nil {
		t.Fatal(err)
	}
	live, err := store.ListLiveWorkers(ctx, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if len(live) != 0 {
		t.Fatalf("stale worker still live: %+v", live)
	}
}

func TestRequeueOrphaned_skipsJobsOfLiveWorker(t *testing.T) {
	store := NewStore(testDB(t))
	ctx := context.Background()
	if err := store.TouchWorker(ctx, WorkerPulse{ID: "alive"}); err != nil {
		t.Fatal(err)
	}
	if _, err := Enqueue(ctx, store, Options{Kind: "Owned"}); err != nil {
		t.Fatal(err)
	}
	job, err := store.ClaimFor(ctx, DefaultQueue, "alive")
	if err != nil || job == nil {
		t.Fatalf("claim: %v %+v", err, job)
	}

	n, err := store.RequeueOrphaned(ctx, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("requeued live worker job: %d", n)
	}
	got, err := store.Get(ctx, job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != StatusRunning {
		t.Fatalf("status = %q, want running", got.Status)
	}
}

func TestRequeueOrphaned_recoversDeadWorkerJobs(t *testing.T) {
	store := NewStore(testDB(t))
	ctx := context.Background()
	if _, err := Enqueue(ctx, store, Options{Kind: "Orphan"}); err != nil {
		t.Fatal(err)
	}
	job, err := store.ClaimFor(ctx, DefaultQueue, "dead")
	if err != nil || job == nil {
		t.Fatal(err)
	}

	n, err := store.RequeueOrphaned(ctx, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("requeued = %d, want 1", n)
	}
	got, err := store.Get(ctx, job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != StatusReady {
		t.Fatalf("status = %q, want ready", got.Status)
	}
}

func TestList_filtersByKind(t *testing.T) {
	store := NewStore(testDB(t))
	ctx := context.Background()
	if _, err := Enqueue(ctx, store, Options{Kind: "Mail"}); err != nil {
		t.Fatal(err)
	}
	if _, err := Enqueue(ctx, store, Options{Kind: "Ping"}); err != nil {
		t.Fatal(err)
	}
	got, err := store.List(ctx, ListFilter{Kind: "Mail"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Kind != "Mail" {
		t.Fatalf("list by kind = %+v", got)
	}
}

// #111: ClaimFor increments attempts but never checks max_attempts, so a job
// whose handler kills the process was requeued by RequeueOrphaned and claimed
// again forever (poison job loop). An orphan at the attempt cap must fail.
func TestRequeueOrphaned_failsJobAtMaxAttempts(t *testing.T) {
	store := NewStore(testDB(t))
	ctx := context.Background()
	if _, err := Enqueue(ctx, store, Options{Kind: "Poison", MaxAttempts: 1}); err != nil {
		t.Fatal(err)
	}
	job, err := store.ClaimFor(ctx, DefaultQueue, "dead")
	if err != nil || job == nil {
		t.Fatal(err)
	}

	if _, err := store.RequeueOrphaned(ctx, time.Minute); err != nil {
		t.Fatal(err)
	}
	got, err := store.Get(ctx, job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != StatusFailed {
		t.Fatalf("status = %q, want failed (attempt cap reached)", got.Status)
	}
	if got.LastError == "" {
		t.Error("failed orphan should explain why")
	}

	again, err := store.ClaimFor(ctx, DefaultQueue, "w2")
	if err != nil {
		t.Fatal(err)
	}
	if again != nil {
		t.Fatal("exhausted job was claimed again")
	}
}

// Jobs with attempts left keep the old behavior: back to ready and claimable.
func TestRequeueOrphaned_requeuesJobWithRemainingAttempts(t *testing.T) {
	store := NewStore(testDB(t))
	ctx := context.Background()
	if _, err := Enqueue(ctx, store, Options{Kind: "Transient", MaxAttempts: 3}); err != nil {
		t.Fatal(err)
	}
	job, err := store.ClaimFor(ctx, DefaultQueue, "dead")
	if err != nil || job == nil {
		t.Fatal(err)
	}

	if _, err := store.RequeueOrphaned(ctx, time.Minute); err != nil {
		t.Fatal(err)
	}
	got, err := store.Get(ctx, job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != StatusReady {
		t.Fatalf("status = %q, want ready", got.Status)
	}
	again, err := store.ClaimFor(ctx, DefaultQueue, "w2")
	if err != nil || again == nil || again.ID != job.ID {
		t.Fatalf("job with attempts left should be claimable again: %v %v", again, err)
	}
}

// #138: the retry UPDATE left worker_id/started_at set, so the dashboard showed
// a ready job as if it belonged to a worker.
func TestMarkFailed_retryClearsWorkerBinding(t *testing.T) {
	store := NewStore(testDB(t))
	ctx := context.Background()
	if _, err := Enqueue(ctx, store, Options{Kind: "Flaky", MaxAttempts: 3}); err != nil {
		t.Fatal(err)
	}
	job, err := store.ClaimFor(ctx, DefaultQueue, "worker-1")
	if err != nil || job == nil {
		t.Fatal(err)
	}
	bound, err := store.Get(ctx, job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if bound.WorkerID != "worker-1" || bound.StartedAt == "" {
		t.Fatalf("claim should bind the worker: %+v", bound)
	}

	if err := store.MarkFailed(ctx, job.ID, errTestFail, job.Attempts, job.MaxAttempts); err != nil {
		t.Fatal(err)
	}
	rec, err := store.Get(ctx, job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Status != StatusReady {
		t.Fatalf("status = %q, want ready", rec.Status)
	}
	if rec.WorkerID != "" {
		t.Errorf("worker_id = %q, want cleared on retry", rec.WorkerID)
	}
	if rec.StartedAt != "" {
		t.Errorf("started_at = %q, want cleared on retry", rec.StartedAt)
	}
}
