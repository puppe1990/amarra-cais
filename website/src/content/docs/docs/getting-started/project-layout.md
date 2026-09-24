---
title: Project layout
description: What amarra-cais new writes, file by file.
sidebar:
  order: 4
---

A generated app is a small Go program with HTML templates and static assets beside it. There is no build step for the frontend beyond Tailwind.

```
myapp/
  cmd/server/main.go          # boot: config, store, migrations, seeds, router
  cmd/worker/main.go          # background jobs worker (when jobs are used)
  internal/app/               # app.go + routes.go (registerRoutes)
  internal/handlers/          # one file per domain, plus *_test.go
  internal/store/             # Store interface + SQLite implementation
  internal/store/migrations/  # NNN_name.sql (-- up / -- down sections)
  internal/db/seeds.go        # idempotent seeds, dev-only when noted
  internal/jobs/              # job handlers + registry.go
  web/templates/layouts/      # app.html — the shell with #amarra-main
  web/templates/pages/        # one HTML file per page
  web/templates/partials/     # flat partials, addressed by name
  web/templates/components/   # kit overrides (optional)
  web/static/                 # css, js/amarra.js, icons, sw.js, manifest
```

## The request path

1. `internal/app/routes.go` registers routes on a `cais.Router`; middleware wires sessions, flash, CSRF and security headers.
2. A handler builds page data and calls `view.Write(w, r, views, view.Page{Layout, Name, Data}, cfg)`.
3. Templates loaded once at boot render the page inside the layout; Drive morphs `#amarra-main` on navigation instead of reloading the shell.

Templates are parsed a single time with `view.Load` — an unknown `<.component>` tag fails at boot, not on the first request. The exact globs and addressing rules are in [Views and kit](/amarra-cais/docs/reference/views-and-kit/).

## Conventions worth keeping

- One handler file per domain with a matching `*_test.go` (SQLite `:memory:`, no mocks).
- Handlers take dependencies (`Store`, `*view.Renderer`, `cais.Config`) via constructor — no globals.
- `registerRoutes`, `Close() error` and the `<!-- cais:nav -->` marker in the layout are what generators and `amarra-cais destroy` patch; keep them.
- Pages are HTML, not JSON: no `HX-Request` checks, no Inertia payloads.

Next: [Pages and views](/amarra-cais/docs/how-to/pages-and-views/).
