---
title: Configuração
description: Variáveis de ambiente, precedência do cais.Load(), gates de validação de produção e flags de segurança de cookies.
sidebar:
  order: 8
---

A configuração é lida uma vez no boot por `cais.Load()`, que também aplica um arquivo `.env` local. O resultado é um valor `cais.Config` passado para handlers, middleware e o router.

## Variáveis de ambiente

| Variável          | Finalidade                                                                                        | Padrão                 |
| ----------------- | ------------------------------------------------------------------------------------------------- | ---------------------- |
| `ENV`             | Ambiente: `development` ou `production`. Controla `CookieSecure()`, HSTS e as ferramentas de dev. | `development`          |
| `PORT`            | Endereço de escuta. `cais.ResolvePort` passa para a próxima porta livre em desenvolvimento.       | `:8080`                |
| `APP_URL`         | URL base absoluta para as URLs de imagem de OG/Twitter. Obrigatória em produção.                  | —                      |
| `ADMIN_TOKEN`     | Bearer token para `middleware.AdminAuth`. Obrigatório em produção.                                | —                      |
| `TRUSTED_PROXIES` | IPs/CIDRs de proxy separados por vírgula; `X-Forwarded-For` só é confiável vindo destes.          | —                      |
| `LOCALE`          | Idioma da UI para `pkg/cais/i18n` (`en` ou `pt`).                                                 | `en`                   |
| `STATIC_DIR`      | Diretório de arquivos estáticos quando o `WorkingDirectory` do processo não é a raiz do app.      | `web/static`           |
| `TEMPLATES_DIR`   | Diretório de templates, com a mesma regra de override.                                            | `web/templates`        |
| `LOG_FORMAT`      | Formato do log de requisições/SQL: `json` ou `text`.                                              | JSON em dev e produção |

:::note
Existem overrides adicionais para os security headers: `DB_PATH`, `PERMISSIONS_POLICY` e `CSP_STYLE_SRC` / `CSP_CONNECT_SRC` / `CSP_MEDIA_SRC` / `CSP_IMG_SRC` / `CSP_FONT_SRC` (uma webfont hospedada precisa tanto de `CSP_FONT_SRC` quanto de `CSP_STYLE_SRC`).
:::

## Carregamento e precedência

```go
cfg := cais.Load()
if err := cfg.Validate(); err != nil {
    log.Fatal(err)
}
```

`cais.Load()` aplica um `.env` local se presente (via `dotenv.LoadFile`). Chaves já presentes no ambiente do processo vencem — `t.Setenv`, `Environment=` do systemd e secrets de CI não são sobrescritos. Um `.env` ausente não faz nada.

Exemplo de `.env` (o env do processo ainda vence):

```bash
ENV=development
PORT=:8080
LOCALE=pt
```

`amarra-cais doctor` lê o mesmo `.env` quando verifica o app.

## Gates de produção

`cfg.Validate()` falha no boot quando `ENV=production` e um valor obrigatório está ausente:

- `ADMIN_TOKEN` — o `AdminAuth` rejeita toda requisição em produção quando o token está vazio.
- `APP_URL` — obrigatória para que as URLs de imagem de OG/Twitter sejam absolutas.

Pelo mesmo interruptor, `SanitizeErrors()` é true em produção e os seeds exclusivos de dev (o usuário de demonstração) não rodam. Use `amarra-cais db seed` para dados de catálogo.

`APP_URL` também alimenta `meta.SiteFrom` para que as pré-visualizações do Open Graph e do Twitter usem URLs de imagem absolutas.

## Cookies

`cfg.CookieSecure()` retorna true quando `ENV=production`. Ele é propagado para os escritores de cookie de session, CSRF e flash:

```go
session.SignIn(w, sessions, r, userID, session.CookieOptionsFromConfig(cfg))
flash.Set(w, "notice", "Saved!", cfg.CookieSecure())
```

Os cookies de sessão (e suas linhas no SQLite) expiram após 7 dias; linhas expiradas são ignoradas na busca e podem ser removidas com `amarra-cais db prune-sessions`.

`ENV=production` também ativa o HSTS e a sanitização de erros, e oculta as ferramentas de desenvolvimento.

## Seleção de porta

`cais.ResolvePort(cfg.Port, cfg.Env)` retorna o endereço no qual escutar, passando para a próxima porta livre em desenvolvimento quando a preferida está ocupada. Defina `PORT_STRICT=1` para desativar a troca e falhar em vez disso. Atrás de um reverse proxy, defina `TRUSTED_PROXIES` para que `middleware.ClientIP` confie em `X-Forwarded-For` para rate limiting e logs.

Veja [Deploy](/amarra-cais/pt-br/docs/how-to/deploy/) para os valores de produção e [Middleware](/amarra-cais/pt-br/docs/reference/middleware/) para como a config flui para headers e auth.
