package view

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

const expandMaxPasses = 10000

var innerAction = regexp.MustCompile(`\{\{\s*\.Inner\s*\}\}`)

// ExpandAll rewrites <.component> tags in src into html/template source.
// components maps filename stems to component HTML. Nested tags expand
// innermost-first so slot content is already template source.
func ExpandAll(src string, components map[string]string) (string, error) {
	for pass := 0; pass < expandMaxPasses; pass++ {
		call, ok, err := findInnermostCall(src)
		if err != nil {
			return "", err
		}
		if !ok {
			return src, nil
		}
		expanded, err := expandCall(call, components)
		if err != nil {
			return "", err
		}
		src = src[:call.Start] + expanded + src[call.End:]
	}
	return "", fmt.Errorf("amarra component expansion exceeded %d passes", expandMaxPasses)
}

func expandCall(call componentCall, components map[string]string) (string, error) {
	body, ok := components[call.Name]
	if !ok {
		return "", fmt.Errorf("unknown amarra component %q", call.Name)
	}
	var b strings.Builder
	// html/template $vars live until the enclosing if/with/range/end (or the
	// whole template). Isolate assigns so nested/sibling attrs cannot leak.
	b.WriteString("{{ if true }}")
	for _, attr := range call.Attrs {
		b.WriteString(attrAssign(attr))
	}
	b.WriteString(rewriteComponentBody(body, call.Attrs, call.Inner))
	b.WriteString("{{ end }}")
	return b.String(), nil
}

func attrAssign(attr componentAttr) string {
	if expr, ok := dynamicAttrExpr(attr.Value); ok {
		return fmt.Sprintf("{{ $%s := %s }}", attr.Name, expr)
	}
	return fmt.Sprintf("{{ $%s := %q }}", attr.Name, attr.Value)
}

func dynamicAttrExpr(value string) (string, bool) {
	trim := strings.TrimSpace(value)
	if !strings.HasPrefix(trim, "{{") || !strings.HasSuffix(trim, "}}") {
		return "", false
	}
	inner := strings.TrimSpace(trim[2 : len(trim)-2])
	if inner == "" || strings.Contains(inner, "{{") {
		return "", false
	}
	return inner, true
}

func rewriteComponentBody(body string, attrs []componentAttr, inner string) string {
	for _, attr := range attrs {
		pascal := attrPascal(attr.Name)
		re := regexp.MustCompile(`\{\{\s*\.` + regexp.QuoteMeta(pascal) + `\s*\}\}`)
		body = re.ReplaceAllLiteralString(body, "{{ $"+attr.Name+" }}")
	}
	return innerAction.ReplaceAllLiteralString(body, inner)
}

func attrPascal(name string) string {
	r, size := utf8.DecodeRuneInString(name)
	if r == utf8.RuneError && size == 0 {
		return name
	}
	return string(unicode.ToUpper(r)) + name[size:]
}
