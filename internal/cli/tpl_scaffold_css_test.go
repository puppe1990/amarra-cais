package cli

import (
	"strings"
	"testing"
)

func TestTplInputCSS_documentsBaseSelectSupports(t *testing.T) {
	for _, want := range []string{
		"@supports (appearance: base-select)",
		"appearance: base-select",
		":user-invalid",
	} {
		if !strings.Contains(tplInputCSS, want) {
			t.Errorf("input.css missing %q", want)
		}
	}
}
