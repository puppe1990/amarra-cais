---
title: PWA and mobile
description: Test on your phone over the LAN, bump the service worker cache, and pass the mobile doctor checklist.
sidebar:
  order: 9
---

Amarra apps ship as installable PWAs: manifest, service worker, offline page, icons, and fullscreen. A few commands keep that experience fresh while you develop on a phone.

## Test on your phone over the LAN

`boot.Print` shows the **LAN** URLs the server is reachable on for phone testing over Wi‑Fi. When the banner has scrolled away, `GET /health` returns the same list as `lan_urls` via `netutil.HealthPayload`:

```bash
curl -s http://localhost:8080/health
```

Use the `lan_urls` array — never concatenate `APP_URL` and the port by hand. In production `lan_urls` is omitted, since the app is only reachable on its public host.

## Bump the service worker cache

The service worker caches your CSS, JS, and pages. After you change templates or HTML, bump its cache so phones pick up the new assets:

```bash
amarra-cais pwa --bump
```

That refreshes the vendored assets and increments `CACHE_VERSION` in `sw.js`. `amarra-cais dev` auto-bumps the cache whenever `sw.js` exists, so local iteration stays fresh without a manual step. The service worker is network-first for `/static/js/amarra.js` and `/static/css/`.

## Run the mobile checks

```bash
amarra-cais doctor --mobile
```

The mobile doctor verifies flash markup, Google Fonts CSP, `amarra.js`, the service-worker cache (network-first `/static/js/amarra.js`), chat form CSS, the `#chat-messages` scroll container, and the health `lan_urls`.

:::caution
The default CSP blocks `fonts.googleapis.com`. Scaffold `input.css` uses system fonts for this reason — keep it that way unless you deliberately widen the CSP.
:::

## Brand assets

Scaffolds ship the docs boat mark as `web/static/favicon.svg` (the tab icon), a **neutral placeholder** tile (`pkg/cais/pwa/assets/icon.png`, 512×512) plus a generated `icons/icon-512-maskable.png`. The manifest splits `any` icons (192 + 512) from the `maskable` one. `amarra-cais doctor` warns while `web/static/favicon.svg`, `web/static/icons/*` and `og.png` are still the shipped defaults. Replace them with your own brand before you ship; `amarra-cais pwa --force` resets brand assets back to the defaults if you need a clean slate.

## Mobile checklist

1. `amarra-cais doctor --mobile`
2. `amarra-cais pwa --bump`
3. Open the boot **LAN** URL on your phone.
4. Stay on the page while an SSE stream runs — Drive morphs `#amarra-main`, and a full navigation drops the `EventSource`.

:::tip
If the phone shows a stale page, force-close the installed PWA so the service worker re-fetches assets after the cache bump.
:::

## Related

- [Streaming chat](/amarra-cais/docs/how-to/streaming-chat/) — why a live page must not navigate mid-stream.
- [Deploy to production](/amarra-cais/docs/how-to/deploy/) — ship the PWA assets beside the binary.
- [CLI reference](/amarra-cais/docs/reference/cli/) — `pwa`, `doctor`, and `dev`.
- [Your first app](/amarra-cais/docs/getting-started/your-first-app/) — the scaffold that ships this PWA setup.
