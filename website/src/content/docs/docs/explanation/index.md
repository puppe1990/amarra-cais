---
title: Explanation
description: Why Amarra is HTML-first, how Cais and Amarra split, and the trade-offs behind its core pieces.
sidebar:
  order: 1
---

The how-to guides show you _how_ to do something. This section explains _why_ Amarra works the way it does — the decisions behind the framework's shape and the trade-offs that come with them.

Read these when you are deciding whether Amarra fits a project, or when a behavior surprises you.

## Pages

- [Views and Drive](/amarra-cais/docs/explanation/views-and-drive/) — why the server renders full HTML and how Drive morphs the page instead of a SPA.
- [Cais vs Amarra](/amarra-cais/docs/explanation/cais-vs-amarra/) — the Cais v0.11.x (Inertia + Svelte) and Amarra (HTML-first) split, and what migrating means.
- [Security model](/amarra-cais/docs/explanation/security-model/) — how CSRF, sessions, headers, rate limits and production gates fit together.
- [Jobs and SQLite](/amarra-cais/docs/explanation/jobs-and-sqlite/) — why one SQLite file handles requests, streams and the background queue.

:::tip
Looking for steps instead of rationale? Start with the [how-to guides](/amarra-cais/docs/how-to/) or the [reference](/amarra-cais/docs/reference/).
:::
