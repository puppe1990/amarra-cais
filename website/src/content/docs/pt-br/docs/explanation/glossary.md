---
title: Glossário
description: Definições em uma frase dos termos do produto Amarra, cada uma apontando para a página-contrato.
sidebar:
  order: 6
---

Palavras do produto. Cada termo é uma frase mais a página que é o contrato. Nomes de componentes do kit e de hooks ficam nas tabelas que essas páginas já pinam — esta página não as copia.

| Termo | Significado |
| --- | --- |
| Cais vs Amarra | Dois produtos: Cais v0.11.x é Inertia + Svelte; Amarra é o fork HTML-first. Veja [Cais vs Amarra](/amarra-cais/pt-br/docs/explanation/cais-vs-amarra/). |
| Drive | Intercepta cliques e submits same-origin e faz morph de `#amarra-main`. Veja [Views e Drive](/amarra-cais/pt-br/docs/explanation/views-and-drive/) e [amarra.js](/amarra-cais/pt-br/docs/reference/amarra-js/). |
| Frame | Região nomeada (`<amarra-frame>` / `Amarra-Frame`) que morfiza em vez do main inteiro. Veja [amarra.js](/amarra-cais/pt-br/docs/reference/amarra-js/). |
| Stream | Ops SSE nomeadas (`append`, `morph`, …) em vez de um swap Drive da página. Veja [Chat com stream](/amarra-cais/pt-br/docs/how-to/streaming-chat/). |
| Live | Hub WebSocket opt-in in-process que empurra HTML renderizado no servidor. Veja [Atualizações ao vivo](/amarra-cais/pt-br/docs/how-to/live-updates/). |
| kit | Tags `<.component>` inclusas (`form`, `input`, …) expandidas no boot. Veja [Views e o kit incluso](/amarra-cais/pt-br/docs/reference/views-and-kit/). |
| morph | O Drive troca o inner HTML de `#amarra-main` sem recarregar o shell. Veja [Views e Drive](/amarra-cais/pt-br/docs/explanation/views-and-drive/). |
| amarra-hook | `amarra-hook="name"` liga um comportamento incluso do `amarra.js`. Veja [amarra.js](/amarra-cais/pt-br/docs/reference/amarra-js/). |
| data-amarra-shell | Marcador no `<body>` (`auth` / `app`); valor diferente força visita completa em vez de morph (#249). Veja [amarra.js](/amarra-cais/pt-br/docs/reference/amarra-js/). |
| #amarra-main | Elemento do layout que o Drive morfiza; o resto do shell fica. Veja [Estrutura do projeto](/amarra-cais/pt-br/docs/getting-started/project-layout/). |
| g resource | Gerador que escreve CRUD de admin de tabela (model, store, HTML, rotas). Veja [Geradores](/amarra-cais/pt-br/docs/reference/generators/). |
| store | Interface SQLite + implementação SQL por trás dos handlers. Veja [Banco de dados e migrations](/amarra-cais/pt-br/docs/how-to/database-and-migrations/). |
| doctor | `amarra-cais doctor` confere o app contra o contrato do scaffold. Veja [CLI](/amarra-cais/pt-br/docs/reference/cli/). |
| /jobs | Dashboard localhost da fila de jobs em SQLite. Veja [Jobs e SQLite](/amarra-cais/pt-br/docs/explanation/jobs-and-sqlite/). |
| destroy | CLI que desfaz um gerador e desfaz patches de rotas, store e nav. Veja [Geradores](/amarra-cais/pt-br/docs/reference/generators/). |
| patch marker | Âncoras que os geradores alteram (`registerRoutes`, `Close() error`, `<!-- cais:nav -->`). Veja [Estrutura do projeto](/amarra-cais/pt-br/docs/getting-started/project-layout/). |
