---
title: Configuration
description: Environment variables, cais.Load() precedence, production validation gates and cookie security flags.
sidebar:
  order: 8
---

Configuration is read once at boot by `cais.Load()`, which also applies a local `.env` file. The result is a `cais.Config` value passed to handlers, middleware and the router.

## Environment variables

| Variable          | Purpose                                                                                    | Default                    |
| ----------------- | ------------------------------------------------------------------------------------------ | -------------------------- |
| `ENV`             | Environment: `development` or `production`. Drives `CookieSecure()`, HSTS and dev tooling. | `development`              |
| `PORT`            | Listen address. `cais.ResolvePort` shifts to the next free port in development.            | `:8080`                    |
| `APP_URL`         | Absolute base URL for OG/Twitter image URLs. Required in production.                       | —                          |
| `ADMIN_TOKEN`     | Bearer token for `middleware.AdminAuth`. Required in production.                           | —                          |
| `TRUSTED_PROXIES` | Comma-separated proxy IPs/CIDRs; `X-Forwarded-For` is trusted only from these.             | —                          |
| `LOCALE`          | UI language for `pkg/cais/i18n` (`en` or `pt`).                                            | `en`                       |
| `STATIC_DIR`      | Static files directory when the process `WorkingDirectory` is not the app root.            | `web/static`               |
| `TEMPLATES_DIR`   | Templates directory, same override rule.                                                   | `web/templates`            |
| `LOG_FORMAT`      | Request/SQL log format: `json` or `text`.                                                  | JSON in dev and production |

:::note
Additional overrides exist for the security headers: `DB_PATH`, `PERMISSIONS_POLICY`, and `CSP_STYLE_SRC` / `CSP_CONNECT_SRC` / `CSP_MEDIA_SRC` / `CSP_IMG_SRC` / `CSP_FONT_SRC` (a hosted webfont needs both `CSP_FONT_SRC` and `CSP_STYLE_SRC`).
:::

## Loading and precedence

```go
cfg := cais.Load()
if err := cfg.Validate(); err != nil {
    log.Fatal(err)
}
```

`cais.Load()` applies a local `.env` if present (via `dotenv.LoadFile`). Keys already present in the process environment win — `t.Setenv`, systemd `Environment=`, and CI secrets are not overwritten. A missing `.env` is a no-op.

Example `.env` (process env still wins):

```bash
ENV=development
PORT=:8080
LOCALE=pt
```

`amarra-cais doctor` reads the same `.env` when it checks the app.

## Production gates

`cfg.Validate()` fails on boot when `ENV=production` and a required value is missing:

- `ADMIN_TOKEN` — `AdminAuth` rejects every request in production when the token is empty.
- `APP_URL` — required so OG/Twitter image URLs are absolute.

By the same switch, `SanitizeErrors()` is true in production and dev-only seeds (the demo user) do not run. Use `amarra-cais db seed` for catalog data.

`APP_URL` also feeds `meta.SiteFrom` so Open Graph and Twitter previews use absolute image URLs.

## Cookies

`cfg.CookieSecure()` returns true when `ENV=production`. It is threaded into the session, CSRF and flash cookie writers:

```go
session.SignIn(w, sessions, r, userID, session.CookieOptionsFromConfig(cfg))
flash.Set(w, "notice", "Saved!", cfg.CookieSecure())
```

Session cookies (and their SQLite rows) expire after 7 days; expired rows are ignored on lookup and can be pruned with `amarra-cais db prune-sessions`.

`ENV=production` also turns on HSTS and error sanitization, and hides the development tooling.

## Port selection

`cais.ResolvePort(cfg.Port, cfg.Env)` returns the address to listen on, shifting to the next free port in development when the preferred one is busy. Set `PORT_STRICT=1` to disable the shift and fail instead. Behind a reverse proxy, set `TRUSTED_PROXIES` so `middleware.ClientIP` trusts `X-Forwarded-For` for rate limiting and logs.

See [Deploy](/amarra-cais/docs/how-to/deploy/) for the production values and [Middleware](/amarra-cais/docs/reference/middleware/) for how config flows into headers and auth.
