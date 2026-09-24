---
title: Security model
description: How CSRF, sessions, headers, rate limits and production gates combine in an Amarra app.
sidebar:
  order: 4
---

Amarra ships a small set of defensive pieces rather than a single security feature. Each layer covers a specific class of risk, and they are meant to be composed on the router in a known order.

## CSRF — double-submit cookie

`middleware.CSRF(cfg)` validates state-changing methods (`POST`, `PUT`, `DELETE`, `PATCH`). It uses a double-submit cookie named `cais_csrf`: the token lives in a cookie and is also submitted as a form field or an `X-CSRF-Token` header. There is no server-side token store — the server only compares the two values, which keeps the check stateless.

The layout renders `<meta name="csrf-token">`, and `amarra.js` sends `X-CSRF-Token` on Drive requests. Kit `<.form>` injects the hidden field from the root page data (`$.CSRFToken`), so you do not duplicate it inside the slot.

## Sessions

Sessions are cookie-based (`pkg/cais/session`) with `SignIn` / `SignOut`. A session rotates on login, invalidating the previous token. Cookies and their SQLite rows expire after 7 days (`sessionTTL` / `defaultMaxAge`); the store keeps `expires_at` and ignores expired rows on lookup. Passwords are stored as bcrypt hashes (`session.HashPassword` / `session.VerifyPassword`).

`session.CookieOptionsFromConfig(cfg)` marks cookies `Secure` when `cfg.CookieSecure()` is true, which happens with `ENV=production`. Prune expired rows with `amarra-cais db prune-sessions` or `session.Store.PruneExpired()`.

## Security headers

`middleware.SecurityHeaders(cfg)` runs after `Recover` and sets `X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy`, and `Permissions-Policy`. In production it also adds `Strict-Transport-Security`.

The Content-Security-Policy keeps `script-src 'self' 'unsafe-inline'`. That is a deliberate trade-off: the theme snippet runs before paint, and Drive adds small inline initialization, so escaping remains the primary defense. The roadmap is to replace `'unsafe-inline'` with a per-request nonce or SRI hashes (#97).

## Rate limiting

Wrap sensitive POST routes with per-IP token buckets:

```go
loginLimit := middleware.NewRateLimiter(10, cfg)   // 10 req/min
r.Post("/login", loginLimit.Middleware(http.HandlerFunc(auth.LoginPost)).ServeHTTP)
```

The limiter resolves the client address with `middleware.ClientIP(r, cfg)`. Behind a reverse proxy, set `TRUSTED_PROXIES` so `X-Forwarded-For` is trusted instead of spoofable.

## Production gates

`cfg.Validate()` fails the boot when `ENV=production` and required values are missing. In particular, `ADMIN_TOKEN` must be set for bearer-protected admin APIs, and `APP_URL` is required (it is also used for absolute Open Graph URLs). Missing configuration stops the process rather than degrading silently.

:::caution
The in-memory rate limiter and the double-submit CSRF check both assume a single app process. If you run more than one replica, move rate limiting to a shared store and revisit how the CSRF token is validated.
:::

Related reading: [auth and sessions](/amarra-cais/docs/how-to/auth-and-sessions/), [middleware reference](/amarra-cais/docs/reference/middleware/), and [configuration reference](/amarra-cais/docs/reference/configuration/).
