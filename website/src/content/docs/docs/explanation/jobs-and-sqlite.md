---
title: Jobs and SQLite
description: Why one SQLite file can serve requests, streams and a background job queue at the same time.
sidebar:
  order: 5
---

Amarra uses a single SQLite file for page data, sessions, and the background job queue. That is deliberate: a small app on one box does not need Redis or a second database, and keeping everything in one file keeps deployment to a binary plus `web/static`.

## How one file stays responsive

Scaffold `NewSQLiteStore` calls `sqlite.Configure`, which sets:

- `journal_mode=WAL` — concurrent readers proceed while one writer holds the lock briefly.
- `busy_timeout=5000` — a blocked writer waits and retries instead of failing immediately with `SQLITE_BUSY`.
- `foreign_keys=ON` — referential integrity is enforced.
- `MaxOpenConns(1)` — a single connection serializes writes.

SSE handlers poll the database in a loop, so the rule during a stream is to keep writes short and avoid long transactions. A write that holds the lock is what makes `busy_timeout` matter. For genuinely heavy write concurrency, route writes through a dedicated queue.

## Jobs share the database

`pkg/cais/jobs` is a Solid Queue–shaped queue backed by the same file. It has `jobs` (ready / running / finished / failed), `scheduled_jobs` for delayed work that a dispatcher promotes when `run_at <= now`, and `recurring_tasks` driven by a cron expression. Workers claim work with an atomic `UPDATE ... RETURNING`, since the `modernc.org/sqlite` driver has no `SKIP LOCKED`.

Enqueue from a handler with `jobs.Enqueue(ctx, store, jobs.Options{Kind: "SendWelcome", Payload: data})` and run the worker as a separate process: `amarra-cais jobs work --concurrency 2`. The dispatcher and heartbeat run inside that worker.

## Heartbeats and orphans

A worker writes a heartbeat (`job_workers`), so `RequeueOrphaned` can return only `running` rows whose worker is missing or stale. A live worker keeps its in-flight jobs; recovery never steals work that is still running. If two live heartbeats appear, the dashboard warns — because one SQLite file should have exactly one `amarra-cais jobs work`.

## The single-replica caveat

One database file means one host. Multiple goroutines via `--concurrency` on that host are supported, but multiple workers across replicas are not, and cross-replica fan-out is out of scope. The same constraint applies to Live: the WebSocket hub broadcasts in-process only, so two app replicas do not share sockets. Jobs also stay in the worker process.

## Pruning

Expired sessions accumulate, so prune them with `session.Store.PruneExpired()` or `amarra-cais db prune-sessions`; the built-in `PruneSessions` job does this on a schedule. Finished jobs are cleared by `PruneFinished` (`amarra-cais jobs prune [--older 24h]`), and the worker registers a daily `0 4 * * *` recurring task for it.

:::caution
Scaling past one host means moving the queue, the rate limiter, and the Live hub off in-process state. Until then, keep requests and the job worker on the same machine and database file.
:::

See the [background jobs how-to](/amarra-cais/docs/how-to/background-jobs/), the [live updates guide](/amarra-cais/docs/how-to/live-updates/), and the [jobs API reference](/amarra-cais/docs/reference/jobs-api/).
