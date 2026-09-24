---
title: Autenticação e sessões
description: Proteja rotas com sessões de cookie respaldadas por SQLite, escolha entre modo admin de sessão ou bearer e faça a limpeza de logins expirados.
sidebar:
  order: 5
---

O Amarra entrega sessões baseadas em cookie respaldadas por SQLite. Todo app criado com `amarra-cais new` já inclui login, logout e um `/dashboard` protegido. Adicione o mesmo scaffolding a um app existente com `amarra-cais g auth` e depois rode `amarra-cais db migrate` para criar as tabelas de sessão.

## Configure o middleware de sessão

Registre `LoadSession` antes de qualquer coisa que leia o usuário atual, e `Flash` logo em seguida para que avisos one-shot sobrevivam aos redirects:

```go
r.Use(middleware.LoadSession(deps.Store.Sessions()))
r.Use(middleware.Flash(cfg))
r.Get("/dashboard", middleware.RequireAuthFunc("/login", dashboard.ServeHTTP))
```

`RequireAuthFunc` redireciona visitantes não autenticados para o caminho de login que você passar. Envolva um prefixo inteiro com `middleware.RequireAuth("/login")` quando todas as rotas abaixo dele estiverem protegidas.

## Faça o login de um usuário

Depois de validar as credenciais, chame `session.SignIn` com o store respaldado por SQLite e defina um flash notice:

```go
session.SignIn(w, sessions, r, userID, session.CookieOptionsFromConfig(cfg))
flash.Set(w, "notice", "Welcome back!", cfg.CookieSecure())
```

`session.NewSQLiteStore` persiste as sessões no arquivo SQLite do app, então os logins sobrevivem a um restart. Faça logout com `session.SignOut`. A sessão gira no login, o que invalida o token anterior — uma defesa contra fixation. Em um scaffold novo, o usuário de seed de desenvolvimento é `demo@example.com` / `password`.

## Expiração e limpeza

Cookies e linhas do banco expiram após 7 dias (`sessionTTL` / `defaultMaxAge`). O SQLite armazena `expires_at`, e as linhas expiradas são ignoradas na consulta. Delete linhas obsoletas periodicamente:

```bash
amarra-cais db prune-sessions
```

Você também pode chamar `session.Store.PruneExpired()` do seu próprio código de manutenção; o job `PruneSessions` embutido faz o mesmo de forma agendada.

## Cookies em produção

`session.CookieOptionsFromConfig(cfg)` define `Secure` quando `cfg.CookieSecure()` é true, o que acontece sob `ENV=production`. Sempre construa as opções de cookie a partir da config para que o tráfego de produção nunca receba um cookie não-Secure.

## Proteja áreas de admin

Escolha um de dois modos:

| Modo                        | Middleware                         | Flag do gerador                 |
| --------------------------- | ---------------------------------- | ------------------------------- |
| Admin no navegador (padrão) | `middleware.RequireAuth("/login")` | `amarra-cais g resource` padrão |
| API com token bearer        | `middleware.AdminAuth(cfg)`        | `--admin-auth bearer`           |

`amarra-cais g resource` usa autenticação de sessão por padrão (`--admin-auth session`). Escolha `--admin-auth bearer` para APIs de admin somente com token, que não têm páginas de login.

`AdminAuth` aceita apenas um header Bearer — ele não lê query params. Defina `ADMIN_TOKEN` em produção; `cfg.Validate()` falha no boot quando ele está ausente. Quando `ADMIN_TOKEN` não está definido, o middleware é um no-op, que é o comportamento em desenvolvimento.

:::caution
Trate `ADMIN_TOKEN` como um segredo de produção. Armazene-o no ambiente do processo (por exemplo, `/etc/myapp/env` com `chmod 600`) e nunca faça commit dele.
:::

:::tip
A ordem do middleware importa: `LoadSession` → `Flash` → `CSRF` → `SecurityHeaders`. Veja a [Middleware reference](/amarra-cais/pt-br/docs/reference/middleware/) para a stack completa.
:::

## Relacionados

- [Banco de dados e migrations](/amarra-cais/pt-br/docs/how-to/database-and-migrations/) — as tabelas onde as sessões vivem.
- [Deploy em produção](/amarra-cais/pt-br/docs/how-to/deploy/) — os gates de produção que exigem `ADMIN_TOKEN`.
- [Modelo de segurança](/amarra-cais/pt-br/docs/explanation/security-model/) — como CSRF, cookies e sessões interagem.
- [Referência da CLI](/amarra-cais/pt-br/docs/reference/cli/) — `g auth`, `db migrate` e `db prune-sessions`.
