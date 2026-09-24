---
title: Jobs em segundo plano
description: "Enfileire trabalho no SQLite: gere jobs, rode um processo worker, enfileire a partir de handlers e gerencie jobs pelo dashboard."
sidebar:
  order: 6
---

O Amarra roda trabalho em segundo plano em uma fila SQLite em `pkg/cais/jobs`, armazenada no mesmo arquivo de banco do seu app — sem Redis, sem serviço extra. Gere seu primeiro job:

```bash
amarra-cais g job send_welcome --cron "0 3 * * *"
```

Isso escreve `internal/jobs/*.go`, registra o handler em `internal/jobs/registry.go` e faz o scaffold de `cmd/worker`. A flag `--cron` agenda uma tarefa recorrente. Depois crie as tabelas `jobs` e `recurring_tasks`:

```bash
amarra-cais db migrate
```

## Rode o worker

Em produção, rode o worker como um processo separado ao lado de `bin/server`:

```bash
amarra-cais jobs work --concurrency 2
```

O worker também roda o dispatcher de jobs atrasados e publica um heartbeat. Restrinja-o a filas específicas com `--queues`:

```bash
amarra-cais jobs work --queues default,mail --concurrency 2
```

Como a fila é um único arquivo SQLite, dois workers ativos escrevendo nele geram um aviso no dashboard.

## Enfileire trabalho a partir de um handler

```go
jobs.Enqueue(ctx, jobStore, jobs.Options{Kind: "SendWelcome", Payload: data})
```

`Kind` deve corresponder ao nome registrado em `internal/jobs/registry.go`. `PruneSessions` já vem embutido, então a limpeza de sessões pode rodar como job em vez de um script cron.

## Inspecione e recupere

O mesmo store expõe a API operacional para scripts e o console:

```go
store.List(ctx, jobs.ListFilter{Status: jobs.StatusFailed, Kind: "SendWelcome"})
store.RetryFailed(ctx, id)
store.Discard(ctx, id)
store.PruneFinished(ctx, 24*time.Hour) // 0 = all finished rows
store.RequeueOrphaned(ctx, jobs.DefaultWorkerStale)
store.ListLiveWorkers(ctx, jobs.DefaultWorkerStale)
```

Pela CLI:

```bash
amarra-cais jobs status      # counts + queues + workers + recurring
amarra-cais jobs retry 12
amarra-cais jobs discard 12
amarra-cais jobs prune --older 24h
```

## O dashboard /jobs

`jobsui.Register` em `app.New` monta `GET /jobs` (somente localhost, em todos os ambientes), com `GET /jobs/{id}` para um job individual e `?kind=` para filtrar. De lá você pode repetir ou descartar jobs com falha, reenfileirar órfãos travados (o que ignora jobs de workers ativos) e limpar linhas finalizadas. Os heartbeats dos workers mostram quais workers estão vivos.

:::tip
Quando o app é remoto, acesse o dashboard por um túnel SSH: `ssh -L 8080:127.0.0.1:8080 user@host`. `amarra-cais routes` lista as rotas `/jobs` quando `jobsui.Register` está presente, e `amarra-cais doctor` avisa quando ela está ausente.
:::

:::note
Tarefas recorrentes são reivindicadas com um compare-and-set a cada tick, então apenas um worker executa um dado agendamento mesmo quando vários workers compartilham o banco.
:::

## Relacionados

- [Jobs e SQLite](/amarra-cais/pt-br/docs/explanation/jobs-and-sqlite/) — por que a fila vive no banco do app.
- [API de jobs](/amarra-cais/pt-br/docs/reference/jobs-api/) — toda a superfície de store e worker.
- [Referência da CLI](/amarra-cais/pt-br/docs/reference/cli/) — `jobs work|status|retry|discard|prune`.
- [Banco de dados e migrations](/amarra-cais/pt-br/docs/how-to/database-and-migrations/) — as migrations adicionam as tabelas da fila.
