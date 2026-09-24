---
title: Deploy to production
description: Cross-compile, ship web/static with a systemd unit, and satisfy the production environment gates.
sidebar:
  order: 11
---

Amarra deploys as a single Go binary plus a `web/static` directory. There is no Vite `web/static/build/` step and no Node process in production.

## Build the release

```bash
amarra-cais css     # Tailwind → web/static/css/styles.css
amarra-cais build --os linux --arch amd64 -o bin/server-linux
tar czf release.tar.gz bin/server-linux web/static
```

Ship `web/static` (CSS, `js/amarra.js`, the manifest, icons) beside the binary. `amarra-cais doctor` checks that `web/static`, `manifest.webmanifest`, and `amarra.js` are present before you ship.

## Server layout and systemd

```text
/opt/myapp/
  current/
    bin/server          # renamed from server-linux
    web/static/         # CSS, JS, manifest, icons
  data/app.db           # persistent SQLite
/etc/myapp/env          # production variables (chmod 600)
```

Copy `deploy/systemd/cais-app.service.example` and point systemd `WorkingDirectory` at `/opt/myapp/current`, where `web/static/` lives. If the working directory is not the app root, set `STATIC_DIR` and `TEMPLATES_DIR` explicitly so the binary finds its assets. The repo's `docs/deploy/lightsail-systemd.md` walks through a full Lightsail + Caddy setup.

```bash
sudo systemctl daemon-reload
sudo systemctl enable myapp
sudo systemctl start myapp
curl -s http://127.0.0.1:4006/health
```

## Production environment

```bash
PORT=:4006
ENV=production
APP_URL=https://myapp.example.com
DB_PATH=/opt/myapp/data/app.db
ADMIN_TOKEN=<strong-token>
LOCALE=pt
TRUSTED_PROXIES=127.0.0.1
# STATIC_DIR=/opt/myapp/current/web/static   # optional
```

`ENV=production` turns on the production gate: `cfg.Validate()` fails on boot when `ADMIN_TOKEN` or `APP_URL` is missing, so the app refuses to start rather than run insecurely. Set `TRUSTED_PROXIES` (comma-separated IPs) behind a reverse proxy so `middleware.ClientIP` trusts `X-Forwarded-For` for rate limiting and logging.

## Seeds in production

Development seeds — including the demo user — do **not** run when `ENV=production`. For idempotent catalog data, register seeds in `internal/db/seeds.go` and run:

```bash
amarra-cais db seed
```

## Security headers and logs

`middleware.SecurityHeaders(cfg)` (registered after `Recover`) sets `X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy`, and `Permissions-Policy`, and adds `Strict-Transport-Security` in production. CSRF and flash cookies use `Secure` when `cfg.CookieSecure()` is true, and the session rotates on login.

In `ENV=development`, request and SQL logs stream as JSON (`kind: request`, `kind: sql`); `LOG_FORMAT=text` opts out. `/logs` is localhost-only, and `/jobs` is localhost-only in every environment — reach it over an SSH tunnel on a remote box.

## Verify

```bash
curl -sI https://myapp.example.com/ | grep -i permissions-policy
curl -s https://myapp.example.com/static/manifest.webmanifest | grep display
curl -s https://myapp.example.com/health
```

## Related

- [PWA and mobile](/amarra-cais/docs/how-to/pwa-and-mobile/) — the assets you are shipping.
- [Configuration](/amarra-cais/docs/reference/configuration/) — every environment variable above.
- [Security model](/amarra-cais/docs/explanation/security-model/) — headers, cookies, and the production gate.
- [Upgrade the framework](/amarra-cais/docs/how-to/upgrade/) — bump versions before you redeploy.
