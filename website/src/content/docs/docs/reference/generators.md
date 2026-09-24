---
title: Generators
description: Reference for amarra-cais g and destroy — generators, resource options, field types and the files each one writes.
sidebar:
  order: 3
---

Generators are the `amarra-cais g` family. They run inside a generated app (the CLI checks for a `go.mod` that depends on `amarra-cais`) and write HTML + Amarra scaffolds, patching `routes.go`, `store.go`, `seeds.go` and the layout nav as needed.

## Generators

| Generator                           | Files and patches                                                                                                                                                                                                                                                                             |
| ----------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `g handler <name>`                  | `internal/handlers/<name>.go` + `<name>_test.go`, `web/templates/pages/<name>.html`, and a route added to `internal/app/routes.go`.                                                                                                                                                           |
| `g page <name>`                     | `web/templates/pages/<name>.html` only.                                                                                                                                                                                                                                                       |
| `g resource <name>`                 | `internal/models/<name>.go`; `internal/handlers/admin_<plural>.go` (+`_test.go`); `web/templates/pages/admin_<plural>.html`, `admin_<name>_show.html`, `admin_<name>_form.html`; `web/templates/partials/admin_<name>_form_errors.html`; a new migration; store, route, seed and nav patches. |
| `g model <name>`                    | `internal/models/<name>.go`, a migration, and store methods — no handlers, templates or routes.                                                                                                                                                                                               |
| `g migration <name>`                | `internal/store/migrations/NNN_<name>.sql` (next number in sequence).                                                                                                                                                                                                                         |
| `g auth`                            | Login/logout handlers, pages and tests, a protected `/dashboard`, and session middleware wired in `app.go`.                                                                                                                                                                                   |
| `g console`                         | `cmd/console/main.go`.                                                                                                                                                                                                                                                                        |
| `g ci`                              | GitHub Actions workflow, pre-commit hooks, golangci-lint config and Prettier (patches `Makefile` and `package.json`).                                                                                                                                                                         |
| `g job <name> [--cron "0 3 * * *"]` | `internal/jobs/<name>.go` (+`_test.go`), `internal/jobs/registry.go`, and `cmd/worker/main.go`.                                                                                                                                                                                               |
| `g stream chat [--live]`            | Conversation/message models, `internal/handlers/chat.go` (+`_test.go`), `conversations.html` / `chat.html`, and `chat_sse*.html` partials.                                                                                                                                                    |
| `g live <name>`                     | `internal/handlers/<name>_live.go` (+`_test.go`) and `web/templates/pages/<name>.html`.                                                                                                                                                                                                       |
| `g component <name>`                | `web/templates/components/<stem>.html`. A shipped kit name seeds that component's real markup; other names get a generic slot.                                                                                                                                                                |

The CLI also ships `g sitemap` (blog posts + a dynamic `/sitemap.xml`).

## Resource options

`amarra-cais g resource bookmark --fields title:string,url:url,notes:text? --public --paginate`

| Option                         | Effect                                                                              |
| ------------------------------ | ----------------------------------------------------------------------------------- |
| `--fields <spec>`              | Comma-separated field list. Defaults to `name:string`.                              |
| `--public`                     | Also generate a public list page (kit `<.filters>` / `<.empty>` / `<.pagination>`). |
| `--paginate`                   | Paginate the admin index, 25 rows per page.                                         |
| `--no-seed`                    | Skip the `SeedDemo*` demo data.                                                     |
| `--admin-auth session\|bearer` | Admin protection mode: browser session (default) or bearer token.                   |
| `--force`                      | Overwrite files that already exist.                                                 |

Generated store methods take `(search, sort, dir string, ...)`; sort is whitelisted per column and search is a `LIKE` on the display field. The admin index renders `<.filters>` + `<.table>` + `<.empty>` + `<.pagination>`.

## Field types

Each field is `name:type`; append `?` to make it optional. `name:belongs_to` is shorthand for `name_id:references`.

| Type         | Column                                      | Input                       |
| ------------ | ------------------------------------------- | --------------------------- |
| `string`     | `TEXT [NOT NULL]`                           | text input                  |
| `text`       | `TEXT` (textarea)                           | textarea                    |
| `url`        | `TEXT [NOT NULL]`                           | `type="url"` input          |
| `bool`       | `INTEGER NOT NULL DEFAULT 0`                | checkbox                    |
| `int`        | `INTEGER [NOT NULL DEFAULT 0]`              | number input                |
| `float`      | `REAL [NOT NULL DEFAULT 0]`                 | number input (`step="any"`) |
| `date`       | `TEXT [NOT NULL]`                           | `type="date"` input         |
| `references` | `INTEGER [NOT NULL] REFERENCES <table>(id)` | select (parent row)         |

An optional field maps to a pointer (`*string`, `*int64`, `*float64`). For a `references` field, generate the parent resource first and give the parent table a `name` or `title` column for labels.

```bash
amarra-cais g resource post --fields title:string,category_id:references
# or: --fields title:string,category:belongs_to
```

## Component overrides

`amarra-cais g component --list` prints the shipped kit components an app can override. The list is framework-side, so it works without an app. Overriding a kit name seeds its real markup so you restyle the contract instead of recreating it; the file keeps the kit stem (for example `locale-toggle.html`).

## Dry-run

`amarra-cais g --dry-run …` and `amarra-cais destroy --dry-run …` print the planned changes without writing files.

## Destroy

`amarra-cais destroy resource|handler|model|component <name>` removes the generated files and unpatches `routes.go`, `store.go`, `seeds.go` and the layout nav where applicable. `destroy auth` also reverts the `app.go` session middleware. `destroy migration <name>` removes the matching `*_<name>.sql` only — it does not roll back `schema_migrations`.

:::caution
After `g resource`, `g model` or `g auth`, run `amarra-cais db migrate`. If a patch fails with `could not patch routes.go` or `could not patch store`, check that the `registerRoutes` and `Close() error` markers still exist; public nav needs the `<!-- cais:nav -->` marker (or `</nav>`).
:::

See [Database and migrations](/amarra-cais/docs/how-to/database-and-migrations/) for the SQL the resource generator writes, and [CLI](/amarra-cais/docs/reference/cli/) for the full command list.
