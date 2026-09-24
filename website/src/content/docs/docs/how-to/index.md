---
title: How-to guides
description: Task-focused guides for adding pages, forms, a database, auth, jobs, and shipping an Amarra app.
sidebar:
  order: 1
---

These guides are task-focused walkthroughs for a scaffolded Amarra app. Each one assumes you generated an app with `amarra-cais new` and can run it with `amarra-cais dev`. Commands run from the app directory unless a step says otherwise, and the file paths are the ones the scaffold creates (`internal/handlers/`, `web/templates/`, `internal/store/`).

Start with [Pages and views](/amarra-cais/docs/how-to/pages-and-views/) if you have never added a route, then follow the guide that matches your current task.

:::tip
Amarra is test-driven. Write the Go test first, confirm it fails for the right reason, then add the minimal code to pass it. Scaffolded apps run tests with `go test ./...` or `amarra-cais test`.
:::

## Guides

- [Pages and views](/amarra-cais/docs/how-to/pages-and-views/) — generate a handler and a page, render with `view.Write`, and re-render on validation.
- [Forms and validation](/amarra-cais/docs/how-to/forms-and-validation/) — the shipped kit form tags, CSRF, server-side field errors, and file uploads.
- [Database and migrations](/amarra-cais/docs/how-to/database-and-migrations/) — add a SQLite table, write the migration, and run it from the CLI.
- [Auth and sessions](/amarra-cais/docs/how-to/auth-and-sessions/) — login, logout, protected routes, and session storage.
- [Background jobs](/amarra-cais/docs/how-to/background-jobs/) — enqueue work on the SQLite queue and run a worker.
- [Streaming chat](/amarra-cais/docs/how-to/streaming-chat/) — scaffold an SSE chat with the named stream ops.
- [Live updates](/amarra-cais/docs/how-to/live-updates/) — the opt-in WebSocket hub for real-time views.
- [PWA and mobile](/amarra-cais/docs/how-to/pwa-and-mobile/) — installability, the service worker, and LAN testing on a phone.
- [i18n](/amarra-cais/docs/how-to/i18n/) — switch UI strings between English and Portuguese.
- [Deploy](/amarra-cais/docs/how-to/deploy/) — cross-compile, ship static assets, and run under systemd.
- [Upgrade](/amarra-cais/docs/how-to/upgrade/) — bump the framework and apply the migration checklist.

## Related

- [CLI reference](/amarra-cais/docs/reference/cli/) — every `amarra-cais` command and subcommand.
- [Generators](/amarra-cais/docs/reference/generators/) — `g handler`, `g page`, `g resource`, and the rest.
- [Views and the kit](/amarra-cais/docs/reference/views-and-kit/) — the shipped components and hooks these guides use.
