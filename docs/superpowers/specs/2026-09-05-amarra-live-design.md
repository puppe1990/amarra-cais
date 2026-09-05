# Amarra Live (Slice B)

**Date:** 2026-09-05
**Status:** Approved for implementation
**Parent:** `docs/superpowers/specs/2026-09-05-amarra-cais-design.md`

## Goal

Replace `GET /amarra/live` 501 with an opt-in WebSocket LiveView: server-side view state, HTML morph via Idiomorph, in-process hub. Drive/Frame/Stream stay the default for CRUD.

## Protocol

Upgrade: `GET /amarra/live?view=<name>&topic=<topic>` (topic defaults to view name).

Client → server JSON:

- `{ "type": "join", "csrf": "<token>" }`
- `{ "type": "event", "event": "inc", "payload": {}, "ref": "1" }`

Server → client JSON:

- `{ "type": "ok", "html": "...", "target": "count" }` after join
- `{ "type": "morph", "html": "...", "target": "count" }` after Handle
- `{ "type": "ack", "ref": "1" }`
- `{ "type": "error", "message": "...", "ref": "1" }`

CSRF: cookie from the handshake vs `join.csrf` (constant-time). Missing/mismatch closes the socket.

## Server API

```go
type View interface {
    Mount(ctx context.Context, sock Socket) error
    Handle(ctx context.Context, ev Event) error
    Render() Rendered
}

type Hub struct { /* MaxConns, Idle, PingInterval */ }
func NewHub(cfg Config) *Hub
func (h *Hub) Register(name string, fn func() View)
func (h *Hub) Handler() http.Handler
func (h *Hub) Broadcast(topic string, ev Event)
```

- One goroutine owns each connection (event channel serializes Handle/Render/write).
- Unknown view → 404 before upgrade. Hub full → 503. Non-WebSocket GET → 426.
- Panic in Handle → recover, log, close. Unknown event name → `{type:error}`, stay up.
- Pub/sub is in-process only. Two binaries do not share Live state.

## JS

`[amarra-live="counter"]` auto-connects. `amarra-click` / `amarra-change` / `amarra-submit` inside that root go over WS (capture phase; Drive does not see them). Reconnect remounts.

## Scaffold

- `app.New` builds a `live.Hub` and `hub.Handler()` (empty register is fine).
- `amarra-cais g live <name>` writes a Live view + page + route + `hub.Register`.
- `amarra-cais g stream chat --live` adds a Live view for the conversation; SSE remains the default without the flag.

## Out of scope

Redis/NATS, Phoenix static/dynamic diffs, sticky state across deploys, Live on resource CRUD.
