package view

import (
	"fmt"
	"html/template"
	"strings"
)

// LinkTo renders an anchor for Drive. href and label are HTML-escaped.
// Drive intercepts same-origin links by default, so no attribute is needed;
// opt out with data-amarra-skip on the link when a full visit is required.
// An optional map may set method, confirm, and frame (Turbo-shaped attrs).
func LinkTo(href, label string, opts ...any) template.HTML {
	var method, confirm, frame string
	if len(opts) > 0 {
		if m, ok := opts[0].(map[string]any); ok {
			method = fmt.Sprint(m["method"])
			if method == "<nil>" || method == "" {
				method = ""
			}
			confirm = fmt.Sprint(m["confirm"])
			if confirm == "<nil>" || confirm == "" {
				confirm = ""
			}
			frame = fmt.Sprint(m["frame"])
			if frame == "<nil>" || frame == "" {
				frame = ""
			}
		}
	}
	var b strings.Builder
	b.WriteString(`<a href="`)
	template.HTMLEscape(&b, []byte(href))
	b.WriteString(`"`)
	writeDataAttr(&b, "data-amarra-method", method)
	writeDataAttr(&b, "data-amarra-confirm", confirm)
	writeDataAttr(&b, "data-amarra-frame", frame)
	b.WriteString(`>`)
	template.HTMLEscape(&b, []byte(label))
	b.WriteString(`</a>`)
	return template.HTML(b.String())
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
		"linkTo": LinkTo,
		"dict":   Dict,
	}
}
