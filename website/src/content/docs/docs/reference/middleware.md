---
title: Middleware
description: Built-in middleware — recover, security headers, CSRF, sessions, flash, auth, logging and rate limiting.
sidebar:
  order: 9
---

Middleware lives in `pkg/cais/middleware`. Add it with `r.Use(mw)` on a `cais.Router`; the first `Use` is the outermost wrapper. `middleware.Middleware` is `func(http.Handler) http.Handler`.

## Reference

| Middleware                                | Purpose                                                                                                                         |
| ----------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| `middleware.Recover`                      | Catches panics, logs the stack, returns `500`.                                                                                  |
| `middleware.SecurityHeaders(cfg)`         | Sets `X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy`, `Permissions-Policy` and the CSP; adds HSTS in production. |
| `middleware.CSRF(cfg)`                    | Double-submit cookie (`cais_csrf`); validates `POST`/`PUT`/`PATCH`/`DELETE` and skips `/health` and `/static/`.                 |
| `middleware.LoadSession(store)`           | Reads the session cookie and attaches the user ID to the request context.                                                       |
| `middleware.Flash(cfg)`                   | Consumes the one-shot flash cookie into the request context.                                                                    |
| `middleware.RequireAuth(loginURL)`        | `303` redirect to `loginURL` when unauthenticated.                                                                              |
| `middleware.RequireAuthFunc(loginURL, h)` | Wraps a single handler with `RequireAuth`.                                                                                      |
| `middleware.AdminAuth(cfg)`               | Bearer token from `ADMIN_TOKEN`; no-op in development when unset, rejects everything in production when unset.                  |
| `middleware.LoggerTo(cfg, w)`             | Request logs (JSON when `cfg.LogJSON()`).                                                                                       |
| `middleware.NewRateLimiter(limit, cfg)`   | Per-IP token bucket, `limit` requests per minute.                                                                               |
| `middleware.ClientIP(r, cfg)`             | Resolves the client IP, trusting `X-Forwarded-For` only from `TRUSTED_PROXIES`.                                                 |

## Usage order

The scaffold wires the router in this order:

```go
r := cais.NewRouter()
r.Use(middleware.CSRF(cfg))
r.Use(middleware.LoadSession(deps.Store.Sessions()))
r.Use(middleware.Flash(cfg))
r.Use(i18n.LocaleMiddleware(catalogs, cfg.Locale))
if buf != nil {
  r.Use(middleware.LoggerTo(cfg, devlog.MirrorDefault(log.Writer())))
} else {
  r.Use(middleware.Logger(cfg))
}
r.Use(middleware.Recover)
r.Use(middleware.SecurityHeaders(cfg))
```

`Flash` runs after `LoadSession` (it needs the session), and `SecurityHeaders` is registered after `Recover`. CSRF covers the state-changing methods for every route.

The CSP keeps `script-src 'self' 'unsafe-inline'` for the FOUC theme snippet and Drive inline init; escaping is the primary defense, with a per-request nonce on the roadmap.

## Route-level middleware

Auth and rate limits are applied where they are needed rather than globally:

```go
loginLimit := middleware.NewRateLimiter(10, cfg)   // 10 req/min
r.Post("/login", loginLimit.Middleware(http.HandlerFunc(auth.LoginPost)).ServeHTTP)

r.Get("/dashboard", middleware.RequireAuthFunc("/login", dashboard.ServeHTTP))
```

Limiters key on `middleware.ClientIP(r, cfg)` plus the path, so set `TRUSTED_PROXIES` when the app sits behind a reverse proxy. Both browser admin (`RequireAuth`) and bearer-token APIs (`AdminAuth`) are options for `amarra-cais g resource`.

## CSRF and flash details

CSRF is a double-submit cookie: the token lives in the `cais_csrf` cookie and is echoed in a `csrf_token` form field or the `X-CSRF-Token` header (Drive sends the header from the `<meta name="csrf-token">` tag). Reading a message after a redirect goes through `middleware.FlashMessage(r)`, which returns the message the `Flash` middleware consumed.

Session rotates on login, invalidating the previous token, and the CSRF and flash cookies use `Secure` when `cfg.CookieSecure()` is true.

## Localhost-only tooling

Two development helpers mount their own routes, guarded to loopback connections:

```go
devlog.Register(r, cfg.Env, buf)        // /logs — development only
jobsui.Register(r, deps.Store.DB())     // /jobs — all envs, loopback only
```

`/logs` shows request and SQL logs in development. `/jobs` is the queue dashboard (counts, failed retry/discard, recurring tasks). Both reject proxied requests, so reach them over an SSH tunnel in production.

See [Configuration](/amarra-cais/docs/reference/configuration/) for `cfg` values and [Auth and sessions](/amarra-cais/docs/how-to/auth-and-sessions/) for session and flash flows.
