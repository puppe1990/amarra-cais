package cli

// migrationStep is a curated breaking change an app owner must apply when moving
// between framework versions. Versions are pinned to CHANGELOG sections (#192).
type migrationStep struct {
	Version string
	Title   string
	Action  string
}

// frameworkMigrations is maintained per release, newest first.
var frameworkMigrations = []migrationStep{
	{Version: "0.10.0", Title: "netutil.HealthPayload gained an env argument", Action: "call netutil.HealthPayload(status, port, cfg.Env)"},
	{Version: "0.10.0", Title: "jobs dashboard needs Store.DB()", Action: "add DB() *sql.DB to the Store interface in internal/store/store.go"},
	{Version: "0.10.0", Title: "HTMX assets deprecated", Action: "migrate hx-* templates to Amarra Views + Drive"},
	{Version: "0.9.0", Title: ".cais-generated.json tracks generated files", Action: "regenerate resources or accept that destroy skips untracked files"},
	{Version: "0.6.1", Title: "view.Write defaults dynamic pages to Cache-Control: no-store", Action: "set Page.CacheControl on pages that are safe to cache"},
	{Version: "0.5.0", Title: "writeView argument order changed", Action: "call writeView(w, r, views, cfg, layout, name, data, status)"},
	{Version: "0.3.0", Title: "doctor FAILs on hx-* / gonertia", Action: "finish the HTML-first (Amarra) migration — see docs/migrate-inertia.md"},
}

// migrationsBetween returns the steps strictly newer than from and up to to.
// An unknown from (semverCore{}) returns every step up to to; an unknown to
// (e.g. `latest`) returns every step newer than from (#192).
func migrationsBetween(from, to semverCore) []migrationStep {
	var out []migrationStep
	for _, step := range frameworkMigrations {
		v := parseSemverCore(step.Version)
		if !v.OK {
			continue
		}
		if from.OK && compareSemverCore(v, from) <= 0 {
			continue
		}
		if to.OK && compareSemverCore(v, to) > 0 {
			continue
		}
		out = append(out, step)
	}
	return out
}
