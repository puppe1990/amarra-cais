---
title: Pages and views
description: Add a page to a scaffolded Amarra app and render it with view.Write, writeView, and amarraData.
sidebar:
  order: 2
---

Amarra renders HTML on the server, so the browser does not mount a SPA. Handlers call `view.Write`, and Drive morphs the layout's `#amarra-main` element in place. Templates load once at boot, so an unknown `<.component>` tag fails at startup rather than on the first request.

## Add a page

The quickest path is the generator:

```bash
amarra-cais g handler foo
```

This creates the handler, its test, `web/templates/pages/foo.html`, and patches `internal/app/routes.go`. `amarra-cais g page about` does the same for a simple page.

To wire it by hand, follow the same four steps:

1. Go test in `internal/handlers/foo_test.go` — use `setupTestViews(t)` and assert the rendered HTML (`id="amarra-main"`, form fields).
2. HTML page in `web/templates/pages/foo.html`, defining a `{{ define "content" }}` block.
3. Handler that calls `view.Write`, or `writeView` when the page can fail validation.
4. Register the route in `internal/app/routes.go`.

## Render with view.Write

```go
view.Write(w, r, h.views, view.Page{
  Layout: "app",
  Name:   "contact",
  Data: map[string]any{
    "Title":     h.catalog.T("contact.title"),
    "Site":      meta.ForRequest(h.site, r),
    "CSRFToken": csrf.TokenFromRequest(r),
    "Flash":     flashMsg,
  },
}, h.cfg)
```

`Layout` names a file in `web/templates/layouts/` (at least one is required). `Name` is the page's path under `web/templates/pages/` without the `.html` suffix, so `pages/blog/post.html` is `view.Page{Name: "blog/post"}`.

The layout owns the shell — navigation, the `#amarra-main` container, and the single `/static/js/amarra.js` script. A page template only fills the `content` block:

```html
{{ define "content" }}
<h1>{{ .Title }}</h1>
{{ end }}
```

A page can also define `{{ define "frame:<id>" }}` for frame requests (`Amarra-Frame: <id>`), which render just that block.

## Page data

Build page data with the same helpers the scaffold uses. `amarraData` bundles the request-scoped values (site, flash, and anything you pass), and `meta.ForRequest(h.site, r)` produces the Open Graph / Twitter preview.

```go
writeView(w, r, h.views, h.cfg, "app", "contact", amarraData(r, h.site, map[string]any{
  "Title":  h.catalog.T("contact.title"),
  "Errors": errs,
  "Name":   name,
  "Email":  email,
}), http.StatusUnprocessableEntity)
```

`writeView` is the validation variant: it names the layout explicitly (pass a second layout such as `"landing"` if the app has one) and lets you set the status — `422` when re-rendering a form with errors.

:::note
Pass `meta.SiteFrom(appName, cfg.AppURL)` from bootstrap so page data has a `Site` value for OG/Twitter tags.
:::

## Drive vs full HTML

Drive intercepts same-origin clicks and submits by default and morphs `#amarra-main`; it still renders through the layout, so the swap is seamless. First loads, `curl`, and crawlers get the full page. Add `data-amarra-skip` to opt an element out.

## Next

- [Forms and validation](/amarra-cais/docs/how-to/forms-and-validation/) — handle a submitted form with the kit.
- [Views and the kit](/amarra-cais/docs/reference/views-and-kit/) — the shipped components and hooks.
