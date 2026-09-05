# Amarra-cais — AI Conventions

Primary reader of this repo is often an LLM agent: cheap `rg`, expensive full reads, file truncation ~2k lines. Optimize for grep-unique names, small modules, and headless tests.

This is the HTML-first fork of Cais. Generated apps use **Amarra Views + Drive**, not Inertia + Svelte. The CLI binary is `amarra-cais` (it does not overwrite `cais`). Cais v0.11.x remains the Inertia product.

## Rule #1: TDD is mandatory

Before writing production code:

1. Write the test in `*_test.go` (framework JS: `pkg/amarra/js/*.test.mjs` or `pkg/cais/js/**/*.test.mjs`)
2. Run: `go test ./... -v -run TestName` (framework JS: `npm run js:test`)
3. Confirm it **fails** for the right reason (missing feature, not a typo)
4. Write the **minimal** code to make it pass
5. Run: `make test` and/or `make js-test` when JS changed
6. Only then refactor

Scaffolded apps: Go tests only (`go test ./...` / `amarra-cais test`). Pages are `web/templates/pages/*.html`, not Svelte.

## Clean Code for Agents (Akita / ranked)

Imperative. Prefer these when trading off effort.

| Priority | Rule                                                                                                                                                                               |
| -------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1        | **Small units** — functions ~4–20 lines; files target 200–300 lines, hard cap ~500. Split god files before adding more logic.                                                      |
| 2        | **SRP** — one reason to change per file/package. Prefer three focused modules over one multi-concern dump.                                                                         |
| 3        | **Greppable names** — unique domain nouns. Avoid `data`, `handler`, `Manager`, `Service`, `util`, `helper` as primary names. If `rg Name` floods, rename.                          |
| 4        | **Comments = WHY / provenance** — security tradeoffs, SQLite limits, issue IDs (`#123`), upstream constraints. Never strip intent comments on refactor. No `// increment i` noise. |
| 5        | **Explicit contracts** — Go public APIs fully typed; HTML pages and kit tags named; no silent `interface{}`/`any` on boundaries.                                                   |
| 6        | **DRY** — extract shared logic (generators, handlers, scaffold templates). No copy-variant drift.                                                                                  |
| 7        | **Headless tests** — one command: `make ci` (or focused `go test ./pkg/... -run X`). SQLite `:memory:`; no manual seed for unit tests. Bug fix → regression test.                  |
| 8        | **Predictable layout** — table below + generator patch markers. Mirror `foo.go` ↔ `foo_test.go`.                                                                                   |
| 9        | **Inject deps** — handlers take `Store`, `*view.Renderer`, `cais.Config` via constructor; no hidden globals.                                                                       |
| 10       | **Early returns** — max ~2 nesting levels for control flow.                                                                                                                        |
| 11       | **Errors with values** — `fmt.Errorf("tailwind watch: %w", err)` not bare `"failed"`.                                                                                              |
| 12       | **Format without debate** — `gofmt` / `make format` (Prettier).                                                                                                                    |
| 13       | **Structured logs** — JSON request/SQL in dev via `devlog`/`sqllog`; plain text only for CLI UX.                                                                                   |

### Repo red flags (split when you touch them)

| Path                                          | Notes                                                         |
| --------------------------------------------- | ------------------------------------------------------------- |
| `internal/cli/tpl_scaffold_handlers_*.go`     | Split done: home/contact/dashboard/auth + tests + testhelpers |
| `internal/cli/doctor.go` + `_env` + `_mobile` | Split done: core / env production / mobile checks             |
| `internal/cli/tpl_scaffold_handlers_auth.go`  | ~420 — split login vs signup/reset if it grows further        |
| `internal/cli/tpl_scaffold_tooling.go`        | ~600 — CI vs package.json vs Makefile templates               |
| `pkg/cais/pwa/assets/cais*.js`                | Leftover HTMX runtime; generated apps ship `amarra.js` only   |

Scaffold `const tpl*` blobs: one family per file (`tpl_scaffold_handlers_auth.go`), never a mega template dump.

## Structure

| Directory                 | Responsibility                                                                                               |
| ------------------------- | ------------------------------------------------------------------------------------------------------------ |
| `pkg/cais/`               | Framework: config, router, render, middleware, sessions, jobs                                                |
| `pkg/amarra/view/`        | HTML views: `Load`, `<.component>` expand, kit (`form`/`input`/`select`/`textarea`/`checkbox`), `view.Write` |
| `pkg/amarra/`             | Drive/Frame headers (`IsDrive`, `FrameID`)                                                                   |
| `pkg/amarra/stream/`      | Named SSE ops (`append`, `prepend`, `replace`, `morph`, `remove`, `toast`)                                   |
| `pkg/amarra/live/`        | Live WebSocket hub (`Mount`/`Handle`/`Render`, in-process broadcast)                                         |
| `pkg/amarra/js/`          | Drive / Frame / Stream / Hook sources; esbuild → `pkg/cais/pwa/assets/amarra.js`                             |
| `pkg/cais/httpx/`         | Render and redirect helpers for handlers                                                                     |
| `pkg/cais/meta/`          | Open Graph / Twitter preview (`Site`, `PreviewHTML`)                                                         |
| `pkg/cais/session/`       | Cookie sessions (`SignIn`, `SignOut`, `Store`)                                                               |
| `pkg/cais/boot/`          | Rails-style startup banner                                                                                   |
| `pkg/cais/devlog/`        | Development log buffer + `/logs` viewer                                                                      |
| `pkg/cais/sqllog/`        | SQL query logging wrapper (`Wrap`, `EnabledForEnv`)                                                          |
| `pkg/cais/console/`       | Interactive REPL (yaegi + SQL)                                                                               |
| `pkg/cais/csrf/`          | CSRF tokens (double-submit cookie)                                                                           |
| `pkg/cais/validate/`      | Form field validation helpers                                                                                |
| `pkg/cais/forms/`         | Template helpers (`csrfField`, `fieldError`, `makeField`, `fieldInput`, `fieldPassword`)                     |
| `pkg/cais/dotenv/`        | `.env` parser (`Parse`, `LoadFile`) used by `cais.Load()`                                                    |
| `pkg/cais/i18n/`          | Locale catalogs (`LOCALE` env, `t` template func)                                                            |
| `pkg/cais/testutil/`      | Test helpers (`NewRenderer`, `NewRequest`, `AssertHTMLContains`, `AssertChatMarkers`)                        |
| `pkg/cais/pwa/`           | Default PWA assets (`InstallForAmarra` ships `amarra.js`)                                                    |
| `pkg/cais/cache/`         | In-memory TTL cache + stable `Key`/`Hash` for ETags                                                          |
| `pkg/cais/pagination/`    | Offset/limit helpers for list pages                                                                          |
| `pkg/cais/stream/`        | Legacy SSE relay (prefer `pkg/amarra/stream` in new code)                                                    |
| `pkg/cais/chat/`          | Chat bubbles, tool UI (`LiveBubble`, `MessageBubble`)                                                        |
| `pkg/cais/middleware/`    | CSRF, sessions, auth, rate limits, security headers                                                          |
| `pkg/cais/passwordreset/` | Password-reset tokens + notifier interface                                                                   |
| `pkg/cais/sqlite/`        | WAL, busy timeout, foreign keys (`Configure`)                                                                |
| `pkg/cais/netutil/`       | Health payload + LAN URLs for mobile testing                                                                 |
| `pkg/cais/migrate/`       | `schema_migrations` runner (idempotent on boot)                                                              |
| `pkg/cais/flash/`         | One-shot flash cookies                                                                                       |
| `pkg/cais/jobs/`          | SQLite background job queue                                                                                  |
| `pkg/cais/jobsui/`        | Localhost `/jobs` dashboard (counts, failed retry/discard, recurring)                                        |
| `pkg/cais/testdata/`      | Fixture HTML (legacy HTMX layouts + chat_sse partials for framework tests)                                   |
| `pkg/cais/htmx.go`        | Leftover HTMX helpers — not the public generated-app contract                                                |
| `internal/cli/`           | Generators (`amarra-cais new`, `g`, `destroy`) — **HTML + Amarra scaffolds**                                 |
| `cmd/amarra-cais/`        | CLI entry point                                                                                              |
| `cmd/pwagen/`             | Write PWA assets into a target directory                                                                     |

This repo is **framework + CLI only** (no dogfood app). Apps live outside; create with `amarra-cais new`.

### Generated apps (`amarra-cais new`)

Layout: `web/templates/layouts/app.html` (`#amarra-main`) + `web/templates/pages/*.html` + `web/static/js/amarra.js`. Default scaffold uses `pwa.InstallForAmarra` (no Vite, no gonertia, no HTMX JS bundles).

`amarra-cais g handler` / `amarra-cais g page` generate **HTML** pages in `web/templates/pages/`.

`amarra-cais g resource` generates HTML admin CRUD (`view.Write` + kit `<.form>`). `amarra-cais g stream chat` uses Amarra Stream SSE (`data-amarra-stream`). `amarra-cais g component` writes `web/templates/components/<name>.html`.

Handlers do **not** check `HX-Request`. They call `view.Write`.

## Router path params and groups

```go
r.Get("/blog/{slug}", cais.StringParam("slug", blog.Show))
r.Post("/chat/{id}/permissions/{permID}/approve", cais.StringParams("id", "permID", chat.ApprovePermission))
r.Get("/items/{id}/{slug}", cais.IntStringParams("id", "slug", items.Show))
r.Group(middleware.Protect, func(g *cais.Router) {
  g.Post("/admin/items/{id}", cais.IntParam("id", admin.Update))
})
```

**ServeMux conflicts (Go 1.22+)** — registration panics when two patterns under the same method can match the same path and neither is more specific:

```go
// ❌ panic: both match /webhooks/kiwify/delete
r.Post("/webhooks/kiwify/{token}", receiver)
r.Post("/webhooks/{id}/delete", deleteHandler)

// ✅ prefer REST method or a distinct prefix
r.Delete("/webhooks/{id}", deleteHandler)
r.Post("/webhooks/incoming/{token}", receiver)
```

`cais.Router` rewrites the panic with a short hint. `amarra-cais routes` also warns when it can detect this statically from `routes.go`.

## Admin protection

| Mode                    | Middleware                         | Generator flag                   |
| ----------------------- | ---------------------------------- | -------------------------------- |
| Browser admin (default) | `middleware.RequireAuth("/login")` | `amarra-cais g resource` default |
| Bearer token API        | `middleware.AdminAuth(cfg)`        | `--admin-auth bearer`            |

Set `ADMIN_TOKEN` in production (`cfg.Validate()` fails on boot if missing). `AdminAuth` accepts Bearer header only, no query params. No-op in development when unset.

`amarra-cais g resource` defaults to session auth (`--admin-auth session`). Use `--admin-auth bearer` for token-only admin APIs without login pages.

## Session auth

`amarra-cais new` includes login/logout and protects `/dashboard`. Add to existing apps with `amarra-cais g auth`.

```go
r.Use(middleware.LoadSession(deps.Store.Sessions()))
r.Use(middleware.Flash(cfg))
r.Get("/dashboard", middleware.RequireAuthFunc("/login", dashboard.ServeHTTP))
auth := handlers.NewAuthHandler(views, store, site, store.Sessions(), cfg, catalog)
session.SignIn(w, sessions, r, userID, session.CookieOptionsFromConfig(cfg))
flash.Set(w, "notice", "Bem-vindo!", cfg.CookieSecure())
```

Dev seed user: `demo@example.com` / `password`. Sessions persist in SQLite via `session.NewSQLiteStore`.

**Session expiry** — cookies and DB rows expire after 7 days (`sessionTTL` / `defaultMaxAge`). SQLite stores `expires_at`; expired rows are ignored on lookup. Prune stale rows with `amarra-cais db prune-sessions` (or call `session.Store.PruneExpired()`).

**Production cookies** — `session.CookieOptionsFromConfig(cfg)` sets `Secure` when `cfg.CookieSecure()` is true (`ENV=production`).

## Amarra Views + Drive (default frontend)

Handlers in generated apps render **HTML** via `pkg/amarra/view`. Boot loads templates once (`view.Load`); unknown `<.component>` tags fail at boot, not on the first request.

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

// Validation — same page, status 422, `.Errors` on inputs
writeView(w, r, h.views, h.cfg, "contact", amarraData(r, h.site, map[string]any{
  "Title":  h.catalog.T("contact.title"),
  "Errors": errs,
  "Name":   name,
  "Email":  email,
}), http.StatusUnprocessableEntity)

// Flash on redirect — cais cookie API only
flash.Set(w, "notice", "Saved!", cfg.CookieSecure())
http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
```

Drive requests still render the layout so JS can morph `#amarra-main`. Frame requests (`Amarra-Frame: <id>`) render `{{ define "frame:<id>" }}` only.

**Public HTML contract** (`amarra.js`): `data-amarra-drive`, `data-amarra-confirm`, `data-amarra-method`, `data-amarra-disable-with`, `data-amarra-frame`, `amarra-click` / `amarra-change` / `amarra-submit` / `amarra-debounce` / `amarra-hook` / `amarra-live`, `<amarra-frame loading="lazy">`. HTMX is not a public dependency. Layout loads a single script: `/static/js/amarra.js`.

```html
{{ define "content" }}
<.form action="/items" method="post">
  <.input name="title" label="Title" value="{{ .Item.Title }}" error="{{ fieldError .Errors "title" }}" />
  <.button type="submit">Save</.button>
</.form>
{{ linkTo "/items/1" "Delete" (dict "method" "delete" "confirm" "Delete this item?") }}
<div amarra-hook="clipboard" data-amarra-copy="hi">Copy</div>
<.input type="password" name="password" />
<button type="button" amarra-hook="password" data-amarra-password-for="#password">Show</button>
<select amarra-hook="reveal" data-amarra-reveal-show="access_keys" data-amarra-reveal-target="#aws-keys">
  <option value="default_chain">Default chain</option>
  <option value="access_keys">Access keys</option>
</select>
<div id="aws-keys" hidden>…keys…</div>
<button type="button" amarra-hook="theme" data-amarra-theme-color="#f5f5f4" data-amarra-theme-color-off="#0f172a">Theme</button>
{{ end }}
```

Shipped `amarra-hook` builtins: `clipboard`, `password` (toggle `type` + `aria-pressed`), `reveal` (show/hide a target when the control value matches, no Drive round-trip), `theme` (toggle `html.light`, persist `localStorage["amarra-theme"]`, optional `theme-color` meta). Put this FOUC snippet in the layout `<head>` before CSS so a stored light theme does not flash dark:

```html
<script>
  try {
    if (localStorage.getItem("amarra-theme") === "light")
      document.documentElement.classList.add("light");
  } catch (e) {}
</script>
```

Shipped kit (override in `web/templates/components/`): `form`, `input`, `select`, `textarea`, `checkbox`, `button`, `flash`, `nav`, `pagination`, `modal`, `locale-toggle`. One slot: `.Inner`. Self-closing `<.flash />` is allowed. `<.form>` injects `csrf_token` from `.CSRFToken` and `data-amarra-drive="true"`.

`{{ linkTo "/x" "Label" }}` emits a Drive-enabled `<a>`. Optional `(dict "method" "delete" "confirm" "Sure?" "frame" "cart")`.

**Live** (`GET /amarra/live`) is an opt-in WebSocket hub. CRUD stays on Drive. Register views with `hub.Register`; pages use `amarra-live` + `amarra-click`. CSRF is the join payload vs the handshake cookie. Hub is in-process only. `sock.Patch`, `sock.Stream`, and `sock.Push` ride on the morph message.

**Stream HTTP:** `stream.WriteHTTP(w, stream.Op{Kind: "append", Target: "list", HTML: row})` with `Content-Type: text/vnd.amarra-stream` applies ops instead of morphing `#amarra-main`.

**Frontend TDD** — framework JS only:

```bash
npm run js:test   # pkg/cais/js/**/*.test.mjs + pkg/amarra/js/**/*.test.mjs
```

## New page (scaffolded app)

1. Go test in `internal/handlers/foo_test.go` — `setupTestViews(t)` + assert HTML (`id="amarra-main"`, form fields)
2. HTML page in `web/templates/pages/foo.html` (`{{ define "content" }}`)
3. Handler — `view.Write` / `writeView` with `amarraData`
4. Register the route in `internal/app/routes.go`

Or use `amarra-cais g handler foo` (creates handler + test + `web/templates/pages/foo.html` + route patch).

Pass `meta.SiteFrom(appName, cfg.AppURL)` from bootstrap for OG/Twitter in page data (`amarraData` / `meta.ForRequest`).

**Integration tests** (`internal/app/` when present, or handler tests): GET page first; POSTs return `303` on success and `422` HTML with `.Errors` on validation. Do not set `X-Inertia`.

## CSRF

- `middleware.CSRF(cfg)` on the router (validates POST/PUT/DELETE/PATCH)
- Double-submit cookie (`cais_csrf`) — token in cookie + form field or `X-CSRF-Token` header; no server-side token store

**HTML templates** — pass `meta.ForRequest(site, r)` / `amarraData`; layout renders `<meta name="csrf-token">` + `amarra.js` sends `X-CSRF-Token` on Drive requests. Forms: `{{ csrfField .CSRFToken }}` or hidden `csrf_token` input. Kit `<.form>` still needs the CSRF field in the body.

**Integration tests** — GET page first (read `csrf` cookie), then POST with matching `csrf_token` field + cookie.

## Flash messages

- `middleware.Flash(cfg)` on the router (after `LoadSession`)
- **One API:** `flash.Set(w, "notice", "Saved!", cfg.CookieSecure())` then redirect.
- Read on the next request: `flash.MessageFromRequest(r)` → put into page data (scaffold `amarraData` copies `s.Flash`). Layout: `<.flash />` **inside** `#amarra-main` so Drive morph keeps notices.
- Helpers: `{{ flashMessage .Flash }}` (`pkg/cais/forms`) — never `{{ .Flash }}` (stringifies the struct)
- One-shot: consumed on the next request

## Mobile PWA

- `boot.Print` shows **LAN** URLs for phone testing on Wi‑Fi
- `GET /health` returns `lan_urls` via `netutil.HealthPayload` — use this array, never concatenate `APP_URL` + port manually
- `amarra-cais pwa --bump` increments `CACHE_VERSION` in `sw.js` after template/HTML changes
- `amarra-cais dev` auto-bumps `CACHE_VERSION` when `sw.js` exists (fresh assets on phone without a manual bump)
- `amarra-cais doctor --mobile` checks flash markup, Google Fonts CSP, `amarra.js`, SW cache (network-first `/static/js/amarra.js`), chat form CSS, `#chat-messages` scroll container, and health `lan_urls`
- Scaffold `input.css` uses system fonts (no `fonts.googleapis.com` — blocked by default CSP)

**Mobile checklist:** `amarra-cais doctor --mobile` → `amarra-cais pwa --bump` → open boot **LAN** URL on phone → stay on the page while SSE streams (Drive morphs `#amarra-main`; a full navigation drops the EventSource)

## Security headers

- `middleware.SecurityHeaders(cfg)` on the router (after `Recover`)
- Sets `X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy`, `Permissions-Policy`
- Adds `Strict-Transport-Security` in production (`ENV=production`)
- CSRF and flash cookies use `Secure` when `cfg.CookieSecure()` is true
- Session rotates on login (invalidates previous token)

## Rate limiting

Wrap sensitive POST routes with per-IP token buckets:

```go
loginLimit := middleware.NewRateLimiter(10, cfg)   // 10 req/min
contactLimit := middleware.NewRateLimiter(20, cfg) // 20 req/min
r.Post("/login", loginLimit.Middleware(http.HandlerFunc(auth.LoginPost)).ServeHTTP)
r.Post("/contact", contactLimit.Middleware(http.HandlerFunc(contact.Post)).ServeHTTP)
```

Rate limiters use `middleware.ClientIP(r, cfg)` — set `TRUSTED_PROXIES` when behind a reverse proxy.

## SSE / streaming

Amarra Stream (`pkg/amarra/stream`) is the named-op contract. Long-lived streams need server and handler setup:

| Setting        | Normal handlers | SSE routes                                             |
| -------------- | --------------- | ------------------------------------------------------ |
| `WriteTimeout` | `30s` ok        | **`0`** (disabled) — set on chat/stream routes         |
| Flush          | N/A             | `stream.Flush(w)` — never assert `http.Flusher` on `w` |

```go
import (
    "github.com/puppe1990/amarra-cais/pkg/amarra/stream"
    "github.com/puppe1990/amarra-cais/pkg/cais/chat"
)

func streamHandler(w http.ResponseWriter, r *http.Request) {
    stream.RelaySSE(w)
    _ = stream.WriteOp(w, stream.Op{Kind: "morph", Target: "chat-live", HTML: chat.LiveBubble("token…")})
    _ = stream.WriteOp(w, stream.Op{Kind: "append", Target: "chat-history", HTML: chat.MessageBubble(chat.RoleAssistant, "done", time.Now().UTC())})
    _ = stream.Flush(w) // never w.(http.Flusher) — middleware may hide Flusher
}
```

Named ops: `append`, `prepend`, `replace`, `morph`, `remove`, `toast`. Chat partials use `data-amarra-stream` (not `hx-ext` / `sse-ext`).

**Chat template patterns** — two partials ship with `amarra-cais g stream chat`:

| Partial               | Use case                        | DOM                                                                    |
| --------------------- | ------------------------------- | ---------------------------------------------------------------------- |
| `chat_sse.html`       | Echo / simple bots              | `#chat-history` + `#chat-live`; `data-amarra-stream`                   |
| `chat_sse_agent.html` | Agent streaming (tokens, tools) | `#chat-history` + `#chat-stream` + `#chat-live` under `#chat-messages` |

**Server helpers** — `pkg/cais/chat`: `LiveBubble`, `MessageBubble`, `SafeMessageBubble`, `ToolCallBubble`, `ToolResultBubble`, `DetailBubble`, `Truncate`, `SelectWindowWithLastUser`, `UnsafeLiveHTML` / `UnsafeMessageHTML` (caller sanitizes HTML).

**Chat handler tests** — `amarra-cais g stream chat` generates handler tests using `testutil.AssertChatMarkers` (Show) and `testutil.AssertHTMLContains` (PostMessage bubble). Missing records return `http.NotFound` (not 500):

```go
conv, err := h.store.FindConversationByID(id)
if err != nil {
    http.NotFound(w, r)
    return
}
```

## Drive / Frame / Hook

Used by `amarra-cais new` / `g resource` / `g stream chat`.

- Pages in `web/templates/pages/` — `{{ define "content" }}` plus optional `{{ define "frame:<id>" }}`
- Layout: `#amarra-nav`, `#amarra-main`, `#amarra-toast-host`; one script `/static/js/amarra.js`
- Do not test with `HX-Request`. Drive sets `Amarra-Drive: true`; frames set `Amarra-Frame: <id>`
- List caching — combine `cache.Key(...)` + `cache.Hash(version)` with `httpx.NotModified` / `httpx.SetETag` for 304 responses on stable list pages

`pkg/cais/htmxattrs` and `httpx.WritePage` still exist for leftover HTMX apps. New generators must not emit `hx-*`.

## Form validation

Use `validate.Email`, `validate.URL`, `validate.Required`, `validate.MinLength`, `validate.MaxLength` for single-field checks. For multiple fields, collect errors in `validate.FieldErrors`:

```go
var errs validate.FieldErrors
if item.Name == "" {
  errs.Add("name", "Name is required")
}
if errs.Any() {
  // re-render form with errs map — templates use {{ fieldError .Errors "name" }}
}
```

Pass `errs` as `.Errors` in page data when re-rendering forms.

**Form helpers** (`pkg/cais/forms`, registered on the view renderer):

```html
{{ fieldInput (makeField "name" "Name" .Name "text" true .Errors) }}
```

`makeField` returns `forms.FieldData`; `fieldInput` renders input/textarea/checkbox + error. Resource generator admin forms use these plus kit `<.form>` / `<.button>`.

Password fields: always `fieldPassword` (eye show/hide).

Foreign-key selects:

```html
{{ fieldSelect (makeSelectField "category_id" "Category" .Item.CategoryID .CategoryOptions true
.Errors) }}
```

## Foreign keys in generators

`amarra-cais g resource post --fields title:string,category_id:references` (or `category:belongs_to`):

- Migration column: `INTEGER [NOT NULL] REFERENCES categories(id)`
- Store: `ListCategoryOptions()` — `SELECT id, COALESCE(name, title, id) FROM categories`
- Admin form: `fieldSelect` / `makeSelectField` populated from options

Generate the parent resource first (`amarra-cais g resource category --fields name:string`). Referenced table needs a `name` or `title` column for labels.

## Integration tests (auth + contact)

Multi-step flows belong in `internal/app/` (full router + CSRF + session + HTML) when you add them; scaffold ships focused handler tests:

1. `GET /login` or `/contact` — read `csrf` cookie
2. `POST` with matching `csrf_token` field + cookie
3. Follow session/flash cookies on subsequent requests

Success: `303`. Validation: `422` HTML containing the field error. Handler tests use `setupTestViews` + `httptest` (see generated `home_test.go` / `contact_test.go` / `auth_test.go`).

## New table

1. Store test with `":memory:"` before the migration
2. SQL in `internal/store/migrations/NNN_name.sql`
3. Methods on the `store.Store` interface
4. Wrap DB with `sqllog.Wrap` in `NewSQLiteStore` for development query logs
5. Migrations tracked in `schema_migrations` via `pkg/cais/migrate` (idempotent on boot)

**SQLite concurrency (SSE / chat)** — scaffold `NewSQLiteStore` calls `sqlite.Configure`: `journal_mode=WAL`, `busy_timeout=5000`, `foreign_keys=ON`, `MaxOpenConns(1)`. WAL allows concurrent readers while a writer holds the lock briefly; `busy_timeout` retries instead of immediate `SQLITE_BUSY`. SSE handlers poll the DB in a loop — keep writes short and avoid long transactions during streams. For heavy write concurrency, consider a dedicated writer queue.

**Migration down sections** — use `-- up` / `-- down` markers in `.sql` files (generator default for resources):

```sql
-- up
CREATE TABLE bookmarks (...);

-- down
DROP TABLE IF EXISTS bookmarks;
```

`amarra-cais db rollback` executes the `-- down` SQL when present, then removes the `schema_migrations` row. Without a down section, only the record is removed.

## Development logging

In `ENV=development`:

- `middleware.LoggerTo(devlog.MirrorDefault(...))` — JSON request logs when `cfg.LogJSON()` (`kind: request`); `LOG_FORMAT=text` opts out
- `sqllog.ConfigForEnv(env)` — SQL JSON logs in development (`kind: sql`); plain text when `JSON: false`
- `devlog.Register(r, cfg.Env, buf)` — mounts `/logs` (localhost only)
- `jobsui.Register(r, db)` — mounts `/jobs` (localhost only, all envs; SSH tunnel in production)

Boot banner via `boot.Print` in `cmd/server/main.go`. Port auto-pick via `cais.ResolvePort` when preferred port is busy.

Set `APP_URL` for absolute OG image URLs. **`APP_URL` is required when `ENV=production`** — `cfg.Validate()` fails on boot if missing.

Set `TRUSTED_PROXIES` (comma-separated IPs) when behind a reverse proxy so `middleware.ClientIP` trusts `X-Forwarded-For` for rate limiting and logging.

`cais.Load()` applies a local `.env` if present (`pkg/cais/dotenv.LoadFile`). Keys already in the process env win — `t.Setenv`, systemd `Environment=`, and CI secrets are not overwritten. Missing `.env` is a no-op.

Set `LOCALE=en` (default) or `LOCALE=pt` for UI strings via `pkg/cais/i18n`. See [i18n design](docs/superpowers/specs/2026-07-01-i18n-design.md).

## CLI generators

```bash
amarra-cais new myapp              # includes AGENTS.md, GitHub Actions CI, pre-commit, golangci-lint, Prettier
amarra-cais new myapp --minimal
amarra-cais new myapp --blank
amarra-cais new myapp --module github.com/acme/myapp
amarra-cais g [--dry-run] stream chat              # SSE chat: conversations, messages, stream relay scaffold
amarra-cais g [--dry-run] resource bookmark --fields title:string,url:url,notes:text? --public --paginate --force
amarra-cais destroy [--dry-run] resource bookmark   # undo generator output
amarra-cais destroy [--dry-run] model bookmark      # remove model + migration + store methods
amarra-cais destroy [--dry-run] handler settings
amarra-cais destroy [--dry-run] auth                # remove login/auth scaffolding
amarra-cais destroy [--dry-run] migration add_tags  # remove *_add_tags.sql
amarra-cais g [--dry-run] model bookmark --fields title:string,url:url
amarra-cais g [--dry-run] handler settings
amarra-cais g [--dry-run] page about
amarra-cais g [--dry-run] component card
amarra-cais g [--dry-run] migration add_tags
amarra-cais g [--dry-run] auth       # login/logout + protected dashboard
amarra-cais g [--dry-run] console    # scaffold cmd/console/main.go
amarra-cais g [--dry-run] ci         # add CI/pre-commit to existing apps
amarra-cais g [--dry-run] job send_welcome --cron "0 3 * * *"
amarra-cais doctor [--mobile]        # verify amarra.js, layouts/app.html, air, go.mod, PWA/mobile
amarra-cais pwa [--bump]             # write/refresh PWA assets; --bump invalidates SW cache
amarra-cais link [path] [--unlink]   # go.mod replace for local framework dev (do not commit; unlink before push)
amarra-cais routes                   # list routes from internal/app/routes.go
```

**Aliases:** `amarra-cais g` → generate, `amarra-cais i` → install, `amarra-cais b` → build, `amarra-cais s` → server, `amarra-cais c` → console.

Field types: `string`, `text`, `url`, `bool`, `int`, `date`, `references` (or `name:belongs_to`). Suffix `?` for optional.

**Resource options:** `--public` (public list page), `--paginate` (admin index pagination, 25/page), `--no-seed` (skip demo data), `--admin-auth session|bearer` (default: session).

**Model generator** — `amarra-cais g model` creates model struct, migration, and store methods only (no handlers, templates, or routes). Use for data layer without admin CRUD.

**Dry-run** — `amarra-cais g --dry-run ...` and `amarra-cais destroy --dry-run ...` print planned changes without writing files.

**Destroy** — `amarra-cais destroy resource|handler|model <name>` removes generated files and unpatches `routes.go`, `store.go`, `seeds.go`, and layout nav where applicable. `destroy auth` also reverts `app.go` session middleware. `destroy migration` removes matching `*_<name>.sql` only (does not roll back `schema_migrations`).

**Demo seed** — `amarra-cais g resource` (unless `--no-seed`) generates `SeedDemo*` store methods and wires them into `cmd/server/main.go` at boot.

## App commands (run from an Amarra app)

```bash
amarra-cais install  # npm install + go mod tidy
amarra-cais css      # build Tailwind
amarra-cais dev      # hot reload + tailwind watch
amarra-cais build    # bin/server
amarra-cais server   # go run ./cmd/server
amarra-cais test     # go test ./...
amarra-cais doctor [--mobile]  # verify amarra.js, air, go.mod, PWA/mobile
amarra-cais pwa [--bump]       # write/refresh PWA assets
amarra-cais console  # Rails-style REPL (store, cfg, db + sql)
amarra-cais routes   # list HTTP routes from internal/app/routes.go
amarra-cais link [../amarra-cais] [--unlink]  # go.mod replace for local framework dev (unlink before push)
amarra-cais db migrate        # run pending migrations
amarra-cais db status         # list applied/pending migrations
amarra-cais db rollback       # roll back last migration (runs -- down SQL when present)
amarra-cais db prune-sessions # delete expired login sessions from SQLite
amarra-cais db seed           # run internal/db/seeds.go (idempotent; safe in production for catalog data)
amarra-cais db seed --list    # list seed helpers referenced in seeds.go
amarra-cais jobs work [--queues default,mail] [--concurrency 2]
amarra-cais jobs status
amarra-cais routes --verbose  # routes with handler names and middleware
amarra-cais version           # print framework version
```

## Background jobs

SQLite queue in `pkg/cais/jobs` (same DB file as the app). See [jobs design](docs/superpowers/specs/2026-07-01-jobs-design.md).

**Dashboard:** `GET /jobs` (localhost only, all envs) via `jobsui.Register` in `app.New`. Detail: `GET /jobs/{id}`. Filter `?kind=`. Failed: Retry / Discard. Orphans: Requeue stuck (skips live worker jobs). Finished: Clear finished. Worker heartbeats show liveness; two live workers warn (one SQLite file). `amarra-cais routes` lists these when `jobsui.Register` is present.

```bash
amarra-cais g job prune_sessions --cron "0 3 * * *"  # internal/jobs/*.go + registry + cmd/worker
amarra-cais db migrate                                # jobs + recurring_tasks tables
amarra-cais jobs work --concurrency 2                   # worker + delayed-job dispatcher + heartbeat
amarra-cais jobs status                               # counts + queues + workers + recurring
amarra-cais jobs retry 12
amarra-cais jobs discard 12
amarra-cais jobs prune [--older 24h]
```

Enqueue from handlers:

```go
jobs.Enqueue(ctx, jobStore, jobs.Options{Kind: "SendWelcome", Payload: data})
```

Inspect / recover:

```go
store.List(ctx, jobs.ListFilter{Status: jobs.StatusFailed, Kind: "SendWelcome"})
store.RetryFailed(ctx, id)
store.Discard(ctx, id)
store.PruneFinished(ctx, 24*time.Hour) // 0 = all finished rows
store.RequeueOrphaned(ctx, jobs.DefaultWorkerStale)
store.ListLiveWorkers(ctx, jobs.DefaultWorkerStale)
```

Register handlers in `internal/jobs/registry.go`. Built-in: `PruneSessions`. Production: run `amarra-cais jobs work` as a separate process next to `bin/server`. Open `http://127.0.0.1:<port>/jobs` (SSH tunnel if remote).

**Generator troubleshooting** — if `could not patch routes.go` or `could not patch store`, check that `registerRoutes` and `Close() error` markers exist. Public nav links need `<!-- cais:nav -->` in the layout (or `</nav>`). Run `amarra-cais db migrate` after `g resource` / `g model` / `g auth`.

Console bindings: `store`, `cfg`, `db`, plus any custom keys in `Bindings`. Commands: `help`, `sql`, `reload`, `history`, `!N`/`!!`, `exit`. Arrow keys when stdin is a TTY.

`/logs` — development-only log viewer (localhost). Shows request + SQL logs.

`/jobs` — queue dashboard (localhost, all envs). Counts, failed retry/discard, scheduled, recurring. `amarra-cais doctor` warns if `jobsui.Register` is missing.

## CLI generator layout

The `amarra-cais` CLI lives in `internal/cli/`. Scaffold templates are split by responsibility so agents can grep a single file instead of loading a 2400-line monolith.

| Path                                                            | Responsibility                                                       |
| --------------------------------------------------------------- | -------------------------------------------------------------------- |
| `internal/cli/cli.go`                                           | Command routing (`new`, `g`, `destroy`, `db`, …)                     |
| `internal/cli/tpl_scaffold_handlers_*.go`                       | HTML handler templates (home/contact/auth; not `*_test.go`)          |
| `internal/cli/doctor.go` / `doctor_env.go` / `doctor_mobile.go` | `amarra-cais doctor` checks (core / env / mobile)                    |
| `internal/cli/scaffold.go`                                      | `amarra-cais new` orchestration and `writeTemplate`                  |
| `internal/cli/resource.go`                                      | `amarra-cais g resource` orchestration (writes files, calls patches) |
| `internal/cli/resource_patch.go`                                | Patches store, routes, layout nav, seeds, main for resources         |
| `internal/cli/resource_gen_*.go`                                | Resource code generation (store, admin, public, HTML, fields)        |
| `internal/cli/tpl_scaffold_*.go`                                | Embedded `const tpl*` for `amarra-cais new` scaffolding              |
| `internal/cli/tpl_scaffold_main.go`                             | `cmd/server/main.go` (full + blank)                                  |
| `internal/cli/tpl_scaffold_app_core.go`                         | `internal/app/app.go` (full + blank)                                 |
| `internal/cli/tpl_scaffold_routes.go`                           | `internal/app/routes.go` (full, minimal, blank)                      |
| `internal/cli/tpl_scaffold_console.go`                          | `cmd/console/main.go`                                                |
| `internal/cli/tpl_scaffold_auth.go`                             | Auth Go templates (handler, store, model, migration, tests)          |
| `internal/cli/tpl_scaffold_auth_pages.go`                       | Auth HTML page templates (`login`, `signup`, reset)                  |
| `internal/cli/tpl_scaffold_handlers_{home,contact,auth,...}.go` | HTML handler scaffolds (one domain per file)                         |
| `internal/cli/tpl_scaffold_web.go`                              | Amarra layout (`#amarra-main` + `amarra.js`)                         |
| `internal/cli/tpl_stream_chat.go`                               | `amarra-cais g stream chat` templates and handlers                   |
| `internal/cli/stream.go`                                        | `amarra-cais g stream chat` orchestration                            |
| `internal/cli/scaffold_auth.go`                                 | `amarra-cais g auth` orchestration and store/app/route patches       |
| `internal/cli/patch.go`                                         | AST-safe patches into generated apps (`routes.go`, `store.go`)       |
| `internal/cli/patch/`                                           | `go/ast` helpers — regex patches break nested `cais.IntParam`        |
| `internal/cli/pwa_cmd.go`                                       | `amarra-cais pwa` asset writer                                       |
| `internal/cli/commands.go`                                      | `amarra-cais install`, `css`, `dev` (air + Tailwind, no Vite)        |
| `internal/cli/destroy.go`                                       | `amarra-cais destroy` — reverses generators                          |

**Generator tests** (split by domain — run focused suites while editing generators):

| Test file                     | Scope                                    |
| ----------------------------- | ---------------------------------------- |
| `cli_help_test.go`            | `amarra-cais help` output                |
| `cli_new_test.go`             | `amarra-cais new` (full, minimal, blank) |
| `resource_scaffold_*_test.go` | `amarra-cais g resource`                 |
| `scaffold_handler_test.go`    | `amarra-cais g handler` route patching   |
| `scaffold_model_test.go`      | `amarra-cais g model`                    |
| `scaffold_migration_test.go`  | `amarra-cais g migration` numbering      |
| `generate_dryrun_test.go`     | `--dry-run` generators                   |
| `patch_gomod_test.go`         | `replace` directive in scaffolded apps   |

```bash
go test ./internal/cli/... -run TestScaffoldResource -count=1
go test ./internal/cli/... -run TestCLI_New -count=1
go test ./internal/cli/... -count=1
```

**Patch markers** — generated apps must keep `registerRoutes`, `Close() error`, and `<!-- cais:nav -->` (or `</nav>`) for destroy/generator patches to work.

## Framework commands (this repo)

```bash
make test-v         # TDD: verbose Go tests
make test           # Go validation with -race (agent default for backend)
make js-test        # pkg/cais/js + pkg/amarra/js unit tests
make lint           # golangci-lint
make format         # prettier --write
make ci             # test + js-test + lint + format-check (full gate)
make build          # bin/amarra-cais
make install-cli    # go install ./cmd/amarra-cais
```

**One-shot validation:** `make ci`  
**Focused TDD:** `go test ./pkg/amarra/view/ -run TestWrite -count=1 -v`  
CI runs Go tests + `npm run js:test` + lint + Prettier + smoke (`amarra-cais new`).

## Production deploy (generated apps)

Cross-compile and ship static assets beside the binary:

```bash
amarra-cais css     # Tailwind → web/static/css/styles.css
amarra-cais build --os linux --arch amd64 -o bin/server-linux
tar czf release.tar.gz bin/server-linux web/static
```

- Guide: `docs/deploy/lightsail-systemd.md`
- Template: `deploy/systemd/cais-app.service.example`
- `amarra-cais doctor` checks `web/static` + `manifest.webmanifest` + `amarra.js`
- Set `STATIC_DIR` / `TEMPLATES_DIR` when `WorkingDirectory` is not the app root
- Dev-only seeds (demo user) do not run when `ENV=production`; use `amarra-cais db seed` for catalog data
- No Vite `web/static/build/` — ship `web/static/js/amarra.js` + CSS

Pre-commit (tests, lint, prettier): `make pre-commit-install` once, then hooks run on every commit.

## Agent code style

See **Clean Code for Agents** above. Project extras:

- Comment **why** on security, SQLite concurrency, CSRF/cookie, Drive vs full HTML, boot-time vs per-request.
- Do not narrate WHAT the code does — names and tests already cover that.
- Prefer file- or func-level provenance over inline noise; agents load whole files.
- Keep generator templates (`tpl_*`) accurate — there is no dogfood app to drift against; rely on CLI scaffold tests.
- Amarra: `view.Write`, `amarraData`, kit `<.form>` / `<.input>` / `<.button>` / `<.flash />`, CSRF `cais_csrf` / `X-CSRF-Token`.
- Password fields: always `fieldPassword` / `fieldInput` with type `password` — eye show/hide is default.
- Do not emit `vite.config.js`, `.svelte`, gonertia, or `HX-Request` checks in new generators.

## Defensive categories (implement only these)

Agents implement the categories listed — do not invent extra ops policy.

- [x] **Rate limits** — `middleware.NewRateLimiter` on login/contact (and sensitive POSTs)
- [x] **Timeouts** — HTTP server Read/Write/Idle; SSE uses `WriteTimeout: 0` + stream package
- [x] **SQLite busy** — `sqlite.Configure` WAL + `busy_timeout`; short writes during SSE
- [x] **CSRF** — double-submit cookie on state-changing methods
- [x] **Production gate** — `cfg.Validate()` requires `ADMIN_TOKEN` + `APP_URL`
- [ ] **Retries with backoff** — only if adding outbound HTTP clients (mail, webhooks); not for request handlers
- [ ] **Circuit breaker** — not required unless proxying flaky upstreams

## Do not

- Parse templates per request (use `view.Load` once at boot)
- Use inline CSS (use Tailwind classes in templates)
- Mock the database (use SQLite `:memory:`)
- Import `internal/` from `pkg/cais/` or `pkg/amarra/` (avoids import cycles)
- Grow files past ~500 lines without splitting
- Ship features without a test the agent can run headless
- Strip WHY/provenance comments “to clean up”
- Check `HX-Request` or render Inertia JSON in generated apps
- Add `vite.config.js` / Svelte pages to `amarra-cais new`
- Use `$form` store syntax or Svelte 4 `new App()` in scaffolds
