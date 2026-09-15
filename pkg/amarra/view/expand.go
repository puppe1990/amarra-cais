package view

import (
	"fmt"
	"regexp"
	"strconv"
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
	attrs := defaultComponentAttrs(call.Name, call.Attrs)
	var b strings.Builder
	// html/template $vars live until the enclosing if/with/range/end (or the
	// whole template). Isolate assigns so nested/sibling attrs cannot leak.
	b.WriteString("{{ if true }}")
	for _, attr := range attrs {
		assign, err := attrAssign(attrs, attr)
		if err != nil {
			return "", fmt.Errorf("amarra component %q: %w", call.Name, err)
		}
		b.WriteString(assign)
	}
	b.WriteString(rewriteComponentBody(body, attrs, call.Inner))
	b.WriteString("{{ end }}")
	return b.String(), nil
}

func defaultComponentAttrs(name string, attrs []componentAttr) []componentAttr {
	if name != "form" || hasAttrName(attrs, "enctype") {
		return attrs
	}
	// Optional enctype must be a $var. Bare {{ if .Enctype }} inside
	// {{ range .Items }} looks up Enctype on the row struct and 500s (#43).
	out := make([]componentAttr, len(attrs), len(attrs)+1)
	copy(out, attrs)
	return append(out, componentAttr{Name: "enctype", Value: ""})
}

func hasAttrName(attrs []componentAttr, name string) bool {
	for _, a := range attrs {
		if a.Name == name {
			return true
		}
	}
	return false
}

// attrAssign returns the $var assignment for one attribute. A bare action keeps
// the raw expression (types survive, so if/range still work); a value mixing
// text with actions becomes one printf, so it interpolates like plain markup (#72).
func attrAssign(siblings []componentAttr, attr componentAttr) (string, error) {
	if expr, ok := dynamicAttrExpr(attr.Value); ok {
		return fmt.Sprintf("{{ $%s := %s }}", attr.Name, expr), nil
	}
	if strings.Contains(attr.Value, "{{") {
		expr, err := interpolatedAttrExpr(siblings, attr)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("{{ $%s := %s }}", attr.Name, expr), nil
	}
	return fmt.Sprintf("{{ $%s := %q }}", attr.Name, attr.Value), nil
}

// interpolatedAttrExpr builds `printf "%v%s…" …` for a value that mixes literal
// text with actions: `<.stat value="{{ .Power }} kWp" />` renders "5 kWp"
// instead of printing the action (#72).
func interpolatedAttrExpr(siblings []componentAttr, attr componentAttr) (string, error) {
	parts, err := splitTemplateParts(attr.Value)
	if err != nil {
		return "", fmt.Errorf("attribute %q: %w", attr.Name, err)
	}
	format := ""
	var args []string
	for _, part := range parts {
		if !part.action {
			format += "%s"
			args = append(args, strconv.Quote(part.text))
			continue
		}
		inner := strings.TrimSpace(part.text[2 : len(part.text)-2])
		if templateActionKeyword(inner) {
			return "", fmt.Errorf("attribute %q: %q is a control action — attribute values take expressions ({{ .Field }}, printf, …) or literal text", attr.Name, inner)
		}
		format += "%v"
		args = append(args, "("+rewriteIdentsInAction(inner, siblings)+")")
	}
	return "printf " + strconv.Quote(format) + " " + strings.Join(args, " "), nil
}

type templatePart struct {
	text   string
	action bool
}

// splitTemplateParts splits an attribute value into literal text and actions.
// An unterminated action is an error: it would render literally to the user (#72).
func splitTemplateParts(value string) ([]templatePart, error) {
	var parts []templatePart
	for i := 0; i < len(value); {
		start := strings.Index(value[i:], "{{")
		if start < 0 {
			return append(parts, templatePart{text: value[i:]}), nil
		}
		start += i
		if start > i {
			parts = append(parts, templatePart{text: value[i:start]})
		}
		end := actionEnd(value, start)
		if end < 0 {
			return nil, fmt.Errorf("unclosed action in %q", value)
		}
		parts = append(parts, templatePart{text: value[start:end], action: true})
		i = end
	}
	return parts, nil
}

// actionEnd returns the index just past the "}}" closing the action at start,
// skipping quoted strings so `{{ printf "}}" }}` stays one action.
func actionEnd(src string, start int) int {
	for i := start + 2; i < len(src); {
		switch src[i] {
		case '"', '`':
			i = skipTemplateString(src, i)
		case '}':
			if i+1 < len(src) && src[i+1] == '}' {
				return i + 2
			}
			i++
		default:
			i++
		}
	}
	return -1
}

// templateActionKeyword reports whether inner is a control action, which cannot
// be used as a printf argument.
func templateActionKeyword(inner string) bool {
	fields := strings.Fields(inner)
	if len(fields) == 0 {
		return true
	}
	switch fields[0] {
	case "if", "else", "end", "range", "with", "template", "define", "block", "break", "continue":
		return true
	}
	return false
}

func dynamicAttrExpr(value string) (string, bool) {
	trim := strings.TrimSpace(value)
	if !strings.HasPrefix(trim, "{{") || !strings.HasSuffix(trim, "}}") {
		return "", false
	}
	inner := strings.TrimSpace(trim[2 : len(trim)-2])
	if inner == "" || strings.Contains(inner, "{{") || templateActionKeyword(inner) {
		return "", false
	}
	return inner, true
}

func rewriteComponentBody(body string, attrs []componentAttr, inner string) string {
	body = rewriteAttrIdents(body, attrs)
	return innerAction.ReplaceAllLiteralString(body, inner)
}

func attrPascal(name string) string {
	r, size := utf8.DecodeRuneInString(name)
	if r == utf8.RuneError && size == 0 {
		return name
	}
	return string(unicode.ToUpper(r)) + name[size:]
}
