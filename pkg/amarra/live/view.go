package live

import (
	"context"
	"encoding/json"
	"sync"
)

// Event is a client or broadcast message delivered to View.Handle.
type Event struct {
	Name    string
	Payload json.RawMessage
	Ref     string
}

// Rendered is HTML to morph into Target (CSS id, no #).
type Rendered struct {
	Target string
	HTML   string
}

// StreamOp is a Phoenix-style collection mutation sent with a morph.
type StreamOp struct {
	Kind   string `json:"kind"`
	Target string `json:"target,omitempty"`
	HTML   string `json:"html,omitempty"`
}

// Push is a server-to-hook event (Phoenix push_event).
type Push struct {
	Event   string          `json:"event"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// Socket is the per-connection bag a View sees during Mount/Handle.
type Socket interface {
	Topic() string
	Param(key string) string
	Assign(key string, val any)
	Get(key string) any
	Patch(url string)
	Navigate(url string)
	Push(event string, payload any)
	Stream(kind, target, html string)
}

// View is one Live session. Hub calls Mount once, then Handle for each event,
// then Render after join and after every successful Handle.
type View interface {
	Mount(ctx context.Context, sock Socket) error
	Handle(ctx context.Context, ev Event) error
	Render() Rendered
}

type memSocket struct {
	mu       sync.RWMutex
	topic    string
	params   map[string]string
	assign   map[string]any
	patch    string
	navigate string
	pushes   []Push
	ops      []StreamOp
}

func (s *memSocket) Topic() string { return s.topic }

func (s *memSocket) Param(key string) string {
	if s.params == nil {
		return ""
	}
	return s.params[key]
}

func (s *memSocket) Assign(key string, val any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.assign == nil {
		s.assign = map[string]any{}
	}
	s.assign[key] = val
}

func (s *memSocket) Get(key string) any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.assign == nil {
		return nil
	}
	return s.assign[key]
}

func (s *memSocket) Patch(url string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.patch = url
}

func (s *memSocket) Navigate(url string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.navigate = url
}

func (s *memSocket) Push(event string, payload any) {
	raw, err := json.Marshal(payload)
	if err != nil {
		raw = nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pushes = append(s.pushes, Push{Event: event, Payload: raw})
}

func (s *memSocket) Stream(kind, target, html string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ops = append(s.ops, StreamOp{Kind: kind, Target: target, HTML: html})
}

func (s *memSocket) drain() (patch, nav string, pushes []Push, ops []StreamOp) {
	s.mu.Lock()
	defer s.mu.Unlock()
	patch, nav = s.patch, s.navigate
	pushes, ops = s.pushes, s.ops
	s.patch, s.navigate = "", ""
	s.pushes, s.ops = nil, nil
	return
}
