---
title: Router
description: The cais.Router API — HTTP methods, route groups, path parameters, param helpers and designed 404 handling.
sidebar:
  order: 4
---

`cais.Router` wraps `net/http`'s `http.ServeMux` (Go 1.22+) and adds path-parameter helpers, middleware groups and a designed 404. Register routes in `internal/app/routes.go`.

## Methods

| Call                         | Registers |
| ---------------------------- | --------- |
| `r.Get(pattern, handler)`    | `GET`     |
| `r.Post(pattern, handler)`   | `POST`    |
| `r.Put(pattern, handler)`    | `PUT`     |
| `r.Patch(pattern, handler)`  | `PATCH`   |
| `r.Delete(pattern, handler)` | `DELETE`  |

Handlers are plain `http.HandlerFunc`. The router also exposes `r.Use(mw)`, `r.Handle(pattern, handler)` and `r.Static(prefix, dir)`.

`Middleware` is `func(http.Handler) http.Handler`; `Use` appends it to the chain that wraps every registered route.

## Path parameters

Write parameters as `{name}` in the pattern and read them through a helper, which parses the value and calls your typed handler:

```go
r.Get("/blog/{slug}", cais.StringParam("slug", blog.Show))
r.Post("/chat/{id}/permissions/{permID}/approve", cais.StringParams("id", "permID", chat.ApprovePermission))
r.Get("/items/{id}/{slug}", cais.IntStringParams("id", "slug", items.Show))
r.Group(middleware.RequireAuth("/login"), func(g *cais.Router) {
  g.Post("/admin/items/{id}", cais.IntParam("id", admin.Update))
})
```

| Helper                           | Handler signature                   |
| -------------------------------- | ----------------------------------- |
| `cais.StringParam(name, fn)`     | `func(w, r, value string)`          |
| `cais.StringParams(a, b, fn)`    | `func(w, r, a, b string)`           |
| `cais.IntParam(name, fn)`        | `func(w, r, value int64)`           |
| `cais.IntStringParams(i, s, fn)` | `func(w, r, value int64, s string)` |

`IntParam` rejects non-integers and values `<= 0`; the string helpers reject empty values. A failed parse is routed to the designed 404 (below) instead of a silent zero value.

## Groups

`r.Group(mw, fn)` builds a child router that shares the parent's URL namespace and `ServeMux` but adds one middleware. Only the middleware differs — groups do not create a path prefix. The child copies the parent's current chain and appends its own, so routes in a group run the parent middlewares plus the group middleware. Pass any `Middleware` (`func(http.Handler) http.Handler`), for example `middleware.RequireAuth("/login")` for browser admin or `middleware.AdminAuth(cfg)` for a bearer-token API.

## Designed 404

`r.NotFound(handler)` serves unmatched routes **and** the param helpers' parse failures. The handler owns the status and runs with the router middlewares (session, flash, CSRF), so it can render a normal page:

```go
r.NotFound(notFound.ServeHTTP)   // notFound calls view.Write(..., Status: http.StatusNotFound)
```

Without it, both paths fall back to `http.NotFound` (`text/plain`).

## Exact root and static files

A bare `"/"` pattern is rewritten to the end-of-path wildcard `/{$}`, so the home handler matches only the exact root instead of every unmatched path (such as `/.env` or `/nope`). Real subtree patterns like `"/static/"` keep their prefix behavior. Serve files with `r.Static(prefix, dir)` (or `StaticForEnv` with a `Config`); in development it sets `Cache-Control: no-store` so JS/CSS edits apply without a rebuild.

## ServeMux conflicts

Registration panics when two patterns under the same method can match the same path and neither is more specific:

```go
// ❌ panic: both match /webhooks/kiwify/delete
r.Post("/webhooks/kiwify/{token}", receiver)
r.Post("/webhooks/{id}/delete", deleteHandler)

// ✅ prefer a REST method or a distinct prefix
r.Delete("/webhooks/{id}", deleteHandler)
r.Post("/webhooks/incoming/{token}", receiver)
```

`cais.Router` rewrites the panic with a short hint, and `amarra-cais routes` warns when it can detect the conflict statically from `routes.go`.
