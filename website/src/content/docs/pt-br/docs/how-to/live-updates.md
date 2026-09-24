---
title: Atualizações ao vivo
description: Faça broadcast de HTML renderizado no servidor pelo hub WebSocket opt-in e conheça seu limite de réplica única.
sidebar:
  order: 8
---

Live é um hub WebSocket opt-in (`pkg/amarra/live`) para empurrar HTML renderizado no servidor para páginas abertas. CRUD continua no Drive; recorra ao Live quando o servidor — e não o usuário — é a origem da mudança: dashboards, notificações ou presença.

## Registre views e faça opt-in

No boot você registra as views que o hub pode transmitir com `hub.Register`. Depois, uma página faz opt-in com o atributo `amarra-live` e conecta um trigger com `amarra-click`:

```html
<div amarra-live>…</div>
<button amarra-click>Refresh</button>
```

O endpoint do hub é `GET /amarra/live`. O navegador entra com um payload de join CSRF, que é validado contra o cookie de handshake — o mesmo modelo double-submit que protege as requisições do Drive, aplicado ao handshake do socket.

## Empurre a partir do servidor

Uma vez que um cliente entrou, três métodos de socket trafegam na mesma mensagem morph que o Drive usa:

- `sock.Patch` — aplica um patch a uma região live.
- `sock.Stream` — envia ops nomeadas no estilo SSE para uma região live.
- `sock.Push` — empurra HTML novo para uma região live.

Como eles reutilizam o transporte morph, um broadcast atualiza o DOM com o mesmo caminho de renderização de uma navegação do Drive, em vez de um formato de payload separado.

## Ressalva de réplica única

O hub é apenas in-process: duas réplicas do app não compartilham sockets.

:::caution
Live é **single-replica**. Se você rodar mais de uma instância atrás de um load balancer, um cliente conectado à réplica A não vai receber eventos transmitidos na réplica B. O fan-out entre réplicas precisa de um bus externo.
:::

Clientes lentos são descartados em vez de bloquear o hub. Os descartes são contados por `hub.Dropped()` e registrados no primeiro descarte e depois a cada 100º, para que um celular travado não deixe silenciosamente outros assinantes sem serviço. Fique de olho nesse contador quando suspeitar que um cliente está ficando para trás.

:::note
Fora do desenvolvimento, o Live esconde do cliente o detalhe de erro do handler, para que um broadcast que falhou não vaze internos do servidor.
:::

## Relacionados

- [Chat em streaming](/amarra-cais/pt-br/docs/how-to/streaming-chat/) — streaming SSE para o caso de requisição/resposta.
- [Páginas e views](/amarra-cais/pt-br/docs/how-to/pages-and-views/) — o layout `#amarra-main` que os morphs miram.
- [Amarra JS](/amarra-cais/pt-br/docs/reference/amarra-js/) — os atributos públicos, incluindo `amarra-live` e `amarra-click`.
- [Modelo de segurança](/amarra-cais/pt-br/docs/explanation/security-model/) — o handshake CSRF usado para entrar.
