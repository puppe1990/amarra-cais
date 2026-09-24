---
title: Middleware
description: Middleware embutido — recover, security headers, CSRF, sessões, flash, auth, logging e rate limiting.
sidebar:
  order: 9
---

O middleware fica em `pkg/cais/middleware`. Adicione-o com `r.Use(mw)` em um `cais.Router`; o primeiro `Use` é o wrapper mais externo. `middleware.Middleware` é `func(http.Handler) http.Handler`.

## Referência

| Middleware                                | Finalidade                                                                                                                      |
| ----------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| `middleware.Recover`                      | Captura panics, registra a stack e retorna `500`.                                                                               |
| `middleware.SecurityHeaders(cfg)`         | Define `X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy`, `Permissions-Policy` e a CSP; adiciona HSTS em produção. |
| `middleware.CSRF(cfg)`                    | Cookie double-submit (`cais_csrf`); valida `POST`/`PUT`/`PATCH`/`DELETE` e ignora `/health` e `/static/`.                       |
| `middleware.LoadSession(store)`           | Lê o cookie de sessão e anexa o ID do usuário ao contexto da requisição.                                                        |
| `middleware.Flash(cfg)`                   | Consome o cookie flash de uso único no contexto da requisição.                                                                  |
| `middleware.RequireAuth(loginURL)`        | Redirecionamento `303` para `loginURL` quando não autenticado.                                                                  |
| `middleware.RequireAuthFunc(loginURL, h)` | Envolve um único handler com `RequireAuth`.                                                                                     |
| `middleware.AdminAuth(cfg)`               | Bearer token de `ADMIN_TOKEN`; no-op em desenvolvimento quando não definido, rejeita tudo em produção quando não definido.      |
| `middleware.LoggerTo(cfg, w)`             | Logs de requisição (JSON quando `cfg.LogJSON()`).                                                                               |
| `middleware.NewRateLimiter(limit, cfg)`   | Token bucket por IP, `limit` requisições por minuto.                                                                            |
| `middleware.ClientIP(r, cfg)`             | Resolve o IP do cliente, confiando em `X-Forwarded-For` apenas de `TRUSTED_PROXIES`.                                            |

## Ordem de uso

O scaffold conecta o router nesta ordem:

```go
r := cais.NewRouter()
r.Use(middleware.CSRF(cfg))
r.Use(middleware.LoadSession(deps.Store.Sessions()))
r.Use(middleware.Flash(cfg))
r.Use(i18n.LocaleMiddleware(catalogs, cfg.Locale))
if buf != nil {
  r.Use(middleware.LoggerTo(cfg, devlog.MirrorDefault(log.Writer())))
} else {
  r.Use(middleware.Logger(cfg))
}
r.Use(middleware.Recover)
r.Use(middleware.SecurityHeaders(cfg))
```

`Flash` roda depois de `LoadSession` (precisa da sessão), e `SecurityHeaders` é registrado depois de `Recover`. O CSRF cobre os métodos que alteram estado em toda rota.

A CSP mantém `script-src 'self' 'unsafe-inline'` para o snippet de tema FOUC e o init inline do Drive; o escaping é a defesa principal, com um nonce por requisição no roadmap.

## Middleware no nível da rota

Auth e rate limits são aplicados onde são necessários em vez de globalmente:

```go
loginLimit := middleware.NewRateLimiter(10, cfg)   // 10 req/min
r.Post("/login", loginLimit.Middleware(http.HandlerFunc(auth.LoginPost)).ServeHTTP)

r.Get("/dashboard", middleware.RequireAuthFunc("/login", dashboard.ServeHTTP))
```

Os limiters usam como chave `middleware.ClientIP(r, cfg)` mais o path, então defina `TRUSTED_PROXIES` quando o app está atrás de um reverse proxy. Tanto o admin no navegador (`RequireAuth`) quanto as APIs com bearer token (`AdminAuth`) são opções para `amarra-cais g resource`.

## Detalhes de CSRF e flash

O CSRF é um cookie double-submit: o token fica no cookie `cais_csrf` e é repetido em um campo de formulário `csrf_token` ou no header `X-CSRF-Token` (o Drive envia o header a partir da tag `<meta name="csrf-token">`). Ler uma mensagem após um redirecionamento passa por `middleware.FlashMessage(r)`, que retorna a mensagem que o middleware `Flash` consumiu.

A sessão é rotacionada no login, invalidando o token anterior, e os cookies de CSRF e flash usam `Secure` quando `cfg.CookieSecure()` é true.

## Ferramentas exclusivas de localhost

Dois helpers de desenvolvimento montam suas próprias rotas, restritas a conexões de loopback:

```go
devlog.Register(r, cfg.Env, buf)        // /logs — development only
jobsui.Register(r, deps.Store.DB())     // /jobs — all envs, loopback only
```

`/logs` mostra logs de requisição e SQL em desenvolvimento. `/jobs` é o dashboard da fila (contagens, retry/discard de falhas, tarefas recorrentes). Ambos rejeitam requisições via proxy, então acesse-os por um túnel SSH em produção.

Veja [Configuração](/amarra-cais/pt-br/docs/reference/configuration/) para os valores de `cfg` e [Auth e sessões](/amarra-cais/pt-br/docs/how-to/auth-and-sessions/) para os fluxos de sessão e flash.
