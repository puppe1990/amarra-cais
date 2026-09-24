---
title: Guias práticos
description: Guias focados em tarefas para adicionar páginas, formulários, um banco de dados, autenticação, jobs e publicar um app Amarra.
sidebar:
  order: 1
---

Estes guias são walkthroughs focados em tarefas para um app Amarra gerado pelo scaffold. Cada um assume que você gerou um app com `amarra-cais new` e consegue rodá-lo com `amarra-cais dev`. Os comandos rodam a partir do diretório do app, a menos que um passo diga o contrário, e os caminhos de arquivo são os que o scaffold cria (`internal/handlers/`, `web/templates/`, `internal/store/`).

Comece por [Páginas e views](/amarra-cais/pt-br/docs/how-to/pages-and-views/) se você nunca adicionou uma rota, depois siga o guia que corresponde à sua tarefa atual.

:::tip
Amarra é test-driven. Escreva o teste Go primeiro, confirme que ele falha pelo motivo certo, depois adicione o código mínimo para passá-lo. Apps gerados pelo scaffold rodam os testes com `go test ./...` ou `amarra-cais test`.
:::

## Guias

- [Páginas e views](/amarra-cais/pt-br/docs/how-to/pages-and-views/) — gere um handler e uma página, renderize com `view.Write` e re-renderize na validação.
- [Formulários e validação](/amarra-cais/pt-br/docs/how-to/forms-and-validation/) — as tags de formulário do kit incluído, CSRF, erros de campo no servidor e upload de arquivos.
- [Banco de dados e migrations](/amarra-cais/pt-br/docs/how-to/database-and-migrations/) — adicione uma tabela SQLite, escreva a migration e rode-a pela CLI.
- [Autenticação e sessões](/amarra-cais/pt-br/docs/how-to/auth-and-sessions/) — login, logout, rotas protegidas e armazenamento de sessão.
- [Jobs em background](/amarra-cais/pt-br/docs/how-to/background-jobs/) — enfileire trabalho na fila SQLite e rode um worker.
- [Chat com streaming](/amarra-cais/pt-br/docs/how-to/streaming-chat/) — faça o scaffold de um chat SSE com as stream ops nomeadas.
- [Atualizações ao vivo](/amarra-cais/pt-br/docs/how-to/live-updates/) — o hub WebSocket opt-in para views em tempo real.
- [PWA e mobile](/amarra-cais/pt-br/docs/how-to/pwa-and-mobile/) — instalabilidade, o service worker e testes em LAN no celular.
- [i18n](/amarra-cais/pt-br/docs/how-to/i18n/) — alterne as strings de UI entre inglês e português.
- [Deploy](/amarra-cais/pt-br/docs/how-to/deploy/) — cross-compile, envie assets estáticos e rode sob systemd.
- [Upgrade](/amarra-cais/pt-br/docs/how-to/upgrade/) — atualize o framework e aplique o checklist de migração.

## Relacionado

- [Referência da CLI](/amarra-cais/pt-br/docs/reference/cli/) — todos os comandos e subcomandos do `amarra-cais`.
- [Geradores](/amarra-cais/pt-br/docs/reference/generators/) — `g handler`, `g page`, `g resource` e os demais.
- [Views e o kit](/amarra-cais/pt-br/docs/reference/views-and-kit/) — os componentes e hooks incluídos que estes guias usam.
