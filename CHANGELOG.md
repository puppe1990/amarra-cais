# Changelog

All notable changes to the Cais framework are documented here.

Format based on [Keep a Changelog](https://keepachangelog.com/). Versioning follows [Semantic Versioning](https://semver.org/).

## Unreleased

### Fixed

- `amarra-cais pwa --bump` now refreshes the vendored assets before incrementing `CACHE_VERSION`, matching the help text (#187).
- `amarra-cais pwa` preserves existing app-owned brand assets (`manifest.webmanifest`, `offline.html`, `og.png`, icons) and only refreshes the framework runtime; `--force` overwrites them with defaults (#186).

## [0.10.0] - 2026-09-17

### Added

- `cache.SetMaxEntries(n)` (default 4096) and `RateLimiter.SetMaxBuckets(n)`: high-cardinality keys evict in batches instead of growing the heap.
- `fsutil.RefuseSymlinkWrite`: CLI and PWA writes refuse symlinked targets or ancestors (`internal/`, `store.go`, `amarra.js`).
- `sqlite.DSN(path)`: pragmas ride the DSN, so driver reconnects keep `foreign_keys` and `busy_timeout`.
- `jobs.WorkerConfig.HeartbeatStore` plus a dedicated liveness pool in the scaffolded worker.
- `live.Hub.Dropped()` counts slow-client drops (logged on the first and every 100th).
- `httpx.ServerError(w, err, cfg)`: sanitized 500s for generated handlers.
- `make js-bundle-check` and a CI step that rebuilds and diffs the committed bundles.

### Changed

- `jobs`: the worker retries transient store errors with exponential backoff, drains handlers before dropping the heartbeat, and `DispatchDue` runs `BEGIN IMMEDIATE`; the recurring scheduler claims each tick with compare-and-set.
- `migrate`: each version is claimed with `INSERT OR IGNORE` and schema setup retries `SQLITE_BUSY`, so concurrent boots (server + worker) stop failing with `UNIQUE constraint`.
- Drive/Frame drop superseded responses (monotonic sequence per document/element); `bulk`, `dialog` and `dropdown` hooks rebind on `amarra:morphed`.
- `netutil.HealthPayload(status, port, env)` omits `lan_urls` in production; `RequireAuth` always redirects 303 for Drive.
- `pagination` clamps `page` to the last page and saturates offsets; generated SQL quotes identifiers and field names are validated (reserved words, `id`/`created_at`, duplicates rejected).
- Generated handlers use `httpx.ServerError`; `UnsafeMessageHTML` maps roles through an allowlist; `linkTo` unchanged.
- `live` hides handler error detail outside development.
- HTMX legacy packages and assets are marked deprecated with removal at v1.0; generators are guarded against emitting `hx-*`.

### Fixed

- Generator patches validate markers before writing and roll back on failure (no more half-patched apps); `MarkFailed` clears `worker_id`/`started_at`; `RelaySSE` logs unsupported `SetWriteDeadline`; the nav hook ignores `#`/fragment links; the password toggle only flips `type=password` fields.
- Files over 500 lines split by domain, with a test enforcing the cap.

## [0.9.0] - 2026-09-16

### Added

- `amarra-cais new` records every generated file in `.cais-generated.json`; `destroy` skips untracked or modified files unless `--force` (missing manifest fails closed).
- `destroy` refuses parents still referenced by other resources and drops orphan `List<Ref>Options`; generators record their own files after gofmt.

### Changed

- Drive morphs remount `data-amarra-stream` and `amarra-live` nodes (rescan on `amarra:morphed`, replaced nodes disconnect) — chat/live reached through a link now connects.
- Service worker skips `no-store`/`private` responses, never falls back to cached pages for navigations or Drive fragments, and clears the cache on `POST /logout`.
- `jobs` worker drains in-flight handlers (bounded by `DrainTimeout`, default 30s) before removing its heartbeat and records the final status with a detached context.
- `linkTo` allowlists `http`, `https`, `mailto`, `tel` and relative hrefs; unsafe schemes (e.g. `javascript:`) render `href="#"`.

### Fixed

- Kit tags accept hyphenated attributes (`data-*`, `aria-*`): expansion maps them to `attr_*` variables instead of failing the template parse at boot.
- `g resource`: prefix collisions no longer skip store patches (`post_comment` → `post`); bool-only fields compile; `references` without the generated parent fails early pointing at the parent command.
- `destroy`: removing `boolInt` no longer corrupts `store.go`.
- `jobs`: handler panics fail the job instead of killing the worker; orphaned jobs at `max_attempts` become `failed` instead of looping forever.
- `Live`: events broadcast before the join handshake are dropped and `Mount` is serialized before `Handle`.
- SSE: lone `\r` in payloads is folded into `data:` lines; `WriteOp` rejects unknown op kinds.

### Security

- `g`/`destroy` reject names that resolve outside the app dir (path traversal); write/remove helpers refuse escaping rel paths.
- `/jobs` rejects requests carrying forwarding headers, so a same-host reverse proxy cannot expose the dashboard.
- Scaffold `.gitignore` ignores `.env`/`.env.*` while keeping `.env.example`.
- `linkTo` neutralizes `javascript:`/`vbscript:`/`data:` hrefs.
- SSE CR field injection and service-worker caching of authenticated HTML (`no-store`) closed.

## [0.8.1] - 2026-09-16

### Changed

- Drive veil transition swaps the favicon for a system-color spinner (copper `#c9893a`, same as the progress bar): no icon lookup, no image request.

## [0.8.0] - 2026-09-16

### Added

- `amarra-cais g sitemap`: scaffolds blog posts (title/slug/body/published, public + paginate + seed) plus a dynamic `/sitemap.xml` (static `/` + `/posts` plus published posts with lastmod, XML-escaped, 1h cache). Reruns reuse; friendly error when posts lacks slug/published.
- Drive default transition: soft dark veil + the current favicon (`link[rel=icon]`, fallback `/static/icons/icon.png`) pulsing center-screen beside the progress bar. No layout change; honors `prefers-reduced-motion`.

### Changed

- `g resource` rows use an Edit button plus a trash icon opening a native "Delete this …?" modal (index and show). The Actions dropdown, confirm-via-link, and the show's unconfirmed Delete are gone; row confirms post via Drive link (no nested forms inside bulk-delete).
- Auth pages (login, signup, forgot, reset) center on screen (`min-h-[calc(100vh-12rem)]`, same as home).

## [0.7.0] - 2026-09-16

### Added

- Scaffold shell troca a nav horizontal por sidebar fixa à esquerda (`fixed w-60`, drawer no mobile via checkbox + `peer-checked`): nasce com Dashboard + logout + locale-toggle, `<!-- cais:nav -->` dentro para `g resource --public`; full, minimal e blank idênticos (a faixa vazia do `--minimal` deixa de existir).
- `amarra.js` avisa no console quando uma resposta Drive 200/422 não tem `#amarra-main` extraível (HTML desbalanceado), com status + URL (#83).

### Fixed

- `amarra-cais new` gera `package.json` com script `build` (Tailwind `input.css` → `web/static/css/styles.css --minify`) para o CSS chegar ao deploy (#82).

### Documentation

- `AGENTS.md` do scaffold: scripts inline re-executam a cada morph — usar IIFE (sem `let`/`const` global) e event delegation no `document` com flag `window` (#84).

## [0.6.1] - 2026-09-16

### Fixed

- `view.Write` defaults dynamic pages to `Cache-Control: no-store` so browsers never heuristically cache HTML + inline scripts (stale clicks after deploy). Opt out via `Page.CacheControl`, a preset `Cache-Control`, or a preset `ETag` (httpx 304 list flow keeps working) (#80).

## [0.6.0] - 2026-09-16

### Added

- Scaffold `AGENTS.md` documents the fullbleed page pattern: Drive only swaps `#amarra-main`, so page-owned chrome (landing header/footer, dashboard sidebar) must live inside it — next to the nav hook (#27), skip opt-out (#31), and second layout (#66) references (#76).

### Fixed

- `amarra-cais dev` restarts no longer leave `tmp/main` zombies holding `:8080` and `data/app.db`: the generated `.air.toml` sets `send_interrupt` + `kill_delay`, and the generated `cmd/server/main.go` shuts down gracefully via `signal.NotifyContext` + `RunContext` (#77).

## [0.5.2] - 2026-09-15

### Fixed

- `amarra-cais new app --minimal` (and `--blank`) produced an app that did not compile: `//go:embed migrations/*.sql` matched no file because the scaffold shipped only `.gitkeep`. Both scaffolds now ship `001_init.sql` with the `-- up` / `-- down` markers, and a compile smoke covers both variants (#74).

## [0.5.1] - 2026-09-15

### Fixed

- Kit attributes interpolate template actions and mixed text: `<.stat value="{{ .Power }} kWp" />` renders `5 kWp` instead of printing `{{ .Power }} kWp` to the user; a control action (`{{ if }}`) in an attribute value fails at boot naming the component and attribute (#72).

## [0.5.0] - 2026-09-15

### Added

- `r.NotFound(handler)` serves unmatched routes and the path-param parse failures (`IntParam`, `StringParam`, …) with the app's own page; the handler runs with the router middlewares and owns the status (#62).
- `amarra-cais g component <kit-name>` seeds the shipped kit markup, so an app restyles the real contract (attributes, error slot, hooks) and keeps the kit stem (`locale-toggle.html`); `amarra-cais g component --list` prints the 16 overridable components (#63).
- `amarra-cais doctor` warns while `web/static/icons/*` and `og.png` are still the scaffold placeholders (#64).
- The template loader contract is documented in `AGENTS.md` and in the scaffold README: nested pages are addressable as `blog/post`, partials and components are flat, and an unknown `<.x>` fails at boot (#65).

### Changed

- `writeView` takes the layout before the name — `writeView(w, r, views, cfg, layout, name, data, status)` — so an app with a second layout renders through the shared helper instead of bypassing it (#66).
- Scaffolds ship neutral placeholder brand assets instead of the framework marks, and the manifest splits `any` (192 + 512) from a dedicated padded `maskable` entry (#64).

## [0.4.0] - 2026-09-15

### Added

- `amarra-cais doctor --mobile` reads Google Fonts from `web/templates/**` as well as `input.css`, and passes when `CSP_STYLE_SRC` / `CSP_FONT_SRC` cover the referenced hosts (#57).

### Fixed

- `amarra-cais install` runs `npm install --include=dev`, so `NODE_ENV=production` no longer skips `tailwindcss`/`prettier`; a failed Tailwind build fails the command instead of exiting 0 on an unstyled app (#54).
- A fresh scaffold passes the Prettier job shipped in its own CI: `.prettierignore` skips `web/static/` (built CSS + vendored PWA assets) and the `AGENTS.md` / `README.md` templates are formatted (#55).
- `amarra-cais new` no longer writes a machine-local `replace` into `go.mod` when it finds a sibling Cais checkout (#56).

### Changed

- `amarra-cais new` links the local framework only through `CAIS_REPLACE`, and prints the "do not commit this replace" notice when it does; `amarra-cais link` still discovers a sibling checkout on purpose (#56).

## [0.3.0] - 2026-09-11

### Added

- `amarra-cais g resource` admin index uses bulk select-all, row `dropdown`, and a native `dialog` for bulk delete; `POST /admin/{plural}/bulk-delete` (#36).
- Kit `<.input type="file">` (`accept`, no `value`) and `<.form enctype="multipart/form-data">` (#43).
- Scaffold dashboard KPIs use kit `<.stat>` (#42).
- CI smoke compiles a generated resource with FK + `--public` + `--paginate` (#34).
- `amarra-cais doctor` FAILs on `hx-*` in templates and `gonertia` in `go.mod` (#40).

### Fixed

- Table sort links keep `q` and drop `page` via `sortHref`; generated admin tables pass `base` (#35).
- Kit `<.form>` slots no longer duplicate `csrfField` (login, contact, dashboard, stream chat) (#37).
- Empty resource indexes render `<.empty>` instead of a hollow table plus empty state (#38).
- `toSnake("Category")` is `category`; `BlogPost` is `blog_post` (#39).
- Theme hook writes `[data-amarra-theme-label]` and does not wipe SVG children with `textContent` (#41).
- Optional form `enctype` is bound as `$enctype` so `<.form>` inside `{{ range }}` does not 500 (#43).

## [0.2.2] - 2026-09-10

### Fixed

- Worker registers `PruneFinished` recurring before the first heartbeat, so a live worker is not visible while `recurring_tasks` is still empty (Linux CI flake).

## [0.2.1] - 2026-09-10

### Fixed

- Reference parent model lookup uses `category.go`, not `Category.go` — `toSnake("Category")` does not lowercase, so Linux CI 404'd the file and `List*Options` always queried `name`.

## [0.2.0] - 2026-09-10

### Added

- Kit `<.table>` with server-side sort headers (`?sort=` / `?dir=`, `aria-sort`) (#18).
- Kit `<.filters>` GET form (search + hidden sort/page) (#19).
- Kit `<.stat>` KPI card and `<.empty>` empty state (#20).
- Kit `<.password>` with eye toggle wired to `amarra-hook="password"` (#29).
- Hooks: `dialog` (native `<dialog>`), `dropdown`, `bulk` (select-all on the current page), `nav` (re-sync active link after Drive morph) (#21, #22, #23, #27).
- Theme hook: storage key, class, colors, and labels from `data-amarra-theme-*` (#30).
- Password hook: sibling input fallback, `aria-label` swap, `data-amarra-password-icon` (#28).
- `amarra-cais g resource` admin/public indexes use the kit (`<.filters>` / `<.table>` / `<.empty>` / `<.pagination>`) with search + sort (#24).

### Fixed

- `Router.Get("/")` is no longer a ServeMux catch-all; root is registered as `/{$}` so `/.env` and unknown paths 404 (#32).
- CSRF middleware parses multipart bodies via `ParseFormOrJSON` so classic file uploads are not 403 (#26).
- `ParseFormOrJSON` rewinds JSON bodies so CSRF and the handler can both read the payload.
- Kit `<.pagination>` joins a bare Base with `?page=` (not `&page=`).
- Reference `List*Options` SQL and parent seed literals use the parent `name` or `title` column, not both.
- Generated public `--paginate` List no longer emits an undefined `Total` when there is no int column.

### Changed

- Generators and kit no longer emit the no-op `data-amarra-drive="true"`; Drive is opt-out via `data-amarra-skip` (#31).
- Kit `<.form>` / locale-toggle read CSRF from the root (`$.CSRFToken`) so nested slots work.

## [0.1.0] - 2026-09-05

### Added

- Amarra Live WebSocket hub at `GET /amarra/live` (`Mount`/`Handle`/`Render`, CSRF join, MaxConns, idle ping, in-process `Broadcast`).
- JS auto-connect for `[amarra-live]`; `amarra-click` / `change` / `submit` inside the live root go over WS.
- `amarra-cais g live <name>` counter view; `amarra-cais g stream chat --live`.
- Stimulus-shaped `amarra-hook` registry (`connect` / `updated` / `disconnect` / `handleEvent`) plus clipboard, password reveal, color-scheme, and client reveal/toggle builtins (#5, #7).
- Kit `<.locale-toggle />` (Drive POST `/locale`) and configurable i18n cookie name (`SetCookieOpts`, `CatalogForRequestNamed`) (#6).
- Drive: title/CSRF head merge, progress bar, `data-amarra-confirm`, `data-amarra-method` / `_method`, `data-amarra-disable-with`, 422 `aria-invalid` focus, scroll restore.
- Frame: `data-amarra-frame` targeting from links, `loading="lazy"`.
- Stream: `stream.WriteHTTP` (`text/vnd.amarra-stream`) and `before` / `after` ops.
- Live socket `Patch` / `Navigate` / `Stream` / `Push`; JS `amarra-debounce`, `amarra-click-loading`.
- Kit: `<.form>` injects CSRF; `<.select>` / `<.textarea>` / `<.checkbox>`; `linkTo` dict opts (`method`, `confirm`, `frame`).
- `docs/migrate-inertia.md` — Cais Inertia → Amarra rewrite guide; `doctor` vite FAIL points at it (#8).

## [0.0.3] - 2026-09-05

### Removed

- Generated `input.css` no longer ships dead `.htmx-*` / indigo auth-screen utilities.
- `/logs` no longer loads `htmx.min.js`; it polls with `fetch("/logs?partial=1")`.
- Dev banner no longer mentions Vite.
- `g resource` / `g stream chat` nav links no longer emit `use:inertia` or indigo classes.

### Changed

- Auth, contact, dashboard, kit button/input/pagination, and PWA theme use ink/foam/copper instead of the old indigo cards.

## [0.0.2] - 2026-09-05

### Changed

- Harbor greeting on `amarra-cais new`: ink/foam/copper shell, “You made landfall.” / “Você atracou.”, Come aboard CTA, and a manifest ticket with the first generator command.
- Layout header links to `/login`; home/contact/dashboard set `ActiveNav`.

## [0.0.1] - 2026-09-05

First public cut of **amarra-cais**: HTML-first fork of Cais. Generated apps use Amarra Views + Drive (no Inertia/Svelte/Vite). Live is a 501 stub.

## [0.11.0] - 2026-08-28

### Added

- `GET /jobs` queue dashboard (`pkg/cais/jobsui`) — localhost only, all envs; retry/discard failed jobs; job detail; kind filter; prune finished; worker heartbeats (#185).
- Jobs inspect APIs: `List` (status/queue/kind), `Get`, `RetryFailed`, `Discard`, `ListScheduled`, `CountByQueue`, `PruneFinished`, `TouchWorker` / `ListLiveWorkers`, `RequeueOrphaned` (#185).
- Worker heartbeat (`job_workers`) so `/jobs` shows live processes; `RequeueOrphaned` skips in-flight jobs of live workers (#185).
- Built-in `PruneFinished` job (daily 04:00 UTC when a worker runs) plus `cais jobs prune [--older 24h]` (#185).
- `cais jobs retry|discard <id>`; `cais jobs status` prints workers (#185).
- `cais routes` lists `/jobs` when `jobsui.Register` is in `app.go` (#185).
- Boot banner lists `http://127.0.0.1:<port>/jobs` (#185).
- `cais doctor` checks `jobsui.Register` in `app.go` (#185).
- Dashboard warns when two workers share one SQLite file (#185).

## [0.10.0] - 2026-08-25

### Fixed

- `cais g auth` compiles on minimal/blank apps: routes patched via the AST helper, `errors` import ordered correctly, dev demo user seeded so generated auth tests pass (#166).
- SSE `WriteEvent` splits multi-line payloads into `data:` lines — newlines in chat HTML no longer truncate events or let user text forge extra SSE events (#167).
- `cais destroy` unpatches via `go/ast`: exact statement/name removal only; user comments, custom `/admin/<plural>` routes and lookalike methods survive (#168).
- `ClientIP` walks `X-Forwarded-For` right-to-left; leftmost entries were spoofable to evade rate limits (#170).
- `mail.SMTPSender.Send` rejects CRLF header injection and invalid recipients via `net/mail.ParseAddress` (#171).
- Jobs stranded in `running` by a crashed worker are requeued on worker boot (`Store.RequeueStuck`) (#172).
- CSRF cookie uses the `__Host-cais_csrf` prefix when secure (production); generated Inertia apps pick the matching `xsrfCookieName` per Vite `PROD` (#173).
- Low-severity batch: atomic devlog default buffer, mutex-guarded jobs `Registry`, recurring enqueue + `last_run` in one transaction, cache sweep past 1024 keys, rune-safe `chat.Truncate`, `Recover` logs stack traces, destroy/stream patches propagate I/O errors (#174).

### Added

- `cais destroy resource|model|handler` warns before deleting files modified since generation. Generators record SHA-256 hashes in `.cais-generated.json`; `--force` overrides (#169).

### Changed

- `middleware.Flash` takes a config (`Flash(cfg)`) so flash deletion cookies mirror the production `Secure` flag (#174).

## [0.9.0] - 2026-08-25

### Added

- `pkg/cais/i18n` — per-request catalog: `NormalizeLocale`, `CatalogForRequest` (`?lang=` then `cais_locale` cookie then fallback), `LocaleMiddleware`, `CatalogFromRequest`, `SetCookie`. Accept-Language stays out of scope (#160).
- Locale tags `es` and `zh` in `normalizeLocale`.
- `cais.Load()` reads `.env` when present (`pkg/cais/dotenv`); does not override process env (tests, systemd, CI). Doctor shares the same parser (#153).
- `cais doctor` warns (fails when `CI=true` / `GITHUB_ACTIONS=true`) if `go.mod` still has a local `cais link` replace; `cais link` says not to commit it (#154).
- `cais new` ships `AGENTS.md` with the app (#150).
- Auth pages use shared `AuthLayout` (centered) by default (#151).
- Password inputs default to show/hide eye (#137).
- `cais routes` / router panic hint when two ServeMux patterns collide (#145).

### Changed

- Scaffold handlers are **Inertia-only** (no HTMX HTML fallbacks in home/contact/auth/dashboard templates) (#138).
- `cais g auth` writes Svelte pages (`Login`/`Signup`/…) instead of `web/templates/pages/*.html`.
- `cais destroy auth` removes Svelte auth pages (and leftover HTMX login HTML if present).
- Framework repo is CLI + `pkg/cais` only — smoke production boots a scaffolded app via `cais new` (#138).
- Flash on Inertia redirects uses `flash.Set` cookies, not `inertia.SetFlash` (#143).
- Scaffold `go.mod` default pin when CLI is `dev`: `v0.9.0`.

### Fixed

- `cais doctor` fails (not a warning) when an Inertia app has no `web/static/build/assets/main.js` — scaffold `.gitkeep` is not a bundle (#159).
- `cais build` / Vite step fails if `npm run build` did not emit `web/static/build/assets/main.js` (#159).
- `cais new --help` / `-h` print usage and do not create a directory; unknown flags error instead of becoming the app name (#158).
- Fresh `cais new` apps pass their own CI: goimports local-prefixes, unused `bootstrap()`, console `log.Fatal` after defer, Prettier on Vite JS configs (#157, #152).
- `cais doctor` / install / server detect unbuilt `styles.css` and auto-build on install/server (#144).

### Removed

- **Dogfood app** from this repository (`cmd/server`, `internal/app|handlers|store|models|db`, `web/` SPA) (#138).
- Root Vite/Svelte/Tailwind dogfood tooling (apps bring their own via `cais new`).

## [0.8.1] - 2026-07-27

### Added

- `cais doctor` — **cais CLI version** check: warn when installed binary is older than `go.mod` or predates Vite watch (`≥ 0.8.0`)
- `cais dev` — warn on CLI/module version mismatch; explicit SPA-not-watched message if Vite watch did not start
- `pkg/cais/pwa.SyncServiceWorker` — migrate existing apps to network-first `/static/build/` + `/static/css/`
- `cais pwa` — always re-syncs `sw.js` (Inertia apps use `InstallForInertia`); prints migration message
- Docs: Svelte 5 + `useForm` reactive-write footgun (local state + assign on submit)
- README: reinstall CLI after tag (`go install …@v0.8.1`)

### Changed

- Dev banner lists Vite (`build --watch`) alongside air/Tailwind so stale CLIs are obvious
- Scaffold `go.mod` default pin when CLI is `dev`: `v0.8.1`

### Fixed

- Packaging gap after #128: vite watch + network-first SW were on `main` only until this tag (#132)
- Existing apps stuck on cache-first SW can upgrade with `cais pwa` without hand-editing (#135)

## [0.8.0] - 2026-07-12

### Added

- `internal/cli/frontend.go` — detect Vite apps; `cais build` runs `npm run build` when present
- `cais dev` and `make dev` — initial Vite build plus `vite build --watch` for Svelte pages
- `cais g resource` on Inertia scaffolds — Svelte admin CRUD (`Admin*.svelte`), Inertia handlers, and public list pages instead of HTMX templates
- `web/src/components/AppLayout.svelte` — shared nav, flash, and slot for Inertia pages
- `cais new` scaffolds `AppLayout` with `<!-- cais:nav -->` marker for public resource link patching
- Home handler passes translated `labels` props to the Svelte home page (i18n keys + formatted subtitle)
- Pre-commit hook runs `npm run test:fe` on `*.{js,mjs,svelte}` changes

### Changed

- `AGENTS.md` — Inertia + Svelte conventions, Vite build path, expanded CLI generator layout table
- `home.stack` locale — `Go · Inertia · Svelte · SQLite` (was HTMX/Tailwind)
- Public nav patching prefers `AppLayout.svelte` over `Home.svelte` on Inertia apps
- Dogfood app removes legacy HTMX page templates (`web/templates/pages/*.html`)

## [0.7.0] - 2026-07-05

### Added

- **Inertia + Svelte default frontend** — `cais new` scaffolds gonertia, Vite, and Svelte 5 pages (`Home`, `Contact`, `Login`, `Dashboard`, auth flows)
- `cais g handler` / `cais g page` — generate Svelte pages in `web/src/pages/`
- Inertia integration tests (`X-Inertia: true`, validation 422, redirect 303) in `internal/app/app_test.go`
- Vitest + Testing Library for Svelte pages (`npm run test:fe`)
- CI builds Vite assets before `go test`
- `pkg/cais/cache`: `Key(parts ...any)` and `Hash(v any)` — stable key building and short content hashing. Helps avoid embedding full lists (e.g. 80 sessions) in cache keys for list pages.
- `pkg/cais/httpx`: `NotModified(w, r, etag)` and `SetETag(w, etag)` — simple ETag / 304 conditional response support for cacheable pages and lists.

### Added

- `pkg/cais/chat`: `Truncate`, `SafeMessageBubble`, `TrimForDisplay`, `MaxMessageChars` — server-side safety and perf for large/polluted agent histories (addresses loading 1-2s, 500s on huge turns)
- `cais g stream chat` scaffold now demonstrates TrimForDisplay + SafeMessageBubble in Show/ListMessages/Stream
- `pkg/cais/chat`: `UnsafeLiveHTML` + `UnsafeMessageHTML` + `WriteUnsafe*` helpers — enables first-class streaming agent UIs with rich pre-rendered content (Markdown, media) in #chat-live and #chat-stream without duplicating bubble wrappers in the app.
- `pkg/cais/sqlite` package docs — WAL / busy_timeout guidance for SSE chat apps
- README — `testutil` chat assertion examples
- `cais.js`: remove optimistic user bubbles on SSE error/close + when assistant streaming starts; `window.caisRemoveOptimisticUserBubble` (better rollback during streaming for #86)
- `pkg/cais/chat`: `DetailBubbleWithTitle`, `ToolCallBubble`, `ToolResultBubble` — basic primitives for tool-calling, permissions flows and distinguishing tool output (#87)
- `pkg/cais/chat`: `SelectWindowWithLastUser` — robust history window with pinned last user (for #85)

### Added

- `testutil.AssertHTMLContains` / `testutil.AssertChatMarkers` — chat handler HTML assertions
- `cais dev` auto-bumps PWA `CACHE_VERSION` when `sw.js` is present
- `cais doctor --mobile` — chat enter-submit JS (`bindChatEnterSubmit`) and chat form CSS checks
- `cais g stream chat` handler tests — Show 404, PostMessage user bubble, `AssertChatMarkers`

### Added

- `cais.js` `bindChatEnterSubmit` — delegated Enter-to-send on `form[data-cais-chat-form]` (Shift+Enter newline)
- `cais.js` `dedupOptimisticUserBubble` — drops optimistic user bubble when server partial already includes it
- Chat form submit CSS — `inline-flex` button + scoped `htmx-indicator` / `htmx-request-hide` swap

### Changed

- `hxChatForm` no longer uses inline `hx-on:keydown` or `this.reset()` — input clear and Enter handled in `cais.js`
- `cais g stream chat` submit button uses `htmx-request-hide` + `htmx-indicator` pattern
- `cais g stream chat` uses mobile `cais-chat-shell` layout (sticky footer, viewport height)
- `#chat-messages` includes `overflow-x-hidden` — `cais doctor --mobile` warns when missing
- `finalizeChatStream` runs after `#chat-history` swaps; `pruneEmptyChatNodes` removes empty SSE slots

### Added

- `chat.DetailBubble` — collapsible tool/log output for agent streams

### Added

- `hxChatForm` calls `window.caisFinalizeChatStream` before submit to merge SSE stream slots
- `cais doctor --mobile` — chat agent JS finalize check and `#chat-messages` scroll container check
- `cais g stream chat` demo uses `pkg/cais/chat` — `event: stream` typing preview + timestamped `event: message`
- Scaffold `input.css` — generic chat styles (`.cais-chat-scroll-down`, `.cais-msg-time`, `.cais-thinking-dots`)
- `cais.js` agent chat module — `finalizeChatStream`, device-local timestamps, stick-to-bottom scroll, poll guard (opt-in `data-cais-chat`)
- `chat_sse_agent.html` partial — multi-slot agent chat (`#chat-history` + `#chat-stream` + `#chat-live`, `data-cais-chat`)
- `pkg/cais/chat` — generic SSE chat HTML helpers (`LiveBubble`, `MessageBubble`, `ThinkingHTML`, `WriteStream`, `WriteMessage`)
- `pkg/cais/stream` — `Flush`, `RelaySSE`, and `RelayAndCopy` for HTMX SSE through middleware-wrapped `ResponseWriter`s
- Logger skips misleading `Completed` log line for `/stream` and `/event` paths
- `cais.js` reconnects SSE after `hx-boost` swaps (`data-cais-sse-persist`, `htmx:sseClose` handler)
- `hxChatForm` template helper — Enter-to-send chat forms with thinking indicator
- `chat_sse.html` partial — `#chat-thinking` indicator and optional `data-cais-poll-url` fallback
- `cais g stream chat` — conversations + messages migration, SSE handler, HTMX chat UI
- `netutil.HealthPayload` — `/health` exposes `lan_urls` for mobile testing
- `cais doctor --mobile` — chat SSE pattern, SSE reconnect JS, health `lan_urls` checks
- `{{ flashMessage .Flash }}` template helper in `pkg/cais/forms`
- `cais pwa [--bump]` — refresh PWA assets; `--bump` increments `CACHE_VERSION` in `sw.js`
- `cais doctor --mobile` — flash template, Google Fonts CSP, and PWA cache version checks
- `boot.Print` LAN URL line via `pkg/cais/netutil`
- `cais.PortBusy` and dev-server warning when the configured port is already in use
- `cais.StringParams` — ergonomic two-param routes without nested `StringParam` callbacks
- `cais.IntStringParams` — int + string path params in one wrapper
- `cais link [path] [--unlink]` — go.mod `replace` for local framework development
- Scaffold partial `chat_sse.html` — append-only SSE chat pattern (`#chat-history` + `#chat-sse`)
- `cais doctor` warns when `sse-ext.min.js` is installed but `WriteTimeout > 0`

### Changed

- Scaffold and reference app default `WriteTimeout: 0` so long-lived SSE connections are not killed at 30s
- Scaffold uses system font stack (no Google Fonts `@import`) to avoid CSP console errors
- Scaffold layouts use `{{ flashMessage .Flash }}` instead of struct stringification

## [0.6.0] - 2026-07-04

### Added

- `pkg/cais/barcode` — Open Food Facts lookup client
- `pkg/cais/money` — `FormatBRL` for cent-based prices
- `middleware.LoadUserStats` / `UserStatsFrom` — gamification chrome in layouts
- `meta.Site.LoggedIn` — session flag for layout auth chrome
- `Config` security knobs: `PERMISSIONS_POLICY`, `CSP_MEDIA_SRC`, `CSP_CONNECT_SRC` (camera + barcode scan in PWA)
- `Router.StaticForEnv` — `no-store` for static assets in development
- `NewRendererForEnv` — disk template reload in development
- Scaffold partials: `icons.html`, `nav_links.html` on `cais new`

### Removed

- `cais g app supermarket` and `internal/cli/app_templates/supermarket/` — app UI belongs in apps, not the framework
- `pkg/cais/ui` — nav/icon HTML helpers (use app templates instead)

## [0.5.0] - 2026-07-03

### Added

#### HTMX UX (app shell)

- `pkg/cais/ui` — `navTab`, `makeNavTab`, `icon` helpers; `Site.ActiveNav` for tab highlighting
- `pkg/cais/htmxattrs` — `hxForm`, `hxDelete`, `hxBoostLink`, `hxPaginate`, `hxMorphOuter`
- Idiomorph extension bundled for `hx-swap="morph"` (`hx-ext="morph"` on layout body)
- Supermarket-style scaffold layout with `#cais-nav`, `#cais-main`, and hx-boost navigation
- Resource generator HTMX admin CRUD (inline delete, form partials, `RenderPageOrPartial` on 422)
- HTMX pagination with morph swap for admin and public lists (`--paginate`)
- `float` / `float?` field type in `cais g resource`
- `cais.SetToast`, `SetFocus`, `SetRetarget`, `SetTrigger` response header helpers

#### Security & sessions

- Session expiry in SQLite (`expires_at`, 7-day TTL, reject expired `Get`)
- `session.PruneExpired()` and `cais db prune-sessions`
- Secure cookies in production (`Config.CookieSecure()`, `session.CookieOptionsFromConfig`)
- Security headers middleware (`middleware.SecurityHeaders`) — CSP, HSTS, X-Frame-Options, Referrer-Policy
- Per-IP rate limiting (`middleware.NewRateLimiter`) on login and contact POST routes
- Trusted proxy support (`TRUSTED_PROXIES`, `middleware.ClientIP`)
- Production error sanitization (`Config.SanitizeErrors()`, `httpx.RenderOrError`)
- CSRF protection (`middleware.CSRF`, `meta.WithCSRF`, double-submit cookie)
- Session auth scaffold (`cais g auth`, login/logout, protected dashboard)
- Flash messages (`pkg/cais/flash`, `middleware.Flash`, one-shot redirect feedback)

#### HTMX & UI

- `httpx.RenderPageOrPartial` for HTMX-aware form responses
- `cais.js` — CSRF header injection, focus restore, optimistic toggles
- HTMX swap/loading CSS utilities (`.htmx-swapping`, `.htmx-settling`, `.htmx-request-hide`)
- `cais.SetTrigger` and `cais.SetRetarget` response helpers
- Optimistic bool toggles in generated resource admin (`data-cais-optimistic="toggle"`)
- Contact form HTMX polish (loading indicator, swap transitions)

#### Forms & validation

- `validate.FieldErrors` map with `Add`, `Has`, `First`, `Any`
- `validate.MinLength` and `validate.MaxLength`
- `pkg/cais/forms` — `csrfField`, `fieldError`, `makeField`, `fieldInput` template helpers
- `forms.FieldData` for labeled inputs, textareas, and checkboxes with inline errors
- `fieldSelect`, `makeSelectField`, `makeSelectFieldPtr` for foreign-key dropdowns
- Resource generator `category_id:references` and `category:belongs_to` field types (FK migration, `ListCategoryOptions`, admin select)

#### Router & render

- `Router.Put` and `Router.Patch` methods
- Partials parsed into the page template tree (`{{ template "name" . }}` works in pages and layouts)
- `pkg/cais/pagination` — offset/limit helpers for list pages
- `pkg/cais/cache` — in-memory TTL cache (`New`, `Get`, `Set`, `Delete`)

#### i18n

- `pkg/cais/i18n` — key-based locale catalogs (`LOCALE=en` default, `LOCALE=pt` supported)
- Template funcs `t`, `htmlLang`, `ogLocale` registered on the renderer
- i18n wired through scaffolds, handlers, and `meta` OG locale defaults

#### Background jobs

- `pkg/cais/jobs` — SQLite-backed queue (enqueue, delay, worker, dispatcher)
- Recurring cron scheduler (`recurring_tasks`, `jobs.RunScheduler`)
- `cais jobs work [--queues ...] [--concurrency N]` and `cais jobs status`
- `cais g job <name> [--cron "0 3 * * *"]` — scaffolds handler, registry, and `cmd/worker`
- Built-in `PruneSessions` job handler
- Jobs migration (`003_jobs.sql`, `004_recurring_tasks.sql`)

#### CLI — generators

- `cais g resource` defaults to session auth (`--admin-auth session|bearer`)
- `cais g resource --paginate` — admin index pagination (25/page, HTMX controls)
- `cais g resource --force` — overwrite existing generated files
- Resource admin show page (handler, template, route, tests)
- `cais g model` — model struct + migration + store methods (no handlers/UI)
- `cais g --dry-run` on all generators (resource, model, handler, page, migration, auth, console, ci, job)
- `cais destroy [--dry-run]` — resource, handler, model, auth, migration (unpatch routes, store, seeds, nav)
- `cais g ci` — add GitHub Actions, pre-commit, golangci-lint, Prettier to existing apps
- `cais new --module <path>` — override Go module path
- Nav marker `<!-- cais:nav -->` for reliable public link patching
- AST-based route patching (`internal/cli/patch`) via `insertBeforeFunctionEnd`
- `nextMigrationFile` — sequential migration numbering across generators
- Migration `-- up` / `-- down` sections in generated SQL files
- Welcome screen and i18n catalogs in `cais new` scaffolds

#### CLI — database & tooling

- Versioned migrations (`pkg/cais/migrate`, `cais db migrate`, `cais db status`)
- `cais db rollback` — roll back last migration (runs `-- down` SQL when present)
- `cais db seed` and `cais db seed --list`
- `cais routes` and `cais routes --verbose` (handler names + middleware)
- `cais version` — print framework version
- `cais doctor` — production readiness checks (`ADMIN_TOKEN`, `APP_URL`, CI tooling hints)
- Smoke scaffold test (`scripts/smoke-scaffold.sh`, `generate_smoke_test.go`)
- App-level integration tests: login → dashboard (flash) → logout; contact validation with CSRF (`internal/app/app_test.go`)

#### Config & health

- `validate.Email`, `validate.URL`, `validate.Required`
- `APP_URL` required in production (`cfg.Validate()`)
- DB-aware `/health` endpoint (503 `degraded` when SQLite is down)
- SQLite production defaults (`PRAGMA foreign_keys=ON`, WAL)

### Changed

- Admin auth requires `ADMIN_TOKEN` in production; Bearer header only (no query params)
- Generated resource admin routes use `RequireAuth` by default instead of `AdminAuth`
- Reference app aligned with `cais new` output (routes.go, auth, dashboard, contact validation)
- Blank app scaffold includes `Recover`, `SecurityHeaders`, server timeouts, and `/health`
- Contact scaffold uses `validate.FieldErrors` with name validation
- Auth migration template includes `expires_at` column (7-day default)
- `pwa.FS()` returns `(fs.FS, error)` instead of panicking on failure
- README, AGENTS.md, and `.env.example` synced with all new commands and env vars
- Jobs documentation in README and AGENTS (`cais g job`, `cais jobs work/status`, deploy notes)
- CSP `'unsafe-inline'` tradeoff documented (required for HTMX and inline service-worker script)

### Security

- Removed admin token via query string
- Constant-time admin token comparison
- Rate limiter bucket cleanup for stale entries
- Flash cookies use `HttpOnly` and `Secure` in production

### Fixed

- `cais g auth` migration includes session expiry column on fresh installs
- Local `CAIS_REPLACE` resolves from cwd for remote app directories
- Generated `dashboard_test` scaffold includes `cais` import
- AST route patch preserves `cais.IntParam(...)` lines through gofmt
- CAIS startup banner renders block Unicode logo clearly (no longer reads as "COTS")

### Deprecated

- `middleware.TokenAuth` — use `AdminAuth(cfg)` instead; scheduled for removal in a future release

## [0.4.7] - 2026-07-01

### Added

- Default OG preview and fullscreen PWA (`pkg/cais/meta`, `pkg/cais/pwa`)
- Go on Cais hero image and harbor PNG icon in scaffold assets

### Fixed

- Startup banner redrawn to spell CAIS clearly with block Unicode art

## [0.4.0] - 2026-07-01

### Added

- Interactive console REPL (`cais console`, `pkg/cais/console`) with SQL, history, reload, and typed bindings
- `/logs` development log viewer (`pkg/cais/devlog`, HTMX auto-refresh, localhost only)
- Cais-branded air startup banner (`pkg/cais/boot`)
