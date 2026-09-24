---
title: Instalação
description: Instale a CLI amarra-cais a partir de uma tag de release ou do código-fonte.
sidebar:
  order: 2
---

## Requisitos

- Go 1.26 ou mais recente (o framework é construído sobre a stdlib `net/http`).
- Node.js 22+ para as ferramentas de Tailwind dentro dos apps gerados.
- O SQLite é embutido via `modernc.org/sqlite` — sem CGO, sem pacote do sistema.

## Instale a CLI

```bash
export PATH="$HOME/go/bin:$PATH"
go install github.com/puppe1990/amarra-cais/cmd/amarra-cais@v0.11.0
amarra-cais version   # expect 0.11.0
```

O `amarra-cais` é instalado ao lado do `cais` — ele nunca sobrescreve o binário existente. O Cais v0.11.x continua sendo o produto Inertia + Svelte; veja [Cais vs Amarra](/amarra-cais/pt-br/docs/explanation/cais-vs-amarra/).

## Instale a partir do código-fonte

Use isto quando estiver trabalhando no próprio framework:

```bash
git clone https://github.com/puppe1990/amarra-cais.git
cd amarra-cais
make install-cli
```

## Rode as verificações do framework

```bash
make test      # go test ./... -race
make js-test   # pkg/cais/js + pkg/amarra/js unit tests
make ci        # test + js-test + lint + format-check
```

Para apontar um app para um checkout local em vez de uma versão lançada, use `amarra-cais link .` no diretório do app. Lembre-se de executar `--unlink` antes de fazer push — a diretiva `replace` não deve ser commitada.

Próximo: [Seu primeiro app](/amarra-cais/pt-br/docs/getting-started/your-first-app/).
