---
title: Páginas e views
description: Adicione uma página a um app Amarra gerado pelo scaffold e renderize-a com view.Write, writeView e amarraData.
sidebar:
  order: 2
---

Amarra renderiza HTML no servidor, então o navegador não monta uma SPA. Os handlers chamam `view.Write`, e o Drive faz o morph do elemento `#amarra-main` do layout no lugar. Os templates carregam uma vez no boot, então uma tag `<.component>` desconhecida falha na inicialização em vez de na primeira requisição.

## Adicionar uma página

O caminho mais rápido é o gerador:

```bash
amarra-cais g handler foo
```

Isso cria o handler, seu teste, `web/templates/pages/foo.html` e faz patch de `internal/app/routes.go`. `amarra-cais g page about` faz o mesmo para uma página simples.

Para ligar tudo na mão, siga os mesmos quatro passos:

1. Teste Go em `internal/handlers/foo_test.go` — use `setupTestViews(t)` e verifique o HTML renderizado (`id="amarra-main"`, campos de formulário).
2. Página HTML em `web/templates/pages/foo.html`, definindo um bloco `{{ define "content" }}`.
3. Handler que chama `view.Write`, ou `writeView` quando a página pode falhar na validação.
4. Registre a rota em `internal/app/routes.go`.

## Renderizar com view.Write

```go
view.Write(w, r, h.views, view.Page{
  Layout: "app",
  Name:   "contact",
  Data: map[string]any{
    "Title":     h.catalog.T("contact.title"),
    "Site":      meta.ForRequest(h.site, r),
    "CSRFToken": csrf.TokenFromRequest(r),
    "Flash":     flashMsg,
  },
}, h.cfg)
```

`Layout` nomeia um arquivo em `web/templates/layouts/` (pelo menos um é obrigatório). `Name` é o caminho da página sob `web/templates/pages/` sem o sufixo `.html`, então `pages/blog/post.html` é `view.Page{Name: "blog/post"}`.

O layout é dono do shell — navegação, o container `#amarra-main` e o único script `/static/js/amarra.js`. Um template de página só preenche o bloco `content`:

```html
{{ define "content" }}
<h1>{{ .Title }}</h1>
{{ end }}
```

Uma página também pode definir `{{ define "frame:<id>" }}` para requisições de frame (`Amarra-Frame: <id>`), que renderizam apenas esse bloco.

## Dados da página

Construa os dados da página com os mesmos helpers que o scaffold usa. `amarraData` agrupa os valores no escopo da requisição (site, flash e qualquer coisa que você passar), e `meta.ForRequest(h.site, r)` produz a prévia de Open Graph / Twitter.

```go
writeView(w, r, h.views, h.cfg, "app", "contact", amarraData(r, h.site, map[string]any{
  "Title":  h.catalog.T("contact.title"),
  "Errors": errs,
  "Name":   name,
  "Email":  email,
}), http.StatusUnprocessableEntity)
```

`writeView` é a variante de validação: ela nomeia o layout explicitamente (passe um segundo layout, como `"landing"`, se o app tiver um) e permite definir o status — `422` ao re-renderizar um formulário com erros.

:::note
Passe `meta.SiteFrom(appName, cfg.AppURL)` a partir do bootstrap, para que os dados da página tenham um valor `Site` para as tags OG/Twitter.
:::

## Drive vs HTML completo

O Drive intercepta cliques e submits da mesma origem por padrão e faz o morph de `#amarra-main`; ele ainda renderiza através do layout, então a troca é fluida. Primeiros carregamentos, `curl` e crawlers recebem a página completa. Adicione `data-amarra-skip` para isentar um elemento.

## Próximos passos

- [Formulários e validação](/amarra-cais/pt-br/docs/how-to/forms-and-validation/) — trate um formulário enviado com o kit.
- [Views e o kit](/amarra-cais/pt-br/docs/reference/views-and-kit/) — os componentes e hooks incluídos.
