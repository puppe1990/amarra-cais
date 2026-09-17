// Legacy HTMX helpers. The public contract is Drive/Amarra (pkg/amarra):
// generated apps ship amarra.js and must not emit hx-* attributes (#99).
// Deprecated: kept only for leftover HTMX apps; slated for removal at v1.0.
// Use pkg/amarra (IsDrive/HeaderDrive), pkg/amarra/stream, and pkg/cais/httpx.

package cais

import (
	"fmt"
	"net/http"
	"strings"
)

// Deprecated: use amarra.IsDrive; HTMX is not part of the generated contract (#99).
func IsHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// Deprecated: legacy HTMX helper, remove at v1.0 (#99).
// SetTrigger sets HX-Trigger for client-side events after swap.
func SetTrigger(w http.ResponseWriter, id string) {
	w.Header().Set("HX-Trigger", id)
}

// Deprecated: legacy HTMX helper, remove at v1.0 (#99).
// SetToast sets HX-Trigger so cais.js shows a transient toast after the HTMX swap.
// Non-ASCII runes are escaped as \uXXXX so browsers read the header without mojibake.
func SetToast(w http.ResponseWriter, message string) {
	w.Header().Set("HX-Trigger", hxTriggerPayload("caisToast", message))
}

// Deprecated: legacy HTMX helper, remove at v1.0 (#99).
// SetRetarget sets HX-Retarget to change the swap target from the response.
func SetRetarget(w http.ResponseWriter, selector string) {
	w.Header().Set("HX-Retarget", selector)
}

// Deprecated: legacy HTMX helper, remove at v1.0 (#99).
// SetFocus sets HX-Trigger so cais.js focuses a field after swap (e.g. first invalid input).
func SetFocus(w http.ResponseWriter, selector string) {
	w.Header().Set("HX-Trigger", hxTriggerPayload("caisFocus", selector))
}

func hxTriggerPayload(key, value string) string {
	return fmt.Sprintf(`{%s:%s}`, jsonStringASCII(key), jsonStringASCII(value))
}

func jsonStringASCII(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 2)
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\b':
			b.WriteString(`\b`)
		case '\f':
			b.WriteString(`\f`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if r < 0x20 || r > 0x7e {
				fmt.Fprintf(&b, `\u%04x`, r)
			} else {
				b.WriteRune(r)
			}
		}
	}
	b.WriteByte('"')
	return b.String()
}
