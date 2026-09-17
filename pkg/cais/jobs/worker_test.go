package jobs

import (
	"bytes"
	"context"
	"errors"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestWorker_runsRegisteredJob(t *testing.T) {
	db := testDB(t)
	store := NewStore(db)
	ctx, cancel := context.WithCancel(context.Background())

	var ran atomic.Bool
	reg := NewRegistry()
	reg.Register("Ping", func(ctx context.Context, payload []byte) error {
		ran.Store(true)
		cancel()
		return nil
	})

	if _, err := Enqueue(ctx, store, Options{Kind: "Ping"}); err != nil {
		t.Fatal(err)
	}

	w := NewWorker(WorkerConfig{
		Store:            store,
		Registry:         reg,
		Concurrency:      1,
		PollInterval:     20 * time.Millisecond,
		DispatchInterval: time.Hour,
	})
	go func() { _ = w.Run(ctx) }()

	deadline := time.After(2 * time.Second)
	for !ran.Load() {
		select {
		case <-deadline:
			t.Fatal("job did not run")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func TestWorker_writesHeartbeatAndPruneRecurring(t *testing.T) {
	db := testDB(t)
	store := NewStore(db)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w := NewWorker(WorkerConfig{
		Store:             store,
		Registry:          NewRegistry(),
		Concurrency:       1,
		PollInterval:      20 * time.Millisecond,
		DispatchInterval:  time.Hour,
		SchedulerInterval: time.Hour,
	})
	done := make(chan struct{})
	go func() {
		_ = w.Run(ctx)
		close(done)
	}()

	deadline := time.After(2 * time.Second)
	for {
		live, err := store.ListLiveWorkers(ctx, time.Minute)
		if err != nil {
			t.Fatal(err)
		}
		tasks, err := store.ListRecurring(ctx)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, task := range tasks {
			if task.Kind == KindPruneFinished {
				found = true
			}
		}
		if len(live) == 1 && found {
			break
		}
		select {
		case <-deadline:
			t.Fatalf("heartbeat/recurring never appeared live=%d recurring=%+v, want %s", len(live), tasks, KindPruneFinished)
		default:
			time.Sleep(20 * time.Millisecond)
		}
	}

	cancel()
	<-done
	live, err := store.ListLiveWorkers(context.Background(), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if len(live) != 0 {
		t.Fatalf("heartbeat survived shutdown: %+v", live)
	}
}

func TestPruneFinishedHandler_deletesOldFinished(t *testing.T) {
	db := testDB(t)
	store := NewStore(db)
	ctx := context.Background()
	id, err := Enqueue(ctx, store, Options{Kind: "Done"})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.MarkFinished(ctx, id); err != nil {
		t.Fatal(err)
	}
	h := PruneFinishedHandler(db)
	if err := h(ctx, []byte(`{"older_than_hours":0}`)); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Get(ctx, id); err != ErrNotFound {
		t.Fatalf("expected pruned, err=%v", err)
	}
}

func TestPruneSessionsHandler(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	h := PruneSessionsHandler(db)
	if err := h(ctx, nil); err != nil {
		t.Fatal(err)
	}
}

// #109: a handler panic used to escape the worker goroutine and kill the whole
// process, taking every queue down. It must fail the job (panic in last_error,
// stack in the log) and leave the worker running.
func TestWorker_panicInHandlerMarksJobFailed(t *testing.T) {
	db := testDB(t)
	store := NewStore(db)
	ctx := context.Background()

	reg := NewRegistry()
	reg.Register("Boom", func(ctx context.Context, payload []byte) error {
		panic("handler exploded")
	})

	id, err := Enqueue(ctx, store, Options{Kind: "Boom", MaxAttempts: 1})
	if err != nil {
		t.Fatal(err)
	}

	var logBuf bytes.Buffer
	w := NewWorker(WorkerConfig{
		Store:            store,
		Registry:         reg,
		Concurrency:      1,
		PollInterval:     time.Millisecond,
		DispatchInterval: time.Hour,
		Logger:           log.New(&logBuf, "", 0),
	})
	if err := w.pollOnce(ctx); err != nil {
		t.Fatal(err)
	}

	rec, err := store.Get(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Status != StatusFailed {
		t.Fatalf("status = %q, want failed", rec.Status)
	}
	if !strings.Contains(rec.LastError, "panic") {
		t.Fatalf("last_error = %q, want panic details", rec.LastError)
	}
	if !strings.Contains(logBuf.String(), "handler exploded") {
		t.Errorf("log should carry the panic, got: %s", logBuf.String())
	}

	// The worker must keep processing jobs after the panic.
	var ran atomic.Bool
	reg.Register("Ping", func(ctx context.Context, payload []byte) error {
		ran.Store(true)
		return nil
	})
	if _, err := Enqueue(ctx, store, Options{Kind: "Ping"}); err != nil {
		t.Fatal(err)
	}
	if err := w.pollOnce(ctx); err != nil {
		t.Fatal(err)
	}
	if !ran.Load() {
		t.Fatal("worker stopped processing jobs after a handler panic")
	}
}

// #110: Run returned as soon as ctx was done and removed the heartbeat while
// handlers were still running. RequeueOrphaned then re-enqueued the live job —
// duplicate execution of non-idempotent work (mail, billing) on every deploy.
func TestWorker_shutdownDrainsInFlightJob(t *testing.T) {
	db := testDB(t)
	store := NewStore(db)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	started := make(chan struct{})
	release := make(chan struct{})
	reg := NewRegistry()
	reg.Register("Slow", func(ctx context.Context, payload []byte) error {
		close(started)
		<-release
		return nil
	})
	id, err := Enqueue(ctx, store, Options{Kind: "Slow"})
	if err != nil {
		t.Fatal(err)
	}

	w := NewWorker(WorkerConfig{
		Store:             store,
		Registry:          reg,
		Concurrency:       1,
		PollInterval:      10 * time.Millisecond,
		DispatchInterval:  time.Hour,
		SchedulerInterval: time.Hour,
		DrainTimeout:      2 * time.Second,
	})
	done := make(chan error, 1)
	go func() { done <- w.Run(ctx) }()

	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("job never started")
	}

	cancel()

	select {
	case err := <-done:
		t.Fatalf("Run returned before draining the in-flight job (err=%v)", err)
	case <-time.After(150 * time.Millisecond):
	}

	close(release)
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Run did not return after the handler finished")
	}

	rec, err := store.Get(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Status != StatusFinished {
		t.Fatalf("status = %q, want finished (final mark must survive cancel)", rec.Status)
	}
	live, err := store.ListLiveWorkers(context.Background(), time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if len(live) != 0 {
		t.Fatalf("heartbeat should be removed after drain: %+v", live)
	}
	n, err := store.RequeueOrphaned(context.Background(), time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("RequeueOrphaned requeued %d job(s) after a clean shutdown", n)
	}
}

// #118: any pollOnce error (e.g. SQLITE_BUSY under lock contention) was pushed
// to errCh and killed Run — the generated cmd/worker then log.Fatal'd, taking
// job processing down until a manual restart.
func TestWorker_pollErrorDoesNotKillWorker(t *testing.T) {
	db := testDB(t)
	store := NewStore(db)
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logBuf := &syncBuffer{}
	w := NewWorker(WorkerConfig{
		Store:             store,
		Registry:          NewRegistry(),
		Concurrency:       1,
		PollInterval:      5 * time.Millisecond,
		DispatchInterval:  time.Hour,
		SchedulerInterval: time.Hour,
		Logger:            log.New(logBuf, "", 0),
	})
	done := make(chan error, 1)
	go func() { done <- w.Run(ctx) }()

	select {
	case err := <-done:
		t.Fatalf("Run returned on a transient store error (err=%v)", err)
	case <-time.After(150 * time.Millisecond):
	}

	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Run returned %v, want context.Canceled", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Run did not stop after cancel")
	}
	if !strings.Contains(logBuf.String(), "jobs poll") {
		t.Errorf("poll errors should be logged, got: %s", logBuf.String())
	}
}

// syncBuffer is a race-safe log sink: worker goroutines keep logging after
// Run returns.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}
