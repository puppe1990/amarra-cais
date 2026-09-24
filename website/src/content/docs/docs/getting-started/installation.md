---
title: Installation
description: Install the amarra-cais CLI from a release tag or from source.
sidebar:
  order: 2
---

## Requirements

- Go 1.26 or newer (the framework builds on the `net/http` stdlib).
- Node.js 22+ for the Tailwind tooling inside generated apps.
- SQLite is bundled via `modernc.org/sqlite` — no CGO, no system package.

## Install the CLI

```bash
export PATH="$HOME/go/bin:$PATH"
go install github.com/puppe1990/amarra-cais/cmd/amarra-cais@v0.11.0
amarra-cais version   # expect 0.11.0
```

`amarra-cais` installs alongside `cais` — it never overwrites the existing binary. Cais v0.11.x remains the Inertia + Svelte product; see [Cais vs Amarra](/amarra-cais/docs/explanation/cais-vs-amarra/).

## Install from source

Use this when you are working on the framework itself:

```bash
git clone https://github.com/puppe1990/amarra-cais.git
cd amarra-cais
make install-cli
```

## Run the framework checks

```bash
make test      # go test ./... -race
make js-test   # pkg/cais/js + pkg/amarra/js unit tests
make ci        # test + js-test + lint + format-check
```

To point an app at a local checkout instead of a released version, use `amarra-cais link .` from the app directory. Remember to `--unlink` before pushing — the `replace` directive must not be committed.

Next: [Your first app](/amarra-cais/docs/getting-started/your-first-app/).
