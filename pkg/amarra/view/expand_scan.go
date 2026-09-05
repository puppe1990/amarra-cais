package view

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type componentCall struct {
	Name  string
	Attrs []componentAttr
	Inner string
	Start int
	End   int
}

type componentAttr struct {
	Name  string
	Value string
}

type componentOpen struct {
	Name        string
	Attrs       []componentAttr
	SelfClosing bool
	TagEnd      int
}

func findInnermostCall(src string) (componentCall, bool, error) {
	pos := 0
	for {
		start, ok := indexComponentOpen(src, pos)
		if !ok {
			return componentCall{}, false, nil
		}
		open, err := parseComponentOpen(src, start)
		if err != nil {
			return componentCall{}, false, err
		}
		if open.SelfClosing {
			return makeCall(open, start, "", open.TagEnd), true, nil
		}
		closeStart, closeEnd, err := findMatchingClose(src, open.TagEnd, open.Name)
		if err != nil {
			return componentCall{}, false, err
		}
		inner := src[open.TagEnd:closeStart]
		if _, nested := indexComponentOpen(inner, 0); !nested {
			return makeCall(open, start, inner, closeEnd), true, nil
		}
		pos = open.TagEnd
	}
}

func makeCall(open componentOpen, start int, inner string, end int) componentCall {
	return componentCall{
		Name:  open.Name,
		Attrs: open.Attrs,
		Inner: inner,
		Start: start,
		End:   end,
	}
}

func indexComponentOpen(src string, from int) (int, bool) {
	for i := from; i+2 < len(src); i++ {
		if src[i] != '<' || src[i+1] != '.' {
			continue
		}
		r, _ := utf8.DecodeRuneInString(src[i+2:])
		if isNameStart(r) {
			return i, true
		}
	}
	return 0, false
}

func parseComponentOpen(src string, start int) (componentOpen, error) {
	if start+2 >= len(src) || src[start] != '<' || src[start+1] != '.' {
		return componentOpen{}, fmt.Errorf("invalid amarra component tag")
	}
	name, nameEnd, err := readIdent(src, start+2)
	if err != nil {
		return componentOpen{}, err
	}
	attrs, tagEnd, selfClosing, err := parseComponentAttrs(src, nameEnd)
	if err != nil {
		return componentOpen{}, err
	}
	return componentOpen{Name: name, Attrs: attrs, SelfClosing: selfClosing, TagEnd: tagEnd}, nil
}

func parseComponentAttrs(src string, i int) ([]componentAttr, int, bool, error) {
	var attrs []componentAttr
	for {
		i = skipSpace(src, i)
		if i >= len(src) {
			return nil, 0, false, fmt.Errorf("unclosed amarra component tag")
		}
		if src[i] == '>' {
			return attrs, i + 1, false, nil
		}
		if src[i] == '/' {
			if i+1 < len(src) && src[i+1] == '>' {
				return attrs, i + 2, true, nil
			}
			return nil, 0, false, fmt.Errorf("invalid amarra component tag")
		}
		name, next, err := readIdent(src, i)
		if err != nil {
			return nil, 0, false, err
		}
		next = skipSpace(src, next)
		if next >= len(src) || src[next] != '=' {
			return nil, 0, false, fmt.Errorf("invalid amarra component tag")
		}
		next = skipSpace(src, next+1)
		value, next, err := parseQuotedAttr(src, next)
		if err != nil {
			return nil, 0, false, err
		}
		attrs = append(attrs, componentAttr{Name: name, Value: value})
		i = next
	}
}

func parseQuotedAttr(src string, i int) (string, int, error) {
	if i >= len(src) || (src[i] != '"' && src[i] != '\'') {
		return "", 0, fmt.Errorf("invalid amarra component tag")
	}
	quote := src[i]
	i++
	start := i
	for i < len(src) {
		if i+1 < len(src) && src[i] == '{' && src[i+1] == '{' {
			end := strings.Index(src[i+2:], "}}")
			if end < 0 {
				return "", 0, fmt.Errorf("unclosed action in amarra component attribute")
			}
			i += 2 + end + 2
			continue
		}
		if src[i] == quote {
			return src[start:i], i + 1, nil
		}
		i++
	}
	return "", 0, fmt.Errorf("unclosed amarra component attribute")
}

func findMatchingClose(src string, from int, name string) (int, int, error) {
	closeTag := "</." + name + ">"
	openPrefix := "<." + name
	depth := 1
	i := from
	for i < len(src) {
		j := strings.IndexByte(src[i:], '<')
		if j < 0 {
			return 0, 0, fmt.Errorf("unclosed amarra component %q", name)
		}
		i += j
		if strings.HasPrefix(src[i:], closeTag) {
			depth--
			if depth == 0 {
				return i, i + len(closeTag), nil
			}
			i += len(closeTag)
			continue
		}
		if !strings.HasPrefix(src[i:], openPrefix) {
			i++
			continue
		}
		after := i + len(openPrefix)
		if after < len(src) {
			r, _ := utf8.DecodeRuneInString(src[after:])
			if isNameChar(r) {
				i++
				continue
			}
		}
		open, err := parseComponentOpen(src, i)
		if err != nil {
			return 0, 0, err
		}
		if !open.SelfClosing {
			depth++
		}
		i = open.TagEnd
	}
	return 0, 0, fmt.Errorf("unclosed amarra component %q", name)
}

func readIdent(src string, i int) (string, int, error) {
	if i >= len(src) {
		return "", 0, fmt.Errorf("invalid amarra component tag")
	}
	r, size := utf8.DecodeRuneInString(src[i:])
	if !isNameStart(r) {
		return "", 0, fmt.Errorf("invalid amarra component tag")
	}
	j := i + size
	for j < len(src) {
		r, size = utf8.DecodeRuneInString(src[j:])
		if !isNameChar(r) {
			break
		}
		j += size
	}
	return src[i:j], j, nil
}

func skipSpace(src string, i int) int {
	for i < len(src) {
		r, size := utf8.DecodeRuneInString(src[i:])
		if !unicode.IsSpace(r) {
			return i
		}
		i += size
	}
	return i
}

func isNameStart(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

func isNameChar(r rune) bool {
	return r == '_' || r == '-' || unicode.IsLetter(r) || unicode.IsDigit(r)
}
