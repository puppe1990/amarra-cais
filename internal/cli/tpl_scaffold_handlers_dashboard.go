package cli

const tplDashboardHandler = `package handlers

import (
	"net/http"
	"strconv"

	"github.com/puppe1990/amarra-cais/pkg/amarra/view"
	"github.com/puppe1990/amarra-cais/pkg/cais"
	"github.com/puppe1990/amarra-cais/pkg/cais/i18n"
	"github.com/puppe1990/amarra-cais/pkg/cais/meta"

	"{{.ModulePath}}/internal/store"
)

type DashboardHandler struct {
	views   *view.Renderer
	store   store.Store
	site    meta.Site
	catalog *i18n.Catalog
	cfg     cais.Config
}

func NewDashboardHandler(views *view.Renderer, s store.Store, site meta.Site, catalog *i18n.Catalog, cfg cais.Config) *DashboardHandler {
	return &DashboardHandler{views: views, store: s, site: site, catalog: catalog, cfg: cfg}
}

// t translates Go-side copy with the request locale (#211); h.catalog is the
// boot fallback for requests that skipped LocaleMiddleware (unit tests).
func (h *DashboardHandler) t(r *http.Request, key string, args ...any) string {
	return i18n.CatalogOr(r, h.catalog).T(key, args...)
}

func (h *DashboardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	taps := tapsFromRequest(r)
	writeView(w, r, h.views, h.cfg, "app", "dashboard", amarraData(r, h.site, map[string]any{
		"Title":     h.t(r, "dashboard.title"),
		"ActiveNav": "dashboard",
		"Env":       h.cfg.Env,
		"Locale":    i18n.CatalogOr(r, h.catalog).Locale(),
		"Taps":      taps,
		"Outlets":   view.Outlet("amarra-taps", strconv.Itoa(taps)),
	}), 0)
}

// TapPost increments the shell counter cookie and 303s back so Drive can
// apply the outlet from the following GET (#397).
func (h *DashboardHandler) TapPost(w http.ResponseWriter, r *http.Request) {
	n := tapsFromRequest(r) + 1
	http.SetCookie(w, &http.Cookie{
		Name:     "amarra_taps",
		Value:    strconv.Itoa(n),
		Path:     "/",
		MaxAge:   86400 * 365,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   h.cfg.CookieSecure(),
	})
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}
`
