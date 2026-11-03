package view

import (
	"fmt"
	"html/template"
	"net/url"
	"strings"
)

// LinkTo renders an anchor for Drive. href and label are HTML-escaped.
// Drive intercepts same-origin links by default, so no attribute is needed.
// An optional map may set method, confirm, frame, skip (emits
// data-amarra-skip when truthy — a full visit, #398), and disable-with.
func LinkTo(href, label string, opts ...any) template.HTML {
	var method, confirm, frame, skip, disableWith string
	if len(opts) > 0 {
		if m, ok := opts[0].(map[string]any); ok {
			method = optString(m["method"])
			confirm = optString(m["confirm"])
			frame = optString(m["frame"])
			skip = truthyDataAttr(m["skip"])
			disableWith = optString(m["disable-with"])
		}
	}
	var b strings.Builder
	b.WriteString(`<a href="`)
	template.HTMLEscape(&b, []byte(safeHref(href)))
	b.WriteString(`"`)
	writeDataAttr(&b, "data-amarra-method", method)
	writeDataAttr(&b, "data-amarra-confirm", confirm)
	writeDataAttr(&b, "data-amarra-frame", frame)
	writeDataAttr(&b, "data-amarra-skip", skip)
	writeDataAttr(&b, "data-amarra-disable-with", disableWith)
	b.WriteString(`>`)
	template.HTMLEscape(&b, []byte(label))
	b.WriteString(`</a>`)
	return template.HTML(b.String())
}

func optString(v any) string {
	if v == nil {
		return ""
	}
	s := fmt.Sprint(v)
	if s == "<nil>" {
		return ""
	}
	return s
}

// truthyDataAttr returns "true" so writeDataAttr emits a present attribute
// (Drive uses hasAttribute). false / "" / nil stay empty.
func truthyDataAttr(v any) string {
	switch t := v.(type) {
	case bool:
		if t {
			return "true"
		}
		return ""
	case string:
		if t == "" || t == "false" || t == "0" || t == "<nil>" {
			return ""
		}
		return "true"
	}
	return ""
}

// safeHref returns href when it is a relative reference or an allowed scheme,
// and "#" otherwise. LinkTo emits raw template.HTML, bypassing html/template's
// URL filtering, so javascript:/data:/vbscript: must be neutralized here (#116).
// Control characters are stripped for the probe because browsers ignore them
// inside schemes ("java\tscript:").
func safeHref(href string) string {
	probe := strings.Map(func(r rune) rune {
		if r <= ' ' {
			return -1
		}
		return r
	}, href)
	if probe == "" {
		return href
	}
	u, err := url.Parse(probe)
	if err != nil {
		return "#"
	}
	switch strings.ToLower(u.Scheme) {
	case "", "http", "https", "mailto", "tel":
		return href
	}
	return "#"
}

func writeDataAttr(b *strings.Builder, name, value string) {
	if value == "" {
		return
	}
	b.WriteString(` `)
	b.WriteString(name)
	b.WriteString(`="`)
	template.HTMLEscape(b, []byte(value))
	b.WriteString(`"`)
}

// Dict builds a map for template helpers: {{ dict "method" "delete" }}.
func Dict(kv ...any) map[string]any {
	out := make(map[string]any, len(kv)/2)
	for i := 0; i+1 < len(kv); i += 2 {
		out[fmt.Sprint(kv[i])] = kv[i+1]
	}
	return out
}

func helperFuncs() template.FuncMap {
	return template.FuncMap{
		"linkTo":         LinkTo,
		"dict":           Dict,
		"paginationHref": PaginationHref,
		"sortHref":       SortHref,
	}
}

// PaginationHref joins Base with page. Bare paths use ?page=; paths that
// already have a query string use &page=. Empty Base stays relative (?page=).
func PaginationHref(base any, page any) string {
	b := ""
	if s, ok := base.(string); ok {
		b = s
	}
	p := 0
	switch v := page.(type) {
	case int:
		p = v
	case int64:
		p = int(v)
	}
	q := fmt.Sprintf("page=%d", p)
	if b == "" {
		return "?" + q
	}
	sep := "?"
	if strings.Contains(b, "?") {
		sep = "&"
	}
	return b + sep + q
}

// SortHref builds a sort header URL from Base. It keeps existing query
// params (q, etc.), sets sort/dir, and drops page so a new sort starts
// at page 1 (#35).
func SortHref(base, field, currentSort, currentDir any) string {
	b := stringifyAttr(base)
	col := stringifyAttr(field)
	curSort := stringifyAttr(currentSort)
	curDir := stringifyAttr(currentDir)
	nextDir := "asc"
	if col == curSort && strings.EqualFold(curDir, "asc") {
		nextDir = "desc"
	}

	path, vals := splitBaseQuery(b)
	vals.Del("page")
	vals.Del("sort")
	vals.Del("dir")

	var out strings.Builder
	out.WriteString(path)
	out.WriteString("?sort=")
	out.WriteString(url.QueryEscape(col))
	out.WriteString("&dir=")
	out.WriteString(url.QueryEscape(nextDir))
	if enc := vals.Encode(); enc != "" {
		out.WriteByte('&')
		out.WriteString(enc)
	}
	return out.String()
}

func stringifyAttr(v any) string {
	s, ok := v.(string)
	if !ok || s == "<nil>" {
		return ""
	}
	return s
}

func splitBaseQuery(base string) (path string, vals url.Values) {
	vals = url.Values{}
	if base == "" {
		return "", vals
	}
	u, err := url.Parse(base)
	if err != nil {
		return base, vals
	}
	if u.RawQuery != "" {
		vals, _ = url.ParseQuery(u.RawQuery)
	}
	if u.Path != "" || strings.HasPrefix(base, "/") {
		return u.Path, vals
	}
	return "", vals
}
