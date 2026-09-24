---
title: CLI
description: Todos os comandos e aliases do amarra-cais — scaffold, generate, destroy, dev, build, banco de dados, jobs e targets make do framework.
sidebar:
  order: 2
---

`amarra-cais` é a CLI no estilo Rails para apps Amarra. Execute os comandos de app de dentro de um app gerado (os que precisam disso leem o `go.mod` do diretório atual); `new`, `version`, `help` e `g component --list` funcionam em qualquer lugar.

## Scaffold, geração e desfazer

| Comando                                                            | Descrição                                                                                                                                                         |
| ------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `amarra-cais new <app> [dir] [--minimal\|--blank] [--module path]` | Faz o scaffold de um app (diretório padrão `./<app>`). `--minimal` traz apenas a home, `--blank` não tem conteúdo inicial e `--module` sobrescreve o module path. |
| `amarra-cais g [--dry-run] <generator> [name]`                     | Executa um generator. Veja [Generators](/amarra-cais/pt-br/docs/reference/generators/).                                                                           |
| `amarra-cais destroy [--dry-run] <kind> <name>`                    | Desfaz um generator.                                                                                                                                              |

### Subcomandos de `amarra-cais g`

| Generator                                   | O que faz                                                                           |
| ------------------------------------------- | ----------------------------------------------------------------------------------- |
| `g handler <name>`                          | Handler + teste + `web/templates/pages/<name>.html` + patch de rota.                |
| `g page <name>`                             | Apenas o template da página.                                                        |
| `g resource <name> [options]`               | CRUD de admin baseado no kit, opcionalmente com uma listagem pública.               |
| `g model <name> [--fields …]`               | Model + migration + métodos de store (sem handlers, templates ou rotas).            |
| `g migration <name>`                        | Novo `NNN_<name>.sql`.                                                              |
| `g auth`                                    | Login/logout e um `/dashboard` protegido.                                           |
| `g console`                                 | `cmd/console/main.go`.                                                              |
| `g ci`                                      | Workflow de CI, pre-commit, lint e Prettier.                                        |
| `g job <name> [--cron "0 3 * * *"]`         | Handler de job, registry e worker.                                                  |
| `g stream chat [--live]`                    | Scaffold de chat via SSE.                                                           |
| `g live <name>`                             | Live view + página (WebSocket).                                                     |
| `g component <name>` / `g component --list` | Faz o seed de um override de kit / lista os componentes que podem ser sobrescritos. |
| `g sitemap`                                 | Resource `post` público (blog) mais uma rota dinâmica `/sitemap.xml`.               |

### Subcomandos de `amarra-cais destroy`

| Comando                    | O que desfaz                                                              |
| -------------------------- | ------------------------------------------------------------------------- |
| `destroy resource <name>`  | Arquivos do resource e patches de rota/store/seeds/nav.                   |
| `destroy handler <name>`   | Handler, teste, página e patch de rota.                                   |
| `destroy model <name>`     | Model, migration e métodos de store.                                      |
| `destroy auth`             | Scaffolding de login/auth e o middleware de sessão do `app.go`.           |
| `destroy migration <name>` | Apenas o `*_<name>.sql` correspondente (não reverte `schema_migrations`). |
| `destroy component <name>` | O arquivo de override do kit.                                             |

Adicione `--dry-run` a qualquer um dos comandos para imprimir as mudanças planejadas sem escrever arquivos.

## Ciclo de vida do app

| Comando                                                   | Descrição                                                          |
| --------------------------------------------------------- | ------------------------------------------------------------------ |
| `amarra-cais install`                                     | `npm install` + `go mod tidy` (e um build do Tailwind).            |
| `amarra-cais css`                                         | Compila o Tailwind CSS → `web/static/css/styles.css`.              |
| `amarra-cais dev`                                         | Hot reload (air + Tailwind watch).                                 |
| `amarra-cais build [--os linux] [--arch amd64] [-o path]` | Compila `bin/server`, opcionalmente com cross-compile para deploy. |
| `amarra-cais server`                                      | Executa o app (`go run ./cmd/server`).                             |
| `amarra-cais test`                                        | Executa os testes (`go test ./...`).                               |
| `amarra-cais console`                                     | REPL no estilo Rails (`store`, `cfg`, `db` + SQL).                 |
| `amarra-cais routes [--verbose]`                          | Lista as rotas HTTP de `internal/app/routes.go`.                   |
| `amarra-cais version`                                     | Imprime a versão do framework.                                     |

## Banco de dados

| Comando                                              | Descrição                                                                |
| ---------------------------------------------------- | ------------------------------------------------------------------------ |
| `amarra-cais db migrate`                             | Executa as migrations pendentes.                                         |
| `amarra-cais db status`                              | Lista as migrations aplicadas e pendentes.                               |
| `amarra-cais db rollback`                            | Reverte a última migration (executa o SQL de `-- down` quando presente). |
| `amarra-cais db prune-sessions`                      | Exclui as sessões de login expiradas.                                    |
| `amarra-cais db seed` / `amarra-cais db seed --list` | Executa `internal/db/seeds.go` / lista os helpers de seed referenciados. |

## Jobs em background

| Comando                                                           | Descrição                                                       |
| ----------------------------------------------------------------- | --------------------------------------------------------------- |
| `amarra-cais jobs work [--queues default,mail] [--concurrency 2]` | Executa o worker, o dispatcher de jobs atrasados e o heartbeat. |
| `amarra-cais jobs status`                                         | Contagens, filas, workers e tarefas recorrentes.                |
| `amarra-cais jobs retry <id>` / `jobs discard <id>`               | Atua sobre um job com falha.                                    |
| `amarra-cais jobs prune [--older 24h]`                            | Exclui os jobs finalizados.                                     |

A fila fica no arquivo SQLite do app; o dashboard `/jobs` é servido no próprio processo. Veja [Jobs em background](/amarra-cais/pt-br/docs/how-to/background-jobs/).

## Diagnóstico e ferramentas

| Comando                                     | Descrição                                                                                                                           |
| ------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| `amarra-cais doctor [--mobile]`             | Verifica `amarra.js`, `layouts/app.html`, air, `go.mod` e a configuração de PWA/mobile.                                             |
| `amarra-cais pwa [--bump] [--force]`        | Atualiza os assets de PWA; `--bump` incrementa o cache do SW e `--force` reinicia a marca.                                          |
| `amarra-cais upgrade [version] [--dry-run]` | Atualiza o framework, executa o `doctor` e imprime os passos de migração.                                                           |
| `amarra-cais link [path] [--unlink]`        | Adiciona um `replace` no `go.mod` para desenvolvimento local do framework. Não faça commit disso; use `--unlink` antes de dar push. |

## Aliases

| Alias           | Comando    |
| --------------- | ---------- |
| `amarra-cais g` | `generate` |
| `amarra-cais i` | `install`  |
| `amarra-cais b` | `build`    |
| `amarra-cais s` | `server`   |
| `amarra-cais c` | `console`  |

## Comandos do framework (este repositório)

Estes targets `make` rodam dentro do próprio repositório do framework Amarra, e não em um app gerado:

| Comando            | Descrição                                                     |
| ------------------ | ------------------------------------------------------------- |
| `make test`        | Testes Go com `-race`.                                        |
| `make js-test`     | Testes unitários de `pkg/cais/js` + `pkg/amarra/js`.          |
| `make lint`        | `golangci-lint`.                                              |
| `make format`      | `prettier --write`.                                           |
| `make ci`          | `test` + `js-test` + `lint` + format-check — o gate completo. |
| `make build`       | Compila `bin/amarra-cais`.                                    |
| `make install-cli` | `go install ./cmd/amarra-cais`.                               |

:::tip
Para um passo a passo dos comandos que a maioria dos apps usa, veja [Upgrade](/amarra-cais/pt-br/docs/how-to/upgrade/) e [Deploy](/amarra-cais/pt-br/docs/how-to/deploy/).
:::
