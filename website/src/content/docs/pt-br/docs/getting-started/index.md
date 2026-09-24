---
title: Começando
description: O que é o Amarra, como as peças se encaixam e para onde ir em seguida.
sidebar:
  order: 1
---

Amarra é o framework Go HTML-first por trás da CLI `amarra-cais`. Ele traz um scaffold para mini apps fáceis de hospedar em uma VPS pequena — amigável para Lightsail — e fáceis de manter na cabeça.

## O que você recebe

| Camada         | Escolha                                                           |
| -------------- | ----------------------------------------------------------------- |
| Linguagem      | Go 1.26 (stdlib `net/http`)                                       |
| Frontend       | Amarra Views + Drive (`pkg/amarra/view` + `/static/js/amarra.js`) |
| CSS            | Tailwind CSS 3.x                                                  |
| Banco de dados | SQLite (`modernc.org/sqlite`, sem CGO)                            |
| PWA            | Manifest, service worker, página offline, ícones, fullscreen      |
| Núcleo         | Router, sessões, CSRF, jobs, i18n em `pkg/cais/`                  |

O navegador não monta um SPA. Os handlers chamam `view.Write` e o Drive faz morph de `#amarra-main` — sem Vite, sem Svelte, sem Inertia, sem HTMX nos apps gerados.

UI repetitiva vem do **kit + hooks** que acompanha o framework — `<.table>`, `<.filters>`, `<.stat>`, `<.empty>`, `<.password>`, além de builtins do `amarra-hook` como `dialog`, `dropdown`, `bulk`, `nav`, `theme` e `password`. Ordenação e filtro são round-trips `GET ?q=&sort=` simples; o `amarra-cais g resource` emite tudo isso.

## Para onde ir em seguida

- [Instalação](/amarra-cais/pt-br/docs/getting-started/installation/) — instale a CLI e confira a versão.
- [Seu primeiro app](/amarra-cais/pt-br/docs/getting-started/your-first-app/) — faça o scaffold, rode o servidor de dev e faça login.
- [Estrutura do projeto](/amarra-cais/pt-br/docs/getting-started/project-layout/) — o que o `amarra-cais new` escreve e por quê.
- [Guias práticos](/amarra-cais/pt-br/docs/how-to/) — receitas focadas em tarefas: formulários, auth, jobs, streaming, deploy.
- [Referência](/amarra-cais/pt-br/docs/reference/) — CLI, geradores, router, kit, middleware, API de jobs.
- [Explicação](/amarra-cais/pt-br/docs/explanation/) — por que HTML-first, o Drive e um único arquivo SQLite.
