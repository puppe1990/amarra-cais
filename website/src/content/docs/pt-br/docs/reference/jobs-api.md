---
title: Jobs API
description: A API Go e o dashboard de localhost para a fila de jobs em background no SQLite do Amarra.
sidebar:
  order: 10
---

`pkg/cais/jobs` é uma fila de jobs com suporte a SQLite que fica no mesmo arquivo de banco de dados do app — sem Redis. Enfileire trabalho a partir dos handlers, rode um processo worker e inspecione ou recupere jobs pelo dashboard de localhost e pela CLI.

## Enfileirar

```go
jobs.Enqueue(ctx, jobStore, jobs.Options{Kind: "SendWelcome", Payload: data})
```

`jobs.Options` carrega `Kind` (o nome do handler) e `Payload` (os dados que o handler recebe).

## Métodos de store

Registre os handlers em `internal/jobs/registry.go`; o handler embutido é `PruneSessions`. O store expõe métodos de inspeção e recuperação:

```go
store.List(ctx, jobs.ListFilter{Status: jobs.StatusFailed, Kind: "SendWelcome"})
store.RetryFailed(ctx, id)
store.Discard(ctx, id)
store.PruneFinished(ctx, 24*time.Hour) // 0 = all finished rows
store.RequeueOrphaned(ctx, jobs.DefaultWorkerStale)
store.ListLiveWorkers(ctx, jobs.DefaultWorkerStale)
```

| Método                                     | Finalidade                                                                                    |
| ------------------------------------------ | --------------------------------------------------------------------------------------------- |
| `List(ctx, jobs.ListFilter{Status, Kind})` | Lista jobs, opcionalmente filtrados por status e kind.                                        |
| `RetryFailed(ctx, id)`                     | Reenfileira um job com falha.                                                                 |
| `Discard(ctx, id)`                         | Descarta um job com falha.                                                                    |
| `PruneFinished(ctx, older)`                | Exclui linhas finalizadas mais antigas que a duração (`0` limpa todas as linhas finalizadas). |
| `RequeueOrphaned(ctx, stale)`              | Reenfileira jobs presos em um worker morto.                                                   |
| `ListLiveWorkers(ctx, stale)`              | Lista os heartbeats de worker vistos dentro da janela de staleness.                           |

`jobs.StatusFailed` é a constante de status usada para filtrar jobs com falha. `jobs.DefaultWorkerStale` é a janela padrão de staleness do pacote para detecção de órfãos e vivacidade de worker.

## Worker

```bash
amarra-cais jobs work --queues default,mail --concurrency 2
```

O worker executa os jobs mais o dispatcher de jobs atrasados e escreve um heartbeat para que o dashboard possa mostrar a vivacidade. Flags: `--queues` (lista separada por vírgulas) e `--concurrency` (número de workers). Em produção, rode-o como um processo separado ao lado de `bin/server`.

```bash
amarra-cais jobs status            # counts + queues + workers + recurring
amarra-cais jobs retry 12
amarra-cais jobs discard 12
amarra-cais jobs prune [--older 24h]
```

:::caution
Dois workers vivos em um mesmo arquivo SQLite disparam um aviso — os heartbeats são como o dashboard os detecta.
:::

## Tarefas recorrentes

`amarra-cais g job` faz o scaffold do handler, da entrada no registry e do comando do worker:

```bash
amarra-cais g job prune_sessions --cron "0 3 * * *"
amarra-cais db migrate      # creates the jobs + recurring_tasks tables
```

A expressão cron registra uma tarefa recorrente, e `amarra-cais db migrate` é obrigatório após a geração para criar as tabelas `jobs` e `recurring_tasks`.

## Dashboard

`jobsui.Register(r, db)` monta o dashboard (somente localhost, em todos os ambientes; use um túnel SSH quando estiver remoto). `amarra-cais doctor` avisa se `jobsui.Register` estiver faltando.

| Rota             | Finalidade                                      |
| ---------------- | ----------------------------------------------- |
| `GET /jobs`      | Lista e contagens de jobs. Filtre com `?kind=`. |
| `GET /jobs/{id}` | Detalhe de um job.                              |

No dashboard você pode fazer Retry ou Discard de jobs com falha, Requeue de órfãos presos (pulando jobs cujo worker ainda está vivo) e limpar jobs finalizados. Essas ações correspondem aos métodos de store acima: Retry chama `RetryFailed`, Discard chama `Discard`, Requeue stuck chama `RequeueOrphaned` e Clear finished chama `PruneFinished`. Tarefas agendadas e recorrentes também aparecem lá. Acesse `http://127.0.0.1:<port>/jobs`.

`amarra-cais routes` lista as rotas do dashboard quando `jobsui.Register` está presente.

Veja [Jobs em background](/amarra-cais/pt-br/docs/how-to/background-jobs/) e o [design de jobs e SQLite](/amarra-cais/pt-br/docs/explanation/jobs-and-sqlite/).
