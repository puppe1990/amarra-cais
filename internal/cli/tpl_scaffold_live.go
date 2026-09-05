package cli

const tplLiveView = `package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/puppe1990/amarra-cais/pkg/amarra/live"
	"github.com/puppe1990/amarra-cais/pkg/amarra/view"
	"github.com/puppe1990/amarra-cais/pkg/cais"
	"github.com/puppe1990/amarra-cais/pkg/cais/i18n"
	"github.com/puppe1990/amarra-cais/pkg/cais/meta"
)

type {{.Pascal}}Live struct {
	n int
}

func New{{.Pascal}}Live() *{{.Pascal}}Live {
	return &{{.Pascal}}Live{}
}

func (v *{{.Pascal}}Live) Mount(context.Context, live.Socket) error { return nil }

func (v *{{.Pascal}}Live) Handle(_ context.Context, ev live.Event) error {
	switch ev.Name {
	case "inc":
		v.n++
		return nil
	case "dec":
		v.n--
		return nil
	default:
		return fmt.Errorf("unknown event %s", ev.Name)
	}
}

func (v *{{.Pascal}}Live) Render() live.Rendered {
	return live.Rendered{Target: "count", HTML: "<span id=\"count\">" + strconv.Itoa(v.n) + "</span>"}
}

type {{.Pascal}}Page struct {
	views   *view.Renderer
	site    meta.Site
	catalog *i18n.Catalog
	cfg     cais.Config
}

func New{{.Pascal}}Page(views *view.Renderer, site meta.Site, catalog *i18n.Catalog, cfg cais.Config) *{{.Pascal}}Page {
	return &{{.Pascal}}Page{views: views, site: site, catalog: catalog, cfg: cfg}
}

func (h *{{.Pascal}}Page) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	view.Write(w, r, h.views, view.Page{
		Layout: "app",
		Name:   "{{.Snake}}",
		Data: amarraData(r, h.site, map[string]any{
			"Title":     h.catalog.T("home.title"),
			"ActiveNav": "{{.Snake}}",
		}),
	}, h.cfg)
}
`

const tplLivePage = `{{"{{"}} define "content" {{"}}"}}
<section amarra-live="{{.Snake}}" data-amarra-topic="{{.Snake}}" class="mx-auto max-w-xl py-16 text-center">
  <p class="font-mono text-[11px] uppercase tracking-[0.32em] text-copper">Live</p>
  <p class="mt-6 font-serif text-6xl text-foam"><span id="count">0</span></p>
  <div class="mt-8 flex items-center justify-center gap-4">
    <button type="button" amarra-click="dec" class="bg-copper px-5 py-2 font-mono text-[11px] uppercase tracking-[0.22em] text-ink">Dec</button>
    <button type="button" amarra-click="inc" class="bg-copper px-5 py-2 font-mono text-[11px] uppercase tracking-[0.22em] text-ink">Inc</button>
  </div>
</section>
{{"{{"}} end {{"}}"}}
`

const tplLiveTest = `package handlers

import (
	"context"
	"testing"

	"github.com/puppe1990/amarra-cais/pkg/amarra/live"
)

func Test{{.Pascal}}Live_Inc(t *testing.T) {
	v := New{{.Pascal}}Live()
	if err := v.Handle(context.Background(), live.Event{Name: "inc"}); err != nil {
		t.Fatal(err)
	}
	out := v.Render()
	if out.HTML != ` + "`" + `<span id="count">1</span>` + "`" + ` {
		t.Fatalf("html = %s", out.HTML)
	}
}
`
