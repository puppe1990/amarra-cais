---
title: Background jobs
description: "Queue work in SQLite: generate jobs, run a worker process, enqueue from handlers, and manage jobs from the dashboard."
sidebar:
  order: 6
---

Amarra runs background work on a SQLite queue in `pkg/cais/jobs`, stored in the same database file as your app — no Redis, no extra service. Generate your first job:

```bash
amarra-cais g job send_welcome --cron "0 3 * * *"
```

This writes `internal/jobs/*.go`, registers the handler in `internal/jobs/registry.go`, and scaffolds `cmd/worker`. The `--cron` flag schedules a recurring task. Then create the `jobs` and `recurring_tasks` tables:

```bash
amarra-cais db migrate
```

## Run the worker

In production, run the worker as a separate process next to `bin/server`:

```bash
amarra-cais jobs work --concurrency 2
```

The worker also runs the delayed-job dispatcher and publishes a heartbeat. Restrict it to specific queues with `--queues`:

```bash
amarra-cais jobs work --queues default,mail --concurrency 2
```

Because the queue is one SQLite file, two live workers both writing it produce a warning on the dashboard.

## Enqueue work from a handler

```go
jobs.Enqueue(ctx, jobStore, jobs.Options{Kind: "SendWelcome", Payload: data})
```

`Kind` must match the name registered in `internal/jobs/registry.go`. `PruneSessions` is built in, so session cleanup can run as a job instead of a cron script.

## Inspect and recover

The same store exposes the operational API for scripts and the console:

```go
store.List(ctx, jobs.ListFilter{Status: jobs.StatusFailed, Kind: "SendWelcome"})
store.RetryFailed(ctx, id)
store.Discard(ctx, id)
store.PruneFinished(ctx, 24*time.Hour) // 0 = all finished rows
store.RequeueOrphaned(ctx, jobs.DefaultWorkerStale)
store.ListLiveWorkers(ctx, jobs.DefaultWorkerStale)
```

From the CLI:

```bash
amarra-cais jobs status      # counts + queues + workers + recurring
amarra-cais jobs retry 12
amarra-cais jobs discard 12
amarra-cais jobs prune --older 24h
```

## The /jobs dashboard

`jobsui.Register` in `app.New` mounts `GET /jobs` (localhost only, in every environment), with `GET /jobs/{id}` for a single job and `?kind=` to filter. From there you can retry or discard failed jobs, requeue stuck orphans (which skips live worker jobs), and clear finished rows. Worker heartbeats show which workers are alive.

:::tip
When the app is remote, reach the dashboard over an SSH tunnel: `ssh -L 8080:127.0.0.1:8080 user@host`. `amarra-cais routes` lists the `/jobs` routes when `jobsui.Register` is present, and `amarra-cais doctor` warns when it is missing.
:::

:::note
Recurring tasks are claimed with a compare-and-set each tick, so only one worker runs a given schedule even when several workers share the database.
:::

## Related

- [Jobs and SQLite](/amarra-cais/docs/explanation/jobs-and-sqlite/) — why the queue lives in the app database.
- [Jobs API](/amarra-cais/docs/reference/jobs-api/) — the full store and worker surface.
- [CLI reference](/amarra-cais/docs/reference/cli/) — `jobs work|status|retry|discard|prune`.
- [Database and migrations](/amarra-cais/docs/how-to/database-and-migrations/) — migrations add the queue tables.
