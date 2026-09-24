---
title: CLI
description: Every amarra-cais command and alias — scaffold, generate, destroy, dev, build, database, jobs and framework make targets.
sidebar:
  order: 2
---

`amarra-cais` is the Rails-style CLI for Amarra apps. Run app commands from inside a generated app (the ones that need it read the current directory's `go.mod`); `new`, `version`, `help` and `g component --list` work anywhere.

## Scaffold, generate and undo

| Command                                                            | Description                                                                                                                                |
| ------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------ |
| `amarra-cais new <app> [dir] [--minimal\|--blank] [--module path]` | Scaffold an app (default dir `./<app>`). `--minimal` is home only, `--blank` has no starter content, `--module` overrides the module path. |
| `amarra-cais g [--dry-run] <generator> [name]`                     | Run a generator. See [Generators](/amarra-cais/docs/reference/generators/).                                                                |
| `amarra-cais destroy [--dry-run] <kind> <name>`                    | Undo a generator.                                                                                                                          |

### `amarra-cais g` subcommands

| Generator                                   | What it does                                                          |
| ------------------------------------------- | --------------------------------------------------------------------- |
| `g handler <name>`                          | Handler + test + `web/templates/pages/<name>.html` + route patch.     |
| `g page <name>`                             | Page template only.                                                   |
| `g resource <name> [options]`               | Kit-based admin CRUD, optionally a public list.                       |
| `g model <name> [--fields …]`               | Model + migration + store methods (no handlers, templates or routes). |
| `g migration <name>`                        | New `NNN_<name>.sql`.                                                 |
| `g auth`                                    | Login/logout and a protected `/dashboard`.                            |
| `g console`                                 | `cmd/console/main.go`.                                                |
| `g ci`                                      | CI workflow, pre-commit, lint and Prettier.                           |
| `g job <name> [--cron "0 3 * * *"]`         | Job handler, registry and worker.                                     |
| `g stream chat [--live]`                    | SSE chat scaffold.                                                    |
| `g live <name>`                             | Live view + page (WebSocket).                                         |
| `g component <name>` / `g component --list` | Seed a kit override / list overridable components.                    |
| `g sitemap`                                 | Public `post` resource (blog) plus a dynamic `/sitemap.xml` route.    |

### `amarra-cais destroy` subcommands

| Command                    | What it undoes                                                             |
| -------------------------- | -------------------------------------------------------------------------- |
| `destroy resource <name>`  | Resource files and route/store/seeds/nav patches.                          |
| `destroy handler <name>`   | Handler, test, page and route patch.                                       |
| `destroy model <name>`     | Model, migration and store methods.                                        |
| `destroy auth`             | Login/auth scaffolding and the `app.go` session middleware.                |
| `destroy migration <name>` | The matching `*_<name>.sql` only (does not roll back `schema_migrations`). |
| `destroy component <name>` | The kit override file.                                                     |

Add `--dry-run` to either command to print planned changes without writing files.

## App lifecycle

| Command                                                   | Description                                               |
| --------------------------------------------------------- | --------------------------------------------------------- |
| `amarra-cais install`                                     | `npm install` + `go mod tidy` (plus a Tailwind build).    |
| `amarra-cais css`                                         | Build Tailwind CSS → `web/static/css/styles.css`.         |
| `amarra-cais dev`                                         | Hot reload (air + Tailwind watch).                        |
| `amarra-cais build [--os linux] [--arch amd64] [-o path]` | Build `bin/server`, optionally cross-compiled for deploy. |
| `amarra-cais server`                                      | Run the app (`go run ./cmd/server`).                      |
| `amarra-cais test`                                        | Run tests (`go test ./...`).                              |
| `amarra-cais console`                                     | Rails-style REPL (`store`, `cfg`, `db` + SQL).            |
| `amarra-cais routes [--verbose]`                          | List HTTP routes from `internal/app/routes.go`.           |
| `amarra-cais version`                                     | Print the framework version.                              |

## Database

| Command                                              | Description                                                         |
| ---------------------------------------------------- | ------------------------------------------------------------------- |
| `amarra-cais db migrate`                             | Run pending migrations.                                             |
| `amarra-cais db status`                              | List applied and pending migrations.                                |
| `amarra-cais db rollback`                            | Roll back the last migration (runs the `-- down` SQL when present). |
| `amarra-cais db prune-sessions`                      | Delete expired login sessions.                                      |
| `amarra-cais db seed` / `amarra-cais db seed --list` | Run `internal/db/seeds.go` / list the seed helpers it references.   |

## Background jobs

| Command                                                           | Description                                           |
| ----------------------------------------------------------------- | ----------------------------------------------------- |
| `amarra-cais jobs work [--queues default,mail] [--concurrency 2]` | Run the worker, delayed-job dispatcher and heartbeat. |
| `amarra-cais jobs status`                                         | Counts, queues, workers and recurring tasks.          |
| `amarra-cais jobs retry <id>` / `jobs discard <id>`               | Act on a failed job.                                  |
| `amarra-cais jobs prune [--older 24h]`                            | Delete finished jobs.                                 |

The queue lives in the app's SQLite file; the `/jobs` dashboard is served in-process. See [Background jobs](/amarra-cais/docs/how-to/background-jobs/).

## Diagnostics and tooling

| Command                                     | Description                                                                                |
| ------------------------------------------- | ------------------------------------------------------------------------------------------ |
| `amarra-cais doctor [--mobile]`             | Verify `amarra.js`, `layouts/app.html`, air, `go.mod`, and PWA/mobile setup.               |
| `amarra-cais pwa [--bump] [--force]`        | Refresh PWA assets; `--bump` bumps the SW cache, `--force` resets the brand.               |
| `amarra-cais upgrade [version] [--dry-run]` | Bump the framework, run `doctor`, print migration steps.                                   |
| `amarra-cais link [path] [--unlink]`        | Add a `go.mod` `replace` for local framework dev. Do not commit it; unlink before pushing. |

## Aliases

| Alias           | Command    |
| --------------- | ---------- |
| `amarra-cais g` | `generate` |
| `amarra-cais i` | `install`  |
| `amarra-cais b` | `build`    |
| `amarra-cais s` | `server`   |
| `amarra-cais c` | `console`  |

## Framework commands (this repo)

These `make` targets run inside the Amarra framework repo itself, not in a generated app:

| Command            | Description                                                 |
| ------------------ | ----------------------------------------------------------- |
| `make test`        | Go tests with `-race`.                                      |
| `make js-test`     | `pkg/cais/js` + `pkg/amarra/js` unit tests.                 |
| `make lint`        | `golangci-lint`.                                            |
| `make format`      | `prettier --write`.                                         |
| `make ci`          | `test` + `js-test` + `lint` + format-check — the full gate. |
| `make build`       | Build `bin/amarra-cais`.                                    |
| `make install-cli` | `go install ./cmd/amarra-cais`.                             |

:::tip
For a walkthrough of the commands most apps use, see [Upgrade](/amarra-cais/docs/how-to/upgrade/) and [Deploy](/amarra-cais/docs/how-to/deploy/).
:::
