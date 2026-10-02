package view

import (
	"fmt"
	"strings"
)

// applyPassthroughAttrs writes call attrs the component body never referenced
// onto the first HTML tag. Kit form lists action/method/enctype only; extras
// such as data-amarra-skip and class must reach the <form> for Drive (#301).
func applyPassthroughAttrs(expanded, originalBody string, callAttrs []componentAttr) string {
	extra := unusedCallAttrs(originalBody, callAttrs)
	if len(extra) == 0 {
		return expanded
	}
	var b strings.Builder
	for _, a := range extra {
		fmt.Fprintf(&b, ` %s="{{ $%s }}"`, a.Name, attrVar(a.Name))
	}
	return insertBeforeFirstTagClose(expanded, b.String())
}

func unusedCallAttrs(body string, callAttrs []componentAttr) []componentAttr {
	var extra []componentAttr
	for _, a := range callAttrs {
		if a.Name == "" || attrReferencedInBody(body, a) {
			continue
		}
		extra = append(extra, a)
	}
	return extra
}

func attrReferencedInBody(body string, attr componentAttr) bool {
	pascal := attrPascal(attr.Name)
	if pascal == "" {
		return false
	}
	needle := "." + pascal
	from := 0
	for {
		i := strings.Index(body[from:], needle)
		if i < 0 {
			return false
		}
		i += from
		if !identContinueAt(body, i+len(needle)) {
			return true
		}
		from = i + len(needle)
	}
}

func insertBeforeFirstTagClose(body, extra string) string {
	i := 0
	for i < len(body) {
		if strings.HasPrefix(body[i:], "{{") {
			end := strings.Index(body[i:], "}}")
			if end < 0 {
				return body
			}
			i += end + 2
			continue
		}
		if body[i] == '>' {
			return body[:i] + extra + body[i:]
		}
		i++
	}
	return body
}
