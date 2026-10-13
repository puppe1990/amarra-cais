package cli

const tplGoMod = `module {{.ModulePath}}

go 1.26

require (
	github.com/puppe1990/amarra-cais v{{.CaisVersion}}
	modernc.org/sqlite v1.53.0
)
`

const tplEmptyCSS = `/* Run: amarra-cais css */\n`

const tplAir = `root = "."
tmp_dir = "tmp"

[build]
  cmd = "go build -o ./tmp/main ./cmd/server"
  entrypoint = ["./tmp/main"]
  delay = 1000
  exclude_dir = ["tmp", "data", "bin", "node_modules", "web"]
  # .go only: templates and the built stylesheet are reloaded in-process by the
  # server (devreload, #340) and the browser is refreshed over /_dev/reload, so
  # watching web/ would only add a rebuild+restart per HTML edit (and a second
  # one for the Tailwind output).
  include_ext = ["go"]
  stop_on_error = true
  # SIGINT the old tmp/main and wait before SIGKILL, or the zombie keeps
  # :8080 and the sqlite WAL and the next boot shifts ports / locks (#77).
  send_interrupt = true
  kill_delay = 1000

[log]
  time = false
  main_only = true

[misc]
  clean_on_exit = true
  startup_banner = ""
`

const tplWebEmbed = `package web

import "embed"

//go:embed all:templates
var Templates embed.FS
`

const tplEnv = `# Local overrides for the scaffold defaults (see .env.example for the full list).
# Commit the app's real values here only when they are not secrets.

# The default design uses the docs-site heading font (Fraunces, linked in
# layouts/app.html). The default CSP is 'self' only, so the stylesheet and the
# woff2 hosts must be allowed or the browser blocks the font and the headings
# fall back to the system serif.
CSP_STYLE_SRC=https://fonts.googleapis.com
CSP_FONT_SRC=https://fonts.gstatic.com
`

const tplEnvExample = `# Server
PORT=:8080
# PaaS may inject PORT=8080 (number only); Load canonicalizes to :8080 (#303).
ENV=development
APP_URL=http://localhost:8080
LOCALE=en
# LOG_FORMAT=json (default in dev/production) or text for Rails-style request logs

# Database
DB_PATH=./data/app.db

# Deploy (optional; override when systemd WorkingDirectory is not app root)
# STATIC_DIR=/opt/myapp/current/web/static
# TEMPLATES_DIR=/opt/myapp/current/web/templates

# Security (required when ENV=production)
ADMIN_TOKEN=

# Reverse proxy (comma-separated IPs; trust X-Forwarded-For for client IP)
TRUSTED_PROXIES=

# Total request-body cap before any parse (bytes; default 33554432 = 32 MiB).
# Raise for legitimate large uploads; oversized POSTs get 413. (#221)
# Invalid values fail cfg.Validate() at boot (#286).
# MAX_BODY_BYTES=33554432

# CSP extras (optional; space-separated hosts appended to defaults)
# CSP_STYLE_SRC=https://fonts.googleapis.com
# CSP_FONT_SRC=https://fonts.gstatic.com
# CSP_CONNECT_SRC=
# CSP_MEDIA_SRC=
# CSP_IMG_SRC=
# CSP_SCRIPT_SRC=
# Unsigned inline scripts (pre-#263): CSP_SCRIPT_SRC='unsafe-inline'
# Camera and remote images stay off unless the app needs them (#262):
# PERMISSIONS_POLICY=camera=(self), microphone=(), geolocation=()
# CSP_IMG_SRC=https://images.example.com

# SMTP (optional; password reset emails — logs to stdout when unset)
SMTP_HOST=
SMTP_PORT=587
SMTP_USER=
SMTP_PASSWORD=
SMTP_FROM=
`
