# amarra-cais — HTML-first fork of Cais

**Date:** 2026-09-05
**Status:** Draft — awaiting review
**Source:** Cais v0.11.x (`github.com/puppe1990/cais`), Inertia + Svelte line frozen
**Decision:** Approach B — full framework fork (not an adapter inside Cais)

---

## Goal

Ship a Cais-family framework whose default UI is server-rendered HTML with a Hotwire + LiveView-shaped client, not Inertia + Svelte.

- **Drive / Frame / Stream / Hook** cover CRUD and navigation (Turbo / Stimulus / current Cais HTMX path).
- **Live** covers opt-in realtime (Phoenix LiveView) in slice B, after the HTML scaffold works.
- Views are a real library (Rails ERB + Phoenix HEEx), not leftover `html/template` helpers.

Cais v0.11.x stays the Inertia product. Apps generated there do not migrate automatically.

---

## Identity

| Layer            | Name             | Role                                                      |
| ---------------- | ---------------- | --------------------------------------------------------- |
| Forked framework | **amarra-cais**  | CLI, router, session, CSRF, jobs, SQLite, PWA, generators |
| Front library    | **Amarra**       | Views, Drive, Frame, Stream, Live, Hook                   |
| Frozen line      | **Cais v0.11.x** | Inertia + Svelte                                          |

| Surface                | Value                                         |
| ---------------------- | --------------------------------------------- |
| GitHub / Go module     | `github.com/puppe1990/amarra-cais`            |
| CLI binary             | `amarra-cais` (does not overwrite `cais`)     |
| JS bundle              | `/static/js/amarra.js`                        |
| Core packages (copied) | `github.com/puppe1990/amarra-cais/pkg/cais`   |
| Front packages         | `github.com/puppe1990/amarra-cais/pkg/amarra` |

Do not rename every internal `pkg/cais/*` type in slice A. Docs and CLI copy say **amarra-cais** for the framework and **Amarra** for templates, `amarra-click`, and JS.

Example:

```bash
go install github.com/puppe1990/amarra-cais/cmd/amarra-cais@latest
amarra-cais new myapp
cd myapp && amarra-cais install && amarra-cais dev
```

---

## Architecture

One JS runtime and one Go view stack. The server owns HTML. The browser does not mount Svelte.

```
Browser                         amarra-cais
───────                         ──────────
amarra.js                       pkg/amarra
  Drive  ──HTTP GET/POST──►       drive
  Frame  ──HTTP fragment──►       frame
  Stream ◄─SSE / HTML ops──       stream
  Live   ◄─WebSocket──────►       live     (stub in slice A, real in slice B)
  Hook   (client only)            view     (ERB/HEEx layer)
                                pkg/cais   (router, session, jobs, …)
```

Public HTML contract: `data-amarra-*`, `amarra-click` / `amarra-change` / `amarra-submit` / `amarra-hook` / `amarra-live`, and `<amarra-frame>`. HTMX is not a public dependency. Layout loads a single script: `amarra.js` (Idiomorph bundled). `htmx.min.js`, `sse-ext`, `idiomorph-ext`, and `cais.js` are gone from generated apps.

Handlers do not check `HX-Request`. They call `view.Write`.

| Generator                                              | Runtime                                                        |
| ------------------------------------------------------ | -------------------------------------------------------------- |
| `amarra-cais new`, `g handler`, `g resource`, `g auth` | Drive + Frame + Hook                                           |
| `amarra-cais g stream chat`                            | Stream (SSE) in slice A; Live on the same generator in slice B |
| Live dashboard / counters                              | `amarra.Live` registered explicitly                            |

CRUD does not open a WebSocket.

### Drive

Replaces Turbo Drive and today's `hx-boost`. Clicks and submits become `fetch` with `Amarra-Drive: true`, CSRF header, morph into `#amarra-main`, `history.pushState`. First load, curl, and crawlers get full layout HTML.

### Frame

Replaces Turbo Frames. `<amarra-frame id="cart">` fetches a fragment (`Amarra-Frame: cart`). URL changes only when the frame sets `amarra-push`.

### Stream

Replaces Turbo Streams. Named ops: `append`, `prepend`, `replace`, `morph`, `remove`, `toast`. Transport: SSE or HTTP body. Chat and `/jobs` use this. This is `pkg/cais/stream.WriteEvent` with a named contract, moved under `pkg/amarra/stream`.

### Live (slice B)

Opt-in WebSocket at `/amarra/live`. One goroutine per connection, state struct on the server. Event → `Handle` → re-render template → morph HTML. No Phoenix static/dynamic diffs in slice B: send HTML, Idiomorph applies it. Pub/sub is in-process (one Lightsail binary, no Redis). `MaxConns` + idle timeout. Two processes on the same box do not share Live state.

### Hook

Stimulus-shaped. Small DOM JS (`amarra-hook="clipboard"`). Toast, focus, optimistic UI move out of `cais-core.js` into official hooks.

Engine stays `html/template`. Do not adopt `templ` in this fork.

---

## Amarra Views

`pkg/amarra/view` is the only render path. It is the ERB + HEEx layer.

Do **not** invent `{@count}`, `<%= %>`, or a new language. Go's ERB is `{{ .Title }}`. The missing piece is HEEx: components, slots, form builder, Live bindings as HTML attributes.

### File layout (generated app)

```
web/templates/
  layouts/app.html
  pages/items/index.html
  pages/items/show.html
  pages/items/_form.html
  components/button.html
  components/form.html
  components/input.html
  components/flash.html
```

### Page (ERB)

```html
{{ define "content" }}
<h1>{{ .Item.Title }}</h1>
<.form action="/items" method="post">
  <.input name="title" label="Title" value="{{ .Item.Title }}" error="{{ fieldError .Errors "title" }}" />
  <.button type="submit">Save</.button>
</.form>
{{ end }}
```

### Component (HEEx)

```html
{{/* components/button.html — attrs: type, class, click */}}
<button
  type="{{ .Type }}"
  class="{{ .Class }}"
  {{
  if
  .Click
  }}amarra-click="{{ .Click }}"
  {{
  end
  }}
>
  {{ .Inner }}
</button>
```

At boot, a small preprocessor expands `<.button type="submit">Save</.button>` into `html/template` calls plus a single `.Inner` slot. Self-closing `<.flash />` is allowed. Nested `<.name>` tags expand inside-out. `{{ }}` inside component attributes is left untouched for `html/template`. Unknown component names fail **boot**, not the first request.

### Live bindings

```html
<button amarra-click="inc">{{ .Count }}</button>
<form amarra-change="validate" amarra-submit="save">…</form>
<div amarra-hook="clipboard">…</div>
<amarra-frame id="cart" src="/cart">…</amarra-frame>
```

These are HTML attributes. `amarra.js` listens. Live Go dispatches the event name. The view layer never talks to the socket.

### Helpers

| Rails / Phoenix         | Amarra                               |
| ----------------------- | ------------------------------------ |
| `form_with`             | `<.form>` (injects CSRF)             |
| `f.text_field`          | `<.input>`                           |
| `link_to`               | `{{ linkTo "/x" "Label" }}` (Drive)  |
| `render "form"`         | `{{ render "items/_form" . }}`       |
| `render @items`         | `{{ renderEach "item_row" .Items }}` |
| `yield` / `content_for` | layout `{{ template "content" . }}`  |
| `<.flash>`              | `<.flash />`                         |
| `phx-click`             | `amarra-click`                       |

### Shipped kit

Framework-owned components: `form`, `input`, `button`, `flash`, `nav`, `pagination`, `modal`. An app overrides by placing the same filename in `web/templates/components/`.

`amarra-cais g component card` writes `web/templates/components/card.html` plus a render test. Resource generators use the kit, not one-off Tailwind copies.

### Go API

```go
view.Write(w, r, view.Page{
  Layout: "app",
  Page:   "items/index",
  Frame:  "admin-items", // used when Amarra-Frame is set
  Data:   data,
})
```

Drive requests render `content` without the full layout. Frame requests render the named frame/partial. Full page requests render layout + page.

Slice A views do **not** include: `{@x}` syntax, HEEx compile-time HTML validation, scoped CSS, templ/Go components, named slots (`<:header>`). One slot: `.Inner`.

---

## Packages, scaffold, CLI

```
pkg/cais/                    # copied: router, session, csrf, jobs, migrate, i18n, mail, pwa, forms (kit uses forms internally)
pkg/amarra/
  view/                      # Load, preprocessor, Write, helpers, kit
  drive.go
  frame.go
  stream.go                  # moved from pkg/cais/stream
  live.go                    # slice A: types + 501 upgrade; slice B: hub
  attrs.go
  js/                        # testable sources (from pkg/cais/js)
pkg/cais/pwa/assets/amarra.js
cmd/amarra-cais/
internal/cli/                # HTML templates, no Svelte
```

`pkg/cais/htmx.go` and `htmxattrs` are not public API. Useful bits move into `pkg/amarra` or die.

### Generated app (`amarra-cais new`)

Removed: `web/src/**/*.svelte`, `vite.config.js`, `svelte.config.js`, `vitest-setup.js`, `web/static/build/`, gonertia, `inertia_test.go`.

Present:

```
web/templates/layouts/app.html
web/templates/pages/*.html
web/templates/partials/
web/templates/components/
web/static/js/amarra.js
web/static/css/styles.css
input.css, tailwind.config.js
```

`--minimal` and `--blank` remain, still HTML-only.

### CLI map

| Command           | Slice A behavior                                                          |
| ----------------- | ------------------------------------------------------------------------- |
| `amarra-cais new` | Amarra scaffold (full / `--minimal` / `--blank`)                          |
| `g handler`       | Go handler + `pages/X.html` + HTML test                                   |
| `g resource`      | Drive/Frame CRUD, `_form` partial, kit components                         |
| `g auth`          | HTML auth pages + Amarra forms                                            |
| `g component`     | `components/X.html` + render test                                         |
| `g stream chat`   | Stream SSE (Live in slice B on the same generator)                        |
| `dev`             | air + Tailwind watch. No Vite                                             |
| `install`         | `go mod tidy` + npm only for Tailwind/postcss                             |
| `build`           | `go build` + CSS. No SPA `npm run build`                                  |
| `doctor`          | `amarra.js`, templates, PWA; fail if Vite/Inertia present                 |
| `pwa`             | `InstallForAmarra`: network-first HTML + `amarra.js`; no `/static/build/` |

`link`, `jobs`, `db`, `console`, `routes`, `destroy` stay. Destroy deletes `.html` and routes, not `.svelte`.

Handlers take `*view.Renderer` (or the amarra-cais renderer type) and `cais.Config`. They do not take `*inertia.Inertia`.

PWA service worker is network-first for HTML Drive and `/static/js/amarra.js`, fallback `offline.html`. `amarra-cais pwa --bump` remains.

No dual mode, no `--front inertia`.

---

## Data flow

**Boot:** `view.Load(web/templates)` → preprocess `<.component>` → parse `html/template` once → serve `amarra.js` as a static asset.

**Drive:** click/submit → fetch (`Amarra-Drive`, CSRF, optional `Amarra-Frame`) → handler (session, validate, store) → `view.Write` → morph `#amarra-main` → `pushState`.

Invalid form: **no redirect**. Same page/frame, status **422**, `.Errors` on `<.input>`, focus first invalid field. Success: **303** to show/index; Drive follows and morphs. Flash cookie (existing Cais) rendered by `<.flash />`.

**Frame:** `Amarra-Frame: cart` → fragment only.

**Stream:** `stream.Send(w, stream.Append("chat-history", html))` → JS applies the op. Chat in slice A uses SSE, not Live.

**Live (slice B):** page with `amarra-live="/chat/42"` → WS `/amarra/live?topic=chat:42` → `Mount` → events `{event, payload, ref}` → `Handle` → render → morph. Disconnect → idle timeout → drop from memory. Reconnect remounts; state is not sticky across deploys. `hub.Broadcast("chat:42", …)` is in-process only. Jobs stay in the worker process; Live does not go through the job queue.

**CSRF:** Drive/Frame/Stream = double-submit cookie + `X-CSRF-Token` (current Cais rule). Live = token on WS first message, checked against session. Missing token = 403 / socket close.

**PWA:** network-first, no Vite bundle. Offline Drive does not queue POSTs in slice A; toast on failure.

**Where state lives**

| Channel       | State                                                         |
| ------------- | ------------------------------------------------------------- |
| Drive / Frame | HTTP + SQLite. Server forgets after the response              |
| Stream        | Same, plus a read SSE connection                              |
| Live          | Struct on the connection goroutine. Full page reload remounts |

Resource handlers never see Live. They only `view.Write` + store.

---

## Errors

| Case                                     | Behavior                                                                                                                                   |
| ---------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------ |
| Field validation                         | 422, same view, per-field errors, focus first invalid                                                                                      |
| Business rule (duplicate, FK, forbidden) | Flash `alert` + 303, or 403/404 error page. Product copy, not SQLite text                                                                  |
| Template/preprocessor failure            | Fail at **boot** when parse can see it. Request-time execute: 500 with template error in dev; generic 500 + log in prod (`SanitizeErrors`) |
| Drive 401 / login 303                    | JS follows redirect                                                                                                                        |
| Drive 5xx                                | Rollback optimistic UI + toast                                                                                                             |
| Offline                                  | Toast, no silent retry                                                                                                                     |
| Bad CSRF                                 | 403 + reload                                                                                                                               |
| SSE drop                                 | Backoff reconnect; malformed event logged and skipped; after N failures, chat UI warning (no infinite spinner)                             |
| Live WS drop                             | Reconnect + `Mount`. UI may flash. Hub full → 503 on upgrade; page still works over HTTP                                                   |
| Unknown Live event                       | Log + `{error}` on the socket; connection stays up                                                                                         |
| Panic in Live `Handle`                   | Recover goroutine, close WS, log stack                                                                                                     |
| Hook throw                               | `console.error` + toast; Drive keeps running                                                                                               |

This design does **not** include: offline POST queue, automatic POST retry (would double-create), Phoenix diffs to hide Live reconnect flicker, Inertia-style `Error.svelte`.

---

## Testing

TDD as in Cais `AGENTS.md`. No Svelte/Vitest. Headless HTML and protocol tests.

**Library (`pkg/amarra`)**

- Preprocessor: golden HTML in → expanded template → rendered string
- `view.Write`: `httptest` chooses full / Drive / Frame from headers
- Kit: `<.form>` / `<.input>` emit CSRF, `for=`, 422 error text
- Stream ops: move existing `pkg/cais/stream` tests
- Live (slice B): WS `httptest` for Mount → click → HTML; panic isolation; `MaxConns`
- CSRF: Drive POST without token = 403; WS without token = close

JS: `pkg/amarra/js/**/*.test.mjs` via `node --test`. Drive morph, 422, optimistic rollback, Stream parse. Idiomorph faked in unit tests; one integration fixture with testdata HTML.

**CLI / scaffold**

- `new` (full, minimal, blank) emits no `.svelte`, `vite.config.js`, or gonertia; emits `web/templates/**` and `amarra.js`
- `g resource|handler|auth|component` and `destroy` file + AST assertions
- `doctor` fails on Inertia leftovers or missing `<.foo>`
- Smoke (`scripts/smoke-scaffold.sh`): `GET /` is HTML with `#amarra-main` and `/static/js/amarra.js`, not Inertia JSON

**Generated app tests**

`httptest` + `AssertHTMLContains`. No `setupTestInertia`. Login GET has email + CSRF; bad POST = 422 + error text; good POST = 303.

**CI**

`make ci` = `go test ./... -race` + Amarra `js-test` + lint + format + smoke. No Vitest `test:fe`. No Playwright in the framework CI.

---

## Milestones

First published tag of the fork is **v0.1.0** (not Cais v1). Numbers below are product slices, not semver of Cais.

### Slice A — HTML-first (tag v0.1.0)

1. New repo `github.com/puppe1990/amarra-cais` copied from Cais v0.11, module path rewritten, CLI `cmd/amarra-cais`. This spec is copied to the new repo.
2. Delete Inertia/Svelte/Vite scaffold and generators.
3. `pkg/amarra/view` + preprocessor + kit.
4. Drive + Frame + Stream + Hook in `amarra.js`.
5. Rewrite `new`, `g handler|resource|auth|component`, `dev`, `install`, `build`, `doctor`, `pwa`, `destroy`.
6. Chat stays on Stream SSE.
7. `live.go` exists as types + 501 on `/amarra/live` so the JS namespace `amarra.live` is reserved.

### Slice B — Live (tag v0.2.0)

1. Hub, Conn, Mount/Handle/Render, CSRF on upgrade.
2. `amarra-click` / `change` / `submit` over WS.
3. In-process broadcast, `MaxConns`, idle timeout, reconnect remount.
4. `g stream chat` can opt into Live; resources stay on Drive.

---

## Out of scope (this design)

- Keeping Inertia as a default or `--front inertia` flag
- Migrating existing v0.11 Svelte apps
- Renaming Cais v0.11
- `templ`, `{@x}`, ERB `<%= %>`, named slots
- Redis / multi-process Live pub/sub
- Offline mutation queue
- Playwright as a framework CI gate
- Replacing SQLite, jobs, or the Lightsail deploy story

---

## Copy from Cais v0.11

Keep: router, middleware, session, CSRF, flash, jobs, migrate, i18n, mail, PWA machinery, console, doctor idea, `html/template` renderer (replaced at the API boundary by `pkg/amarra/view`), pagination, `pkg/cais/stream` (moved). `pkg/cais/forms` may remain as the HTML field builder behind `<.input>` / `<.form>`; generated apps only see Amarra Views.

Drop: Inertia scaffold, Svelte pages, Vite watch, `resource_gen_inertia.go`, `pwa.InstallForInertia` as default, public `hx*` helpers, gonertia in generated `go.mod`.
