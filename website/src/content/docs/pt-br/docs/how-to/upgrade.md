---
title: Atualize o framework
description: Incremente o framework no go.mod, leia o checklist de migração e faça link de um checkout local do framework.
sidebar:
  order: 12
---

`amarra-cais upgrade` move um app para uma release mais nova do framework e informa o que mudou:

```bash
amarra-cais upgrade            # bump to the latest version
amarra-cais upgrade v0.11.0    # pin a specific version
amarra-cais upgrade --dry-run  # preview the changes without writing
```

## O que ele faz

1. Incrementa `github.com/puppe1990/amarra-cais` no `go.mod` (padrão `latest`).
2. Roda novamente `npm install` e `go mod tidy`.
3. Roda `doctor`.
4. Imprime um checklist de migração selecionado para o intervalo de versões que você atravessou.

`--dry-run` mostra o plano sem aplicá-lo, do mesmo jeito que os geradores pré-visualizam suas mudanças.

## Quando rodá-lo

Rode-o depois de uma nova release do framework e leia o checklist antes de fazer deploy. O `CHANGELOG.md` do repositório agrupa cada release em `Added`, `Changed`, `Fixed` e `Security`, e o comando upgrade ajusta o checklist ao intervalo entre sua versão atual e a de destino. Se uma release mudou templates ou HTML, incremente o cache do PWA depois para que os celulares peguem os novos assets:

```bash
amarra-cais pwa --bump
```

## Fixe versões

Mantenha a CLI e o `go.mod` do app em sincronia. O quick start instala uma CLI fixada:

```bash
go install github.com/puppe1990/amarra-cais/cmd/amarra-cais@v0.11.0
amarra-cais version   # expect 0.11.0
```

Gerar código com uma release enquanto depende de outra é a fonte usual de erros confusos de "unknown component" ou "missing method". `amarra-cais version` imprime a versão do framework a partir da qual a CLI foi compilada, para que você possa conferir os dois lados.

## Desenvolva contra um checkout local

`amarra-cais link` adiciona um `replace` no `go.mod` para que um app compile contra sua cópia local do framework:

```bash
amarra-cais link ../amarra-cais   # from an app directory
```

O replace é apenas para desenvolvimento local — não faça commit dele. Desfaça-o antes de dar push:

```bash
amarra-cais link --unlink
```

:::caution
Uma diretiva `replace` commitada fixa o app a um caminho que não existe no seu build server ou no runner de CI. Faça unlink antes de dar push para que o app publicado resolva o módulo a partir do proxy.
:::

:::note
Você também pode apontar para um checkout local através da variável de ambiente `CAIS_REPLACE` em vez de editar o `go.mod` diretamente.
:::

## Relacionados

- [Referência da CLI](/amarra-cais/pt-br/docs/reference/cli/) — `upgrade`, `link` e `version`.
- [Deploy em produção](/amarra-cais/pt-br/docs/how-to/deploy/) — faça redeploy após um upgrade.
- [Instalação](/amarra-cais/pt-br/docs/getting-started/installation/) — como instalar a própria CLI.
- [Layout do projeto](/amarra-cais/pt-br/docs/getting-started/project-layout/) — o que os geradores mantêm à medida que as versões avançam.
