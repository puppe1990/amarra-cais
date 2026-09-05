package stream

import "net/http"

// Op is a named stream mutation. Kind is the SSE event name (append, prepend,
// replace, morph, remove, toast). Target is the DOM id for later clients;
// the wire payload is encodeOp (HTML for DOM ops, message text for toast).
type Op struct {
	Kind   string // append, prepend, replace, morph, remove, toast
	Target string
	HTML   string
}

func WriteOp(w http.ResponseWriter, op Op) error {
	return WriteEvent(w, op.Kind, encodeOp(op))
}

func encodeOp(op Op) string {
	return op.HTML
}
