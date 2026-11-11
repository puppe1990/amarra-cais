// Resource admin/public HTML template generation for cais g resource.
package cli

import (
	"fmt"
	"strings"
)

func buildAdminFormHTML(data scaffoldData) string {
	var fields strings.Builder
	for _, f := range data.Fields {
		switch f.Widget {
		case "select":
			if f.GoType == "*int64" {
				fmt.Fprintf(&fields, `    {{ fieldSelect (makeSelectFieldPtr "%s" "%s" .Item.%s .%sOptions %t .Errors) }}
`, f.Name, f.RefPascal, f.Pascal, f.RefPascal, f.Required)
			} else {
				fmt.Fprintf(&fields, `    {{ fieldSelect (makeSelectField "%s" "%s" .Item.%s .%sOptions %t .Errors) }}
`, f.Name, f.RefPascal, f.Pascal, f.RefPascal, f.Required)
			}
		case "textarea":
			fmt.Fprintf(&fields, `    {{ fieldInput (makeField "%s" "%s" .Item.%s "textarea" %t .Errors) }}
`, f.Name, f.Pascal, f.Pascal, f.Required)
		case "checkbox":
			fmt.Fprintf(&fields, `    {{ if .Item.%[3]s }}<.checkbox name="%[1]s" label="%[2]s" checked="checked" error="{{ fieldError .Errors "%[1]s" }}" />{{ else }}<.checkbox name="%[1]s" label="%[2]s" error="{{ fieldError .Errors "%[1]s" }}" />{{ end }}
`, f.Name, f.Pascal, f.Pascal)
		default:
			fmt.Fprintf(&fields, `    {{ fieldInput (makeField "%s" "%s" .Item.%s "%s" %t .Errors) }}
`, f.Name, f.Pascal, f.Pascal, f.HTMLType, f.Required)
		}
	}
	return fmt.Sprintf(`{{ define "title" }}{{ if .IsNew }}New %s{{ else }}Edit %s{{ end }}{{ end }} {{ define "content" }}
<div class="mx-auto max-w-md">
  <a href="/admin/%s" class="inline-flex items-center gap-2 text-sm text-foam/60 hover:text-foam"><svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 12H5m0 0l6-6m-6 6l6 6" /></svg>Back</a>
  <h1 class="mt-4 font-serif text-3xl font-medium text-foam">{{ if .IsNew }}New %s{{ else }}Edit %s{{ end }}</h1>
  {{ $action := "/admin/%s" }}{{ if not .IsNew }}{{ $action = printf "/admin/%s/%%d" .Item.ID }}{{ end }}
  <.form action="{{ $action }}" method="post" class="mt-8 space-y-5">
    {{ if not .IsNew }}<input type="hidden" name="_method" value="put" />{{ end }}
    <div id="admin-%s-errors"></div>
%s
    <.button type="submit" class="amarra-btn amarra-btn-primary">{{ if .IsNew }}Create{{ else }}Save{{ end }}</.button>
  </.form>
</div>
{{ end }}
`, data.Title, data.Title, data.Plural, data.Title, data.Title, data.Plural, data.Plural, data.Snake, fields.String())
}

func buildAdminFormErrorsPartial(data scaffoldData) string {
	return fmt.Sprintf(`{{- define "admin_%s_form_errors" -}}
{{ range $field, $msg := .Errors }}
<p class="mb-2 text-sm text-danger">{{ $msg }}</p>
{{ end }}
{{- end -}}
`, data.Snake)
}

// Row trash icon (heroicons trash outline) shared by index rows and show page.
const adminTrashSVG = `<svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>`

// adminDeleteModal renders a trash button opening a "tem certeza?" modal.
// Confirm posts via a Drive link (CSRF rides the X-CSRF-Token header), not a
// form: rows live inside the bulk-delete form and nested forms are invalid
// HTML. deleteLink is the full {{ linkTo ... }} delete expression.
func adminDeleteModal(plural, title, deleteLink string) string {
	return fmt.Sprintf(`<div amarra-hook="dialog" class="relative inline-block">
              <button type="button" data-amarra-dialog-open aria-label="Delete %[2]s" class="inline-flex h-8 w-8 items-center justify-center rounded-lg border border-foam/10 text-foam/50 hover:border-danger/40 hover:text-danger">%[3]s</button>
              <.modal>
                <p class="text-sm text-foam/80">Delete this %[2]s?</p>
                <div class="mt-6 flex items-center gap-4">%[4]s<button type="button" data-amarra-dialog-close class="text-sm text-foam/60 hover:text-foam">Cancel</button></div>
              </.modal>
            </div>`, plural, title, adminTrashSVG, deleteLink)
}

func buildAdminPaginationBlock(data scaffoldData) string {
	if !data.Paginate {
		return ""
	}
	return `    <.pagination base="{{ .Base }}" />`
}

func adminIndexRowCell(f FieldDef) string {
	switch f.GoType {
	case "bool":
		return fmt.Sprintf("          <td class=\"px-4 py-3 text-sm text-foam/80\">{{ if .%s }}Yes{{ else }}No{{ end }}</td>\n", f.Pascal)
	case "*int64", "*float64":
		return fmt.Sprintf("          <td class=\"px-4 py-3 text-sm tabular-nums text-foam/80\">{{ if .%s }}{{ .%s }}{{ end }}</td>\n", f.Pascal, f.Pascal)
	default:
		// Long URLs would widen the table past the frame and clip the row
		// actions; clamp them and keep the full value in the title.
		if f.HTMLType == "url" {
			return fmt.Sprintf("          <td class=\"px-4 py-3 text-sm text-foam/80\"><div class=\"max-w-[16rem] truncate\" title=\"{{ .%s }}\">{{ .%s }}</div></td>\n", f.Pascal, f.Pascal)
		}
		return fmt.Sprintf("          <td class=\"px-4 py-3 text-sm text-foam/80\">{{ .%s }}</td>\n", f.Pascal)
	}
}

func buildAdminIndexPanel(data scaffoldData) string {
	var rowCells strings.Builder
	for _, f := range data.Fields {
		rowCells.WriteString(adminIndexRowCell(f))
	}
	return fmt.Sprintf(`    <.filters action="/admin/%[1]s" clear="/admin/%[1]s">
      {{ if .Sort }}<input type="hidden" name="sort" value="{{ .Sort }}" />{{ end }}
      {{ if .Dir }}<input type="hidden" name="dir" value="{{ .Dir }}" />{{ end }}
      <.input name="q" label="Search" value="{{ .Q }}" />
    </.filters>
    {{ if .Items }}
    <div amarra-hook="bulk">
      <.form action="/admin/%[1]s/bulk-delete" method="post">
        <div class="mb-3 hidden items-center gap-4" data-amarra-bulk-bar hidden>
          <span class="text-sm text-foam/60"><span class="font-mono tabular-nums" data-amarra-bulk-count>0</span> selected</span>
          <div amarra-hook="dialog">
            <button type="button" data-amarra-dialog-open class="text-sm text-danger hover:underline">Delete selected</button>
            <.modal>
              <p class="text-sm text-foam/80">Delete the selected %[3]s?</p>
              <.button type="submit" class="amarra-btn amarra-btn-danger">Delete</.button>
              <button type="button" data-amarra-dialog-close class="ml-4 text-sm text-foam/60 hover:text-foam">Cancel</button>
            </.modal>
          </div>
        </div>
        <label class="mb-3 inline-flex items-center gap-2 text-sm text-foam/70">
          <input type="checkbox" class="accent-copper" data-amarra-bulk-all />
          Select page
        </label>
        <.table cols="{{ .Cols }}" sort="{{ .Sort }}" dir="{{ .Dir }}" base="{{ .Base }}">
          {{ range .Items }}
          <tr>
            <td class="px-4 py-3"><input type="checkbox" class="accent-copper" name="ids" value="{{ .ID }}" data-amarra-bulk-row /></td>
%[2]s            <td class="px-4 py-3 text-right whitespace-nowrap">
              <a href="/admin/%[1]s/{{ .ID }}/edit" class="amarra-btn amarra-btn-secondary h-8 px-3">Edit</a>
              %[5]s
            </td>
          </tr>
          {{ end }}
        </.table>
      </.form>
    </div>
     {{ else }}
    <.empty title="No %[3]s yet" href="/admin/%[1]s/new" action="New %[3]s">Nothing here yet.</.empty>
    {{ end }}
%[4]s`, data.Plural, rowCells.String(), data.Title, buildAdminPaginationBlock(data),
		adminDeleteModal(data.Plural, data.Title, fmt.Sprintf(`{{ linkTo (printf "/admin/%s/%%d" .ID) "Delete" (dict "method" "delete") }}`, data.Plural)))
}

func buildAdminIndexPartial(data scaffoldData) string {
	if !data.Paginate {
		return ""
	}
	return fmt.Sprintf(`{{- define "admin_%s_index" -}}
%s
{{- end -}}
`, data.Plural, buildAdminIndexPanel(data))
}

func buildAdminIndexHTML(data scaffoldData) string {
	panel := buildAdminIndexPanel(data)
	if data.Paginate {
		panel = fmt.Sprintf(`{{ template "admin_%s_index" . }}`, data.Plural)
	}
	return fmt.Sprintf(`{{ define "title" }}Admin — %s{{ end }} {{ define "content" }}
<div class="mx-auto max-w-3xl">
  <div class="mb-6 flex items-center justify-between gap-4">
    <h1 class="font-serif text-3xl font-medium text-foam">%s</h1>
    <a href="/admin/%s/new" class="amarra-btn amarra-btn-primary"><svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 5v14M5 12h14" /></svg>New</a>
  </div>
  <amarra-frame id="admin-%s" class="block overflow-hidden rounded-xl border border-foam/10 bg-tide/40">
%s
  </amarra-frame>
</div>
{{ end }}
`, data.Title, data.Title, data.Plural, data.Plural, panel)
}

func buildAdminShowHTML(data scaffoldData) string {
	var fields strings.Builder
	for _, f := range data.Fields {
		if f.GoType == "bool" {
			fmt.Fprintf(&fields, `    <div>
      <dt class="text-sm text-foam/60">%s</dt>
      <dd class="mt-1 text-sm text-foam">{{ if .Item.%s }}Yes{{ else }}No{{ end }}</dd>
    </div>
`, f.Pascal, f.Pascal)
			continue
		}
		fmt.Fprintf(&fields, `    <div>
      <dt class="text-sm text-foam/60">%s</dt>
      <dd class="mt-1 text-sm text-foam">{{ .Item.%s }}</dd>
    </div>
`, f.Pascal, f.Pascal)
	}
	return fmt.Sprintf(`{{ define "title" }}%s{{ end }} {{ define "content" }}
<div class="mx-auto max-w-md">
  <a href="/admin/%s" class="inline-flex items-center gap-2 text-sm text-foam/60 hover:text-foam"><svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 12H5m0 0l6-6m-6 6l6 6" /></svg>Back</a>
  <h1 class="mt-4 font-serif text-3xl font-medium text-foam">%s</h1>
  <dl class="mt-8 space-y-4 rounded-xl border border-foam/10 bg-tide/40 p-6">
%s  </dl>
  <div class="mt-6 flex items-center gap-3">
    <a href="/admin/%s/{{ .Item.ID }}/edit" class="amarra-btn amarra-btn-secondary">Edit</a>
    %s
  </div>
</div>
{{ end }}
`, data.Title, data.Plural, data.Title, fields.String(), data.Plural,
		adminDeleteModal(data.Plural, data.Title, fmt.Sprintf(`{{ linkTo (printf "/admin/%s/%%d" .Item.ID) "Delete" (dict "method" "delete") }}`, data.Plural)))
}

func buildPublicShowHTML(data scaffoldData) string {
	var fields strings.Builder
	for _, f := range data.Fields {
		if f.Widget == "textarea" {
			fmt.Fprintf(&fields, `    <div>
      <dt class="text-sm text-foam/60">%s</dt>
      <dd class="mt-1 whitespace-pre-wrap text-sm text-foam">{{ .Item.%s }}</dd>
    </div>
`, f.Pascal, f.Pascal)
			continue
		}
		fmt.Fprintf(&fields, `    <div>
      <dt class="text-sm text-foam/60">%s</dt>
      <dd class="mt-1 text-sm text-foam">{{ .Item.%s }}</dd>
    </div>
`, f.Pascal, f.Pascal)
	}
	return fmt.Sprintf(`{{ define "title" }}%s{{ end }} {{ define "content" }}
<div class="mx-auto max-w-2xl">
  {{ linkTo "/%s" "Back" }}
  <h1 class="mt-4 font-serif text-3xl font-medium text-foam">{{ .Item.%s }}</h1>
  <dl class="mt-8 space-y-4 rounded-xl border border-foam/10 bg-tide/40 p-6">
%s  </dl>
</div>
{{ end }}
`, data.Title, data.Plural, displayFieldForList(data.Fields).Pascal, fields.String())
}

func publicBoolBadge(f FieldDef) string {
	// #261: public list is read-only; mutating the bool belongs on the admin form.
	return fmt.Sprintf(`<span class="inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium {{ if .%s }}bg-success/15 text-success{{ else }}bg-foam/10 text-foam/60{{ end }}">{{ if .%s }}%s{{ else }}No{{ end }}</span>`, f.Pascal, f.Pascal, f.Pascal)
}

func buildPublicListItemHTML(data scaffoldData) string {
	display := displayFieldForList(data.Fields)
	var linkField *FieldDef
	for i, f := range data.Fields {
		if f.HTMLType == "url" {
			linkField = &data.Fields[i]
			break
		}
	}

	var b strings.Builder
	if linkField != nil {
		fmt.Fprintf(&b, `<a href="{{ .%s }}" target="_blank" rel="noopener" class="text-[15px] font-medium text-foam underline decoration-foam/40 underline-offset-4 hover:decoration-copper/70">{{ .%s }}</a>`, linkField.Pascal, display.Pascal)
	} else {
		fmt.Fprintf(&b, `<p class="text-[15px] font-medium text-foam">{{ .%s }}</p>`, display.Pascal)
	}

	var meta []string
	for _, f := range data.Fields {
		if f.Pascal == display.Pascal {
			continue
		}
		switch f.GoType {
		case "bool":
			meta = append(meta, publicBoolBadge(f))
		case "int64", "*int64", "float64", "*float64":
			meta = append(meta, fmt.Sprintf(`<span class="font-mono text-xs tabular-nums text-foam/60">%s: {{ .%s }}</span>`, f.Pascal, f.Pascal))
		}
	}
	if len(meta) > 0 {
		b.WriteString(`<div class="mt-2 flex flex-wrap items-center gap-2">`)
		b.WriteString(strings.Join(meta, "\n"))
		b.WriteString(`</div>`)
	}

	for _, f := range data.Fields {
		if f.Pascal == display.Pascal || f.Widget != "textarea" {
			continue
		}
		fmt.Fprintf(&b, `{{ if .%s }}<p class="mt-2 text-sm text-foam/60 line-clamp-2">{{ .%s }}</p>{{ end }}`, f.Pascal, f.Pascal)
	}

	return b.String()
}

func buildPublicPaginationBlock(data scaffoldData) string {
	if !data.Paginate {
		return ""
	}
	return `  <.pagination base="{{ .Base }}" />`
}

func buildPublicListPanel(data scaffoldData) string {
	itemBlock := buildPublicListItemHTML(data)
	return fmt.Sprintf(`  <.filters action="/%[1]s" clear="/%[1]s">
    <.input name="q" label="Search" value="{{ .Q }}" />
  </.filters>
  {{ if .Items }}
  <ul id="%[1]s-list" class="mt-4 space-y-3">
    {{ range .Items }}
    <li class="rounded-xl border border-foam/10 bg-tide/40 p-5">%[2]s</li>
    {{ end }}
  </ul>
  {{ else }}
  <.empty title="No %[3]s yet">Nothing here yet.</.empty>
  {{ end }}
%[4]s`, data.Plural, itemBlock, data.Title, buildPublicPaginationBlock(data))
}

func buildPublicListPartial(data scaffoldData) string {
	if !data.Paginate {
		return ""
	}
	return fmt.Sprintf(`{{- define "%s_list" -}}
%s
{{- end -}}
`, data.Plural, buildPublicListPanel(data))
}

func buildPublicListHTML(data scaffoldData) string {
	pluralTitle := toTitle(data.Plural)
	intField := firstIntField(data.Fields)
	totalBlock := ""
	if intField != nil {
		totalVar := "Total"
		if data.Paginate {
			totalVar = "Sum"
		}
		totalBlock = fmt.Sprintf(`  <div class="mb-6 flex items-center justify-between rounded-xl border border-foam/10 bg-tide/40 px-5 py-4">
    <span class="text-sm text-foam/60">Total %s</span>
    <span class="font-mono text-2xl tabular-nums leading-none text-foam">{{ .%s }}</span>
  </div>
`, intField.Pascal, totalVar)
	}
	listBlock := buildPublicListPanel(data)
	if data.Paginate {
		listBlock = fmt.Sprintf(`  <amarra-frame id="%s-panel">
{{ template "%s_list" . }}
  </amarra-frame>`, data.Plural, data.Plural)
	} else {
		listBlock = fmt.Sprintf("  <amarra-frame id=\"%s-panel\">\n%s  </amarra-frame>\n", data.Plural, listBlock)
	}
	return fmt.Sprintf(`{{ define "title" }}%s{{ end }} {{ define "content" }}
<div class="mx-auto max-w-2xl">
  <h1 class="mb-6 font-serif text-3xl font-medium text-foam">%s</h1>
%s%s
</div>
{{ end }}
`, pluralTitle, pluralTitle, totalBlock, listBlock)
}
