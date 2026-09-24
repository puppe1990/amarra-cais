---
title: Views e Drive
description: Por que o Amarra renderiza HTML completo no servidor e deixa o Drive fazer o morph da página em vez de enviar uma SPA.
sidebar:
  order: 2
---

O frontend do Amarra é HTML-first. Handlers renderizam HTML completo no servidor, e um pequeno script no navegador transforma a navegação e o envio de formulários em uma atualização parcial. Nenhuma SPA é montada, e nenhum framework de componentes no lado do cliente é dono da página.

## O servidor renderiza a página inteira

Todo handler chama `view.Write`. O boot carrega os templates uma vez com `view.Load`; um `<.component>` desconhecido falha no boot, e não na primeira request. Isso significa que um primeiro carregamento, um `curl` e um crawler recebem todos o mesmo documento completo — layout, página e dados já renderizados.

O layout define `#amarra-nav`, `#amarra-main` e `#amarra-toast-host`, e carrega um único script, `/static/js/amarra.js`.

## O Drive aprimora a navegação

O Drive intercepta por padrão todos os cliques e envios de mesma origem — um `<a>` simples, um `<form>`, `{{ linkTo }}` ou um `<.form>`. Não existe atributo de opt-in; você faz opt-out com `data-amarra-skip`. Cada interação vira um `fetch` carregando `Amarra-Drive: true` e o header CSRF. A resposta faz o morph de `#amarra-main` e faz push do estado do histórico.

Requests do Drive ainda renderizam o layout para que o alvo do morph exista e as flash messages sobrevivam à troca. Uma frame request (`Amarra-Frame: <id>`) é mais estreita: ela renderiza apenas o bloco `{{ define "frame:<id>" }}`, um fragmento para uma região da página.

## Por que não uma SPA

Não há Inertia, nem Svelte, nem Vite nos apps gerados. O navegador carrega um script e nunca monta uma árvore de componentes. O servidor é dono do roteamento, dos dados e da marcação, então você depura uma única linguagem e faz o deploy sem etapa de bundler. Crawlers, `curl` e o primeiro paint veem a página real.

UI repetitiva não é trabalho de um framework aqui. Ela vem do kit incluído — `<.form>`, `<.input>`, `<.table>`, `<.filters>`, `<.stat>`, `<.empty>`, `<.password>` — além dos builtins do `amarra-hook`, como `dialog`, `dropdown`, `bulk`, `nav`, `theme` e `password`. Um app reestiliza um contrato real jogando um arquivo de mesmo nome em `web/templates/components/` em vez de recriar um widget. Ordenação e filtro são round-trips simples de `GET ?q=&sort=` que o servidor re-renderiza.

:::note
O shell da sidebar fica fora de `#amarra-main`, então um morph do Drive nunca o perturba. `amarra-hook="nav"` com `data-amarra-nav-on` / `data-amarra-nav-off` ressincroniza o link ativo após cada morph.
:::

## Os trade-offs

HTML-first não sai de graça, e o Amarra deixa os custos explícitos:

- **Cache de listas** — uma página de lista estável pode cachear seu markup renderizado com `cache.Key` e `cache.Hash(version)`, e então servir um `304` via `httpx.NotModified` / `httpx.SetETag`. Essa é a resposta pretendida para renderização por request, não um cache no cliente.
- **Estabilidade do morph** — o Drive faz morph casando elementos, então ids e estrutura precisam permanecer estáveis entre as respostas.
- **Sem fila de mutações offline** — um POST do Drive offline não é enfileirado. A UI mostra um toast em vez de tentar de novo silenciosamente.
- **Estado no lado do servidor** — Drive e Frame não guardam estado no cliente; cada interação é um round-trip HTTP contra o SQLite.

Para a superfície de API por trás disso, veja a [referência de views e kit](/amarra-cais/pt-br/docs/reference/views-and-kit/) e a [referência do amarra.js](/amarra-cais/pt-br/docs/reference/amarra-js/).
