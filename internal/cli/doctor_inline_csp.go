package cli

import (
	"fmt"
	"regexp"
	"strings"
)

// #319: production CSP is script-src 'self' + nonce — inline <script> without
// nonce and inline handlers (onclick...) die silently in production only.
func checkInlineScripts(dir string) doctorCheck {
	const name = "inline scripts/handlers (CSP)"
	markup, err := readTemplateMarkup(dir)
	if err != nil {
		return doctorCheck{Name: name, OK: true, Detail: err.Error()}
	}
	var problems []string
	if n := countScriptTagsWithoutNonce(markup); n > 0 {
		problems = append(problems, fmt.Sprintf("%d <script> sem nonce", n))
	}
	if n := countInlineHandlers(markup); n > 0 {
		problems = append(problems, fmt.Sprintf("%d handler inline (onclick/onchange/...)", n))
	}
	if swRegistersWithoutScope(markup) {
		problems = append(problems, "serviceWorker.register sem { scope: \"/\" }")
	}
	if len(problems) == 0 {
		return doctorCheck{Name: name, OK: true}
	}
	return doctorCheck{
		Name:     name,
		Optional: true,
		Detail:   strings.Join(problems, "; "),
		FixHint:  `script inline: nonce="{{ .CSPNonce }}" (ou CSP_SCRIPT_SRC='unsafe-inline'); handlers: hooks Amarra (sidebar, reveal, dialog) ou data-amarra-*; SW: register("/static/js/sw.js", { scope: "/" })`,
	}
}

var (
	scriptTagOpen     = regexp.MustCompile(`(?i)<script\b[^>]*>`)
	inlineHandlerAttr = regexp.MustCompile(`(?i)\son(?:click|change|submit|input|load)\s*=`)
	swRegisterCall    = regexp.MustCompile(`serviceWorker\.register\(([^)]*)\)`)
)

func countScriptTagsWithoutNonce(markup string) int {
	n := 0
	for _, open := range scriptTagOpen.FindAllString(markup, -1) {
		if strings.Contains(open, "src=") || strings.Contains(strings.ToLower(open), "nonce") {
			continue
		}
		n++
	}
	return n
}

func countInlineHandlers(markup string) int {
	return len(inlineHandlerAttr.FindAllStringIndex(markup, -1))
}

func swRegistersWithoutScope(markup string) bool {
	for _, m := range swRegisterCall.FindAllStringSubmatch(markup, -1) {
		if !strings.Contains(m[1], "scope") {
			return true
		}
	}
	return false
}
