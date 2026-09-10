#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

cd "$ROOT"
go build -o "$TMP/amarra-cais" ./cmd/amarra-cais

export CAIS_REPLACE="$ROOT"
export CAIS_SKIP_TIDY=1

APP="$TMP/smokeapp"
"$TMP/amarra-cais" new smokeapp "$APP"
cd "$APP"
go mod tidy

# #34: the empty app compiling is not enough — g resource with a parent FK,
# --public, and --paginate is what shipped undefined Total and Category.go.
"$TMP/amarra-cais" g resource category --fields name:string
"$TMP/amarra-cais" g resource bookmark --fields title:string,url:url,category_id:references,read:bool --public --paginate

go test ./... -count=1
go build -o "$TMP/server" ./cmd/server

grep -q 'id="amarra-main"' web/templates/layouts/app.html
test -f web/static/js/amarra.js
! test -f vite.config.js

echo "smoke scaffold: ok"
