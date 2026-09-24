---
title: Database and migrations
description: Add a SQLite table to an Amarra app, write the migration, and run it from the CLI.
sidebar:
  order: 4
---

An Amarra app keeps its data in SQLite (via `modernc.org/sqlite`, no CGO). Tables are created by ordered SQL migration files that the app tracks in `schema_migrations` and applies idempotently on boot.

## Add a table

Follow the same order the generators use — test first:

1. Write a store test in `internal/store/` against SQLite `:memory:`.
2. Add the SQL file `internal/store/migrations/NNN_name.sql`.
3. Add methods to the `store.Store` interface and its implementation.
4. Wrap the DB with `sqllog.Wrap` in `NewSQLiteStore` so development query logs are captured.
5. The migration is recorded in `schema_migrations` by `pkg/cais/migrate`, which is idempotent on boot.

Or scaffold the data layer:

```bash
amarra-cais g model bookmark --fields title:string,url:url
```

`g model` writes the model struct, the migration, and the store methods — no handlers, templates, or routes. For an empty migration, use `amarra-cais g migration add_tags`.

## Write the migration

Each file uses `-- up` and `-- down` markers so rollback has SQL to run:

```sql
-- up
CREATE TABLE bookmarks (
  id    INTEGER PRIMARY KEY,
  title TEXT    NOT NULL,
  url   TEXT    NOT NULL
);

-- down
DROP TABLE IF EXISTS bookmarks;
```

`amarra-cais db rollback` executes the `-- down` SQL when present, then removes the row from `schema_migrations`. Without a down section, only the record is removed.

## Run migrations

```bash
amarra-cais db migrate        # run pending migrations
amarra-cais db status         # list applied and pending migrations
amarra-cais db rollback       # roll back the last migration
```

Run `amarra-cais db migrate` after `g resource`, `g model`, or `g auth`.

:::note
Store tests use SQLite `:memory:` — do not mock the database. Run them with `amarra-cais test` or `go test ./...`.
:::

## SQLite concurrency

The scaffold's `NewSQLiteStore` calls `sqlite.Configure`, which sets `journal_mode=WAL`, `busy_timeout=5000`, `foreign_keys=ON`, and `MaxOpenConns(1)`. WAL lets readers run while a writer briefly holds the lock, and `busy_timeout` retries instead of failing immediately with `SQLITE_BUSY`.

SSE handlers poll the database in a loop, so keep writes short and avoid long transactions during a stream. For heavy write concurrency, consider a dedicated writer queue.

## Related

- [Jobs and SQLite](/amarra-cais/docs/explanation/jobs-and-sqlite/) — why SQLite, and how the job queue shares the file.
- [CLI reference](/amarra-cais/docs/reference/cli/) — the `db` subcommands.
