package view

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// rewriteAttrIdents rewrites .Attr identifiers inside {{ }} actions to $attr.
// if/with/range/pipelines need this; print-only rewrite left {{ if .Error }} on
// page data so kit input errors never rendered. Unspecified fields stay on `.`
// (page data). $.Field is left alone so templates can opt into the root.
func rewriteAttrIdents(body string, attrs []componentAttr) string {
	if len(attrs) == 0 {
		return body
	}
	var b strings.Builder
	i := 0
	for i < len(body) {
		start := strings.Index(body[i:], "{{")
		if start < 0 {
			b.WriteString(body[i:])
			break
		}
		start += i
		b.WriteString(body[i:start])
		endRel := strings.Index(body[start+2:], "}}")
		if endRel < 0 {
			b.WriteString(body[start:])
			break
		}
		end := start + 2 + endRel + 2
		b.WriteString(rewriteIdentsInAction(body[start:end], attrs))
		i = end
	}
	return b.String()
}

func rewriteIdentsInAction(action string, attrs []componentAttr) string {
	var b strings.Builder
	i := 0
	for i < len(action) {
		if action[i] == '"' || action[i] == '`' {
			next := skipTemplateString(action, i)
			b.WriteString(action[i:next])
			i = next
			continue
		}
		if action[i] == '.' && (i == 0 || action[i-1] != '$') {
			if dollar, n, ok := matchAttrIdent(action, i+1, attrs); ok {
				b.WriteByte('$')
				b.WriteString(dollar)
				i += 1 + n
				continue
			}
		}
		b.WriteByte(action[i])
		i++
	}
	return b.String()
}

func matchAttrIdent(action string, identStart int, attrs []componentAttr) (string, int, bool) {
	bestDollar := ""
	bestLen := 0
	rest := action[identStart:]
	for _, attr := range attrs {
		pascal := attrPascal(attr.Name)
		if len(pascal) <= bestLen || !strings.HasPrefix(rest, pascal) {
			continue
		}
		after := identStart + len(pascal)
		if identContinueAt(action, after) {
			continue
		}
		bestDollar = attr.Name
		bestLen = len(pascal)
	}
	if bestLen == 0 {
		return "", 0, false
	}
	return bestDollar, bestLen, true
}

func skipTemplateString(action string, i int) int {
	quote := action[i]
	i++
	if quote == '`' {
		if j := strings.IndexByte(action[i:], '`'); j >= 0 {
			return i + j + 1
		}
		return len(action)
	}
	for i < len(action) {
		if action[i] == '\\' && i+1 < len(action) {
			i += 2
			continue
		}
		if action[i] == '"' {
			return i + 1
		}
		i++
	}
	return len(action)
}

func identContinueAt(s string, i int) bool {
	if i >= len(s) {
		return false
	}
	r, _ := utf8.DecodeRuneInString(s[i:])
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}
