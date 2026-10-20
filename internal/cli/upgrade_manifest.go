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
	{Version: "0.23.1", Title: "Drive same-id morph reuse docs", Action: "a stable id keeps the same DOM node; listeners and JS properties survive, while value/attributes come from the new HTML and JS-only data-* attributes are removed. Replace dataset.ready init guards with one-time document delegation or teardown (removeEventListener/disconnect/cloneNode) before re-init. No runtime or API change"},
	{Version: "0.23.0", Title: "CSRF Bearer validation; proxied loopback; Drive POST download", Action: "CSRF no longer exempts /api/ or an unvalidated Bearer; session routes under /api/ need csrf_token or X-CSRF-Token, and bearer clients need ADMIN_TOKEN (set it in dev too) or middleware.BearerAPI(prefix, verify). /logs, /jobs and /_dev/reload reject forwarding headers. Drive saves POST file responses as a blob. Run amarra-cais pwa --bump"},
	{Version: "0.22.2", Title: "revalidate CSS/JS; Drive 404 and GET query", Action: "production CSS/JS/webmanifest/sw.js send must-revalidate; fonts/icons keep the year TTL. Drive morphs 404/500 HTML with #amarra-main (404 pushState, 500 does not) and GET forms replace the action query. Run amarra-cais pwa --bump"},
	{Version: "0.22.1", Title: "production static Cache-Control", Action: "production /static 200s send public, max-age=31536000; sw.js is must-revalidate. Drop app wrappers that only set those headers. doctor warns when web/static/fonts exceeds 200 KiB"},
	{Version: "0.22.0", Title: "SignOut and Delete return error", Action: "handle session.Store.Delete and session.SignOut errors in logout and login rotation before redirecting; see docs/security-upgrade.md"},
	{Version: "0.22.0", Title: "admin CRUD requires is_admin", Action: "generated /admin/* CRUD requires users.is_admin plus middleware.RequireAdmin; public signup no longer promotes admin"},
	{Version: "0.21.2", Title: "flash override needs amarra-flash id", Action: "apps overriding web/templates/components/flash.html must keep id=\"amarra-flash\""},
	{Version: "0.15.0", Title: "HTMX removed", Action: "migrate leftover hx-* apps to Views + Drive; Cais v0.11 remains the Inertia product. amarra.js is the only shipped JS bundle"},
	{Version: "0.14.1", Title: "password kit overlay", Action: "rebuild CSS (amarra-cais css) so cais-password-wrap / cais-password-toggle overlay the eye inside <.password>; copy the wrap CSS if the app overrode input.css"},
	{Version: "0.14.0", Title: "PORT required in production", Action: "set PORT=:8080 or PORT=8080 (bare 12-factor numbers work); missing PORT no longer binds :8080 on a shared host"},
	{Version: "0.14.0", Title: "locale-toggle extra locales", Action: "pass locales=\"en,pt,es\" or page .Locales to emit extra language buttons; forms use data-amarra-skip so the cookie applies on a full navigation"},
	{Version: "0.13.3", Title: "Service worker claims /", Action: "in the layout, register(\"/static/js/sw.js\", { scope: \"/\" }); without it the worker stays scoped to /static/js/ and offline.html never intercepts navigations. amarra-cais pwa refreshes amarra.js (Live named-target outerHTML)"},
	{Version: "0.13.3", Title: "g stream chat List/Show maps", Action: "regenerate g stream chat, or pass amarraData maps from List/Show so the layout can read .CSPNonce"},
	{Version: "0.13.2", Title: "Validate rejects invalid explicit env", Action: "fix MAX_BODY_BYTES / PORT / APP_URL / TRUSTED_PROXIES / CSP extras if boot now fails; unset variables still keep their defaults"},
	{Version: "0.13.1", Title: "TokenAuth removed", Action: "replace middleware.TokenAuth with AdminAuth(cfg); the old helper called cais.Load() per request"},
	{Version: "0.13.1", Title: "CSP script-src drops unsafe-inline", Action: "add nonce=\"{{ .CSPNonce }}\" to every inline script in layouts (view.Write injects CSPNonce on map[string]any data); or set CSP_SCRIPT_SRC='unsafe-inline' to restore the old policy"},
	{Version: "0.13.1", Title: "g resource update/delete use PUT and DELETE", Action: "new generators emit PUT /admin/{plural}/{id} and DELETE /admin/{plural}/{id} (edit form _method=put, delete link method delete); existing POST routes keep working until you regenerate"},
	{Version: "0.13.1", Title: "formatMoney and pkg/cais/barcode removed", Action: "if templates call formatMoney, register a local helper; copy Open Food Facts lookup into the app if it imported pkg/cais/barcode"},
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
