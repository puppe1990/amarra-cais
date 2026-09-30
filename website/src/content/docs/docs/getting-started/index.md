---
title: Getting started
description: What Amarra is, how the pieces fit together, and where to go next.
sidebar:
  order: 1
---

Amarra is the HTML-first Go framework behind the `amarra-cais` CLI. It ships a scaffold for mini apps that are easy to host on a small VPS — Lightsail-friendly — and easy to keep in your head.

## What you get

| Layer    | Choice                                                            |
| -------- | ----------------------------------------------------------------- |
| Language | Go 1.26 (`net/http` stdlib)                                       |
| Frontend | Amarra Views + Drive (`pkg/amarra/view` + `/static/js/amarra.js`) |
| CSS      | Tailwind CSS 3.x                                                  |
| DB       | SQLite (`modernc.org/sqlite`, no CGO)                             |
| PWA      | Manifest, service worker, offline page, icons, fullscreen         |
| Core     | Router, sessions, CSRF, jobs, i18n under `pkg/cais/`              |

The browser does not mount a SPA. Handlers call `view.Write` and Drive morphs `#amarra-main` — no Vite, no Svelte, no Inertia, no HTMX in generated apps.

Repeating UI comes from the shipped **kit + hooks**. Full tables live in [Views and the shipped kit](/amarra-cais/docs/reference/views-and-kit/#shipped-kit) (`amarra-cais g component --list`) and [amarra.js](/amarra-cais/docs/reference/amarra-js/). Sort and filter are plain `GET ?q=&sort=` round-trips; `amarra-cais g resource` emits the kit.

## One SQLite file, one replica

The scaffold opens one SQLite file with `sqlite.Configure`: `journal_mode=WAL`, `busy_timeout=5000`, `foreign_keys=ON`, and `MaxOpenConns(1)`. That is a single writer. The HTTP server and `amarra-cais jobs work` are meant to share that file on one host.

Live (`GET /amarra/live`) is opt-in and in-process — **single-replica**. Two app replicas do not share sockets; cross-replica fan-out needs an external bus.

This is the mini-app cut. Leaving it means another store, or generated sqlc on top of `database/sql` — not a second replica of the same SQLite file. The default scaffold stays one file. See [Jobs and SQLite](/amarra-cais/docs/explanation/jobs-and-sqlite/).

## Where to go next

- [Installation](/amarra-cais/docs/getting-started/installation/) — install the CLI and check the version.
- [Your first app](/amarra-cais/docs/getting-started/your-first-app/) — scaffold, run the dev server, log in.
- [Project layout](/amarra-cais/docs/getting-started/project-layout/) — what `amarra-cais new` writes and why.
- [How-to guides](/amarra-cais/docs/how-to/) — task-focused recipes: forms, auth, jobs, streaming, deploy.
- [Reference](/amarra-cais/docs/reference/) — CLI, generators, router, kit, middleware, jobs API.
- [Explanation](/amarra-cais/docs/explanation/) — why HTML-first, Drive, and one SQLite file.
