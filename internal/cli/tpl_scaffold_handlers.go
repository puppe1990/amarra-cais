package cli

const tplViewData = `package handlers

import (
	"net/http"

	"github.com/puppe1990/amarra-cais/pkg/amarra/view"
	"github.com/puppe1990/amarra-cais/pkg/cais"
	"github.com/puppe1990/amarra-cais/pkg/cais/meta"
)

func amarraData(r *http.Request, site meta.Site, extra map[string]any) map[string]any {
	s := meta.ForRequest(site, r)
	data := map[string]any{
		"Site":      s,
		"CSRFToken": s.CSRFToken,
		"Flash":     s.Flash,
	}
	for k, v := range extra {
		data[k] = v
	}
	return data
}

func writeView(w http.ResponseWriter, r *http.Request, views *view.Renderer, cfg cais.Config, name string, data any, status int) {
	view.Write(w, r, views, view.Page{Layout: "app", Name: name, Data: data, Status: status}, cfg)
}
`

// Generic handler generator templates (HTML + Amarra).
const tplGenericHandler = `package handlers

import (
	"net/http"

	"github.com/puppe1990/amarra-cais/pkg/amarra/view"
	"github.com/puppe1990/amarra-cais/pkg/cais"
	"github.com/puppe1990/amarra-cais/pkg/cais/i18n"
	"github.com/puppe1990/amarra-cais/pkg/cais/meta"
)

type {{.Pascal}}Handler struct {
	views   *view.Renderer
	site    meta.Site
	catalog *i18n.Catalog
	cfg     cais.Config
}

func New{{.Pascal}}Handler(views *view.Renderer, site meta.Site, catalog *i18n.Catalog, cfg cais.Config) *{{.Pascal}}Handler {
	return &{{.Pascal}}Handler{views: views, site: site, catalog: catalog, cfg: cfg}
}

func (h *{{.Pascal}}Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	view.Write(w, r, h.views, view.Page{
		Layout: "app",
		Name:   "{{.Snake}}",
		Data: amarraData(r, h.site, map[string]any{
			"Title": "{{.Title}}",
		}),
	}, h.cfg)
}
`

const tplGenericHandlerTest = `package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/puppe1990/amarra-cais/pkg/cais"

	appi18n "{{.ModulePath}}/internal/i18n"
)

func Test{{.Pascal}}Handler_RendersHTML(t *testing.T) {
	h := New{{.Pascal}}Handler(setupTestViews(t), testSite(), appi18n.DefaultCatalog(), cais.Config{})

	req := httptest.NewRequest(http.MethodGet, "/{{.Snake}}", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "{{.Title}}") {
		t.Errorf("body missing title, got: %s", body)
	}
}
`

const tplGenericPage = `{{"{{"}} define "content" {{"}}"}}
<div class="p-8">
  <h1 class="text-2xl font-bold">{{.Title}}</h1>
  <p>{{.Title}} page — customize this HTML template.</p>
</div>
{{"{{"}} end {{"}}"}}
`
