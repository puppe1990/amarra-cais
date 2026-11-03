#!/usr/bin/env bash
# Boot a freshly scaffolded app in production mode (replaces dogfood server smoke).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
TMP="$(mktemp -d)"
PORT_NUM=$((18000 + RANDOM % 8000))
trap 'kill "${PID:-}" 2>/dev/null || true; rm -rf "$TMP"' EXIT

cd "$ROOT"
go build -o "$TMP/amarra-cais" ./cmd/amarra-cais

export CAIS_REPLACE="$ROOT"
export CAIS_SKIP_TIDY=1

APP="$TMP/smokeprod"
"$TMP/amarra-cais" new smokeprod "$APP"
cd "$APP"

# #396: CSS as Cleat does it — Tailwind before the module is in GOMODCACHE.
# GH runners keep `go` in /usr/bin, so hiding it from PATH is not enough:
# stub the binary so `go list -m` fails and shippedKitContent() falls back
# to kit-classes.txt (#391).
test -f web/static/css/kit-classes.txt
EMPTY_MOD="$TMP/gomod-empty"
HIDE_GO="$TMP/hide-go"
mkdir -p "$EMPTY_MOD" "$HIDE_GO"
printf '#!/bin/sh\nexit 127\n' >"$HIDE_GO/go"
chmod +x "$HIDE_GO/go"
NODE_BIN="$(dirname "$(command -v node)")"
NPM_BIN="$(dirname "$(command -v npm)")"
NPX_BIN="$(dirname "$(command -v npx)")"
CSS_PATH="${HIDE_GO}:${NODE_BIN}:${NPM_BIN}:${NPX_BIN}:/usr/bin:/bin"
if PATH="$CSS_PATH" go list -m github.com/puppe1990/amarra-cais >/dev/null 2>&1; then
  echo "go list -m still resolves the module; CSS gate is invalid" >&2
  exit 1
fi
GOMODCACHE="$EMPTY_MOD" PATH="$CSS_PATH" npm install
GOMODCACHE="$EMPTY_MOD" PATH="$CSS_PATH" npx tailwindcss -i input.css -o web/static/css/styles.css --minify
if ! grep -q 'bg-copper' web/static/css/styles.css; then
  echo "styles.css missing .bg-copper (kit CSS pruned)" >&2
  exit 1
fi
if ! grep -q 'amarra-btn' web/static/css/styles.css; then
  echo "styles.css missing .amarra-btn" >&2
  exit 1
fi

go mod tidy
go build -o "$TMP/server" ./cmd/server

grep -q 'id="amarra-main"' web/templates/layouts/app.html
test -f web/static/js/amarra.js
! test -f vite.config.js

export ENV=production
export APP_URL=https://example.com
export ADMIN_TOKEN=ci-smoke-secret
export PORT=":${PORT_NUM}"
export DB_PATH="$TMP/app.db"

"$TMP/server" &
PID=$!

READY=0
for _ in $(seq 1 60); do
  if curl -sf "http://127.0.0.1:${PORT_NUM}/health" >/dev/null 2>&1; then
    READY=1
    break
  fi
  sleep 0.2
done
if [ "$READY" -ne 1 ]; then
  echo "server did not become ready on port ${PORT_NUM}" >&2
  exit 1
fi

BODY=$(curl -sf "http://127.0.0.1:${PORT_NUM}/health")
echo "$BODY" | grep -q '"status":"ok"'

CODE=$(curl -s -o /dev/null -w "%{http_code}" "http://127.0.0.1:${PORT_NUM}/login")
if [ "$CODE" != "200" ]; then
  echo "GET /login status = $CODE, want 200" >&2
  exit 1
fi

echo "smoke production: ok"
