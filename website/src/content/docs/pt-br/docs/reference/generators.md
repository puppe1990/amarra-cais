---
title: Generators
description: Referência de amarra-cais g e destroy — generators, opções de resource, tipos de campo e os arquivos que cada um escreve.
sidebar:
  order: 3
---

Generators são a família `amarra-cais g`. Eles rodam dentro de um app gerado (a CLI verifica se há um `go.mod` que depende de `amarra-cais`) e escrevem scaffolds HTML + Amarra, aplicando patches em `routes.go`, `store.go`, `seeds.go` e na nav do layout conforme necessário.

## Generators

| Generator                           | Arquivos e patches                                                                                                                                                                                                                                                                               |
| ----------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `g handler <name>`                  | `internal/handlers/<name>.go` + `<name>_test.go`, `web/templates/pages/<name>.html` e uma rota adicionada a `internal/app/routes.go`.                                                                                                                                                            |
| `g page <name>`                     | Apenas `web/templates/pages/<name>.html`.                                                                                                                                                                                                                                                        |
| `g resource <name>`                 | `internal/models/<name>.go`; `internal/handlers/admin_<plural>.go` (+`_test.go`); `web/templates/pages/admin_<plural>.html`, `admin_<name>_show.html`, `admin_<name>_form.html`; `web/templates/partials/admin_<name>_form_errors.html`; uma nova migration; patches de store, rota, seed e nav. |
| `g model <name>`                    | `internal/models/<name>.go`, uma migration e métodos de store — sem handlers, templates ou rotas.                                                                                                                                                                                                |
| `g migration <name>`                | `internal/store/migrations/NNN_<name>.sql` (próximo número da sequência).                                                                                                                                                                                                                        |
| `g auth`                            | Handlers, páginas e testes de login/logout, um `/dashboard` protegido e o middleware de sessão configurado em `app.go`.                                                                                                                                                                          |
| `g console`                         | `cmd/console/main.go`.                                                                                                                                                                                                                                                                           |
| `g ci`                              | Workflow do GitHub Actions, hooks de pre-commit, config do golangci-lint e Prettier (aplica patches em `Makefile` e `package.json`).                                                                                                                                                             |
| `g job <name> [--cron "0 3 * * *"]` | `internal/jobs/<name>.go` (+`_test.go`), `internal/jobs/registry.go`, e `cmd/worker/main.go`.                                                                                                                                                                                                    |
| `g stream chat [--live]`            | Models de conversa/mensagem, `internal/handlers/chat.go` (+`_test.go`), `conversations.html` / `chat.html`, e os partials `chat_sse*.html`.                                                                                                                                                      |
| `g live <name>`                     | `internal/handlers/<name>_live.go` (+`_test.go`) e `web/templates/pages/<name>.html`.                                                                                                                                                                                                            |
| `g component <name>`                | `web/templates/components/<stem>.html`. Um nome do kit que acompanha o framework faz o seed do markup real desse componente; outros nomes recebem um slot genérico.                                                                                                                              |

A CLI também inclui `g sitemap` (posts de blog + um `/sitemap.xml` dinâmico).

## Opções de resource

`amarra-cais g resource bookmark --fields title:string,url:url,notes:text? --public --paginate`

| Opção                          | Efeito                                                                                        |
| ------------------------------ | --------------------------------------------------------------------------------------------- |
| `--fields <spec>`              | Lista de campos separados por vírgula. O padrão é `name:string`.                              |
| `--public`                     | Também gera uma página de listagem pública (kit `<.filters>` / `<.empty>` / `<.pagination>`). |
| `--paginate`                   | Pagina o index do admin, 25 linhas por página.                                                |
| `--no-seed`                    | Pula os dados de demonstração `SeedDemo*`.                                                    |
| `--admin-auth session\|bearer` | Modo de proteção do admin: sessão de navegador (padrão) ou bearer token.                      |
| `--force`                      | Sobrescreve arquivos que já existem.                                                          |

Os métodos de store gerados recebem `(search, sort, dir string, ...)`; o sort é permitido por coluna via whitelist e a busca é um `LIKE` no campo de exibição. O index do admin renderiza `<.filters>` + `<.table>` + `<.empty>` + `<.pagination>`.

## Tipos de campo

Cada campo é `name:type`; acrescente `?` para torná-lo opcional. `name:belongs_to` é um atalho para `name_id:references`.

| Tipo         | Coluna                                      | Input                         |
| ------------ | ------------------------------------------- | ----------------------------- |
| `string`     | `TEXT [NOT NULL]`                           | input de texto                |
| `text`       | `TEXT` (textarea)                           | textarea                      |
| `url`        | `TEXT [NOT NULL]`                           | input `type="url"`            |
| `bool`       | `INTEGER NOT NULL DEFAULT 0`                | checkbox                      |
| `int`        | `INTEGER [NOT NULL DEFAULT 0]`              | input numérico                |
| `float`      | `REAL [NOT NULL DEFAULT 0]`                 | input numérico (`step="any"`) |
| `date`       | `TEXT [NOT NULL]`                           | input `type="date"`           |
| `references` | `INTEGER [NOT NULL] REFERENCES <table>(id)` | select (linha pai)            |

Um campo opcional é mapeado para um ponteiro (`*string`, `*int64`, `*float64`). Para um campo `references`, gere o resource pai primeiro e dê à tabela pai uma coluna `name` ou `title` para os rótulos.

```bash
amarra-cais g resource post --fields title:string,category_id:references
# or: --fields title:string,category:belongs_to
```

## Overrides de componente

`amarra-cais g component --list` imprime os componentes do kit que acompanha o framework e que um app pode sobrescrever. A lista fica no lado do framework, então funciona sem um app. Sobrescrever um nome do kit faz o seed do markup real, de modo que você reestiliza o contrato em vez de recriá-lo; o arquivo mantém o stem do kit (por exemplo, `locale-toggle.html`).

## Dry-run

`amarra-cais g --dry-run …` e `amarra-cais destroy --dry-run …` imprimem as mudanças planejadas sem escrever arquivos.

## Destroy

`amarra-cais destroy resource|handler|model|component <name>` remove os arquivos gerados e desfaz os patches de `routes.go`, `store.go`, `seeds.go` e da nav do layout quando aplicável. `destroy auth` também reverte o middleware de sessão do `app.go`. `destroy migration <name>` remove apenas o `*_<name>.sql` correspondente — não reverte `schema_migrations`.

:::caution
Depois de `g resource`, `g model` ou `g auth`, rode `amarra-cais db migrate`. Se um patch falhar com `could not patch routes.go` ou `could not patch store`, verifique se os marcadores `registerRoutes` e `Close() error` ainda existem; a nav pública precisa do marcador `<!-- cais:nav -->` (ou `</nav>`).
:::

Veja [Banco de dados e migrations](/amarra-cais/pt-br/docs/how-to/database-and-migrations/) para o SQL que o generator de resource escreve e [CLI](/amarra-cais/pt-br/docs/reference/cli/) para a lista completa de comandos.
