# Security upgrade

## Administrative authorization (#350)

New auth scaffolds include `users.is_admin`, defaulting to 0. Public signup never grants administration. The development demo account is an administrator; production does not seed that account.

Generated session CRUD routes use `middleware.RequireAdmin("/login", deps.Store.DB())`. Guests redirect to login; ordinary users receive 403 for every admin operation. The role is read on every request, and database failures deny access. Bearer API mode is unchanged.

Existing apps must add a migration before changing their route middleware:

```sql
ALTER TABLE users ADD COLUMN is_admin INTEGER NOT NULL DEFAULT 0 CHECK (is_admin IN (0, 1));
```

Promote only a trusted account explicitly, using a bound email parameter:

```sql
UPDATE users SET is_admin = 1 WHERE email = ?;
```

Replace `RequireAuth` with `RequireAdmin` on existing `/admin/*` groups. A framework/CLI update does not rewrite existing app source. Minimal apps need `g auth` before session admin access can be granted. Personal resources also require app-specific ownership checks.

## Legacy session expiry (#351)

`session.EnsureSQLiteSchema` migrates populated legacy tables transactionally. Existing sessions retain expiry at `created_at + 7 days`; older sessions remain expired. Missing or invalid creation dates expire immediately. Repeating the migration is safe, and a failed backfill rolls back the column addition so it can be retried.

## Session revocation errors (#352)

`session.Store.Delete(token string)` now returns `error`. Custom stores must return deletion failures and return `nil` after successful deletion. `session.SignOut` now returns `error`; `SignIn` also propagates failures while revoking the previous session. No replacement or cleared cookie is sent when revocation fails, so the client can retry. A failed logout does not revoke the server token.

Existing apps must handle the error before redirecting or reporting success:

```go
if err := session.SignOut(w, sessions, r, session.CookieOptionsFromConfig(cfg)); err != nil {
    httpx.ServerError(w, err, cfg)
    return
}
http.Redirect(w, r, "/login", http.StatusSeeOther)
```

The scaffold logs errors in production through `httpx.ServerError` and returns HTTP 500 without a success redirect. Never include session tokens in errors or logs. Existing login handlers should already check the error from `SignIn`.

## CSRF bearer exemption (#364)

`middleware.CSRF` no longer exempts every route under `/api/` or any request with an `Authorization: Bearer` header. A session route under `/api/` must send the double-submit token; an invalid Bearer never falls back to the session cookie.

A machine client skips CSRF only when its Bearer token matches `ADMIN_TOKEN` (the same gate `AdminAuth` uses) or when the app registers the prefix explicitly:

```go
r.Use(middleware.CSRF(cfg, middleware.BearerAPI("/api/hooks/", verifyHookToken)))
```

`verifyHookToken` must compare the presented token in constant time, like `AdminAuth` does. The token is validated before content type, so form-encoded clients (`curl -d`) keep working. Never register a prefix that also serves session routes: the session fallback would turn the exemption into a bypass. In development, set `ADMIN_TOKEN` before exercising a bearer API; `AdminAuth` being a no-op when the variable is unset does not make CSRF skip.
