---
title: Streaming chat
description: "Stream named SSE ops to the browser: relay a chat response, target partials, and test the handler."
sidebar:
  order: 7
---

Amarra Stream (`pkg/amarra/stream`) is a named-op contract for server-sent events. Scaffold a complete example — conversations, messages, and the stream relay — with:

```bash
amarra-cais g stream chat
```

## Long-lived routes need WriteTimeout 0

| Setting        | Normal handlers | SSE routes                                             |
| -------------- | --------------- | ------------------------------------------------------ |
| `WriteTimeout` | `30s` ok        | **`0`** (disabled) — set on chat/stream routes         |
| Flush          | N/A             | `stream.Flush(w)` — never assert `http.Flusher` on `w` |

Set `WriteTimeout: 0` on the server for anything that streams; a non-zero timeout cuts the connection mid-response. Always flush with the package helper — middleware may wrap the response writer and hide `http.Flusher`.

## Relay and write ops

```go
import (
    "github.com/puppe1990/amarra-cais/pkg/amarra/stream"
    "github.com/puppe1990/amarra-cais/pkg/cais/chat"
)

func streamHandler(w http.ResponseWriter, r *http.Request) {
    stream.RelaySSE(w)
    _ = stream.WriteOp(w, stream.Op{Kind: "morph", Target: "chat-live", HTML: chat.LiveBubble("token…")})
    _ = stream.WriteOp(w, stream.Op{Kind: "append", Target: "chat-history", HTML: chat.MessageBubble(chat.RoleAssistant, "done", time.Now().UTC())})
    _ = stream.Flush(w) // never w.(http.Flusher) — middleware may hide Flusher
}
```

`stream.RelaySSE` starts the SSE response, `stream.WriteOp` sends one op, and `stream.Flush` pushes it to the client. The named ops are `append`, `prepend`, `replace`, `morph`, `remove`, and `toast`; unknown kinds are rejected.

For a one-shot batch outside a long connection, `stream.WriteHTTP(w, stream.Op{Kind: "append", Target: "list", HTML: row})` sets `Content-Type: text/vnd.amarra-stream` and applies the ops instead of morphing `#amarra-main`.

## The two chat partials

`amarra-cais g stream chat` ships two partials:

| Partial               | Use case                        | DOM                                                                    |
| --------------------- | ------------------------------- | ---------------------------------------------------------------------- |
| `chat_sse.html`       | Echo / simple bots              | `#chat-history` + `#chat-live`; `data-amarra-stream`                   |
| `chat_sse_agent.html` | Agent streaming (tokens, tools) | `#chat-history` + `#chat-stream` + `#chat-live` under `#chat-messages` |

Markup opts into Amarra Stream with `data-amarra-stream` — not `hx-ext` or `sse-ext`.

## Server helpers and tests

`pkg/cais/chat` renders the bubbles you stream: `LiveBubble`, `MessageBubble`, `SafeMessageBubble`, `ToolCallBubble`, `ToolResultBubble`, `DetailBubble`, `Truncate`, `SelectWindowWithLastUser`, plus the `UnsafeLiveHTML` / `UnsafeMessageHTML` variants (the caller sanitizes the HTML).

The generated handler tests use `testutil.AssertChatMarkers` (Show) and `testutil.AssertHTMLContains` (PostMessage bubble). A missing record returns `http.NotFound`, not a 500:

```go
conv, err := h.store.FindConversationByID(id)
if err != nil {
    http.NotFound(w, r)
    return
}
```

:::caution
Read the streamed response body and assert on the op markers. Do not assert `http.Flusher` on the writer — middleware may wrap it and the assertion will fail even though flushing works.
:::

:::tip
SSE handlers poll the database in a loop. Keep writes short and avoid long transactions during a stream so SQLite's WAL mode keeps readers moving.
:::

## Related

- [Live updates](/amarra-cais/docs/how-to/live-updates/) — the WebSocket counterpart for server-pushed HTML.
- [Pages and views](/amarra-cais/docs/how-to/pages-and-views/) — how the chat partials fit the layout.
- [Configuration](/amarra-cais/docs/reference/configuration/) — `WriteTimeout` and other server settings.
- [CLI reference](/amarra-cais/docs/reference/cli/) — `g stream chat`.
