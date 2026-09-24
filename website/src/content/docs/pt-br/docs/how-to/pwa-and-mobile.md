---
title: PWA e mobile
description: Teste no seu celular pela LAN, incremente o cache do service worker e passe no checklist do doctor mobile.
sidebar:
  order: 9
---

Apps Amarra são entregues como PWAs instaláveis: manifest, service worker, página offline, ícones e fullscreen. Alguns comandos mantêm essa experiência atualizada enquanto você desenvolve em um celular.

## Teste no seu celular pela LAN

`boot.Print` mostra as URLs **LAN** nas quais o servidor está acessível para testes no celular via Wi‑Fi. Quando o banner já rolou para fora da tela, `GET /health` retorna a mesma lista como `lan_urls` através de `netutil.HealthPayload`:

```bash
curl -s http://localhost:8080/health
```

Use o array `lan_urls` — nunca concatene `APP_URL` e a porta à mão. Em produção `lan_urls` é omitido, já que o app só é acessível no seu host público.

## Incremente o cache do service worker

O service worker faz cache do seu CSS, JS e páginas. Depois de mudar templates ou HTML, incremente o cache dele para que os celulares peguem os novos assets:

```bash
amarra-cais pwa --bump
```

Isso atualiza os assets vendored e incrementa `CACHE_VERSION` em `sw.js`. `amarra-cais dev` incrementa o cache automaticamente sempre que `sw.js` existe, então a iteração local fica atualizada sem um passo manual. O service worker é network-first para `/static/js/amarra.js` e `/static/css/`.

## Rode as verificações mobile

```bash
amarra-cais doctor --mobile
```

O doctor mobile verifica o markup de flash, o CSP de Google Fonts, `amarra.js`, o cache do service worker (network-first `/static/js/amarra.js`), o CSS do formulário de chat, o container de scroll `#chat-messages` e o `lan_urls` do health.

:::caution
O CSP padrão bloqueia `fonts.googleapis.com`. O `input.css` do scaffold usa fontes do sistema por esse motivo — mantenha assim a menos que você amplie o CSP de propósito.
:::

## Assets de marca

Os scaffolds entregam um tile **placeholder neutro** (`pkg/cais/pwa/assets/icon.png`, 512×512) mais um `icons/icon-512-maskable.png` gerado. O manifest separa os ícones `any` (192 + 512) do ícone `maskable`. `amarra-cais doctor` avisa enquanto `web/static/icons/*` e `og.png` ainda forem os placeholders entregues. Substitua-os pela sua própria marca antes de publicar; `amarra-cais pwa --force` restaura os assets de marca para os padrões se você precisar de uma folha em branco.

## Checklist mobile

1. `amarra-cais doctor --mobile`
2. `amarra-cais pwa --bump`
3. Abra a URL **LAN** do boot no seu celular.
4. Permaneça na página enquanto um stream SSE roda — o Drive faz morph de `#amarra-main`, e uma navegação completa derruba o `EventSource`.

:::tip
Se o celular mostrar uma página obsoleta, force o fechamento do PWA instalado para que o service worker busque os assets novamente após o incremento do cache.
:::

## Relacionados

- [Chat em streaming](/amarra-cais/pt-br/docs/how-to/streaming-chat/) — por que uma página ao vivo não deve navegar no meio do stream.
- [Deploy em produção](/amarra-cais/pt-br/docs/how-to/deploy/) — entregue os assets do PWA ao lado do binário.
- [Referência da CLI](/amarra-cais/pt-br/docs/reference/cli/) — `pwa`, `doctor` e `dev`.
- [Seu primeiro app](/amarra-cais/pt-br/docs/getting-started/your-first-app/) — o scaffold que entrega essa configuração de PWA.
