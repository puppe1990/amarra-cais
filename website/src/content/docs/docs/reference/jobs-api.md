---
title: Jobs API
description: The Go API and localhost dashboard for Amarra's SQLite background job queue.
sidebar:
  order: 10
---

`pkg/cais/jobs` is a SQLite-backed job queue that lives in the same database file as the app — no Redis. Enqueue work from handlers, run a worker process, and inspect or recover jobs from the localhost dashboard and the CLI.

## Enqueue

```go
jobs.Enqueue(ctx, jobStore, jobs.Options{Kind: "SendWelcome", Payload: data})
```

`jobs.Options` carries `Kind` (the handler name) and `Payload` (the data the handler receives).

## Store methods

Register handlers in `internal/jobs/registry.go`; the built-in handler is `PruneSessions`. The store exposes inspection and recovery methods:

```go
store.List(ctx, jobs.ListFilter{Status: jobs.StatusFailed, Kind: "SendWelcome"})
store.RetryFailed(ctx, id)
store.Discard(ctx, id)
store.PruneFinished(ctx, 24*time.Hour) // 0 = all finished rows
store.RequeueOrphaned(ctx, jobs.DefaultWorkerStale)
store.ListLiveWorkers(ctx, jobs.DefaultWorkerStale)
```

| Method                                     | Purpose                                                                      |
| ------------------------------------------ | ---------------------------------------------------------------------------- |
| `List(ctx, jobs.ListFilter{Status, Kind})` | List jobs, optionally filtered by status and kind.                           |
| `RetryFailed(ctx, id)`                     | Requeue one failed job.                                                      |
| `Discard(ctx, id)`                         | Drop one failed job.                                                         |
| `PruneFinished(ctx, older)`                | Delete finished rows older than the duration (`0` clears all finished rows). |
| `RequeueOrphaned(ctx, stale)`              | Requeue jobs stuck on a dead worker.                                         |
| `ListLiveWorkers(ctx, stale)`              | List worker heartbeats seen within the staleness window.                     |

`jobs.StatusFailed` is the status constant used to filter failed jobs. `jobs.DefaultWorkerStale` is the package's default staleness window for orphan detection and worker liveness.

## Worker

```bash
amarra-cais jobs work --queues default,mail --concurrency 2
```

The worker runs jobs plus the delayed-job dispatcher and writes a heartbeat so the dashboard can show liveness. Flags: `--queues` (comma-separated list) and `--concurrency` (number of workers). In production, run it as a separate process next to `bin/server`.

```bash
amarra-cais jobs status            # counts + queues + workers + recurring
amarra-cais jobs retry 12
amarra-cais jobs discard 12
amarra-cais jobs prune [--older 24h]
```

:::caution
Two live workers on one SQLite file trigger a warning — heartbeats are how the dashboard detects them.
:::

## Recurring tasks

`amarra-cais g job` scaffolds the handler, the registry entry, and the worker command:

```bash
amarra-cais g job prune_sessions --cron "0 3 * * *"
amarra-cais db migrate      # creates the jobs + recurring_tasks tables
```

The cron expression registers a recurring task, and `amarra-cais db migrate` is required after generating to create the `jobs` and `recurring_tasks` tables.

## Dashboard

`jobsui.Register(r, db)` mounts the dashboard (localhost only, all environments; use an SSH tunnel when remote). `amarra-cais doctor` warns if `jobsui.Register` is missing.

| Route            | Purpose                                    |
| ---------------- | ------------------------------------------ |
| `GET /jobs`      | Job list and counts. Filter with `?kind=`. |
| `GET /jobs/{id}` | One job's detail.                          |

From the dashboard you can Retry or Discard failed jobs, Requeue stuck orphans (skipping jobs whose worker is still live), and clear finished jobs. Those actions map to the store methods above: Retry calls `RetryFailed`, Discard calls `Discard`, Requeue stuck calls `RequeueOrphaned`, and Clear finished calls `PruneFinished`. Scheduled and recurring tasks are shown there too. Open `http://127.0.0.1:<port>/jobs`.

`amarra-cais routes` lists the dashboard routes when `jobsui.Register` is present.

See [Background jobs](/amarra-cais/docs/how-to/background-jobs/) and the [jobs and SQLite design](/amarra-cais/docs/explanation/jobs-and-sqlite/).
