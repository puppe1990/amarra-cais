package jobs

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

// WorkerConfig configures the job worker and dispatcher loops.
type WorkerConfig struct {
	Store             *Store
	Registry          *Registry
	Queues            []string
	Concurrency       int
	PollInterval      time.Duration
	DispatchInterval  time.Duration
	SchedulerInterval time.Duration
	// DrainTimeout bounds how long Run waits for in-flight handlers before
	// removing the heartbeat on shutdown (#110). Defaults to 30s.
	DrainTimeout time.Duration
	Logger       *log.Logger
}

// Worker processes jobs from SQLite.
type Worker struct {
	cfg WorkerConfig
	id  string
	wg  sync.WaitGroup
}

func NewWorker(cfg WorkerConfig) *Worker {
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = time.Second
	}
	if cfg.DispatchInterval <= 0 {
		cfg.DispatchInterval = time.Second
	}
	if cfg.SchedulerInterval <= 0 {
		cfg.SchedulerInterval = time.Minute
	}
	if cfg.DrainTimeout <= 0 {
		cfg.DrainTimeout = 30 * time.Second
	}
	if len(cfg.Queues) == 0 {
		cfg.Queues = []string{DefaultQueue}
	}
	if cfg.Concurrency < 1 {
		cfg.Concurrency = 1
	}
	if cfg.Registry == nil {
		cfg.Registry = NewRegistry()
	}
	if cfg.Logger == nil {
		cfg.Logger = log.Default()
	}
	return &Worker{cfg: cfg}
}

// Run starts dispatcher and worker goroutines until ctx is cancelled.
func (w *Worker) Run(ctx context.Context) error {
	w.id = NewWorkerID()
	pulse := w.pulse()
	// Drain before RemoveWorker: dropping the heartbeat while a handler is
	// still running lets RequeueOrphaned duplicate the job (#110).
	defer func() {
		w.drain()
		_ = w.cfg.Store.RemoveWorker(context.Background(), w.id)
	}()

	// Recover jobs whose worker heartbeat is gone (#172) without stealing live work.
	if n, err := w.cfg.Store.RequeueOrphaned(ctx, DefaultWorkerStale); err != nil {
		w.cfg.Logger.Printf("jobs requeue-orphaned: %v", err)
	} else if n > 0 {
		w.cfg.Logger.Printf("jobs requeued %d orphaned job(s) from previous run", n)
	}

	if err := w.ensureFinishedPrune(ctx); err != nil {
		w.cfg.Logger.Printf("jobs prune-finished recurring: %v", err)
	}

	// Heartbeat last so a live worker implies prune recurring is already registered.
	if err := w.cfg.Store.TouchWorker(ctx, pulse); err != nil {
		w.cfg.Logger.Printf("jobs heartbeat: %v", err)
	}

	go func() {
		ticker := time.NewTicker(w.cfg.PollInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := w.cfg.Store.TouchWorker(ctx, pulse); err != nil {
					w.cfg.Logger.Printf("jobs heartbeat: %v", err)
				}
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(w.cfg.DispatchInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if _, err := w.cfg.Store.DispatchDue(ctx, time.Now().UTC()); err != nil {
					w.cfg.Logger.Printf("jobs dispatcher: %v", err)
				}
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(w.cfg.SchedulerInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if n, err := RunScheduler(ctx, w.cfg.Store, time.Now().UTC()); err != nil {
					w.cfg.Logger.Printf("jobs scheduler: %v", err)
				} else if n > 0 {
					w.cfg.Logger.Printf("jobs scheduler: enqueued %d recurring task(s)", n)
				}
			}
		}
	}()

	for i := 0; i < w.cfg.Concurrency; i++ {
		w.wg.Add(1)
		go func() {
			defer w.wg.Done()
			ticker := time.NewTicker(w.cfg.PollInterval)
			defer ticker.Stop()
			backoff := time.Duration(0)
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					err := w.pollOnce(ctx)
					if err == nil {
						backoff = 0
						continue
					}
					if ctx.Err() != nil {
						return
					}
					// Transient store errors (SQLITE_BUSY under contention)
					// must not kill job processing (#118).
					backoff = nextPollBackoff(backoff)
					w.cfg.Logger.Printf("jobs poll: %v (retrying in %s)", err, backoff)
					if !sleepContext(ctx, backoff) {
						return
					}
				}
			}
		}()
	}

	<-ctx.Done()
	return ctx.Err()
}

const (
	pollBackoffMin = time.Second
	pollBackoffMax = 30 * time.Second
)

// nextPollBackoff doubles the retry delay up to pollBackoffMax (#118).
func nextPollBackoff(current time.Duration) time.Duration {
	if current <= 0 {
		return pollBackoffMin
	}
	if next := current * 2; next <= pollBackoffMax {
		return next
	}
	return pollBackoffMax
}

// sleepContext waits for d or ctx cancellation; false means cancelled.
func sleepContext(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func (w *Worker) pollOnce(ctx context.Context) error {
	for _, queue := range w.cfg.Queues {
		job, err := w.cfg.Store.ClaimFor(ctx, queue, w.id)
		if err != nil {
			return err
		}
		if job == nil {
			continue
		}
		w.runJob(ctx, job)
	}
	return nil
}

func (w *Worker) runJob(ctx context.Context, job *Job) {
	err := w.perform(ctx, job)
	if err == nil {
		markCtx, cancel := finalizeContext(ctx)
		defer cancel()
		if markErr := w.cfg.Store.MarkFinished(markCtx, job.ID); markErr != nil {
			w.cfg.Logger.Printf("jobs finish id=%d: %v", job.ID, markErr)
		}
		w.cfg.Logger.Printf("jobs finished id=%d kind=%s", job.ID, job.Kind)
		return
	}
	markCtx, cancel := finalizeContext(ctx)
	defer cancel()
	if markErr := w.cfg.Store.MarkFailed(markCtx, job.ID, err, job.Attempts, job.MaxAttempts); markErr != nil {
		w.cfg.Logger.Printf("jobs fail id=%d: %v (mark: %v)", job.ID, err, markErr)
		return
	}
	w.cfg.Logger.Printf("jobs failed id=%d kind=%s: %v", job.ID, job.Kind, err)
}

// drain waits for in-flight handlers before the heartbeat is removed (#110),
// bounded by DrainTimeout so a stuck handler cannot block shutdown forever.
func (w *Worker) drain() {
	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(w.cfg.DrainTimeout):
		w.cfg.Logger.Printf("jobs drain: timeout after %s; in-flight jobs may be requeued", w.cfg.DrainTimeout)
	}
}

// finalizeContext detaches from cancellation so recording a finished/failed
// job survives shutdown (#110). Without this, MarkFinished(ctx) fails on a
// cancelled context and the drained job stayed "running".
func finalizeContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
}

// perform isolates handler panics (#109): a buggy handler must fail its own
// job, not kill every queue in the worker process. The stack goes to the log;
// last_error carries a short panic message.
func (w *Worker) perform(ctx context.Context, job *Job) (err error) {
	defer func() {
		if r := recover(); r != nil {
			w.cfg.Logger.Printf("jobs panic id=%d kind=%s: %v\n%s", job.ID, job.Kind, r, debug.Stack())
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	return w.cfg.Registry.Perform(ctx, job.Kind, job.Payload)
}

func (w *Worker) pulse() WorkerPulse {
	host, _ := os.Hostname()
	return WorkerPulse{
		ID:          w.id,
		Hostname:    host,
		PID:         os.Getpid(),
		Queues:      strings.Join(w.cfg.Queues, ","),
		Concurrency: w.cfg.Concurrency,
	}
}

func (w *Worker) ensureFinishedPrune(ctx context.Context) error {
	return w.cfg.Store.UpsertRecurring(ctx, RecurringOptions{
		Kind:    KindPruneFinished,
		Cron:    pruneFinishedCron,
		Payload: map[string]any{"older_than_hours": 24},
	})
}
