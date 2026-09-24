---
title: Jobs e SQLite
description: Por que um único arquivo SQLite pode servir requests, streams e uma fila de jobs em background ao mesmo tempo.
sidebar:
  order: 5
---

O Amarra usa um único arquivo SQLite para os dados das páginas, as sessions e a fila de jobs em background. Isso é deliberado: um app pequeno em uma única máquina não precisa de Redis nem de um segundo banco de dados, e manter tudo em um arquivo reduz o deploy a um binário mais `web/static`.

## Como um único arquivo se mantém responsivo

O scaffold `NewSQLiteStore` chama `sqlite.Configure`, que define:

- `journal_mode=WAL` — leitores concorrentes prosseguem enquanto um writer segura o lock por um instante.
- `busy_timeout=5000` — um writer bloqueado espera e tenta de novo em vez de falhar imediatamente com `SQLITE_BUSY`.
- `foreign_keys=ON` — a integridade referencial é garantida.
- `MaxOpenConns(1)` — uma única conexão serializa as escritas.

Handlers de SSE consultam o banco em loop, então a regra durante um stream é manter as escritas curtas e evitar transações longas. Uma escrita que segura o lock é o que faz o `busy_timeout` importar. Para concorrência de escrita realmente pesada, direcione as escritas por uma fila dedicada.

## Os jobs compartilham o banco

`pkg/cais/jobs` é uma fila no formato do Solid Queue, sustentada pelo mesmo arquivo. Ela tem `jobs` (ready / running / finished / failed), `scheduled_jobs` para trabalho atrasado que um dispatcher promove quando `run_at <= now`, e `recurring_tasks` guiadas por uma expressão cron. Os workers pegam trabalho com um `UPDATE ... RETURNING` atômico, já que o driver `modernc.org/sqlite` não tem `SKIP LOCKED`.

Enfileire a partir de um handler com `jobs.Enqueue(ctx, store, jobs.Options{Kind: "SendWelcome", Payload: data})` e rode o worker como um processo separado: `amarra-cais jobs work --concurrency 2`. O dispatcher e o heartbeat rodam dentro desse worker.

## Heartbeats e órfãos

Um worker grava um heartbeat (`job_workers`), então `RequeueOrphaned` pode devolver apenas linhas `running` cujo worker está ausente ou obsoleto. Um worker vivo mantém seus jobs em andamento; a recuperação nunca rouba trabalho que ainda está rodando. Se dois heartbeats vivos aparecerem, o dashboard avisa — porque um arquivo SQLite deve ter exatamente um `amarra-cais jobs work`.

## A ressalva de réplica única

Um arquivo de banco de dados significa um host. Várias goroutines via `--concurrency` nesse host são suportadas, mas vários workers em réplicas diferentes não são, e fan-out entre réplicas está fora de escopo. A mesma restrição vale para o Live: o hub de WebSocket faz broadcast apenas no processo, então duas réplicas do app não compartilham sockets. Os jobs também ficam no processo do worker.

## Poda

As sessions expiradas se acumulam, então faça a poda delas com `session.Store.PruneExpired()` ou `amarra-cais db prune-sessions`; o job `PruneSessions`, já incluído, faz isso em um agendamento. Jobs finalizados são limpos por `PruneFinished` (`amarra-cais jobs prune [--older 24h]`), e o worker registra uma recurring task diária `0 4 * * *` para isso.

:::caution
Escalar além de um host significa tirar a fila, o rate limiter e o hub do Live do estado em processo. Até então, mantenha as requests e o worker de jobs na mesma máquina e no mesmo arquivo de banco de dados.
:::

Veja o [how-to de jobs em background](/amarra-cais/pt-br/docs/how-to/background-jobs/), o [guia de live updates](/amarra-cais/pt-br/docs/how-to/live-updates/) e a [referência da API de jobs](/amarra-cais/pt-br/docs/reference/jobs-api/).
