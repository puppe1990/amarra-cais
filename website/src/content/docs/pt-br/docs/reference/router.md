---
title: Router
description: A API do cais.Router — métodos HTTP, grupos de rotas, path parameters, helpers de param e tratamento de 404 projetado.
sidebar:
  order: 4
---

`cais.Router` envolve o `http.ServeMux` do `net/http` (Go 1.22+) e adiciona helpers de path parameter, grupos de middleware e um 404 projetado. Registre as rotas em `internal/app/routes.go`.

## Métodos

| Chamada                      | Registra |
| ---------------------------- | -------- |
| `r.Get(pattern, handler)`    | `GET`    |
| `r.Post(pattern, handler)`   | `POST`   |
| `r.Put(pattern, handler)`    | `PUT`    |
| `r.Patch(pattern, handler)`  | `PATCH`  |
| `r.Delete(pattern, handler)` | `DELETE` |

Os handlers são `http.HandlerFunc` simples. O router também expõe `r.Use(mw)`, `r.Handle(pattern, handler)` e `r.Static(prefix, dir)`.

`Middleware` é `func(http.Handler) http.Handler`; `Use` o acrescenta à cadeia que envolve toda rota registrada.

## Parâmetros de path

Escreva os parâmetros como `{name}` no pattern e leia-os por meio de um helper, que faz o parse do valor e chama o seu handler tipado:

```go
r.Get("/blog/{slug}", cais.StringParam("slug", blog.Show))
r.Post("/chat/{id}/permissions/{permID}/approve", cais.StringParams("id", "permID", chat.ApprovePermission))
r.Get("/items/{id}/{slug}", cais.IntStringParams("id", "slug", items.Show))
r.Group(middleware.RequireAuth("/login"), func(g *cais.Router) {
  g.Post("/admin/items/{id}", cais.IntParam("id", admin.Update))
})
```

| Helper                           | Assinatura do handler               |
| -------------------------------- | ----------------------------------- |
| `cais.StringParam(name, fn)`     | `func(w, r, value string)`          |
| `cais.StringParams(a, b, fn)`    | `func(w, r, a, b string)`           |
| `cais.IntParam(name, fn)`        | `func(w, r, value int64)`           |
| `cais.IntStringParams(i, s, fn)` | `func(w, r, value int64, s string)` |

`IntParam` rejeita valores não inteiros e valores `<= 0`; os helpers de string rejeitam valores vazios. Um parse que falha é encaminhado para o 404 projetado (abaixo) em vez de um valor zero silencioso.

## Grupos

`r.Group(mw, fn)` cria um router filho que compartilha o namespace de URL e o `ServeMux` do pai, mas adiciona um middleware. Só o middleware muda — os grupos não criam um prefixo de path. O filho copia a cadeia atual do pai e acrescenta a sua própria, então as rotas de um grupo executam os middlewares do pai mais o middleware do grupo. Passe qualquer `Middleware` (`func(http.Handler) http.Handler`), por exemplo `middleware.RequireAuth("/login")` para o admin no navegador ou `middleware.AdminAuth(cfg)` para uma API com bearer token.

## 404 projetado

`r.NotFound(handler)` atende as rotas não correspondidas **e** as falhas de parse dos helpers de param. O handler é dono do status e roda com os middlewares do router (session, flash, CSRF), então ele pode renderizar uma página normal:

```go
r.NotFound(notFound.ServeHTTP)   // notFound calls view.Write(..., Status: http.StatusNotFound)
```

Sem ele, ambos os caminhos caem no `http.NotFound` (`text/plain`).

## Raiz exata e arquivos estáticos

Um pattern `"/"` sozinho é reescrito para o wildcard de fim de path `/{$}`, então o handler da home corresponde apenas à raiz exata em vez de a todo path não correspondido (como `/.env` ou `/nope`). Patterns reais de subárvore como `"/static/"` mantêm o comportamento de prefixo. Sirva arquivos com `r.Static(prefix, dir)` (ou `StaticForEnv` com um `Config`); em desenvolvimento ele define `Cache-Control: no-store` para que edições de JS/CSS valham sem um rebuild.

## Conflitos no ServeMux

O registro entra em panic quando dois patterns sob o mesmo método podem corresponder ao mesmo path e nenhum é mais específico:

```go
// ❌ panic: both match /webhooks/kiwify/delete
r.Post("/webhooks/kiwify/{token}", receiver)
r.Post("/webhooks/{id}/delete", deleteHandler)

// ✅ prefer a REST method or a distinct prefix
r.Delete("/webhooks/{id}", deleteHandler)
r.Post("/webhooks/incoming/{token}", receiver)
```

`cais.Router` reescreve o panic com uma dica curta, e `amarra-cais routes` avisa quando consegue detectar o conflito estaticamente a partir de `routes.go`.
