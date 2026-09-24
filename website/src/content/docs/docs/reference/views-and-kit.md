---
title: Views and the shipped kit
description: The Amarra template loader contract and the built-in kit components you can use and override.
sidebar:
  order: 5
---

Amarra renders HTML on the server. Handlers call `view.Write`, which resolves a layout plus a page body and expands any `<.component>` kit tags. Boot loads every template once with `view.Load`; an unknown `<.component>` tag fails at boot, not on the first request.

## Template loader contract

`view.Load` globs `web/templates/` once at boot (`pkg/amarra/view/renderer.go`). The contract is pinned by `pkg/amarra/view/renderer_contract_test.go`.

| Glob                             | Addressable as                 | Notes                                                                                                      |
| -------------------------------- | ------------------------------ | ---------------------------------------------------------------------------------------------------------- |
| `layouts/*.html`                 | `view.Page{Layout: "app"}`     | Flat. At least one layout is required.                                                                     |
| `pages/*.html`, `pages/*/*.html` | `view.Page{Name: "blog/post"}` | Name is the path under `pages/` minus `.html`. One nesting level, so a blog can keep one file per article. |
| `partials/*.html`                | `{{ template "card" . }}`      | Flat only. `partials/posts/card.html` never loads, and the miss surfaces at render time, not at boot.      |
| `components/*.html`              | `<.input>` override            | Flat only. Shipped kit plus app overrides keyed by file stem. An unknown `<.x>` fails at boot.             |

:::caution
`partials/` is flat only. A nested partial such as `partials/posts/card.html` is silently skipped by `view.Load`, and the `{{ template }}` call fails only when the page that needs it renders.
:::

## Rendering a page

`view.Page` carries the layout name, the page name, and the data map:

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

Drive requests still render the full layout so `amarra.js` can morph `#amarra-main`. Frame requests (`Amarra-Frame: <id>`) render only the matching `{{ define "frame:<id>" }}` block.

Layouts provide the shell markers `#amarra-nav`, `#amarra-main`, and `#amarra-toast-host`, and load a single script at `/static/js/amarra.js`.

## Page bodies and frames

Every page defines a `content` block and may define named frame blocks. The layout renders the page's `content` inside `#amarra-main`.

```html
{{ define "content" }}
<h1>{{ .Title }}</h1>
<.form action="/items" method="post">
  <.input name="title" label="Title" value="{{ .Item.Title }}" error="{{ fieldError .Errors "title" }}" />
  <.button type="submit">Save</.button>
</.form>
{{ end }}

{{ define "frame:cart" }}
<div id="cart">{{ .CartTotal }}</div>
{{ end }}
```

A frame request returns the `frame:<id>` block on its own, so Drive can swap a smaller region instead of the whole `#amarra-main`.

## Shipped kit

The kit lives in `web/templates/components/`. Override a component by writing a file with the same stem (for example `locale-toggle.html`); `amarra-cais g component <kit-name>` seeds the shipped markup and `amarra-cais g component --list` prints the overridable names. See [Generators](/amarra-cais/docs/reference/generators/).

| Component       | Purpose                                                   | Documented attributes                                                                                                                             |
| --------------- | --------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| `form`          | Form shell                                                | `action`, `method`, `enctype`. Injects `csrf_token` from `.CSRFToken`.                                                                            |
| `input`         | Text/email/file input                                     | `name`, `type`, `label`, `value`, `error`; `accept` on file inputs.                                                                               |
| `password`      | Password input with eye toggle (`amarra-hook="password"`) | `name`, `label`, `value`, `error`, `autocomplete`, `required`.                                                                                    |
| `select`        | Select field                                              | Body is the markup contract.                                                                                                                      |
| `textarea`      | Multi-line input                                          | —                                                                                                                                                 |
| `checkbox`      | Checkbox                                                  | —                                                                                                                                                 |
| `button`        | Button                                                    | `type` (for example `type="submit"`).                                                                                                             |
| `flash`         | One-shot flash notice                                     | Self-closing `<.flash />` is allowed.                                                                                                             |
| `nav`           | Navigation container                                      | —                                                                                                                                                 |
| `pagination`    | List pagination                                           | `base`.                                                                                                                                           |
| `modal`         | Dialog target for the `dialog` hook                       | —                                                                                                                                                 |
| `locale-toggle` | Locale / language switcher                                | —                                                                                                                                                 |
| `stat`          | KPI card                                                  | `label`, `value`; optional `delta`, `hint`, `href`.                                                                                               |
| `empty`         | Empty state                                               | `title`, body slot; optional `href` or `action` link.                                                                                             |
| `table`         | Data table                                                | Rows go in the `.Inner` slot; `cols` from page data (`Field`/`Label`/`Sortable` maps); `sort`/`dir` for server-side sort links; optional `frame`. |
| `filters`       | GET filter form                                           | `action`; optional `frame`, `submit`, `clear` href. `.Inner` fields become query params; hidden inputs preserve `sort`/`page`.                    |

Every component has exactly one slot: `.Inner`. `<.form>` injects the CSRF field for you, so do not add another inside the slot. For uploads, use `<.input type="file" accept="image/*" />` inside `<.form enctype="multipart/form-data">` and never set `value` on a file input.

## Kit attribute interpolation

Kit attributes interpolate like plain markup, so a single value can mix static text with page data:

```html
<.stat label="Potência" value="{{ .Power }} kWp" />
```

renders `5 kWp`. A bare `{{ .X }}` keeps the raw expression, so non-string data still works inside the component's own `if`/`range` blocks.

Do not place a control action in an attribute value: a `{{ if }}` or `{{ range }}` inside an attribute fails at boot instead of printing `{{ … }}` to the user.

:::note
Both boot-failure rules are deliberate. An unknown `<.x>` tag and a control action inside an attribute value stop the process at startup, so the mistake never reaches a rendered page.
:::

:::tip
Drive returns the whole layout while frames return one block — see [amarra.js](/amarra-cais/docs/reference/amarra-js/) for the interrupt rules and [template helpers](/amarra-cais/docs/reference/template-helpers/) for `csrfField`, `linkTo`, and the form helpers. The rationale behind the HTML-first model is in [Views and Drive](/amarra-cais/docs/explanation/views-and-drive/).
:::
