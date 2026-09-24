---
title: amarra.js
description: The public HTML contract for Drive, frames, and the shipped amarra-hook builtins.
sidebar:
  order: 6
---

`amarra.js` is the one script a generated app ships (`/static/js/amarra.js`). It powers **Drive** (intercepting navigations and morphing the page) and the built-in `amarra-hook` behaviors. HTMX is not a public dependency.

## How Drive works

Drive intercepts **all same-origin clicks and submits by default** — there is no opt-in attribute. Plain `<a>`, `<form>`, `{{ linkTo }}`, and kit `<.form>` are all captured, turned into a `fetch` with an `Amarra-Drive: true` header plus CSRF, and the response morphs `#amarra-main`.

The first load, `curl`, and crawlers get the full layout. Opt an element out of Drive with `data-amarra-skip`.

The layout renders `<meta name="csrf-token">`, and Drive sends `X-CSRF-Token` on its requests (paired with the double-submit `cais_csrf` cookie).

## Attributes

| Attribute                       | Applies to               | Effect                                                               |
| ------------------------------- | ------------------------ | -------------------------------------------------------------------- |
| `data-amarra-skip`              | `<a>`, `<form>`          | Opt that element out of Drive; it navigates or submits normally.     |
| `data-amarra-confirm`           | click / submit source    | Ask for confirmation (text is the value) before running the request. |
| `data-amarra-method`            | link                     | Use this HTTP method for the link (for example `delete`).            |
| `data-amarra-disable-with`      | submit / button          | Label the control switches to while its request is in flight.        |
| `data-amarra-frame`             | link / form              | Target a named frame instead of the default `#amarra-main` morph.    |
| `amarra-click`                  | element                  | Bind a click action.                                                 |
| `amarra-change`                 | element                  | Bind a change action.                                                |
| `amarra-submit`                 | element                  | Bind a submit action.                                                |
| `amarra-debounce`               | element                  | Debounce the bound action.                                           |
| `amarra-hook`                   | element                  | Attach a built-in hook by name (see below).                          |
| `amarra-live`                   | element                  | Opt the element into the Live WebSocket hub.                         |
| `<amarra-frame loading="lazy">` | `<amarra-frame>` element | Frame element; `loading="lazy"` defers it until needed.              |

## Built-in hooks (`amarra-hook`)

Set `amarra-hook="<name>"` on a container to turn on a built-in behavior.

| Hook        | Markup it expects                                                                                                                                                           | Behavior                                                                                                                                                                                                                                                                                    |
| ----------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `bulk`      | container `amarra-hook="bulk"`, `input[data-amarra-bulk-all]`, row checkboxes `input[data-amarra-bulk-row]`, optional `[data-amarra-bulk-bar]` + `[data-amarra-bulk-count]` | Select-all on the current page. The header shows `indeterminate` when some rows are selected, and the optional bar appears when the count is greater than zero. The action itself stays a Drive POST/DELETE of the selected ids.                                                            |
| `clipboard` | `data-amarra-copy`                                                                                                                                                          | Copy the given text to the clipboard.                                                                                                                                                                                                                                                       |
| `dialog`    | container `amarra-hook="dialog"`, `button[data-amarra-dialog-open]`, `dialog[data-amarra-dialog-target]`, optional `button[data-amarra-dialog-close]`                       | Native `<dialog>` via `showModal()`, so Esc, the focus trap, and `::backdrop` come for free. `aria-modal="true"` is set on connect and focus returns to the opener on close. `<.modal>` renders the dialog target.                                                                          |
| `dropdown`  | container `amarra-hook="dropdown"`, `button[data-amarra-dropdown-button aria-expanded]`, `div[data-amarra-dropdown-menu hidden]`                                            | Row actions / overflow menus. Click toggles; Esc and outside clicks close; a menu-item click closes after acting.                                                                                                                                                                           |
| `nav`       | nav container `amarra-hook="nav"` with `data-amarra-nav-on` / `data-amarra-nav-off` class lists                                                                             | Re-sync the active link after a Drive morph. Matches `a.pathname === location.pathname` (query ignored), sets `aria-current="page"`, and runs on connect, `amarra:morphed`, and `popstate`.                                                                                                 |
| `password`  | control plus optional `data-amarra-password-for`                                                                                                                            | Toggle `type` and `aria-pressed`. Falls back to the closest sibling `input` when `data-amarra-password-for` is missing; swaps `aria-label` via `data-amarra-label-show` / `data-amarra-label-hide`. Icons accept `data-amarra-password-icon` (`data-cais-password-icon` is a legacy alias). |
| `reveal`    | control with `data-amarra-reveal-show` and `data-amarra-reveal-target`                                                                                                      | Show or hide the target when the control value matches — no Drive round-trip.                                                                                                                                                                                                               |
| `theme`     | control with `amarra-hook="theme"`                                                                                                                                          | Toggle `html.light`, persist `localStorage["amarra-theme"]`, and update the optional `theme-color` meta.                                                                                                                                                                                    |

### Hook examples

```html
<div amarra-hook="clipboard" data-amarra-copy="hi">Copy</div>

<button
  type="button"
  amarra-hook="password"
  data-amarra-password-for="#password"
  data-amarra-label-show="Show password"
  data-amarra-label-hide="Hide password"
>
  Show
</button>

<select
  amarra-hook="reveal"
  data-amarra-reveal-show="access_keys"
  data-amarra-reveal-target="#aws-keys"
>
  <option value="default_chain">Default chain</option>
  <option value="access_keys">Access keys</option>
</select>
<div id="aws-keys" hidden>…keys…</div>

<button
  type="button"
  amarra-hook="theme"
  data-amarra-theme-key="app-theme"
  data-amarra-theme-color="#f5f5f4"
  data-amarra-theme-color-off="#0f172a"
  data-amarra-theme-on-label="Dark mode"
  data-amarra-theme-off-label="Light mode"
>
  Light mode
</button>
```

The `theme` hook is configurable per element:

| Attribute                                                    | Default        | Purpose                                      |
| ------------------------------------------------------------ | -------------- | -------------------------------------------- |
| `data-amarra-theme-key`                                      | `amarra-theme` | `localStorage` key.                          |
| `data-amarra-theme-class`                                    | `light`        | Class toggled on `html`.                     |
| `data-amarra-theme-color` / `data-amarra-theme-color-off`    | —              | Hex for the `theme-color` meta.              |
| `data-amarra-theme-on-label` / `data-amarra-theme-off-label` | —              | Swap the button text (keeps `aria-pressed`). |

## Theme FOUC snippet

Put this in the layout `<head>` before CSS so a stored light theme does not flash dark. If you set a custom `data-amarra-theme-key`, copy that key into the snippet so restore matches.

```html
<script>
  try {
    if (localStorage.getItem("amarra-theme") === "light")
      document.documentElement.classList.add("light");
  } catch (e) {}
</script>
```

:::note
The Content Security Policy keeps `script-src 'self' 'unsafe-inline'` so this FOUC snippet and Drive's inline init run; escaping is the primary defense. See the [security model](/amarra-cais/docs/explanation/security-model/).
:::

:::tip
`amarra-live` is opt-in and only needed for real-time pages; CRUD stays on Drive. See [Forms and validation](/amarra-cais/docs/how-to/forms-and-validation/) for the form side of the contract.
:::
