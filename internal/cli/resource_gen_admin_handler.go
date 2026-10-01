// Assembles the admin handler source for amarra-cais g resource (#288).
package cli

import (
	"fmt"
	"strings"
)

func buildResourceAdminHandler(data scaffoldData) string {
	parse := buildAdminParseForm(data)
	hasStrconv := needsStrconv(data.Fields) || data.Paginate || hasReferenceFields(data.Fields) || true // BulkDelete parses ids
	hasRefs := hasReferenceFields(data.Fields)
	indexMethod := buildAdminIndexMethod(data)
	showMethod := buildAdminShowMethod(data)
	formDataMethod := ""
	if hasReferenceFields(data.Fields) {
		formDataMethod = buildAdminFormDataMethod(data) + "\n\n"
	}
	paginationImport := ""
	if data.Paginate {
		paginationImport = "\t\"" + frameworkModule + "/pkg/cais/pagination\"\n"
	}
	formsImport := ""
	if hasRefs {
		formsImport = "\t\"" + frameworkModule + "/pkg/cais/forms\"\n"
	}
	newRender := adminFormRender(data, "models."+data.Pascal+"{}", "true", "nil")
	editRender := adminFormRender(data, "item", "false", "nil")
	createErrRender := adminFormRender(data, "item", "true", "errs")
	updateErrRender := adminFormRender(data, "item", "false", "errs")
	colsVar := buildAdminIndexColsVar(data)
	return fmt.Sprintf(`package handlers

import (
	"net/http"
	"net/url"
%s	"strings"

	"%s/pkg/amarra/view"
	"%s/pkg/cais/validate"
%s%s	"%s/pkg/cais"
	"%s/pkg/cais/httpx"
	"%s/pkg/cais/meta"

	"%s/internal/models"
	"%s/internal/store"
)

type Admin%sHandler struct {
	views *view.Renderer
	store store.Store
	site  meta.Site
	cfg   cais.Config
}

func NewAdmin%sHandler(views *view.Renderer, s store.Store, site meta.Site, cfg cais.Config) *Admin%sHandler {
	return &Admin%sHandler{views: views, store: s, site: site, cfg: cfg}
}

%s

%s

%sfunc (h *Admin%sHandler) New(w http.ResponseWriter, r *http.Request) {
	view.Write(w, r, h.views, view.Page{Layout: "app", Name: "admin_%s_form", Data: %s}, h.cfg)
}

func (h *Admin%sHandler) Edit(w http.ResponseWriter, r *http.Request, id int64) {
	item, err := h.store.Find%sByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	view.Write(w, r, h.views, view.Page{Layout: "app", Name: "admin_%s_form", Data: %s}, h.cfg)
}

func (h *Admin%sHandler) Create(w http.ResponseWriter, r *http.Request) {
	item, errs := h.parseForm(r)
	if errs.Any() {
		view.Write(w, r, h.views, view.Page{
			Layout: "app",
			Name:   "admin_%s_form",
			Data:   %s,
			Status: http.StatusUnprocessableEntity,
		}, h.cfg)
		return
	}
	if _, err := h.store.Insert%s(item); err != nil {
		httpx.ServerError(w, err, h.cfg)
		return
	}
	httpx.SeeOther(w, r, "/admin/%s")
}

func (h *Admin%sHandler) Update(w http.ResponseWriter, r *http.Request, id int64) {
	item, errs := h.parseForm(r)
	item.ID = id
	if errs.Any() {
		view.Write(w, r, h.views, view.Page{
			Layout: "app",
			Name:   "admin_%s_form",
			Data:   %s,
			Status: http.StatusUnprocessableEntity,
		}, h.cfg)
		return
	}
	if err := h.store.Update%s(item); err != nil {
		httpx.ServerError(w, err, h.cfg)
		return
	}
	httpx.SeeOther(w, r, "/admin/%s")
}

func (h *Admin%sHandler) Delete(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.store.Delete%s(id); err != nil {
		httpx.ServerError(w, err, h.cfg)
		return
	}
	httpx.SeeOther(w, r, "/admin/%s")
}

func (h *Admin%sHandler) BulkDelete(w http.ResponseWriter, r *http.Request) {
	if err := httpx.ParseFormOrJSON(r); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	for _, raw := range r.Form["ids"] {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || id <= 0 {
			continue
		}
		if err := h.store.Delete%s(id); err != nil {
			httpx.ServerError(w, err, h.cfg)
			return
		}
	}
	httpx.SeeOther(w, r, "/admin/%s")
}

func (h *Admin%sHandler) parseForm(r *http.Request) (models.%s, validate.FieldErrors) {
	%s
}
`,
		boolImport(hasStrconv, "\t\"strconv\"\n"),
		frameworkModule,
		frameworkModule,
		paginationImport,
		formsImport,
		frameworkModule, frameworkModule, frameworkModule, data.ModulePath, data.ModulePath,
		data.PluralPascal,
		data.PluralPascal, data.PluralPascal, data.PluralPascal,
		colsVar+"\n\n"+indexMethod,
		showMethod,
		formDataMethod,
		data.PluralPascal, data.Snake, newRender,
		data.PluralPascal, data.Pascal, data.Snake, editRender,
		data.PluralPascal, data.Snake, createErrRender,
		data.Pascal, data.Plural,
		data.PluralPascal, data.Snake, updateErrRender,
		data.Pascal, data.Plural,
		data.PluralPascal, data.Pascal, data.Plural,
		data.PluralPascal, data.Pascal, data.Plural,
		data.PluralPascal, data.Pascal, parse,
	)
}

func buildAdminFormDataLoader(data scaffoldData) string {
	var lines []string
	for _, f := range data.Fields {
		if f.RefTable == "" {
			continue
		}
		rawVar := "raw" + f.RefPascal + "Opts"
		lines = append(lines, fmt.Sprintf(`	if %s, err := h.store.List%sOptions(); err == nil {
		options := []forms.SelectOption{}
		for _, opt := range %s {
			options = append(options, forms.SelectOption{
				Value: strconv.FormatInt(opt.ID, 10),
				Label: opt.Label,
			})
		}
		data["%sOptions"] = options
	}`, rawVar, f.RefPascal, rawVar, f.RefPascal))
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

func buildAdminFormDataMethod(data scaffoldData) string {
	loader := buildAdminFormDataLoader(data)
	return fmt.Sprintf(`func (h *Admin%sHandler) formData(r *http.Request, item models.%s, isNew bool, errs validate.FieldErrors) map[string]any {
	data := amarraData(r, h.site, map[string]any{
		"Item":   item,
		"IsNew":  isNew,
		"Errors": errs,
	})
%s	return data
}`, data.PluralPascal, data.Pascal, loader)
}
