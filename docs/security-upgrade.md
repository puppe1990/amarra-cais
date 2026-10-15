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
