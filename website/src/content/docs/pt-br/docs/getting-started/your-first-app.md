---
title: Seu primeiro app
description: Faça o scaffold de um app, instale as dependências e rode o servidor de dev.
sidebar:
  order: 3
---

## Scaffold

```bash
amarra-cais new myapp
cd myapp
amarra-cais install   # npm install + go mod tidy + Tailwind build
amarra-cais dev       # http://localhost:8080
```

O `new` usa por padrão o scaffold HTML (Amarra Views + Drive). Existem duas variantes menores:

```bash
amarra-cais new myapp --minimal   # fewer pages, same stack
amarra-cais new myapp --blank     # bare skeleton
amarra-cais new myapp --module github.com/acme/myapp
```

O gerador também escreve `AGENTS.md`, um workflow de CI do GitHub Actions, config de pre-commit, setup de golangci-lint e Prettier, para que o app comece com as mesmas proteções do framework.

## Faça login

O seed de dev cria um usuário de demonstração:

```
demo@example.com / password
```

O `/dashboard` é protegido por autenticação de sessão por padrão. Os seeds não rodam quando `ENV=production` — use `amarra-cais db seed` para dados de catálogo.

## Primeiras verificações úteis

```bash
amarra-cais doctor           # verifies amarra.js, layouts/app.html, air, go.mod, PWA
amarra-cais routes           # lists HTTP routes from internal/app/routes.go
amarra-cais db migrate       # applies pending migrations
go test ./...                # handler tests run headless against :memory: SQLite
```

Se a porta preferida estiver ocupada, o servidor escolhe uma livre na inicialização. O `amarra-cais dev` roda air + watch do Tailwind, então edições em templates recarregam sem reiniciar.

## Próximos passos

- [Estrutura do projeto](/amarra-cais/pt-br/docs/getting-started/project-layout/) — o que foi gerado.
- [Páginas e views](/amarra-cais/pt-br/docs/how-to/pages-and-views/) — adicione sua primeira tela.
- [Referência de geradores](/amarra-cais/pt-br/docs/reference/generators/) — model, resource, auth, jobs, streams.
