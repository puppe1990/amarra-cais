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
<div class="max-w-md mx-auto">
  {{ linkTo "/admin/%s" "← Back" }}
  <h1 class="text-3xl font-bold text-slate-900 mb-6">{{ if .IsNew }}New %s{{ else }}Edit %s{{ end }}</h1>
  {{ $action := "/admin/%s" }}{{ if not .IsNew }}{{ $action = printf "/admin/%s/%%d" .Item.ID }}{{ end }}
  <.form action="{{ $action }}" method="post">
    <div id="admin-%s-errors"></div>
%s
    <.button type="submit">{{ if .IsNew }}Create{{ else }}Save{{ end }}</.button>
  </.form>
</div>
{{ end }}
`, data.Title, data.Title, data.Plural, data.Title, data.Title, data.Plural, data.Plural, data.Snake, fields.String())
}

func buildAdminFormErrorsPartial(data scaffoldData) string {
	return fmt.Sprintf(`{{- define "admin_%s_form_errors" -}}
{{ range $field, $msg := .Errors }}
<p class="text-red-600 text-sm mb-2">{{ $msg }}</p>
{{ end }}
{{- end -}}
`, data.Snake)
}

func adminIndexDisplayField(fields []FieldDef) FieldDef {
	displayField := fields[0]
	for _, f := range fields {
		if f.Name == "title" || f.Name == "name" {
			return f
		}
	}
	return displayField
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
		return fmt.Sprintf("          <td class=\"px-3 py-2\">{{ if .%s }}Yes{{ else }}No{{ end }}</td>\n", f.Pascal)
	case "*int64", "*float64":
		return fmt.Sprintf("          <td class=\"px-3 py-2\">{{ if .%s }}{{ .%s }}{{ end }}</td>\n", f.Pascal, f.Pascal)
	default:
		return fmt.Sprintf("          <td class=\"px-3 py-2\">{{ .%s }}</td>\n", f.Pascal)
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
    <.table cols="{{ .Cols }}" sort="{{ .Sort }}" dir="{{ .Dir }}">
      {{ range .Items }}
      <tr>
%[2]s        <td class="px-3 py-2 text-right space-x-3">
          {{ linkTo (printf "/admin/%[1]s/%%d" .ID) "View" }}
          {{ linkTo (printf "/admin/%[1]s/%%d/edit" .ID) "Edit" }}
          {{ linkTo (printf "/admin/%[1]s/%%d/delete" .ID) "Delete" (dict "method" "post" "confirm" "Delete this %[3]s?") }}
        </td>
      </tr>
      {{ end }}
    </.table>
    {{ if not .Items }}
    <.empty title="No %[3]s yet" href="/admin/%[1]s/new" action="+ New">Nothing here yet.</.empty>
    {{ end }}
%[4]s`, data.Plural, rowCells.String(), data.Title, buildAdminPaginationBlock(data))
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
<div class="max-w-3xl mx-auto">
  <div class="flex items-center justify-between mb-8">
    <h1 class="text-3xl font-bold text-slate-900">%s</h1>
    {{ linkTo "/admin/%s/new" "+ New" }}
  </div>
  <amarra-frame id="admin-%s" class="block bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden">
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
      <dt class="text-sm font-medium text-slate-500">%s</dt>
      <dd class="mt-1 text-slate-900">{{ if .Item.%s }}Yes{{ else }}No{{ end }}</dd>
    </div>
`, f.Pascal, f.Pascal)
			continue
		}
		fmt.Fprintf(&fields, `    <div>
      <dt class="text-sm font-medium text-slate-500">%s</dt>
      <dd class="mt-1 text-slate-900">{{ .Item.%s }}</dd>
    </div>
`, f.Pascal, f.Pascal)
	}
	return fmt.Sprintf(`{{ define "title" }}%s{{ end }} {{ define "content" }}
<div class="max-w-md mx-auto">
  {{ linkTo "/admin/%s" "← Back" }}
  <h1 class="text-3xl font-bold text-slate-900 mb-6">%s</h1>
  <dl class="bg-white rounded-2xl border border-slate-200 p-6 shadow-sm space-y-4">
%s  </dl>
  <div class="mt-6 flex gap-3">
    {{ linkTo (printf "/admin/%s/%%d/edit" .Item.ID) "Edit" }}
    <.form action="{{ printf "/admin/%s/%%d/delete" .Item.ID }}" method="post">
      <.button type="submit">Delete</.button>
    </.form>
  </div>
</div>
{{ end }}
`, data.Title, data.Plural, data.Title, fields.String(), data.Plural, data.Plural)
}

func publicToggleForm(data scaffoldData, f FieldDef) string {
	// csrf_token comes from the <.form> kit (root $.CSRFToken) — no duplicate field here.
	return fmt.Sprintf(`<.form action="{{ printf "/%s/%%d/toggle" .ID }}" method="post"><button type="submit" class="cursor-pointer inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium {{ if .%s }}bg-green-50 text-green-700{{ else }}bg-slate-100 text-slate-600{{ end }}">{{ if .%s }}%s{{ else }}Pending{{ end }}</button></.form>`, data.Plural, f.Pascal, f.Pascal, f.Pascal)
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
		fmt.Fprintf(&b, `<a href="{{ .%s }}" target="_blank" rel="noopener" class="text-lg font-semibold text-copper hover:text-foam">{{ .%s }}</a>`, linkField.Pascal, display.Pascal)
	} else {
		fmt.Fprintf(&b, `<p class="text-lg font-semibold text-slate-800">{{ .%s }}</p>`, display.Pascal)
	}

	var meta []string
	for _, f := range data.Fields {
		if f.Pascal == display.Pascal {
			continue
		}
		switch f.GoType {
		case "bool":
			meta = append(meta, publicToggleForm(data, f))
		case "int64", "*int64", "float64", "*float64":
			meta = append(meta, fmt.Sprintf(`<span class="text-sm text-slate-500">%s: {{ .%s }}</span>`, f.Pascal, f.Pascal))
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
		fmt.Fprintf(&b, `{{ if .%s }}<p class="mt-2 text-sm text-slate-600 line-clamp-2">{{ .%s }}</p>{{ end }}`, f.Pascal, f.Pascal)
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
  <ul id="%[1]s-list" class="space-y-3 mt-4">
    {{ range .Items }}
    <li class="bg-white rounded-2xl border border-slate-200 p-5 shadow-sm">%[2]s</li>
    {{ end }}
  </ul>
  {{ if not .Items }}
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
		totalBlock = fmt.Sprintf(`  <div class="bg-white rounded-2xl border border-slate-200 p-4 shadow-sm mb-6 flex items-center justify-between">
    <span class="text-sm font-medium text-slate-500">Total %s</span>
    <span class="text-2xl font-bold text-copper">{{ .%s }}</span>
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
<div class="max-w-2xl mx-auto">
  <h1 class="text-3xl font-bold text-slate-900 mb-6">%s</h1>
%s%s
</div>
{{ end }}
`, pluralTitle, pluralTitle, totalBlock, listBlock)
}

func buildPublicTogglePartial(data scaffoldData) string {
	boolField := firstBoolField(data.Fields)
	if boolField == nil {
		return ""
	}
	return fmt.Sprintf(`{{- define "%s_toggle" -}}%s{{- end -}}
`, data.Plural, publicToggleForm(data, *boolField))
}
