---
title: Forms and validation
description: Use the shipped kit form tags, CSRF, field errors, and file uploads in an Amarra app.
sidebar:
  order: 3
---

Amarra ships a small kit for forms, so you write markup instead of assembling inputs by hand. Kit tags expand at boot, and `<.form>` injects the CSRF field for you.

## The kit form tags

```html
{{ define "content" }}
<.form action="/items" method="post">
  <.input name="title" label="Title" value="{{ .Item.Title }}" error="{{ fieldError .Errors "title" }}" />
  <.button type="submit">Save</.button>
</.form>
{{ end }}
```

- `<.form>` renders the `<form>` element and injects the `csrf_token` field from the root data (`$.CSRFToken`). Do not add a second CSRF field inside the slot.
- `<.input>` takes `name`, `label`, `value`, and `error`. Render the error with `{{ fieldError .Errors "title" }}`.
- `<.button>` renders a submit or action button.

The kit also ships `<.select>`, `<.textarea>`, and `<.checkbox>` for those control types; give each the field's `name` and `label`, and pass `error` the same way you do for `<.input>`. Every component accepts app overrides in `web/templates/components/`, keyed by file stem, so you can restyle the real contract instead of recreating it.

## CSRF

CSRF is a double-submit cookie (`cais_csrf`): `middleware.CSRF(cfg)` validates every `POST`, `PUT`, `DELETE`, and `PATCH`. Kit `<.form>` injects the field automatically. In hand-written forms use `{{ csrfField .CSRFToken }}` or a hidden `csrf_token` input. Drive requests send the token as `X-CSRF-Token`, fed by the layout's `<meta name="csrf-token">`.

## Server-side validation

Validate in the handler and re-render the same page with `422` when something is wrong. For a single field, use the helpers in `pkg/cais/validate` — `validate.Email`, `validate.URL`, `validate.Required`, `validate.MinLength`, `validate.MaxLength`. For a whole form, collect errors in `validate.FieldErrors`:

```go
var errs validate.FieldErrors
if item.Name == "" {
  errs.Add("name", "Name is required")
}
if errs.Any() {
  writeView(w, r, h.views, h.cfg, "app", "item", amarraData(r, h.site, map[string]any{
    "Errors": errs,
    "Item":   item,
  }), http.StatusUnprocessableEntity)
  return
}
```

Pass the errors as `.Errors` in page data; inputs read them with `{{ fieldError .Errors "name" }}`.

## Template helpers

`pkg/cais/forms` registers helpers on the view renderer for building fields from Go values:

```html
{{ fieldInput (makeField "name" "Name" .Name "text" true .Errors) }}
```

`makeField` returns a `forms.FieldData`, and `fieldInput` renders the right control (input, textarea, or checkbox) plus its error.

:::tip
Password fields always use `fieldPassword` so they get the eye show/hide toggle. Foreign-key selects use `fieldSelect` with `makeSelectField`:

```html
{{ fieldSelect (makeSelectField "category_id" "Category" .Item.CategoryID .CategoryOptions true
.Errors) }}
```

:::

## File uploads

Set `enctype="multipart/form-data"` on the form and use `<.input type="file" />`. Do not set `value` on a file input.

```html
<.form action="/upload" method="post" enctype="multipart/form-data">
  <.input type="file" name="avatar" accept="image/*" label="Avatar" />
  <.button type="submit">Upload</.button>
</.form>
```

## Next

- [Pages and views](/amarra-cais/docs/how-to/pages-and-views/) — how the handler renders and re-renders the page.
- [Template helpers](/amarra-cais/docs/reference/template-helpers/) — the full helper list.
