package cli

import "fmt"

func buildResourcePublicHandler(data scaffoldData) string {
	boolField := firstBoolField(data.Fields)
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

	toggleMethod := ""
	if boolField != nil {
		toggleMethod = fmt.Sprintf(`

func (h *%sHandler) Toggle(w http.ResponseWriter, r *http.Request, id int64) {
	item, err := h.store.Find%sByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	item.%s = !item.%s
	if err := h.store.Update%s(item); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/%s", http.StatusSeeOther)
}
`, data.PluralPascal, data.Pascal, boolField.Pascal, boolField.Pascal, data.Pascal, data.Plural)
	}

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

%s%s`,
		extraStd,
		paginationImport,
		frameworkModule, frameworkModule, frameworkModule, data.ModulePath,
		data.PluralPascal,
		data.PluralPascal, data.PluralPascal, data.PluralPascal,
		listMethod,
		toggleMethod,
	)
}

func buildPublicListMethod(data scaffoldData, sumField, listSum string) string {
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}%s
	pg := pagination.New(page, perPage, total)
	listData := amarraData(r, h.site, map[string]any{
		"Items":    items,
		"Title":    "%s",
		"Q":        q,
		"Base":     h.indexBase("/%s", "q", q),
		"%s":       %s,
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
`, data.PluralPascal, data.PluralPascal, listSum, data.PluralPascal, data.Plural, sumField, sumField, data.Plural, data.PluralPascal)
	}
	return fmt.Sprintf(`func (h *%sHandler) List(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	items, err := h.store.ListAll%s(q, "", "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}%s
	view.Write(w, r, h.views, view.Page{
		Layout: "app",
		Name:   "%s",
		Data: amarraData(r, h.site, map[string]any{
			"Items": items,
			"Title": "%s",
			"Q":     q,
			"Base":  h.indexBase("/%s", "q", q),
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
`, data.PluralPascal, data.PluralPascal, listSum, data.Plural, data.PluralPascal, data.Plural, data.PluralPascal)
}
