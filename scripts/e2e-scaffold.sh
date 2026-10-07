#!/usr/bin/env bash
# Build the current CLI, scaffold an app, and run Playwright against it (#294).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
TMP="${E2E_TMP:-$(mktemp -d)}"
DEV_PID=""
PROD_PID=""

cleanup() {
  local code=$?
  if [ -n "${DEV_PID}" ]; then kill "${DEV_PID}" 2>/dev/null || true; fi
  if [ -n "${PROD_PID}" ]; then kill "${PROD_PID}" 2>/dev/null || true; fi
  if [ "$code" -ne 0 ]; then
    echo "e2e failed; tmp=$TMP" >&2
    if [ -f "$TMP/dev.log" ]; then
      echo "---- dev.log ----" >&2
      tail -80 "$TMP/dev.log" >&2 || true
    fi
    if [ -f "$TMP/prod.log" ]; then
      echo "---- prod.log ----" >&2
      tail -80 "$TMP/prod.log" >&2 || true
    fi
  fi
  if [ -z "${E2E_KEEP:-}" ]; then rm -rf "$TMP"; fi
}
trap cleanup EXIT

wait_ready() {
  local port=$1
  for _ in $(seq 1 120); do
    if curl -sf "http://127.0.0.1:${port}/health" >/dev/null; then
      return 0
    fi
    sleep 0.5
  done
  echo "server did not become ready on port ${port}" >&2
  return 1
}

cd "$ROOT"
go build -o "$TMP/amarra-cais" ./cmd/amarra-cais

export CAIS_REPLACE="$ROOT"
export CAIS_SKIP_TIDY=1

APP="$TMP/e2eapp"
"$TMP/amarra-cais" new e2eapp "$APP"
cd "$APP"
go mod tidy

"$TMP/amarra-cais" g handler dialog
cp "$ROOT/e2e/fixtures/dialog.html" web/templates/pages/dialog.html
"$TMP/amarra-cais" g handler picker
cp "$ROOT/e2e/fixtures/picker.html" web/templates/pages/picker.html
"$TMP/amarra-cais" g live counter
"$TMP/amarra-cais" g stream chat
go mod tidy
go build -o "$TMP/server" ./cmd/server

DEV_PORT=$((18000 + RANDOM % 8000))
PROD_PORT=$((DEV_PORT + 1))

export DB_PATH="$TMP/app.db"
export ENV=development
export PORT=":${DEV_PORT}"
"$TMP/server" >"$TMP/dev.log" 2>&1 &
DEV_PID=$!
wait_ready "$DEV_PORT"

cd "$ROOT"
export BASE_URL="http://127.0.0.1:${DEV_PORT}"
export PLAYWRIGHT_HTML_OPEN=never
npx playwright test --grep-invert @sw

kill "${DEV_PID}"
wait "${DEV_PID}" 2>/dev/null || true
DEV_PID=""

# Playwright ran from the framework root; the binary resolves web/static from cwd.
cd "$APP"
export ENV=production
export APP_URL="http://127.0.0.1:${PROD_PORT}"
export ADMIN_TOKEN=e2e-secret
export PORT=":${PROD_PORT}"
"$TMP/server" >"$TMP/prod.log" 2>&1 &
PROD_PID=$!
wait_ready "$PROD_PORT"

cd "$ROOT"
export BASE_URL="http://127.0.0.1:${PROD_PORT}"
npx playwright test --grep @sw
