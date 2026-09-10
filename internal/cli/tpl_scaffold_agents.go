package cli

// tplAgentsMD is the LLM/agent conventions file shipped with every cais new app.
// Keep it app-scoped (not framework internals). Agents load this first.
const tplAgentsMD = "# {{.AppName}} — AI Conventions\n\n" +
	"Primary reader is often an LLM agent. Prefer small greps, small modules, and headless tests.\n\n" +
	"## Rule #1: TDD is mandatory\n\n" +
	"Before writing production code:\n\n" +
	"1. Write the test (`*_test.go`)\n" +
	"2. Run focused: `go test ./... -v -run TestName`\n" +
	"3. Confirm it **fails** for the right reason\n" +
	"4. Write the **minimal** code to pass\n" +
	"5. Run `amarra-cais test`\n\n" +
	"## Clean code for agents\n\n" +
	"| Priority | Rule |\n" +
	"| -------- | ---- |\n" +
	"| 1 | **Small units** — functions ~4–20 lines; files target 200–300 lines, hard cap ~500 |\n" +
	"| 2 | **SRP** — one reason to change per file/package |\n" +
	"| 3 | **Greppable names** — unique domain nouns; avoid `data`, `handler`, `Manager`, `util` as primary names |\n" +
	"| 4 | **Comments = WHY** — security, SQLite, CSRF/cookie, Drive vs full HTML. No narrating WHAT |\n" +
	"| 5 | **Inject deps** — handlers take `Store`, `*view.Renderer`, `cais.Config` via constructor |\n" +
	"| 6 | **Early returns** — max ~2 nesting levels |\n" +
	"| 7 | **Errors with values** — `fmt.Errorf(\"...: %w\", err)` |\n" +
	"| 8 | **Headless tests** — SQLite `:memory:`; no manual seed for unit tests |\n\n" +
	"## Layout\n\n" +
	"| Path | Responsibility |\n" +
	"| ---- | -------------- |\n" +
	"| `cmd/server/` | Entry point |\n" +
	"| `internal/app/` | Bootstrap, `registerRoutes` |\n" +
	"| `internal/handlers/` | HTTP handlers (`view.Write`) |\n" +
	"| `internal/store/` | SQLite + migrations |\n" +
	"| `internal/models/` | Domain structs |\n" +
	"| `web/templates/layouts/` | Amarra layout (`#amarra-main`) |\n" +
	"| `web/templates/pages/` | HTML pages |\n" +
	"| `web/templates/components/` | App component overrides |\n" +
	"| `web/static/` | CSS, `amarra.js`, PWA |\n\n" +
	"Patch markers (do not remove): `registerRoutes`, `Close() error`, `<!-- cais:nav -->`, `// cais:live-views`.\n\n" +
	"## Amarra HTML\n\n" +
	"Handlers render HTML via `view.Write`:\n\n" +
	"```go\n" +
	"view.Write(w, r, h.views, view.Page{\n" +
	"  Layout: \"app\",\n" +
	"  Name:   \"contact\",\n" +
	"  Data: map[string]any{\n" +
	"    \"Title\":     h.catalog.T(\"contact.title\"),\n" +
	"    \"Site\":      meta.ForRequest(h.site, r),\n" +
	"    \"CSRFToken\": csrf.TokenFromRequest(r),\n" +
	"    \"Flash\":     flashMsg,\n" +
	"  },\n" +
	"}, h.cfg)\n\n" +
	"// Validation — same page, status 422, `.Errors` on inputs\n" +
	"// Flash on redirect — cais cookie API only\n" +
	"flash.Set(w, \"notice\", \"Saved!\", cfg.CookieSecure())\n" +
	"http.Redirect(w, r, \"/dashboard\", http.StatusSeeOther)\n" +
	"```\n\n" +
	"Pages define a content block plus kit tags `<.form>` / `<.input>` / `<.button>` / `<.flash />` / `<.locale-toggle />`.\n" +
	"Drive morphs `#amarra-main` by default — plain links/forms and `linkTo` need no attribute; opt out with `data-amarra-skip` (#31). Do not check `HX-Request`.\n" +
	"Password fields: `fieldPassword` (eye show/hide via `amarra-hook=\"password\"`).\n" +
	"Shipped hooks: `clipboard`, `password`, `reveal` (client show/hide, no Drive round-trip), `theme` (`html.light` + `localStorage[\"amarra-theme\"]`).\n" +
	"Theme FOUC snippet belongs in the layout `<head>` before CSS:\n\n" +
	"```html\n" +
	"<script>\n" +
	"  try {\n" +
	"    if (localStorage.getItem(\"amarra-theme\") === \"light\") document.documentElement.classList.add(\"light\");\n" +
	"  } catch (e) {}\n" +
	"</script>\n" +
	"```\n" +
	"Parse bodies with `httpx.ParseFormOrJSON`.\n\n" +
	"## Auth, CSRF, flash\n\n" +
	"- Session middleware: `LoadSession` + `Flash` + `CSRF(cfg)`\n" +
	"- Protect routes: `middleware.RequireAuth(\"/login\")` / `RequireAuthFunc`\n" +
	"- CSRF: double-submit cookie `cais_csrf` + form field or `X-CSRF-Token`\n" +
	"- Flash: **only** `flash.Set` + read via `flash.MessageFromRequest`\n" +
	"- Dev demo user (when seeded): `demo@example.com` / `password`\n\n" +
	"## New page / resource\n\n" +
	"```bash\n" +
	"amarra-cais g handler settings     # handler + test + web/templates/pages/settings.html + route\n" +
	"amarra-cais g page about           # HTML page only\n" +
	"amarra-cais g resource bookmark --fields title:string,url:url,notes:text?\n" +
	"amarra-cais g model tag --fields name:string\n" +
	"amarra-cais g migration add_notes\n" +
	"amarra-cais g live counter         # WebSocket Live view + /live/counter\n" +
	"amarra-cais g stream chat --live   # chat with Live WS (SSE is the default)\n" +
	"amarra-cais g auth                 # if app was --blank/--minimal\n" +
	"amarra-cais db migrate\n" +
	"```\n\n" +
	"Or by hand:\n\n" +
	"1. Go test in `internal/handlers/`\n" +
	"2. HTML page in `web/templates/pages/`\n" +
	"3. Handler + route in `internal/app/routes.go`\n\n" +
	"## Commands\n\n" +
	"```bash\n" +
	"amarra-cais install          # npm + go mod tidy (+ Tailwind build)\n" +
	"amarra-cais dev              # air + tailwind watch\n" +
	"amarra-cais test             # go test ./...\n" +
	"make ci                     # test + lint + format-check\n" +
	"amarra-cais doctor [--mobile]\n" +
	"amarra-cais routes\n" +
	"amarra-cais db migrate | status | rollback | seed\n" +
	"amarra-cais jobs work | status\n" +
	"```\n\n" +
	"`GET /jobs` — localhost queue dashboard (heartbeats, retry/discard, prune, `?kind=`). Production: SSH tunnel.\n\n" +
	"## Do not\n\n" +
	"- Parse templates per request (`view.Load` once at boot)\n" +
	"- Use inline CSS (Tailwind classes)\n" +
	"- Mock the database (use SQLite `:memory:`)\n" +
	"- Grow files past ~500 lines without splitting\n" +
	"- Ship features without a headless test\n"
