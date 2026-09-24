---
title: Template helpers
description: The template functions Amarra registers on the view renderer for CSRF, forms, flash, links, and i18n.
sidebar:
  order: 7
---

Amarra registers a set of helper functions on the view renderer so page templates stay declarative. They cover CSRF fields, form fields and their errors, flash messages, links, and localization.

## Helper reference

| Helper            | Example call                                                                              | Renders                                                           |
| ----------------- | ----------------------------------------------------------------------------------------- | ----------------------------------------------------------------- |
| `csrfField`       | `{{ csrfField .CSRFToken }}`                                                              | A hidden CSRF input (double-submit token).                        |
| `fieldError`      | `{{ fieldError .Errors "name" }}`                                                         | The error string for one field from a `validate.FieldErrors` map. |
| `makeField`       | `makeField "name" "Name" .Name "text" true .Errors`                                       | A `forms.FieldData` value passed to `fieldInput`.                 |
| `fieldInput`      | `{{ fieldInput (makeField ...) }}`                                                        | An input, textarea, or checkbox plus its error.                   |
| `fieldPassword`   | `{{ fieldPassword ... }}`                                                                 | A password input with the eye show/hide toggle.                   |
| `makeSelectField` | `makeSelectField "category_id" "Category" .Item.CategoryID .CategoryOptions true .Errors` | A select `forms.FieldData` value.                                 |
| `fieldSelect`     | `{{ fieldSelect (makeSelectField ...) }}`                                                 | A foreign-key `<select>` plus its error.                          |
| `flashMessage`    | `{{ flashMessage .Flash }}`                                                               | The one-shot flash notice.                                        |
| `linkTo`          | `{{ linkTo "/items/1" "Delete" (dict "method" "delete" "confirm" "Delete this item?") }}` | A plain `<a>` that Drive intercepts.                              |
| `t`               | `{{ t "key" }}`                                                                           | A localized string from the catalog.                              |
| `dict`            | `(dict "method" "delete" "confirm" "Sure?" "frame" "cart")`                               | A map, used to pass options.                                      |

## CSRF

Forms pair the double-submit `cais_csrf` cookie with a token field or the `X-CSRF-Token` header. In templates you emit the hidden field yourself:

```html
<form action="/contact" method="post">{{ csrfField .CSRFToken }} ...</form>
```

Kit `<.form>` injects the CSRF field from the root data (`$.CSRFToken`), so do not add `csrfField` inside its slot. The layout renders `<meta name="csrf-token">`, and [amarra.js](/amarra-cais/docs/reference/amarra-js/) sends `X-CSRF-Token` on Drive requests.

## Form fields and validation

Collect errors in a `validate.FieldErrors` map, then pass it to the page as `.Errors`:

```go
var errs validate.FieldErrors
if item.Name == "" {
  errs.Add("name", "Name is required")
}
if errs.Any() {
  // re-render the form with errs as .Errors
}
```

Single-field checks use `validate.Email`, `validate.URL`, `validate.Required`, `validate.MinLength`, and `validate.MaxLength`.

Render a field and its error with the grouped helpers. `makeField` builds a `forms.FieldData` value and `fieldInput` renders input, textarea, or checkbox plus the error:

```html
{{ fieldInput (makeField "name" "Name" .Name "text" true .Errors) }}
```

Always use `fieldPassword` for password fields — the eye show/hide toggle is the default.

Foreign-key selects use `makeSelectField` (options come from a store `List*Options` method) and `fieldSelect`:

```html
{{ fieldSelect (makeSelectField "category_id" "Category" .Item.CategoryID .CategoryOptions true
.Errors) }}
```

The [resource generator](/amarra-cais/docs/reference/generators/) wires these helpers plus kit `<.form>` and `<.button>` into admin forms. See [Forms and validation](/amarra-cais/docs/how-to/forms-and-validation/).

## Flash messages

`flash.Set(w, "notice", "Saved!", cfg.CookieSecure())` writes a one-shot cookie; read it on the next request with `flash.MessageFromRequest(r)` and pass it into page data. In the template, always render it with `flashMessage`:

```html
{{ flashMessage .Flash }}
```

:::caution
Never print `{{ .Flash }}` — it stringifies the struct. Keep `<.flash />` inside `#amarra-main` so a Drive morph keeps the notice.
:::

## Links

`linkTo` emits a plain `<a>`; Drive intercepts it by default. Pass a `dict` to add options:

```html
{{ linkTo "/items/1" "Delete" (dict "method" "delete" "confirm" "Delete this item?") }}
```

Options: `"method"` (for example `delete`), `"confirm"` (confirmation text), and `"frame"` (a frame id such as `cart`).

## i18n

`{{ t "key" }}` looks up a string in the locale catalog. The layout also uses `{{ htmlLang }}` and `{{ ogLocale }}`. Missing keys return the key itself, which is visible in development.

The catalog is selected by the `LOCALE` env var (`en` is the default, `pt` is supported) and wired as `cais.Config.Locale`. Handlers receive the catalog for validation and flash strings (`catalog.T("contact.title")`), and the same functions reach pages and partials. See [i18n](/amarra-cais/docs/how-to/i18n/) for the locale files.
