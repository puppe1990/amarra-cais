---
title: Cais vs Amarra
description: Como o produto Cais v0.11.x com Inertia se relaciona com o fork Amarra HTML-first, e quanto custa migrar.
sidebar:
  order: 3
---

Amarra e Cais compartilham uma linhagem, mas entregam frontends diferentes. Saber qual produto você tem indica qual documentação se aplica e quanto uma migração realmente custa.

## A divisão

| Produto      | Frontend                | Observações                      |
| ------------ | ----------------------- | -------------------------------- |
| Cais v0.11.x | Inertia + Svelte (Vite) | Linha congelada, ainda suportada |
| Amarra       | Amarra Views + Drive    | Fork HTML-first do Cais          |

O binário da CLI é `amarra-cais`, e ele não sobrescreve o `cais`. Você pode ter os dois instalados lado a lado. O Cais v0.11.x continua sendo o produto com Inertia; apps gerados pelo Amarra usam HTML renderizado no servidor em `web/templates/` com `/static/js/amarra.js` e sem Vite.

Este não é um fork de dois modos. O `amarra-cais doctor` falha quando encontra um `vite.config.js`, porque um app é um ou outro — não existe a flag `--front inertia`.

## Migrar é reescrever a UI

Mover um app Cais com Inertia para o Amarra não é um `go get`. A camada de dados permanece; a camada de UI é reescrita.

1. Fixe o framework (`go get github.com/puppe1990/amarra-cais@v0.11.0`) e remova o gonertia. Apague `vite.config.js`, `svelte.config.js`, `web/src/` e `web/static/build/`.
2. Carregue as views uma vez no boot com `view.Load`. Handlers recebem um `*view.Renderer` e chamam `view.Write`; eles não checam mais `HX-Request` nem chamam helpers de render do Inertia.
3. Substitua as páginas: `web/src/pages/*.svelte` vira `web/templates/pages/*.html`, `AppLayout.svelte` vira `web/templates/layouts/app.html`, e o mount `#app` vira `#amarra-main` mais um script. Formulários Svelte viram `<.form>` / `<.input>` / `<.button>` do kit.
4. Reescreva os testes para verificar HTML (`testutil.AssertHTMLContains`) em vez de JSON do Inertia. Descarte `setupTestInertia` e `X-Inertia: true`.
5. Mantenha `httpx.ParseFormOrJSON` — o Drive envia corpos form-encoded ou multipart.
6. Use a API de flash via cookie: `flash.Set(w, kind, msg, cfg.CookieSecure())` e então um `303`.
7. Atualize os assets de PWA com `amarra-cais pwa`, depois `amarra-cais pwa --bump` após mudanças nos templates.
8. Renomeie as chamadas de CLI: `cais new` → `amarra-cais new`, `cais g` → `amarra-cais g`, `cais doctor` → `amarra-cais doctor`, `cais pwa` → `amarra-cais pwa`.

O que permanece: seu store, as migrations do SQLite, os jobs e `/jobs`, os pacotes de domínio, os middlewares de session/CSRF/rate-limit e até o nome do cookie de CSRF (`cais_csrf` / `__Host-cais_csrf`). Você está reescrevendo a UI, não a camada de dados.

:::caution
Planeje a migração página por página. Uma ordem sugerida é criar o scaffold com `amarra-cais new scratch`, copiar o layout e os helpers de view-data, e então mover uma página pública (home ou login) de ponta a ponta antes de repetir para o resto.
:::

Continue com o [guia de upgrade](/amarra-cais/pt-br/docs/how-to/upgrade/) e [páginas e views](/amarra-cais/pt-br/docs/how-to/pages-and-views/), ou leia [Views e Drive](/amarra-cais/pt-br/docs/explanation/views-and-drive/) para entender o design por trás do frontend de destino.
