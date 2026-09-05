# Migrate a Cais Inertia app to Amarra

`amarra-cais doctor` fails on `vite.config.js` because this is not a dual-mode fork. Cais v0.11 stays Inertia + Svelte. Amarra is HTML + Drive. The move is a UI rewrite, not `go get`.

## 1. Pin the framework and drop gonertia

```bash
go get github.com/puppe1990/amarra-cais@latest   # pin a release tag when you can
go get github.com/puppe1990/cais@none            # if the module path still points at Cais
```

Remove `github.com/hotwire-go/gonertia` (or `inertia-go`) from `go.mod`. Delete `vite.config.js`, `svelte.config.js`, `web/src/`, and `web/static/build/`. The CLI binary is `amarra-cais`; it does not overwrite `cais`.

Until a tagged release includes Live + Drive polish, `go get github.com/puppe1990/amarra-cais@main` or `amarra-cais link` against a checkout.

## 2. Load views once at boot

```go
views, err := view.Load(tmplFS, catalog) // once in bootstrap, not per request
```

Handlers take `*view.Renderer` and call `view.Write`. Do not check `HX-Request`. Do not call `inertia.Render` / `inertia.Redirect` / `inertia.Flash`.

```go
view.Write(w, r, h.views, view.Page{
  Layout: "app",
  Name:   "contact",
  Data:   amarraData(r, h.site, map[string]any{"Title": h.catalog.T("contact.title")}),
}, h.cfg)
```

## 3. Pages are HTML

| Cais Inertia             | Amarra                                   |
| ------------------------ | ---------------------------------------- |
| `web/src/pages/*.svelte` | `web/templates/pages/*.html`             |
| `AppLayout.svelte`       | `web/templates/layouts/app.html`         |
| Vite / `#app`            | `#amarra-main` + `/static/js/amarra.js`  |
| Svelte forms             | kit `<.form>` / `<.input>` / `<.button>` |

Layout loads one script: `/static/js/amarra.js`. Put `<.flash />` **inside** `#amarra-main`.

## 4. Tests assert HTML, not Inertia JSON

Drop `setupTestInertia`, `X-Inertia: true`, and JSON component names.

```go
rr := httptest.NewRecorder()
h.Get(rr, httptest.NewRequest(http.MethodGet, "/contact", nil))
testutil.AssertHTMLContains(t, rr.Body.String(), `id="amarra-main"`, `name="email"`)
```

Validation stays **422 HTML** with `.Errors` on inputs (`fieldError .Errors "email"`), not a 422 Inertia payload.

## 5. ParseFormOrJSON still works

Drive POSTs `application/x-www-form-urlencoded` or multipart. Keep `httpx.ParseFormOrJSON`. You do not need JSON body parsing for page forms.

## 6. Flash is the cookie API

```go
r.Use(middleware.Flash(cfg)) // not Flash() without cfg
flash.Set(w, "notice", "Saved!", cfg.CookieSecure())
http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
```

Read with `flash.MessageFromRequest` (scaffold `amarraData` copies `s.Flash`). Layout: `<.flash />`.

## 7. PWA

```bash
amarra-cais pwa          # pwa.InstallForAmarra → web/static/js/amarra.js
amarra-cais pwa --bump   # after HTML/template changes on phones
```

Delete leftover `/static/build/` and HTMX / `cais.js` copies. SW is network-first for `/static/js/amarra.js`.

## 8. CLI

| Old           | New                  |
| ------------- | -------------------- |
| `cais new`    | `amarra-cais new`    |
| `cais g`      | `amarra-cais g`      |
| `cais doctor` | `amarra-cais doctor` |
| `cais pwa`    | `amarra-cais pwa`    |

`amarra-cais g handler` / `g page` / `g resource` emit HTML pages, not Svelte.

## 9. What stays

Store, SQLite migrations, jobs + `/jobs`, AWS SDK (or any domain package), session/CSRF/rate-limit middleware, and the CSRF cookie name (`cais_csrf` / `__Host-cais_csrf`). You are rewriting the UI layer, not the data layer.

## Suggested order

1. `amarra-cais new scratch` and copy `web/templates/layouts/app.html`, `viewdata.go`, and `amarraData`.
2. Replace one public page (home or login) end-to-end: handler test → HTML → `view.Write` → route.
3. Repeat for the rest. Keep store methods unchanged.
4. `amarra-cais doctor` until vite/Inertia checks are gone.
5. `amarra-cais pwa --bump` before phone testing.
