---
title: Deploy em produção
description: Faça cross-compile, entregue web/static com uma unit do systemd e atenda aos gates do ambiente de produção.
sidebar:
  order: 11
---

O Amarra é implantado como um único binário Go mais um diretório `web/static`. Não há um passo de Vite `web/static/build/` nem processo Node em produção.

## Faça o build do release

```bash
amarra-cais css     # Tailwind → web/static/css/styles.css
amarra-cais build --os linux --arch amd64 -o bin/server-linux
tar czf release.tar.gz bin/server-linux web/static
```

Entregue `web/static` (CSS, `js/amarra.js`, o manifest, ícones) ao lado do binário. `amarra-cais doctor` verifica se `web/static`, `manifest.webmanifest` e `amarra.js` estão presentes antes de você publicar.

## Layout do servidor e systemd

```text
/opt/myapp/
  current/
    bin/server          # renamed from server-linux
    web/static/         # CSS, JS, manifest, icons
  data/app.db           # persistent SQLite
/etc/myapp/env          # production variables (chmod 600)
```

Copie `deploy/systemd/cais-app.service.example` e aponte o `WorkingDirectory` do systemd para `/opt/myapp/current`, onde `web/static/` fica. Se o diretório de trabalho não for a raiz do app, defina `STATIC_DIR` e `TEMPLATES_DIR` explicitamente para que o binário encontre seus assets. O `docs/deploy/lightsail-systemd.md` do repositório percorre uma configuração completa de Lightsail + Caddy.

```bash
sudo systemctl daemon-reload
sudo systemctl enable myapp
sudo systemctl start myapp
curl -s http://127.0.0.1:4006/health
```

## Ambiente de produção

```bash
PORT=:4006
ENV=production
APP_URL=https://myapp.example.com
DB_PATH=/opt/myapp/data/app.db
ADMIN_TOKEN=<strong-token>
LOCALE=pt
TRUSTED_PROXIES=127.0.0.1
# STATIC_DIR=/opt/myapp/current/web/static   # optional
```

`ENV=production` ativa o gate de produção: `cfg.Validate()` falha no boot quando `ADMIN_TOKEN` ou `APP_URL` está ausente, então o app se recusa a iniciar em vez de rodar de forma insegura. Defina `TRUSTED_PROXIES` (IPs separados por vírgula) atrás de um reverse proxy para que `middleware.ClientIP` confie no `X-Forwarded-For` para rate limiting e logging.

## Seeds em produção

Seeds de desenvolvimento — incluindo o usuário demo — **não** rodam quando `ENV=production`. Para dados de catálogo idempotentes, registre seeds em `internal/db/seeds.go` e rode:

```bash
amarra-cais db seed
```

## Headers de segurança e logs

`middleware.SecurityHeaders(cfg)` (registrado depois de `Recover`) define `X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy` e `Permissions-Policy`, e adiciona `Strict-Transport-Security` em produção. Cookies de CSRF e flash usam `Secure` quando `cfg.CookieSecure()` é true, e a sessão gira no login.

Em `ENV=development`, logs de requisição e SQL fluem como JSON (`kind: request`, `kind: sql`); `LOG_FORMAT=text` desativa isso. `/logs` é somente localhost, e `/jobs` é somente localhost em todos os ambientes — acesse-o por um túnel SSH em uma máquina remota.

## Verifique

```bash
curl -sI https://myapp.example.com/ | grep -i permissions-policy
curl -s https://myapp.example.com/static/manifest.webmanifest | grep display
curl -s https://myapp.example.com/health
```

## Relacionados

- [PWA e mobile](/amarra-cais/pt-br/docs/how-to/pwa-and-mobile/) — os assets que você está entregando.
- [Configuração](/amarra-cais/pt-br/docs/reference/configuration/) — cada variável de ambiente acima.
- [Modelo de segurança](/amarra-cais/pt-br/docs/explanation/security-model/) — headers, cookies e o gate de produção.
- [Atualize o framework](/amarra-cais/pt-br/docs/how-to/upgrade/) — incremente as versões antes de fazer redeploy.
