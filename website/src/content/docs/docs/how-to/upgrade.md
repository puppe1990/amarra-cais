---
title: Upgrade the framework
description: Bump the framework in go.mod, read the migration checklist, and link a local framework checkout.
sidebar:
  order: 12
---

`amarra-cais upgrade` moves an app onto a newer framework release and tells you what changed:

```bash
amarra-cais upgrade            # bump to the latest version
amarra-cais upgrade v0.11.0    # pin a specific version
amarra-cais upgrade --dry-run  # preview the changes without writing
```

## What it does

1. Bumps `github.com/puppe1990/amarra-cais` in `go.mod` (default `latest`).
2. Re-runs `npm install` and `go mod tidy`.
3. Runs `doctor`.
4. Prints a curated migration checklist for the version range you crossed.

`--dry-run` shows the plan without applying it, the same way the generators preview their changes.

## When to run it

Run it after a new framework release, then read the checklist before you deploy. The repo's `CHANGELOG.md` groups each release by `Added`, `Changed`, `Fixed`, and `Security`, and the upgrade command tailors the checklist to the range between your current version and the target. If a release changed templates or HTML, bump the PWA cache afterwards so phones pick up the new assets:

```bash
amarra-cais pwa --bump
```

## Pin versions

Keep the CLI and the app's `go.mod` in step. The quick start installs a pinned CLI:

```bash
go install github.com/puppe1990/amarra-cais/cmd/amarra-cais@v0.11.0
amarra-cais version   # expect 0.11.0
```

Generating code with one release while depending on another is the usual source of confusing "unknown component" or "missing method" errors. `amarra-cais version` prints the framework version the CLI was built from, so you can check both sides.

## Develop against a local checkout

`amarra-cais link` adds a `go.mod` `replace` so an app builds against your local framework copy:

```bash
amarra-cais link ../amarra-cais   # from an app directory
```

The replace is for local development only — do not commit it. Undo it before you push:

```bash
amarra-cais link --unlink
```

:::caution
A committed `replace` directive pins the app to a path that does not exist on your build server or CI runner. Unlink before you push so the released app resolves the module from the proxy.
:::

:::note
You can also point at a local checkout via the `CAIS_REPLACE` environment variable instead of editing `go.mod` directly.
:::

## Related

- [CLI reference](/amarra-cais/docs/reference/cli/) — `upgrade`, `link`, and `version`.
- [Deploy to production](/amarra-cais/docs/how-to/deploy/) — redeploy after an upgrade.
- [Installation](/amarra-cais/docs/getting-started/installation/) — installing the CLI itself.
- [Project layout](/amarra-cais/docs/getting-started/project-layout/) — what the generators maintain as versions move.
