package cli

import (
	"os"
	"strings"
	"testing"
)

// #405: the committed IIFE is minified so phones pull ~19KB gzip, not the
// pretty 121KB source dump. The version banner stays on line 1 (outside minify)
// so doctor can still match go.mod. `amarra-cais pwa` without --bump is enough
// because the SW already network-firsts /static/js/ (#399).
func TestCommittedAmarraJS_isMinified(t *testing.T) {
	body, err := os.ReadFile("../../pkg/cais/pwa/assets/amarra.js")
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	first, _, _ := strings.Cut(text, "\n")
	if !strings.HasPrefix(first, "/* amarra-cais v") || !strings.HasSuffix(first, " */") {
		t.Errorf("version banner must be the first line, got %q", first)
	}
	if strings.Contains(text, "function shouldInterceptClick") {
		t.Error("source identifier shouldInterceptClick still in the bundle — js-build must minify")
	}
	script, err := os.ReadFile("../../scripts/js-build.mjs")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(script), "minify: true") {
		t.Error("scripts/js-build.mjs must set minify: true")
	}
}
