package cli

const tplREADME = "# {{.AppName}}\n\n" +
	"Full-stack Go app built with [Amarra](https://github.com/puppe1990/amarra-cais): HTML templates, Tailwind, and SQLite.\n\n" +
	"## Stack\n\n" +
	"- Go 1.26 (net/http stdlib) + Amarra views\n" +
	"- HTML pages (`web/templates/pages/`) + `amarra.js` Drive\n" +
	"- Tailwind CSS 3.x\n" +
	"- SQLite (modernc.org/sqlite, no CGO)\n\n" +
	"## Quick start\n\n" +
	"```bash\n" +
	"amarra-cais install  # npm install + go mod tidy\n" +
	"amarra-cais dev        # http://localhost:8080\n" +
	"amarra-cais test       # full test suite\n" +
	"amarra-cais build      # bin/server\n" +
	"```\n\n" +
	"## Amarra CLI\n\n" +
	"This app was scaffolded with the Amarra CLI. Useful commands:\n\n" +
	"```bash\n" +
	"amarra-cais install               # npm install + go mod tidy\n" +
	"amarra-cais css                   # build Tailwind\n" +
	"amarra-cais dev                   # air + tailwind watch\n" +
	"amarra-cais server                # go run ./cmd/server\n" +
	"amarra-cais console               # interactive Go REPL + SQL\n" +
	"amarra-cais g handler <name>      # handler + test + page template\n" +
	"amarra-cais g resource <name>     # model + migration + admin CRUD\n" +
	"amarra-cais g page <name>         # page template only\n" +
	"amarra-cais g migration <name>    # SQL migration file\n" +
	"amarra-cais test                  # go test ./...\n" +
	"amarra-cais doctor                # verify setup\n" +
	"```\n\n" +
	"## CI and pre-commit\n\n" +
	"GitHub Actions runs Go tests, `golangci-lint`, Prettier, and `npm test` on every push/PR to `main`.\n\n" +
	"```bash\n" +
	"make pre-commit-install   # once: installs git hooks\n" +
	"make ci                   # test + lint + format-check locally\n" +
	"```\n\n" +
	"Pre-commit hooks run: trailing whitespace, Prettier, `goimports`, `go test`, `golangci-lint`, and `npm test`.\n" +
	"(Install goimports once: `go install golang.org/x/tools/cmd/goimports@latest`.)\n\n" +
	"## Structure\n\n" +
	"```\n" +
	"AGENTS.md          → conventions for LLM/coding agents\n" +
	"pkg/cais/          → framework (via dependency)\n" +
	"internal/app/      → bootstrap and routes\n" +
	"internal/handlers/ → HTTP handlers (view.Write)\n" +
	"internal/store/    → SQLite + migrations\n" +
	"web/templates/layouts/ → Amarra layout\n" +
	"web/templates/pages/   → HTML pages\n" +
	"web/static/            → CSS + amarra.js + PWA\n" +
	"cmd/server/        → entry point\n" +
	"```\n\n" +
	"See [AGENTS.md](AGENTS.md) for TDD, Amarra views, flash/CSRF, and generator conventions.\n\n" +
	"## Environment variables\n\n" +
	"| Variable  | Default         | Description      |\n" +
	"| --------- | --------------- | ---------------- |\n" +
	"| PORT      | :8080           | Server port      |\n" +
	"| DB_PATH   | ./data/app.db   | SQLite file path |\n" +
	"| ENV       | development     | Environment      |\n\n" +
	"Health check: GET /health → {\"status\":\"ok\"}\n\n" +
	"Jobs dashboard: GET /jobs (localhost only) — queue counts, failed retry/discard.\n\n" +
	"## Testing on phone (LAN)\n\n" +
	"1. Run `amarra-cais dev` and note the **LAN** URL printed at boot (e.g. `http://192.168.1.10:8080`).\n" +
	"2. Open that URL in mobile Safari/Chrome on the same Wi‑Fi.\n" +
	"3. After template or SSE changes, run `amarra-cais pwa --bump` and reinstall the PWA (or clear site data) so the service worker cache refreshes.\n" +
	"4. Run `amarra-cais doctor --mobile` to catch flash markup, font CSP, and SW cache issues.\n"

const tplREADMEBlank = "# {{.AppName}}\n\n" +
	"Full-stack Go app built with [Amarra](https://github.com/puppe1990/amarra-cais): HTML templates, Tailwind, and SQLite.\n\n" +
	"## Quick start\n\n" +
	"```bash\n" +
	"amarra-cais install  # npm install + go mod tidy\n" +
	"amarra-cais dev        # http://localhost:8080\n" +
	"amarra-cais test       # full test suite\n" +
	"make ci         # test + lint + format-check\n" +
	"```\n\n" +
	"## CI and pre-commit\n\n" +
	"```bash\n" +
	"make pre-commit-install   # once: installs git hooks\n" +
	"make ci                   # test + lint + format-check locally\n" +
	"```\n\n" +
	"## Add your first resource\n\n" +
	"```bash\n" +
	"amarra-cais g resource bookmark --fields title:string,url:url,notes:text?\n" +
	"```\n\n" +
	"This generates:\n" +
	"- Model, migration, admin CRUD, and public list page\n" +
	"- Tests for handlers and store\n" +
	"- Routes with admin protection\n"
