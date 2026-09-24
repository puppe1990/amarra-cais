---
title: Chat em streaming
description: "Envie ops SSE nomeadas ao navegador: faça o relay de uma resposta de chat, mire partials e teste o handler."
sidebar:
  order: 7
---

O Amarra Stream (`pkg/amarra/stream`) é um contrato de ops nomeadas para server-sent events. Faça o scaffold de um exemplo completo — conversas, mensagens e o relay do stream — com:

```bash
amarra-cais g stream chat
```

## Rotas de longa duração precisam de WriteTimeout 0

| Configuração   | Handlers normais | Rotas SSE                                                      |
| -------------- | ---------------- | -------------------------------------------------------------- |
| `WriteTimeout` | `30s` ok         | **`0`** (desabilitado) — definido nas rotas de chat/stream     |
| Flush          | N/A              | `stream.Flush(w)` — nunca faça assert de `http.Flusher` em `w` |

Defina `WriteTimeout: 0` no servidor para tudo que faz streaming; um timeout diferente de zero corta a conexão no meio da resposta. Sempre faça flush com o helper do pacote — o middleware pode envolver o response writer e esconder `http.Flusher`.

## Relay e ops de escrita

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

`stream.RelaySSE` inicia a resposta SSE, `stream.WriteOp` envia uma op e `stream.Flush` a empurra para o cliente. As ops nomeadas são `append`, `prepend`, `replace`, `morph`, `remove` e `toast`; kinds desconhecidos são rejeitados.

Para um lote one-shot fora de uma conexão de longa duração, `stream.WriteHTTP(w, stream.Op{Kind: "append", Target: "list", HTML: row})` define `Content-Type: text/vnd.amarra-stream` e aplica as ops em vez de fazer morph de `#amarra-main`.

## Os dois partials de chat

`amarra-cais g stream chat` entrega dois partials:

| Partial               | Caso de uso                        | DOM                                                                    |
| --------------------- | ---------------------------------- | ---------------------------------------------------------------------- |
| `chat_sse.html`       | Echo / bots simples                | `#chat-history` + `#chat-live`; `data-amarra-stream`                   |
| `chat_sse_agent.html` | Streaming de agent (tokens, tools) | `#chat-history` + `#chat-stream` + `#chat-live` under `#chat-messages` |

O markup adere ao Amarra Stream com `data-amarra-stream` — não `hx-ext` ou `sse-ext`.

## Helpers de servidor e testes

`pkg/cais/chat` renderiza as bubbles que você envia no stream: `LiveBubble`, `MessageBubble`, `SafeMessageBubble`, `ToolCallBubble`, `ToolResultBubble`, `DetailBubble`, `Truncate`, `SelectWindowWithLastUser`, além das variantes `UnsafeLiveHTML` / `UnsafeMessageHTML` (o chamador sanitiza o HTML).

Os testes de handler gerados usam `testutil.AssertChatMarkers` (Show) e `testutil.AssertHTMLContains` (bubble do PostMessage). Um registro ausente retorna `http.NotFound`, não um 500:

```go
conv, err := h.store.FindConversationByID(id)
if err != nil {
    http.NotFound(w, r)
    return
}
```

:::caution
Leia o corpo da resposta em streaming e faça assert nos marcadores de op. Não faça assert de `http.Flusher` no writer — o middleware pode envolvê-lo e o assert vai falhar mesmo que o flush funcione.
:::

:::tip
Handlers SSE consultam o banco em loop. Mantenha as escritas curtas e evite transações longas durante um stream para que o modo WAL do SQLite mantenha os leitores fluindo.
:::

## Relacionados

- [Atualizações ao vivo](/amarra-cais/pt-br/docs/how-to/live-updates/) — a contraparte WebSocket para HTML empurrado pelo servidor.
- [Páginas e views](/amarra-cais/pt-br/docs/how-to/pages-and-views/) — como os partials de chat se encaixam no layout.
- [Configuração](/amarra-cais/pt-br/docs/reference/configuration/) — `WriteTimeout` e outras configurações do servidor.
- [Referência da CLI](/amarra-cais/pt-br/docs/reference/cli/) — `g stream chat`.
