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

// Socket is the per-connection bag a View sees during Mount/Handle.
type Socket interface {
	Topic() string
	Param(key string) string
	Assign(key string, val any)
	Get(key string) any
}

// View is one Live session. Hub calls Mount once, then Handle for each event,
// then Render after join and after every successful Handle.
type View interface {
	Mount(ctx context.Context, sock Socket) error
	Handle(ctx context.Context, ev Event) error
	Render() Rendered
}

type memSocket struct {
	mu     sync.RWMutex
	topic  string
	params map[string]string
	assign map[string]any
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
