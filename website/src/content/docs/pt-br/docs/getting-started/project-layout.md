---
title: Estrutura do projeto
description: O que o amarra-cais new escreve, arquivo por arquivo.
sidebar:
  order: 4
---

Um app gerado é um pequeno programa Go com templates HTML e assets estáticos ao lado. Não há etapa de build para o frontend além do Tailwind.

```
myapp/
  cmd/server/main.go          # boot: config, store, migrations, seeds, router
  cmd/worker/main.go          # background jobs worker (when jobs are used)
  internal/app/               # app.go + routes.go (registerRoutes)
  internal/handlers/          # one file per domain, plus *_test.go
  internal/store/             # Store interface + SQLite implementation
  internal/store/migrations/  # NNN_name.sql (-- up / -- down sections)
  internal/db/seeds.go        # idempotent seeds, dev-only when noted
  internal/jobs/              # job handlers + registry.go
  web/templates/layouts/      # app.html — the shell with #amarra-main
  web/templates/pages/        # one HTML file per page
  web/templates/partials/     # flat partials, addressed by name
  web/templates/components/   # kit overrides (optional)
  web/static/                 # css, js/amarra.js, icons, sw.js, manifest
```

## O caminho da requisição

1. O `internal/app/routes.go` registra rotas em um `cais.Router`; o middleware conecta sessões, flash, CSRF e headers de segurança.
2. Um handler monta os dados da página e chama `view.Write(w, r, views, view.Page{Layout, Name, Data}, cfg)`.
3. Templates carregados uma vez no boot renderizam a página dentro do layout; o Drive faz morph de `#amarra-main` na navegação em vez de recarregar o shell.

Os templates são parseados uma única vez com `view.Load` — uma tag `<.component>` desconhecida falha no boot, não na primeira requisição. Os globs exatos e as regras de endereçamento estão em [Views e kit](/amarra-cais/pt-br/docs/reference/views-and-kit/).

## Convenções que vale manter

- Um arquivo de handler por domínio, com um `*_test.go` correspondente (SQLite `:memory:`, sem mocks).
- Os handlers recebem dependências (`Store`, `*view.Renderer`, `cais.Config`) via construtor — sem globais.
- `registerRoutes`, `Close() error` e o marcador `<!-- cais:nav -->` no layout são o que os geradores e o `amarra-cais destroy` alteram; mantenha-os.
- As páginas são HTML, não JSON: sem verificações de `HX-Request`, sem payloads do Inertia.

Próximo: [Páginas e views](/amarra-cais/pt-br/docs/how-to/pages-and-views/).
