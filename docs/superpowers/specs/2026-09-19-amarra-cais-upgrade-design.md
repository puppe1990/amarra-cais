# `amarra-cais upgrade` — design

Guided framework upgrade for existing apps (#192). Decisions validated with the
user (brainstorm 2026-09-19):

- Migration checklist source: **curated manifest embedded in the CLI**.
- Baseline version: **framework version already in `go.mod`, read before the bump**
  (no `.cais-version` file).
- Guard: **abort when `go.mod` has a local `replace`** (from `amarra-cais link`).

## 1. Command surface

```
amarra-cais upgrade [version] [--dry-run]
```

- `version` optional: `0.11.0` or `v0.11.0`; default `latest`.
- `--dry-run` prints the plan + checklist and touches nothing.
- Registered in `internal/cli/cli.go` (`case "upgrade"`) and listed in `help`.

## 2. Flow

1. `c.appDir()`; read `go.mod`.
2. `from := extractCaisVersion(go.mod)` (existing helper).
3. Guard: `go.mod` contains `replace <frameworkModule>` → error telling the user
   to run `amarra-cais link --unlink` first; the file is left untouched.
4. Resolve target: `latest` stays `latest`; otherwise normalize to `vX.Y.Z` via
   `parseSemverCore`. If `to < from` → abort (`target is older than current`).
5. Print `→ upgrade v<from> → <to>` and the migration checklist
   (`migrationsBetween(from, to)`), **before** running commands so the user sees
   what may need manual edits.
6. `--dry-run` stops here.
7. `go get <frameworkModule>@<target>`.
8. `npm install --include=dev` (only when `package.json` exists) + `go mod tidy`,
   via a helper extracted from `cmdInstall` and shared by both commands.
9. Run `doctor` (same check runner as `cmdDoctor`) and print the final summary +
   checklist reminder.

When `from` is unknown (`?`, `dev`, `(devel)`) the manifest prints every step up
to `to` with a note that the baseline could not be detected.

## 3. Migration manifest (`internal/cli/upgrade_manifest.go`)

Data-only file, maintained per release:

```go
type migrationStep struct {
	Version string // "0.10.0"
	Title   string // "Store requires DB() *sql.DB"
	Action  string // "add DB() *sql.DB to the Store interface in internal/store/store.go"
}

var frameworkMigrations = []migrationStep{ /* newest first */ }

func migrationsBetween(from, to semverCore) []migrationStep
```

Selection: `compareSemverCore(from, step) < 0 && compareSemverCore(step, to) <= 0`.
Steps are kept newest-first so the output reads top-down from the target.

Seed entries (versions pinned to the `CHANGELOG` sections where the change
landed; the list is curated, not auto-derived):

- `0.10.0` — `netutil.HealthPayload(status, port, env)`: add the `env` argument.
- `0.10.0` — `jobsui.Register` needs `DB() *sql.DB` on the `Store` interface.
- `0.10.0` — HTMX assets deprecated; generators no longer emit `hx-*`.
- `0.9.0` — `.cais-generated.json` records generated files (`destroy` uses it).
- `0.6.1` — `view.Write` defaults dynamic pages to `Cache-Control: no-store`.
- `0.5.0` — `writeView(w, r, views, cfg, layout, name, data, status)` argument order.
- `0.3.0` — `doctor` FAILs on `hx-*` / `gonertia`; apps must be HTML-first.

## 4. Files

| File                                     | Responsibility                               |
| ---------------------------------------- | -------------------------------------------- |
| `internal/cli/upgrade.go`                | command orchestration, guards, output        |
| `internal/cli/upgrade_manifest.go`       | curated migration data + range selection     |
| `internal/cli/upgrade_test.go`           | flow/guards/dry-run via `fakeToolchain`      |
| `internal/cli/upgrade_manifest_test.go`  | `migrationsBetween` cases                    |
| `internal/cli/commands.go`               | extract shared `install` helper (npm + tidy) |
| `internal/cli/cli.go`                    | `case "upgrade"` + help line                 |
| `README.md`, `AGENTS.md`, `CHANGELOG.md` | document the command                         |

## 5. Tests (headless, `:memory:`/temp dirs, `fakeToolchain`)

- `migrationsBetween`: full range, empty range, `from == to`, unknown `from`.
- `upgrade --dry-run`: prints plan + checklist, records **no** tool calls, leaves
  `go.mod` untouched.
- Upgrade run: calls `go get <module>@vX` / `@latest`, then `go mod tidy`, then
  `npm install` when `package.json` exists (asserted from the fake call log).
- Local `replace` guard: errors with the `link --unlink` hint and does not call
  `go get`.
- Target older than current: aborts.
- Help lists `amarra-cais upgrade`.

## 6. Out of scope (YAGNI)

- `.cais-version` baseline file.
- Parsing `CHANGELOG.md` at runtime.
- Requiring a clean git tree.
- Backing up / rolling back `go.mod` (`--dry-run` covers inspection).
- Auto-applying the migration steps (checklist only; edits stay manual).
