// Resource admin handler Go code generation for cais g resource.
package cli

import (
	"fmt"
	"strings"
)

func buildAdminParseForm(data scaffoldData) string {
	var literal []string
	var after []string
	var validations []string
	for _, f := range data.Fields {
		switch f.GoType {
		case "bool":
			literal = append(literal, fmt.Sprintf("%s: httpx.FormTruthy(r.FormValue(%q))", f.Pascal, f.Name))
		case "int64":
			after = append(after, fmt.Sprintf(`raw%s := strings.TrimSpace(r.FormValue(%q))
	if raw%s == "" {
		errs.Add(%q, %q)
	} else if %sVal, err := strconv.ParseInt(raw%s, 10, 64); err != nil {
		errs.Add(%q, %q)
	} else {
		item.%s = %sVal
	}`, f.Pascal, f.Name, f.Pascal, f.Name, f.Name+" is required", f.Name, f.Pascal, f.Name, f.Name+" must be a number", f.Pascal, f.Name))
		case "*int64":
			after = append(after, fmt.Sprintf(`if raw%s := strings.TrimSpace(r.FormValue(%q)); raw%s != "" {
		if %sVal, err := strconv.ParseInt(raw%s, 10, 64); err != nil {
			errs.Add(%q, %q)
		} else {
			item.%s = &%sVal
		}
	}`, f.Pascal, f.Name, f.Pascal, f.Pascal, f.Pascal, f.Name, f.Name+" must be a number", f.Pascal, f.Pascal))
		case "float64":
			after = append(after, fmt.Sprintf(`raw%s := strings.TrimSpace(r.FormValue(%q))
	if raw%s == "" {
		errs.Add(%q, %q)
	} else if %sVal, err := strconv.ParseFloat(raw%s, 64); err != nil {
		errs.Add(%q, %q)
	} else {
		item.%s = %sVal
	}`, f.Pascal, f.Name, f.Pascal, f.Name, f.Name+" is required", f.Pascal, f.Pascal, f.Name, f.Name+" must be a number", f.Pascal, f.Pascal))
		case "*float64":
			after = append(after, fmt.Sprintf(`if raw%s := strings.TrimSpace(r.FormValue(%q)); raw%s != "" {
		if %sVal, err := strconv.ParseFloat(raw%s, 64); err != nil {
			errs.Add(%q, %q)
		} else {
			item.%s = &%sVal
		}
	}`, f.Pascal, f.Name, f.Pascal, f.Pascal, f.Pascal, f.Name, f.Name+" must be a number", f.Pascal, f.Pascal))
		case "*string":
			if f.HTMLType == "url" {
				after = append(after, fmt.Sprintf(`if raw%s := strings.TrimSpace(r.FormValue(%q)); raw%s != "" {
		if err := validate.URL(raw%s); err != nil {
			errs.Add(%q, err.Error())
		} else {
			item.%s = &raw%s
		}
	}`, f.Pascal, f.Name, f.Pascal, f.Pascal, f.Name, f.Pascal, f.Pascal))
			} else {
				after = append(after, fmt.Sprintf(`if raw%s := strings.TrimSpace(r.FormValue(%q)); raw%s != "" {
		item.%s = &raw%s
	}`, f.Pascal, f.Name, f.Pascal, f.Pascal, f.Pascal))
			}
		default:
			literal = append(literal, fmt.Sprintf("%s: strings.TrimSpace(r.FormValue(%q))", f.Pascal, f.Name))
			if f.Required {
				if f.HTMLType == "url" {
					validations = append(validations, fmt.Sprintf("if item.%s == \"\" {\n\t\terrs.Add(%q, %q)\n\t} else if err := validate.URL(item.%s); err != nil {\n\t\terrs.Add(%q, err.Error())\n\t}", f.Pascal, f.Name, f.Name+" is required", f.Pascal, f.Name))
				} else {
					validations = append(validations, fmt.Sprintf("if item.%s == \"\" {\n\t\terrs.Add(%q, %q)\n\t}", f.Pascal, f.Name, f.Name+" is required"))
				}
			}
		}
	}
	validateBlock := ""
	if len(validations) > 0 {
		validateBlock = "\n\t" + strings.Join(validations, "\n\t") + "\n"
	}
	afterBlock := ""
	if len(after) > 0 {
		afterBlock = "\n\t" + strings.Join(after, "\n\t") + "\n"
	}
	return fmt.Sprintf(`var errs validate.FieldErrors
	if err := httpx.ParseFormOrJSON(r); err != nil {
		errs.Add("_form", err.Error())
		return models.%s{}, errs
	}
	item := models.%s{%s}%s%s	return item, errs`, data.Pascal, data.Pascal, strings.Join(literal, ", "), validateBlock, afterBlock)
}

func buildAdminShowMethod(data scaffoldData) string {
	return fmt.Sprintf(`func (h *Admin%sHandler) Show(w http.ResponseWriter, r *http.Request, id int64) {
	item, err := h.store.Find%sByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	view.Write(w, r, h.views, view.Page{
		Layout: "app",
		Name:   "admin_%s_show",
		Data:   amarraData(r, h.site, map[string]any{"Item": item}),
	}, h.cfg)
}`, data.PluralPascal, data.Pascal, data.Snake)
}

func buildAdminIndexColsVar(data scaffoldData) string {
	var rows []string
	rows = append(rows, "\t{\"Field\": \"\", \"Label\": \"\", \"Sortable\": false},")
	for _, f := range data.Fields {
		rows = append(rows, fmt.Sprintf("\t{\"Field\": %q, \"Label\": %q, \"Sortable\": true},", f.Name, f.Pascal))
	}
	rows = append(rows, "\t{\"Field\": \"\", \"Label\": \"Actions\", \"Sortable\": false},")
	return fmt.Sprintf("var admin%sIndexCols = []map[string]any{\n%s\n}\n", data.PluralPascal, strings.Join(rows, "\n"))
}

func buildAdminIndexMethod(data scaffoldData) string {
	if data.Paginate {
		return fmt.Sprintf(`func (h *Admin%sHandler) Index(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	sort := r.URL.Query().Get("sort")
	dir := r.URL.Query().Get("dir")
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			page = n
		}
	}
	perPage := 25
	items, total, err := h.store.List%s(q, sort, dir, page, perPage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pg := pagination.New(page, perPage, total)
	data := amarraData(r, h.site, map[string]any{
		"Items":    items,
		"Title":    "Admin — %s",
		"Q":        q,
		"Sort":     sort,
		"Dir":      dir,
		"Base":     h.indexBase("/admin/%s", q, sort, dir),
		"Cols":     admin%sIndexCols,
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
		Name:   "admin_%s",
		Data:   data,
	}, h.cfg)
}

// indexBase builds the query string pagination must preserve (filters + sort).
func (h *Admin%sHandler) indexBase(path, q, sort, dir string) string {
	params := url.Values{}
	if q != "" {
		params.Set("q", q)
	}
	if sort != "" {
		params.Set("sort", sort)
		params.Set("dir", dir)
	}
	if encoded := params.Encode(); encoded != "" {
		return path + "?" + encoded
	}
	return path
}`, data.PluralPascal, data.PluralPascal, data.Plural, data.Plural, data.PluralPascal, data.Plural, data.PluralPascal)
	}
	return fmt.Sprintf(`func (h *Admin%sHandler) Index(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	sort := r.URL.Query().Get("sort")
	dir := r.URL.Query().Get("dir")
	items, err := h.store.ListAll%s(q, sort, dir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	view.Write(w, r, h.views, view.Page{
		Layout: "app",
		Name:   "admin_%s",
		Data: amarraData(r, h.site, map[string]any{
			"Items": items,
			"Title": "Admin — %s",
			"Q":     q,
			"Sort":  sort,
			"Dir":   dir,
			"Base":  h.indexBase("/admin/%s", q, sort, dir),
			"Cols":  admin%sIndexCols,
		}),
	}, h.cfg)
}

// indexBase builds the query string pagination must preserve (filters + sort).
func (h *Admin%sHandler) indexBase(path, q, sort, dir string) string {
	params := url.Values{}
	if q != "" {
		params.Set("q", q)
	}
	if sort != "" {
		params.Set("sort", sort)
		params.Set("dir", dir)
	}
	if encoded := params.Encode(); encoded != "" {
		return path + "?" + encoded
	}
	return path
}`, data.PluralPascal, data.PluralPascal, data.Plural, data.Plural, data.Plural, data.PluralPascal, data.PluralPascal)
}

func adminFormRender(data scaffoldData, itemExpr, isNewExpr, errsExpr string) string {
	if hasReferenceFields(data.Fields) {
		return fmt.Sprintf("h.formData(r, %s, %s, %s)", itemExpr, isNewExpr, errsExpr)
	}
	return fmt.Sprintf(`amarraData(r, h.site, map[string]any{
			"Item":   %s,
			"IsNew":  %s,
			"Errors": %s,
		})`, itemExpr, isNewExpr, errsExpr)
}

func buildResourceAdminHandler(data scaffoldData) string {
	parse := buildAdminParseForm(data)
	hasStrconv := true // BulkDelete parses ids; parseForm may too.
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	httpx.SeeOther(w, r, "/admin/%s")
}

func (h *Admin%sHandler) Delete(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.store.Delete%s(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
