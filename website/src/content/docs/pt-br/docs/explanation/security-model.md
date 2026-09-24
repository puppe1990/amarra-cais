---
title: Modelo de segurança
description: Como CSRF, sessions, headers, rate limits e production gates se combinam em um app Amarra.
sidebar:
  order: 4
---

O Amarra entrega um pequeno conjunto de peças defensivas em vez de um único recurso de segurança. Cada camada cobre uma classe específica de risco, e elas foram feitas para serem compostas no router em uma ordem conhecida.

## CSRF — cookie de double-submit

`middleware.CSRF(cfg)` valida métodos que mudam estado (`POST`, `PUT`, `DELETE`, `PATCH`). Ele usa um cookie de double-submit chamado `cais_csrf`: o token fica em um cookie e também é enviado como campo de formulário ou header `X-CSRF-Token`. Não existe store de tokens no servidor — o servidor apenas compara os dois valores, o que mantém a checagem stateless.

O layout renderiza `<meta name="csrf-token">`, e o `amarra.js` envia `X-CSRF-Token` nas requests do Drive. O `<.form>` do kit injeta o campo oculto a partir dos dados da página raiz (`$.CSRFToken`), então você não o duplica dentro do slot.

## Sessions

As sessions são baseadas em cookie (`pkg/cais/session`) com `SignIn` / `SignOut`. Uma session gira no login, invalidando o token anterior. Os cookies e suas linhas no SQLite expiram após 7 dias (`sessionTTL` / `defaultMaxAge`); o store mantém `expires_at` e ignora linhas expiradas na busca. As senhas são armazenadas como hashes bcrypt (`session.HashPassword` / `session.VerifyPassword`).

`session.CookieOptionsFromConfig(cfg)` marca os cookies como `Secure` quando `cfg.CookieSecure()` é true, o que acontece com `ENV=production`. Faça a poda das linhas expiradas com `amarra-cais db prune-sessions` ou `session.Store.PruneExpired()`.

## Headers de segurança

`middleware.SecurityHeaders(cfg)` roda depois do `Recover` e define `X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy` e `Permissions-Policy`. Em produção ele também adiciona `Strict-Transport-Security`.

A Content-Security-Policy mantém `script-src 'self' 'unsafe-inline'`. Isso é um trade-off deliberado: o snippet de tema roda antes do paint, e o Drive adiciona uma pequena inicialização inline, então o escaping continua sendo a defesa primária. O roadmap é substituir `'unsafe-inline'` por um nonce por request ou hashes SRI (#97).

## Rate limiting

Envolva rotas POST sensíveis com token buckets por IP:

```go
loginLimit := middleware.NewRateLimiter(10, cfg)   // 10 req/min
r.Post("/login", loginLimit.Middleware(http.HandlerFunc(auth.LoginPost)).ServeHTTP)
```

O limiter resolve o endereço do cliente com `middleware.ClientIP(r, cfg)`. Atrás de um reverse proxy, defina `TRUSTED_PROXIES` para que o `X-Forwarded-For` seja confiável em vez de falsificável.

## Production gates

`cfg.Validate()` falha o boot quando `ENV=production` e valores obrigatórios estão faltando. Em particular, `ADMIN_TOKEN` precisa estar definido para as APIs de admin protegidas por bearer, e `APP_URL` é obrigatório (ele também é usado para URLs absolutas de Open Graph). Configuração faltando para o processo em vez de degradar silenciosamente.

:::caution
O rate limiter em memória e a checagem de CSRF por double-submit assumem ambos um único processo do app. Se você rodar mais de uma réplica, mova o rate limiting para um store compartilhado e reveja como o token de CSRF é validado.
:::

Leitura relacionada: [auth e sessions](/amarra-cais/pt-br/docs/how-to/auth-and-sessions/), [referência de middleware](/amarra-cais/pt-br/docs/reference/middleware/) e [referência de configuração](/amarra-cais/pt-br/docs/reference/configuration/).
