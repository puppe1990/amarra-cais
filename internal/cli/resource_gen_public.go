package cli

import "fmt"

func buildResourcePublicHandler(data scaffoldData) string {
	intField := firstIntField(data.Fields)

	sumField := "Total"
	if data.Paginate && intField != nil {
		sumField = "Sum"
	}
	listSum := ""
	if intField != nil {
		listSum = fmt.Sprintf(`
	var %s int64
	for _, item := range items {
		%s += item.%s
	}
`, sumField, sumField, intField.Pascal)
	}

	listMethod := buildPublicListMethod(data, sumField, listSum)

	extraStd := "\t\"net/url\"\n\t\"strings\"\n"
	paginationImport := ""
	if data.Paginate {
		extraStd += "\t\"strconv\"\n"
		paginationImport = "\t\"" + frameworkModule + "/pkg/cais/pagination\"\n"
	}

	return fmt.Sprintf(`package handlers

import (
	"net/http"
%s
%s	"%s/pkg/amarra/view"
	"%s/pkg/cais"
	"%s/pkg/cais/httpx"
	"%s/pkg/cais/meta"

	"%s/internal/store"
)

type %sHandler struct {
	views *view.Renderer
	store store.Store
	site  meta.Site
	cfg   cais.Config
}

func New%sHandler(views *view.Renderer, s store.Store, site meta.Site, cfg cais.Config) *%sHandler {
	return &%sHandler{views: views, store: s, site: site, cfg: cfg}
}

%s

%s`,
		extraStd,
		paginationImport,
		frameworkModule, frameworkModule, frameworkModule, frameworkModule, data.ModulePath,
		data.PluralPascal,
		data.PluralPascal, data.PluralPascal, data.PluralPascal,
		listMethod,
		buildPublicShowMethod(data),
	)
}

// buildPublicPublishedFilter keeps drafts out of the public list (#313).
func buildPublicPublishedFilter(data scaffoldData) string {
	if !hasFieldNamed(data.Fields, "published") {
		return ""
	}
	return `
	published := items[:0]
	for _, item := range items {
		if item.Published {
			published = append(published, item)
		}
	}
	items = published
`
}

// buildPublicShowMethod renders a single item by its slug (#313). Drafts 404.
func buildPublicShowMethod(data scaffoldData) string {
	if !hasFieldNamed(data.Fields, "slug") {
		return ""
	}
	publishedCheck := ""
	if hasFieldNamed(data.Fields, "published") {
		publishedCheck = `	if !item.Published {
		http.NotFound(w, r)
		return
	}
`
	}
	return fmt.Sprintf(`func (h *%sHandler) Show(w http.ResponseWriter, r *http.Request, slug string) {
	item, err := h.store.Find%sBySlug(slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}
%s	view.Write(w, r, h.views, view.Page{
		Layout: "app",
		Name:   "%s",
		Data:   amarraData(r, h.site, map[string]any{"Item": item}),
	}, h.cfg)
}`, data.PluralPascal, data.Pascal, publishedCheck, data.Snake)
}

func buildPublicListMethod(data scaffoldData, sumField, listSum string) string {
	sumEntry := ""
	if listSum != "" {
		sumEntry = fmt.Sprintf("\n\t\t%q: %s,", sumField, sumField)
	}
	listFilter := buildPublicPublishedFilter(data)

	if data.Paginate {
		return fmt.Sprintf(`func (h *%sHandler) List(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			page = n
		}
	}
	perPage := 25
	items, total, err := h.store.List%s(q, "", "", page, perPage)
	if err != nil {
		httpx.ServerError(w, err, h.cfg)
		return
	}%s%s
	pg := pagination.New(page, perPage, total)
	listData := amarraData(r, h.site, map[string]any{
		"Items":    items,
		"Title":    "%s",
		"Q":        q,
		"Base":     h.indexBase("/%s", "q", q),%s
		"Page":     pg.Page,
		"Total":    pg.Total,
		"PerPage":  pg.PerPage,
		"HasPrev":  pg.HasPrev,
		"HasNext":  pg.HasNext,
		"PrevPage": pg.PrevPage,
		"NextPage": pg.NextPage,
	})
	view.Write(w, r, h.views, view.Page{
		Layout: "app",
		Name:   "%s",
		Data:   listData,
	}, h.cfg)
}

// indexBase keeps the filters in pagination links.
func (h *%sHandler) indexBase(path string, params ...string) string {
	vals := url.Values{}
	for i := 0; i+1 < len(params); i += 2 {
		if params[i+1] != "" {
			vals.Set(params[i], params[i+1])
		}
	}
	if encoded := vals.Encode(); encoded != "" {
		return path + "?" + encoded
	}
	return path
}
`, data.PluralPascal, data.PluralPascal, listSum, listFilter, data.PluralPascal, data.Plural, sumEntry, data.Plural, data.PluralPascal)
	}
	return fmt.Sprintf(`func (h *%sHandler) List(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	items, err := h.store.ListAll%s(q, "", "")
	if err != nil {
		httpx.ServerError(w, err, h.cfg)
		return
	}%s%s
	view.Write(w, r, h.views, view.Page{
		Layout: "app",
		Name:   "%s",
		Data: amarraData(r, h.site, map[string]any{
			"Items": items,
			"Title": "%s",
			"Q":     q,
			"Base":  h.indexBase("/%s", "q", q),%s
		}),
	}, h.cfg)
}

// indexBase keeps the filters in pagination links.
func (h *%sHandler) indexBase(path string, params ...string) string {
	vals := url.Values{}
	for i := 0; i+1 < len(params); i += 2 {
		if params[i+1] != "" {
			vals.Set(params[i], params[i+1])
		}
	}
	if encoded := vals.Encode(); encoded != "" {
		return path + "?" + encoded
	}
	return path
}
`, data.PluralPascal, data.PluralPascal, listSum, listFilter, data.Plural, data.PluralPascal, data.Plural, sumEntry, data.PluralPascal)
}
