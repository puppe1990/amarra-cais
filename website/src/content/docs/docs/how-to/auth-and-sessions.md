---
title: Auth and sessions
description: Protect routes with SQLite-backed cookie sessions, choose a session or bearer admin mode, and prune expired logins.
sidebar:
  order: 5
---

Amarra ships cookie-based sessions backed by SQLite. Every `amarra-cais new` app already includes login, logout, and a protected `/dashboard`. Add the same scaffolding to an existing app with `amarra-cais g auth`, then run `amarra-cais db migrate` to create the session tables.

## Wire the session middleware

Register `LoadSession` before anything that reads the current user, and `Flash` right after it so one-shot notices survive redirects:

```go
r.Use(middleware.LoadSession(deps.Store.Sessions()))
r.Use(middleware.Flash(cfg))
r.Get("/dashboard", middleware.RequireAuthFunc("/login", dashboard.ServeHTTP))
```

`RequireAuthFunc` redirects unauthenticated visitors to the login path you pass. Wrap a whole prefix with `middleware.RequireAuth("/login")` when every route beneath it is protected.

## Sign a user in

After you validate credentials, call `session.SignIn` with the SQLite-backed store and set a flash notice:

```go
session.SignIn(w, sessions, r, userID, session.CookieOptionsFromConfig(cfg))
flash.Set(w, "notice", "Welcome back!", cfg.CookieSecure())
```

`session.NewSQLiteStore` persists sessions in the app's SQLite file, so logins survive a restart. Sign out with `session.SignOut`. The session rotates on login, which invalidates the previous token — a defense against fixation. In a fresh scaffold the development seed user is `demo@example.com` / `password`.

## Expiry and pruning

Cookies and database rows expire after 7 days (`sessionTTL` / `defaultMaxAge`). SQLite stores `expires_at`, and expired rows are ignored on lookup. Delete stale rows periodically:

```bash
amarra-cais db prune-sessions
```

You can also call `session.Store.PruneExpired()` from your own maintenance code; the built-in `PruneSessions` job does the same on a schedule.

## Production cookies

`session.CookieOptionsFromConfig(cfg)` sets `Secure` when `cfg.CookieSecure()` is true, which it is under `ENV=production`. Always build cookie options from config so production traffic never receives a non-Secure cookie.

## Protect admin areas

Pick one of two modes:

| Mode                    | Middleware                         | Generator flag                   |
| ----------------------- | ---------------------------------- | -------------------------------- |
| Browser admin (default) | `middleware.RequireAuth("/login")` | `amarra-cais g resource` default |
| Bearer token API        | `middleware.AdminAuth(cfg)`        | `--admin-auth bearer`            |

`amarra-cais g resource` defaults to session auth (`--admin-auth session`). Choose `--admin-auth bearer` for token-only admin APIs that have no login pages.

`AdminAuth` accepts a Bearer header only — it does not read query params. Set `ADMIN_TOKEN` in production; `cfg.Validate()` fails on boot when it is missing. When `ADMIN_TOKEN` is unset the middleware is a no-op, which is the development behavior.

:::caution
Treat `ADMIN_TOKEN` as a production secret. Store it in the process environment (for example `/etc/myapp/env` with `chmod 600`) and never commit it.
:::

:::tip
The middleware order matters: `LoadSession` → `Flash` → `CSRF` → `SecurityHeaders`. See [Middleware reference](/amarra-cais/docs/reference/middleware/) for the full stack.
:::

## Related

- [Database and migrations](/amarra-cais/docs/how-to/database-and-migrations/) — the tables sessions live in.
- [Production deploy](/amarra-cais/docs/how-to/deploy/) — the production gates that require `ADMIN_TOKEN`.
- [Security model](/amarra-cais/docs/explanation/security-model/) — how CSRF, cookies, and sessions interact.
- [CLI reference](/amarra-cais/docs/reference/cli/) — `g auth`, `db migrate`, and `db prune-sessions`.
