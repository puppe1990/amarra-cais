---
title: Glossary
description: One-sentence definitions of Amarra product terms, each pointing at the contract page.
sidebar:
  order: 6
---

Product words. Each term is one sentence plus the page that owns the contract. Kit component names and hook names live in the tables those pages already pin — this page does not copy them.

| Term | Meaning |
| --- | --- |
| Cais vs Amarra | Two products: Cais v0.11.x is Inertia + Svelte; Amarra is the HTML-first fork. See [Cais vs Amarra](/amarra-cais/docs/explanation/cais-vs-amarra/). |
| Drive | Intercepts same-origin clicks and submits and morphs `#amarra-main`. See [Views and Drive](/amarra-cais/docs/explanation/views-and-drive/) and [amarra.js](/amarra-cais/docs/reference/amarra-js/). |
| Frame | A named region (`<amarra-frame>` / `Amarra-Frame`) that morphs instead of the whole main. See [amarra.js](/amarra-cais/docs/reference/amarra-js/). |
| Stream | Named SSE ops (`append`, `morph`, …) instead of a full-page Drive swap. See [Streaming chat](/amarra-cais/docs/how-to/streaming-chat/). |
| Live | Opt-in in-process WebSocket hub that pushes server-rendered HTML. See [Live updates](/amarra-cais/docs/how-to/live-updates/). |
| kit | Shipped `<.component>` tags (`form`, `input`, …) expanded at boot. See [Views and the shipped kit](/amarra-cais/docs/reference/views-and-kit/). |
| morph | Drive replaces the inner HTML of `#amarra-main` without reloading the shell. See [Views and Drive](/amarra-cais/docs/explanation/views-and-drive/). |
| amarra-hook | `amarra-hook="name"` attaches a built-in behavior from `amarra.js`. See [amarra.js](/amarra-cais/docs/reference/amarra-js/). |
| data-amarra-shell | Marker on `<body>` (`auth` / `app`); a different value forces a full visit instead of a morph (#249). See [amarra.js](/amarra-cais/docs/reference/amarra-js/). |
| #amarra-main | Layout element Drive morphs; the rest of the shell stays. See [Project layout](/amarra-cais/docs/getting-started/project-layout/). |
| g resource | Generator that writes table-admin CRUD (model, store, HTML, routes). See [Generators](/amarra-cais/docs/reference/generators/). |
| store | SQLite interface + SQL implementation behind handlers. See [Database and migrations](/amarra-cais/docs/how-to/database-and-migrations/). |
| doctor | `amarra-cais doctor` checks the app against the scaffold contract. See [CLI](/amarra-cais/docs/reference/cli/). |
| /jobs | Localhost dashboard for the SQLite job queue. See [Jobs and SQLite](/amarra-cais/docs/explanation/jobs-and-sqlite/). |
| destroy | CLI that undoes a generator and unpatches routes, store, and nav. See [Generators](/amarra-cais/docs/reference/generators/). |
| patch marker | Anchors generators patch (`registerRoutes`, `Close() error`, `<!-- cais:nav -->`). See [Project layout](/amarra-cais/docs/getting-started/project-layout/). |
