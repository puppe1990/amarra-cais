package stream

import (
	"fmt"
	"net/http"
	"strings"
)

// Op is a named stream mutation. Kind is the SSE event name (append, prepend,
// replace, morph, remove, toast). Target is the DOM id, written as SSE id:
// so EventSource.lastEventId can address the node. Last-Event-ID on reconnect
// is therefore the last DOM id, not a resume cursor.
type Op struct {
	Kind   string // append, prepend, replace, morph, remove, toast
	Target string
	HTML   string
}

func WriteOp(w http.ResponseWriter, op Op) error {
	if id := sseFieldID(op.Target); id != "" {
		if _, err := fmt.Fprintf(w, "id: %s\n", id); err != nil {
			return err
		}
	}
	return WriteEvent(w, op.Kind, encodeOp(op))
}

func sseFieldID(target string) string {
	if target == "" || strings.ContainsAny(target, "\r\n") {
		return ""
	}
	return target
}

func encodeOp(op Op) string {
	return op.HTML
}
