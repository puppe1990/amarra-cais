---
title: Views and Drive
description: Why Amarra renders full HTML on the server and lets Drive morph the page instead of shipping a SPA.
sidebar:
  order: 2
---

Amarra's frontend is HTML first. Handlers render complete HTML on the server, and a small script in the browser upgrades navigation and form submission into a partial update. No SPA is mounted, and no client-side component framework owns the page.

## The server renders the whole page

Every handler calls `view.Write`. Boot loads templates once with `view.Load`; an unknown `<.component>` fails at boot, not on the first request. That means a first load, a `curl`, and a crawler all receive the same complete document — layout, page, and data already rendered.

The layout defines `#amarra-nav`, `#amarra-main` and `#amarra-toast-host`, and loads a single script, `/static/js/amarra.js`.

## Drive upgrades navigation

Drive intercepts all same-origin clicks and submits by default — a plain `<a>`, a `<form>`, `{{ linkTo }}`, or a `<.form>`. There is no opt-in attribute; you opt out with `data-amarra-skip`. Each interaction becomes a `fetch` carrying `Amarra-Drive: true` and the CSRF header. The response morphs `#amarra-main` and pushes history state.

Drive requests still render the layout so the morph target exists and flash messages survive the swap. A frame request (`Amarra-Frame: <id>`) is narrower: it renders only the `{{ define "frame:<id>" }}` block, a fragment for one region of the page.

## Why not a SPA

There is no Inertia, no Svelte, and no Vite in generated apps. The browser loads one script and never mounts a component tree. The server owns routing, data, and markup, so you debug one language and ship without a bundler step. Crawlers, `curl`, and the first paint see the real page.

Repeating UI is not a framework's job here. It comes from the shipped kit — `<.form>`, `<.input>`, `<.table>`, `<.filters>`, `<.stat>`, `<.empty>`, `<.password>` — plus `amarra-hook` builtins such as `dialog`, `dropdown`, `bulk`, `nav`, `theme`, and `password`. An app restyles a real contract by dropping a same-named file into `web/templates/components/` instead of recreating a widget. Sort and filter are plain `GET ?q=&sort=` round-trips the server re-renders.

:::note
The sidebar shell lives outside `#amarra-main`, so a Drive morph never disturbs it. `amarra-hook="nav"` with `data-amarra-nav-on` / `data-amarra-nav-off` re-syncs the active link after each morph.
:::

## The trade-offs

HTML-first is not free, and Amarra makes the costs explicit:

- **List caching** — a stable list page can cache its rendered markup with `cache.Key` and `cache.Hash(version)`, then serve a `304` via `httpx.NotModified` / `httpx.SetETag`. That is the intended answer to per-request rendering, not a client cache.
- **Morph stability** — Drive morphs by matching elements, so ids and structure need to stay stable across responses.
- **No offline mutation queue** — an offline Drive POST is not queued. The UI shows a toast instead of silently retrying.
- **Server-side state** — Drive and Frame hold no client state; every interaction is an HTTP round-trip against SQLite.

For the API surface behind this, see the [views and kit reference](/amarra-cais/docs/reference/views-and-kit/) and the [amarra.js reference](/amarra-cais/docs/reference/amarra-js/).
