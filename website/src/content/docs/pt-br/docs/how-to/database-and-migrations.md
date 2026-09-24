---
title: Banco de dados e migrations
description: Adicione uma tabela SQLite a um app Amarra, escreva a migration e rode-a pela CLI.
sidebar:
  order: 4
---

Um app Amarra guarda seus dados em SQLite (via `modernc.org/sqlite`, sem CGO). As tabelas são criadas por arquivos de migration SQL ordenados, que o app rastreia em `schema_migrations` e aplica de forma idempotente no boot.

## Adicionar uma tabela

Siga a mesma ordem que os geradores usam — teste primeiro:

1. Escreva um teste de store em `internal/store/` contra o SQLite `:memory:`.
2. Adicione o arquivo SQL `internal/store/migrations/NNN_name.sql`.
3. Adicione métodos à interface `store.Store` e à sua implementação.
4. Envolva o DB com `sqllog.Wrap` em `NewSQLiteStore`, para que os logs de query de desenvolvimento sejam capturados.
5. A migration é registrada em `schema_migrations` por `pkg/cais/migrate`, que é idempotente no boot.

Ou faça o scaffold da camada de dados:

```bash
amarra-cais g model bookmark --fields title:string,url:url
```

`g model` escreve a struct do model, a migration e os métodos de store — sem handlers, templates ou rotas. Para uma migration vazia, use `amarra-cais g migration add_tags`.

## Escrever a migration

Cada arquivo usa os marcadores `-- up` e `-- down`, para que o rollback tenha SQL para rodar:

```sql
-- up
CREATE TABLE bookmarks (
  id    INTEGER PRIMARY KEY,
  title TEXT    NOT NULL,
  url   TEXT    NOT NULL
);

-- down
DROP TABLE IF EXISTS bookmarks;
```

`amarra-cais db rollback` executa o SQL de `-- down` quando presente, depois remove a linha de `schema_migrations`. Sem uma seção down, apenas o registro é removido.

## Rodar migrations

```bash
amarra-cais db migrate        # run pending migrations
amarra-cais db status         # list applied and pending migrations
amarra-cais db rollback       # roll back the last migration
```

Rode `amarra-cais db migrate` depois de `g resource`, `g model` ou `g auth`.

:::note
Os testes de store usam SQLite `:memory:` — não faça mock do banco de dados. Rode-os com `amarra-cais test` ou `go test ./...`.
:::

## Concorrência no SQLite

O `NewSQLiteStore` do scaffold chama `sqlite.Configure`, que define `journal_mode=WAL`, `busy_timeout=5000`, `foreign_keys=ON` e `MaxOpenConns(1)`. O WAL permite que leitores rodem enquanto um escritor segura o lock por um instante, e o `busy_timeout` tenta novamente em vez de falhar de imediato com `SQLITE_BUSY`.

Handlers SSE consultam o banco em loop, então mantenha as escritas curtas e evite transações longas durante um stream. Para concorrência alta de escrita, considere uma fila de escrita dedicada.

## Relacionado

- [Jobs e SQLite](/amarra-cais/pt-br/docs/explanation/jobs-and-sqlite/) — por que SQLite, e como a fila de jobs compartilha o arquivo.
- [Referência da CLI](/amarra-cais/pt-br/docs/reference/cli/) — os subcomandos `db`.
