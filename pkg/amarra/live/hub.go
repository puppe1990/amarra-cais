package live

import (
	"net/http"
	"sync"
	"time"
)

const (
	defaultMaxConns = 256
	defaultIdle     = 2 * time.Minute
	defaultPing     = 30 * time.Second
)

// Config limits a Hub. Zero values use defaults.
type Config struct {
	MaxConns       int
	Idle           time.Duration
	PingInterval   time.Duration
	OriginPatterns []string
}

func (c Config) withDefaults() Config {
	if c.MaxConns <= 0 {
		c.MaxConns = defaultMaxConns
	}
	if c.Idle <= 0 {
		c.Idle = defaultIdle
	}
	if c.PingInterval <= 0 {
		c.PingInterval = defaultPing
	}
	return c
}

// Hub holds Live views and in-process topic membership.
type Hub struct {
	cfg    Config
	mu     sync.Mutex
	views  map[string]func() View
	conns  map[*conn]struct{}
	topics map[string]map[*conn]struct{}
}

// NewHub returns an empty hub. Register views before serving.
func NewHub(cfg Config) *Hub {
	return &Hub{
		cfg:    cfg.withDefaults(),
		views:  map[string]func() View{},
		conns:  map[*conn]struct{}{},
		topics: map[string]map[*conn]struct{}{},
	}
}

// Register binds a view name (query ?view=) to a factory.
func (h *Hub) Register(name string, fn func() View) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.views[name] = fn
}

func (h *Hub) factory(name string) (func() View, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	fn, ok := h.views[name]
	return fn, ok
}

func (h *Hub) add(c *conn) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.conns) >= h.cfg.MaxConns {
		return false
	}
	h.conns[c] = struct{}{}
	if h.topics[c.topic] == nil {
		h.topics[c.topic] = map[*conn]struct{}{}
	}
	h.topics[c.topic][c] = struct{}{}
	return true
}

func (h *Hub) remove(c *conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.conns, c)
	if members := h.topics[c.topic]; members != nil {
		delete(members, c)
		if len(members) == 0 {
			delete(h.topics, c.topic)
		}
	}
}

// Broadcast delivers ev to every connection on topic (best-effort, non-blocking).
func (h *Hub) Broadcast(topic string, ev Event) {
	h.mu.Lock()
	members := make([]*conn, 0, len(h.topics[topic]))
	for c := range h.topics[topic] {
		members = append(members, c)
	}
	h.mu.Unlock()
	for _, c := range members {
		c.push(ev)
	}
}

// Handler upgrades GET /amarra/live.
func (h *Hub) Handler() http.Handler {
	return http.HandlerFunc(h.serve)
}

// Handler is a hub with no views (unknown ?view= is 404). Kept so existing
// scaffolds calling live.Handler() keep compiling.
func Handler() http.Handler {
	return NewHub(Config{}).Handler()
}
