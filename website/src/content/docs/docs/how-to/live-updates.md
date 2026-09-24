---
title: Live updates
description: Broadcast server-rendered HTML over the opt-in WebSocket hub, and know its single-replica limit.
sidebar:
  order: 8
---

Live is an opt-in WebSocket hub (`pkg/amarra/live`) for pushing server-rendered HTML to open pages. CRUD stays on Drive; reach for Live when the server — not the user — is the source of change: dashboards, notifications, or presence.

## Register views and opt in

At boot you register the views the hub is allowed to broadcast with `hub.Register`. Then a page opts in with the `amarra-live` attribute and wires a trigger with `amarra-click`:

```html
<div amarra-live>…</div>
<button amarra-click>Refresh</button>
```

The hub endpoint is `GET /amarra/live`. The browser joins with a CSRF join payload, which is validated against the handshake cookie — the same double-submit model that guards Drive requests, applied to the socket handshake.

## Push from the server

Once a client is joined, three socket methods ride on the same morph message that Drive uses:

- `sock.Patch` — apply a patch to a live region.
- `sock.Stream` — stream named SSE-style ops to a live region.
- `sock.Push` — push fresh HTML to a live region.

Because they reuse the morph transport, a broadcast updates the DOM with the same rendering path as a Drive navigation instead of a separate payload format.

## Single-replica caveat

The hub is in-process only: two app replicas do not share sockets.

:::caution
Live is **single-replica**. If you run more than one instance behind a load balancer, a client connected to replica A will not receive events broadcast on replica B. Cross-replica fan-out needs an external bus.
:::

Slow clients are dropped rather than blocking the hub. Drops are counted by `hub.Dropped()` and logged on the first drop and then every 100th, so a stalled phone does not silently starve other subscribers. Watch that counter when you suspect a client is falling behind.

:::note
Outside development, Live hides handler error detail from the client, so a failed broadcast does not leak server internals.
:::

## Related

- [Streaming chat](/amarra-cais/docs/how-to/streaming-chat/) — SSE streaming for the request/response case.
- [Pages and views](/amarra-cais/docs/how-to/pages-and-views/) — the `#amarra-main` layout the morphs target.
- [Amarra JS](/amarra-cais/docs/reference/amarra-js/) — the public attributes, including `amarra-live` and `amarra-click`.
- [Security model](/amarra-cais/docs/explanation/security-model/) — the CSRF handshake used to join.
