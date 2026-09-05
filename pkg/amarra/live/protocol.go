package live

import "encoding/json"

const (
	typeJoin  = "join"
	typeEvent = "event"
	typeOK    = "ok"
	typeMorph = "morph"
	typeAck   = "ack"
	typeError = "error"
)

type inMsg struct {
	Type    string          `json:"type"`
	CSRF    string          `json:"csrf"`
	Event   string          `json:"event"`
	Payload json.RawMessage `json:"payload"`
	Ref     string          `json:"ref"`
}

type outMsg struct {
	Type     string     `json:"type"`
	HTML     string     `json:"html,omitempty"`
	Target   string     `json:"target,omitempty"`
	Ref      string     `json:"ref,omitempty"`
	Message  string     `json:"message,omitempty"`
	Patch    string     `json:"patch,omitempty"`
	Navigate string     `json:"navigate,omitempty"`
	Ops      []StreamOp `json:"ops,omitempty"`
	Pushes   []Push     `json:"pushes,omitempty"`
}
