package view

import (
	"html/template"
	"strings"
)

// LinkTo renders a Drive-enabled anchor. href and label are HTML-escaped.
func LinkTo(href, label string) template.HTML {
	var b strings.Builder
	b.WriteString(`<a href="`)
	template.HTMLEscape(&b, []byte(href))
	b.WriteString(`" data-amarra-drive="true">`)
	template.HTMLEscape(&b, []byte(label))
	b.WriteString(`</a>`)
	return template.HTML(b.String())
}

func helperFuncs() template.FuncMap {
	return template.FuncMap{
		"linkTo": LinkTo,
	}
}
