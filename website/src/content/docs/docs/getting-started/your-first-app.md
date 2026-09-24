---
title: Your first app
description: Scaffold an app, install dependencies, and run the dev server.
sidebar:
  order: 3
---

## Scaffold

```bash
amarra-cais new myapp
cd myapp
amarra-cais install   # npm install + go mod tidy + Tailwind build
amarra-cais dev       # http://localhost:8080
```

`new` defaults to the HTML scaffold (Amarra Views + Drive). Two smaller variants exist:

```bash
amarra-cais new myapp --minimal   # fewer pages, same stack
amarra-cais new myapp --blank     # bare skeleton
amarra-cais new myapp --module github.com/acme/myapp
```

The generator also writes `AGENTS.md`, a GitHub Actions CI workflow, pre-commit config, golangci-lint and Prettier setup, so the app starts with the same guardrails as the framework.

## Log in

The dev seed creates a demo user:

```
demo@example.com / password
```

`/dashboard` is protected by session auth out of the box. Seeds do not run when `ENV=production` — use `amarra-cais db seed` for catalog data instead.

## Useful first checks

```bash
amarra-cais doctor           # verifies amarra.js, layouts/app.html, air, go.mod, PWA
amarra-cais routes           # lists HTTP routes from internal/app/routes.go
amarra-cais db migrate       # applies pending migrations
go test ./...                # handler tests run headless against :memory: SQLite
```

If the preferred port is busy, the server picks a free one at boot. `amarra-cais dev` runs air + Tailwind watch, so template edits reload without a restart.

## Next steps

- [Project layout](/amarra-cais/docs/getting-started/project-layout/) — what was generated.
- [Pages and views](/amarra-cais/docs/how-to/pages-and-views/) — add your first screen.
- [Generators reference](/amarra-cais/docs/reference/generators/) — model, resource, auth, jobs, streams.
