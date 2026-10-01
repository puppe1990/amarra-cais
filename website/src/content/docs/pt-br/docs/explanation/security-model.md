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

As sessions são baseadas em cookie (`pkg/cais/session`) com `SignIn` / `SignOut`. Uma session gira no login, invalidando o token anterior. Os cookies e suas linhas no SQLite expiram após 7 dias (`sessionTTL` / `defaultMaxAge`); o store mantém `expires_at` e ignora linhas expiradas na busca. As senhas são armazenadas como hashes bcrypt (`session.HashPassword` / `session.VerifyPassword`). Os tokens de reset de senha são armazenados como digests SHA-256 (`passwordreset.Hash`), nunca o valor puro — um banco, backup ou log SQL vazado não pode ser reutilizado — e uma entrega que falha não revela se a conta existe (#222, #223).

`session.CookieOptionsFromConfig(cfg)` marca os cookies como `Secure` quando `cfg.CookieSecure()` é true, o que acontece com `ENV=production`. Faça a poda das linhas expiradas com `amarra-cais db prune-sessions` ou `session.Store.PruneExpired()`.

## Headers de segurança

`middleware.SecurityHeaders(cfg)` roda depois do `Recover` e define `X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy` e `Permissions-Policy`. Em produção ele também adiciona `Strict-Transport-Security`. Câmera e imagens de terceiros ficam desligadas por default (`camera=()`, `img-src 'self' data:`). Apps que precisam de scanner ou de um CDN de imagens setam `PERMISSIONS_POLICY` e `CSP_IMG_SRC`.

A Content-Security-Policy em `script-src` é `'self'` mais um nonce por requisição. `view.Write` injeta `.CSPNonce` em dados `map[string]any` para o snippet FOUC do tema e os scripts inline do service worker. Coloque `nonce="{{ .CSPNonce }}"` em todo `<script>` inline. Apps que ainda precisam de inline sem nonce setam `CSP_SCRIPT_SRC='unsafe-inline'` (#263).

## Rate limiting

Envolva rotas POST sensíveis com token buckets por IP:

```go
loginLimit := middleware.NewRateLimiter(10, cfg)   // 10 req/min
r.Post("/login", loginLimit.Middleware(http.HandlerFunc(auth.LoginPost)).ServeHTTP)
```

O limiter resolve o endereço do cliente com `middleware.ClientIP(r, cfg)`. Atrás de um reverse proxy, defina `TRUSTED_PROXIES` para que o `X-Forwarded-For` seja confiável em vez de falsificável.

## Production gates

`cfg.Validate()` falha o boot quando `ENV=production` e valores obrigatórios estão faltando. Em particular, `ADMIN_TOKEN` precisa estar definido para as APIs de admin protegidas por bearer, e `APP_URL` é obrigatório (ele também é usado para URLs absolutas de Open Graph). Valores explícitos de env que não podem ser aplicados (`MAX_BODY_BYTES`, `PORT`, forma de `APP_URL`, `TRUSTED_PROXIES`, extras CSP) também falham o Validate. Configuração faltando para o processo.

:::caution
O rate limiter em memória e a checagem de CSRF por double-submit assumem ambos um único processo do app. Se você rodar mais de uma réplica, mova o rate limiting para um store compartilhado e reveja como o token de CSRF é validado.
:::

Leitura relacionada: [auth e sessions](/amarra-cais/pt-br/docs/how-to/auth-and-sessions/), [referência de middleware](/amarra-cais/pt-br/docs/reference/middleware/) e [referência de configuração](/amarra-cais/pt-br/docs/reference/configuration/).
