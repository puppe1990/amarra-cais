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

// ContentType is the HTTP body Turbo-stream equivalent for Drive visits.
const ContentType = "text/vnd.amarra-stream"

func WriteHTTP(w http.ResponseWriter, ops ...Op) error {
	w.Header().Set("Content-Type", ContentType+"; charset=utf-8")
	for _, op := range ops {
		if err := WriteOp(w, op); err != nil {
			return err
		}
	}
	return nil
}

// opKinds mirrors STREAM_KINDS in pkg/amarra/js/stream.mjs. WriteOp rejects
// anything else so Op.Kind cannot inject arbitrary SSE event names (#103).
var opKinds = map[string]bool{
	"append": true, "prepend": true, "replace": true, "morph": true,
	"remove": true, "toast": true, "before": true, "after": true,
}

func WriteOp(w http.ResponseWriter, op Op) error {
	if !opKinds[op.Kind] {
		return fmt.Errorf(
			"stream: unknown op kind %q (use append, prepend, replace, morph, remove, toast, before, after)",
			op.Kind,
		)
	}
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
