# Amarra-cais

![Go on Cais](pkg/cais/pwa/assets/go-on-cais.jpg)

Full-stack Go framework for mini apps (Lightsail-friendly): **Amarra Views + Drive**, Tailwind, and SQLite — with a Rails-style CLI.

This repository is the **framework + CLI** only. Generate apps with `amarra-cais new`. The CLI binary is `amarra-cais` (it does not overwrite `cais`). Cais v0.11.x remains the Inertia + Svelte product.

## Stack

| Layer    | Choice                                                                |
| -------- | --------------------------------------------------------------------- |
| Language | Go 1.26 (`net/http` stdlib; see `go.mod`)                             |
| Frontend | **Amarra Views + Drive** (`pkg/amarra/view` + `/static/js/amarra.js`) |
| CSS      | Tailwind CSS 3.x                                                      |
| DB       | SQLite (`modernc.org/sqlite`, no CGO)                                 |
| PWA      | Manifest, service worker, offline page, icons, fullscreen             |
| Meta     | Open Graph / Twitter via `pkg/cais/meta`                              |
| Core     | Router, session, CSRF, jobs, SQLite in `pkg/cais/`                    |

The browser does not mount a SPA. Handlers call `view.Write`. Drive morphs `#amarra-main`. There is no Vite, Svelte, or Inertia in generated apps. Migrating a Cais Inertia app is a UI rewrite — see [docs/migrate-inertia.md](docs/migrate-inertia.md).

Repeating UI is the shipped **kit + hooks**, not Alpine/Stimulus/HTMX: `<.table>`, `<.filters>`, `<.stat>`, `<.empty>`, `<.password>`, plus `amarra-hook` builtins `dialog`, `dropdown`, `bulk`, `nav`, `theme`, `password`. Sort and filter are `GET ?q=&sort=` (server re-render). `amarra-cais g resource` emits the kit.

## Quick start

```bash
export PATH="$HOME/go/bin:$PATH"
go install github.com/puppe1990/amarra-cais/cmd/amarra-cais@v0.2.1   # or: make install-cli from this repo
amarra-cais version   # expect 0.2.1
amarra-cais new myapp
cd myapp && amarra-cais install && amarra-cais dev   # http://localhost:8080
```

**Developing the framework itself:**

```bash
export PATH="$HOME/go/bin:$PATH"
make install-cli
make test                 # go test ./... -race
make js-test              # pkg/cais/js + pkg/amarra/js
make ci                   # test + js-test + lint + format-check
# Local CLI against this checkout:
amarra-cais link .        # from an app dir, or set CAIS_REPLACE; unlink before push
```

Demo login in a fresh scaffold (dev seed): `demo@example.com` / `password`.

## CLI (Rails-style)

```bash
make install-cli
export PATH="$HOME/go/bin:$PATH"
```

| Command                                                                                                                | Description                                               |
| ---------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------- |
| `amarra-cais new <app> [dir] [--minimal\|--blank] [--module path]`                                                     | Scaffold app (HTML + Amarra Drive by default)             |
| `amarra-cais g [--dry-run] handler\|page\|resource\|model\|migration\|auth\|console\|ci\|job\|stream\|live\|component` | Generators                                                |
| `amarra-cais destroy [--dry-run] resource\|handler\|model\|auth\|migration\|component`                                 | Undo generators                                           |
| `amarra-cais install`                                                                                                  | `npm install` + `go mod tidy` (+ Tailwind build)          |
| `amarra-cais dev`                                                                                                      | **air + Tailwind watch** (HTML templates reload with air) |
| `amarra-cais css` / `amarra-cais build` / `amarra-cais server` / `amarra-cais test`                                    | CSS, binary, run, tests                                   |
| `amarra-cais console`                                                                                                  | REPL (store, cfg, db + SQL)                               |
| `amarra-cais routes [--verbose]`                                                                                       | List routes from `internal/app/routes.go`                 |
| `amarra-cais db migrate\|status\|rollback\|prune-sessions\|seed`                                                       | Migrations & seeds                                        |
| `amarra-cais jobs work\|status\|retry\|discard\|prune`                                                                 | SQLite background jobs + `/jobs` dashboard                |
| `amarra-cais doctor [--mobile]`                                                                                        | Verify amarra.js, `#amarra-main`, PWA, mobile             |
| `amarra-cais pwa [--bump]`                                                                                             | Write/refresh PWA assets; `--bump` cache                  |
| `amarra-cais link [path] [--unlink]`                                                                                   | Local `go.mod replace` for framework dev                  |
| `amarra-cais version`                                                                                                  | Framework version                                         |

Field types for generators: `string`, `text`, `url`, `bool`, `int`, `date`, `references` (or `name:belongs_to`). Suffix `?` for optional.

```bash
amarra-cais g resource bookmark --fields title:string,url:url,notes:text? --public --paginate
amarra-cais g handler settings   # Go handler + test + web/templates/pages/settings.html
amarra-cais g component card     # web/templates/components/card.html
```

## Development experience (in generated apps)

- **Port auto-pick** if `:8080` is busy
- **Boot banner** with LAN URLs for phone testing on Wi‑Fi
- **Logs** — JSON request (`kind: request`) + SQL (`kind: sql`); `LOG_FORMAT=text` for plain text
- **`/logs`** — localhost-only log viewer in development
- **`/jobs`** — localhost-only queue dashboard (counts, failed retry/discard, recurring)
- **Frontend** — server-rendered HTML; Drive morphs `#amarra-main` (no Vite)
- **PWA** — SW is **network-first** for `/static/js/amarra.js` and `/static/css/`; `amarra-cais pwa --bump` after HTML changes on phones

## Structure

```
pkg/cais/              framework packages (router, httpx, session, jobs, pwa, …)
pkg/amarra/            Views, Drive, Frame, Stream, Live hub, amarra.js sources
internal/cli/          amarra-cais CLI + scaffold templates (split by domain)
cmd/amarra-cais/       CLI entry point
cmd/pwagen/            helper to write PWA assets into a directory
scripts/               smoke-scaffold + smoke-production (via amarra-cais new)
```

Scaffolded apps get `cmd/server`, `internal/app`, `internal/handlers`, `web/templates`, `web/static/js/amarra.js` — not this repo.

## Amarra Views + Drive (generated apps)

Handlers render HTML via `view.Write`:

```go
view.Write(w, r, h.views, view.Page{
  Layout: "app",
  Name:   "login",
  Data: map[string]any{
    "Title":     "Login",
    "Site":      meta.ForRequest(h.site, r),
    "CSRFToken": csrf.TokenFromRequest(r),
  },
}, h.cfg)
// Validation: same page, status 422, `.Errors` on inputs
// Flash on redirect: flash.Set(w, kind, msg, secure) + http.Redirect(..., 303)
```

Pages live in `web/templates/pages/*.html` and define a `content` block. Kit tags expand at boot:

```html
{{ define "content" }}
<h1>{{ .Title }}</h1>
<.form action="/contact" method="post">
  {{ csrfField .CSRFToken }}
  <.input name="email" type="email" label="Email" value="{{ .Email }}" error="{{ fieldError .Errors "email" }}" />
  <.button type="submit">Send</.button>
</.form>
{{ end }}
```

**Drive** — intercepts all same-origin clicks/submits by default (plain `<a>`, `<form>`, `{{ linkTo }}`, `<.form>`) and turns them into `fetch` with `Amarra-Drive: true` + CSRF, then morphs `#amarra-main`. First load, curl, and crawlers get the full layout. Opt out with `data-amarra-skip`.

**JSON bodies** — if a client posts JSON, handlers should use:

```go
if err := httpx.ParseFormOrJSON(r); err != nil { /* ... */ }
email := r.FormValue("email")
```

## Framework APIs (highlights)

**Router**

```go
r.Get("/blog/{slug}", cais.StringParam("slug", blog.Show))
r.Group(middleware.RequireAuth("/login"), func(g *cais.Router) {
  g.Get("/dashboard", dashboard.ServeHTTP)
})
```

**httpx** — `RenderOrError`, `WritePage`, `SeeOther`, `ParseFormOrJSON`, `FormTruthy`, ETag helpers.

**Sessions** — cookie auth (7-day TTL), `session.SignIn` / `SignOut`, `amarra-cais db prune-sessions`.

**CSRF** — double-submit cookie `cais_csrf` + form field / `X-CSRF-Token`.

**Jobs** — SQLite queue, no Redis:

```bash
amarra-cais g job send_welcome --cron "0 3 * * *"
amarra-cais jobs work --concurrency 2
```

## Framework commands

```bash
make test           # go test ./... -race
make test-v         # verbose
make js-test        # pkg/cais/js + pkg/amarra/js unit tests
make lint           # golangci-lint
make format         # prettier --write
make ci             # test + js-test + lint + format-check
make build          # bin/amarra-cais
make install-cli    # go install ./cmd/amarra-cais
```

CI runs Go tests, JS unit tests, lint, Prettier, and smoke (`amarra-cais new` + production boot of a scaffolded app).

## Production deploy (generated apps)

```bash
amarra-cais css     # Tailwind → web/static/css/styles.css
amarra-cais build --os linux --arch amd64 -o bin/server-linux
tar czf release.tar.gz bin/server-linux web/static
```

Ship `web/static` (CSS, `js/amarra.js`, PWA) beside the binary. There is no Vite `web/static/build/` step.

- Guide: `docs/deploy/lightsail-systemd.md`
- Template: `deploy/systemd/cais-app.service.example`

## License

See [LICENSE](LICENSE).
