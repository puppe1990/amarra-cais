// Help text for the amarra-cais binary, split from dispatch so cli.go stays
// under the line cap (#288).
package cli

import "fmt"

func (c *CLI) printHelp() {
	_, _ = fmt.Fprintln(c.Out, `Cais — Rails-style CLI for Go full-stack apps

Usage:
  amarra-cais new <app> [dir] [--minimal] [--blank] [--module <path>]
                               Create a new app (default dir: ./<app>)
  amarra-cais new <app> [dir] --minimal   Slim app (home only)
  amarra-cais new <app> [dir] --blank     Empty app (no starter content)
  amarra-cais new <app> [dir] --module <path>   Override go module path
  amarra-cais g [--dry-run] handler <name>      Generate handler + test + HTML page
  amarra-cais g [--dry-run] component <name>    Kit name seeds the shipped markup; other names get a slot
  amarra-cais g component --list                List the shipped kit components an app can override
  amarra-cais g [--dry-run] resource <name> [--fields title:string,url:url] [--public] [--paginate] [--no-seed] [--force] [--admin-auth session|bearer]
  amarra-cais g [--dry-run] model <name> [--fields title:string,url:url]
  amarra-cais g [--dry-run] page <name>         Generate page template only
  amarra-cais g [--dry-run] migration <name>    Generate SQL migration file
  amarra-cais g [--dry-run] live <name>         Generate Live view + page (WebSocket)
  amarra-cais g [--dry-run] stream chat [--live]
  amarra-cais g [--dry-run] sitemap              Blog posts + dynamic /sitemap.xml
                             Generate SSE chat (optional Live WebSocket)
  amarra-cais g [--dry-run] job <name> [--cron "0 3 * * *"]
                             Generate job handler + cmd/worker + registry
  amarra-cais g [--dry-run] auth                Add login/logout and protect dashboard
  amarra-cais g [--dry-run] console             Scaffold cmd/console/main.go
  amarra-cais g [--dry-run] ci                  Add GitHub Actions CI, pre-commit, lint, Prettier
  amarra-cais install               npm install (if package.json) + go mod tidy
  amarra-cais css                   Build Tailwind CSS
  amarra-cais dev                   Hot reload (air + tailwind)
  amarra-cais build [--os linux] [--arch amd64] [-o path]
                               Build bin/server (cross-compile for deploy)
  amarra-cais server                Run the app (go run ./cmd/server)
  amarra-cais test                  Run tests (go test ./...)
  amarra-cais doctor [--mobile]     Check app setup (amarra.js, air, go.mod, PWA/mobile)
  amarra-cais pwa [--bump] [--force]  Write/refresh PWA assets; --bump cache, --force resets brand
  amarra-cais upgrade [version]     Bump the framework in go.mod, tidy, run doctor, print migration steps
  amarra-cais link [path] [--unlink]  Add go.mod replace for local Cais dev (default: sibling ../amarra-cais or ../Cais)
  amarra-cais console               Interactive app console (Go REPL + SQL)
  amarra-cais db migrate            Run pending SQL migrations
  amarra-cais db status             List migration status
  amarra-cais db rollback           Roll back last migration (runs -- down SQL when present)
  amarra-cais db prune-sessions     Delete expired login sessions from SQLite
  amarra-cais db seed               Run internal/db/seeds.go
  amarra-cais jobs work [--queues default,mail] [--concurrency 2]
                             Run background job worker + dispatcher
  amarra-cais jobs status           Show job counts, queues, workers, and recurring tasks
  amarra-cais jobs retry|discard <id>
  amarra-cais jobs prune [--older 24h]
  amarra-cais routes [--verbose]    List HTTP routes from internal/app/routes.go
  amarra-cais destroy [--dry-run] resource|handler|model|component <name>
                             Remove generated resource, handler, model, or component files
  amarra-cais destroy [--dry-run] auth             Remove login/auth scaffolding
  amarra-cais destroy [--dry-run] migration <name> Remove a generated SQL migration file
  amarra-cais version               Print Cais framework version
  amarra-cais help                  Show this help

Aliases:
  amarra-cais g        → amarra-cais generate
  amarra-cais i        → amarra-cais install
  amarra-cais b        → amarra-cais build
  amarra-cais s        → amarra-cais server
  amarra-cais c        → amarra-cais console

Examples:
  amarra-cais new myapp && cd myapp && amarra-cais install && amarra-cais dev
  amarra-cais g handler settings
  amarra-cais console
  amarra-cais css && amarra-cais server`)
}
