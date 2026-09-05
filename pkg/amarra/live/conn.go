package live

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"

	"github.com/puppe1990/amarra-cais/pkg/cais/csrf"
)

type conn struct {
	hub    *Hub
	ws     *websocket.Conn
	view   View
	sock   *memSocket
	topic  string
	inbox  chan Event
	writeM sync.Mutex
	mu     sync.Mutex
	active time.Time
}

func (c *conn) push(ev Event) {
	select {
	case c.inbox <- ev:
	default:
	}
}

func (c *conn) touch() {
	c.mu.Lock()
	c.active = time.Now()
	c.mu.Unlock()
}

func (c *conn) lastActive() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.active.IsZero() {
		return time.Now()
	}
	return c.active
}

func (h *Hub) serve(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Upgrade") != "websocket" {
		http.Error(w, "websocket required", http.StatusUpgradeRequired)
		return
	}
	viewName := r.URL.Query().Get("view")
	if viewName == "" {
		http.Error(w, "view query is required", http.StatusBadRequest)
		return
	}
	factory, ok := h.factory(viewName)
	if !ok {
		http.Error(w, "unknown live view", http.StatusNotFound)
		return
	}

	c := &conn{
		hub:   h,
		view:  factory(),
		topic: r.URL.Query().Get("topic"),
		inbox: make(chan Event, 32),
		sock: &memSocket{
			topic:  r.URL.Query().Get("topic"),
			params: map[string]string{},
		},
	}
	if c.topic == "" {
		c.topic = viewName
		c.sock.topic = viewName
	}
	for k, vals := range r.URL.Query() {
		if k == "view" || k == "topic" || len(vals) == 0 {
			continue
		}
		c.sock.params[k] = vals[0]
	}

	if !h.add(c) {
		http.Error(w, "live hub full", http.StatusServiceUnavailable)
		return
	}
	defer h.remove(c)

	opts := &websocket.AcceptOptions{}
	if len(h.cfg.OriginPatterns) > 0 {
		opts.OriginPatterns = h.cfg.OriginPatterns
	}
	ws, err := websocket.Accept(w, r, opts)
	if err != nil {
		return
	}
	c.ws = ws
	c.touch()
	defer func() { _ = ws.CloseNow() }()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	go c.readLoop(ctx, r, cancel)
	c.run(ctx)
}

func (c *conn) readLoop(ctx context.Context, r *http.Request, cancel context.CancelFunc) {
	defer cancel()
	joined := false
	for {
		_, data, err := c.ws.Read(ctx)
		if err != nil {
			return
		}
		var msg inMsg
		if err := json.Unmarshal(data, &msg); err != nil {
			c.write(ctx, outMsg{Type: typeError, Message: "invalid json"})
			continue
		}
		c.touch()
		switch msg.Type {
		case typeJoin:
			if !csrfMatch(r, msg.CSRF) {
				_ = c.ws.Close(websocket.StatusPolicyViolation, "csrf")
				return
			}
			if err := c.view.Mount(ctx, c.sock); err != nil {
				c.write(ctx, outMsg{Type: typeError, Message: err.Error()})
				_ = c.ws.Close(websocket.StatusInternalError, "mount")
				return
			}
			joined = true
			c.writeRendered(ctx, typeOK, c.view.Render(), "")
		case typeEvent:
			if !joined {
				c.write(ctx, outMsg{Type: typeError, Message: "join first"})
				continue
			}
			c.push(Event{Name: msg.Event, Payload: msg.Payload, Ref: msg.Ref})
		default:
			c.write(ctx, outMsg{Type: typeError, Message: "unknown type"})
		}
	}
}

func (c *conn) run(ctx context.Context) {
	ping := time.NewTicker(c.hub.cfg.PingInterval)
	defer ping.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ping.C:
			if time.Since(c.lastActive()) > c.hub.cfg.Idle {
				_ = c.ws.Close(websocket.StatusGoingAway, "idle")
				return
			}
			pctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err := c.ws.Ping(pctx)
			cancel()
			if err != nil {
				return
			}
		case ev := <-c.inbox:
			c.touch()
			c.dispatch(ctx, ev)
		}
	}
}

func (c *conn) dispatch(ctx context.Context, ev Event) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("amarra live panic: %v", rec)
			_ = c.ws.Close(websocket.StatusInternalError, "panic")
		}
	}()
	if err := c.view.Handle(ctx, ev); err != nil {
		c.write(ctx, outMsg{Type: typeError, Message: err.Error(), Ref: ev.Ref})
		return
	}
	c.writeRendered(ctx, typeMorph, c.view.Render(), ev.Ref)
	if ev.Ref != "" {
		c.write(ctx, outMsg{Type: typeAck, Ref: ev.Ref})
	}
}

func (c *conn) writeRendered(ctx context.Context, kind string, out Rendered, ref string) {
	patch, nav, pushes, ops := c.sock.drain()
	c.write(ctx, outMsg{
		Type:     kind,
		HTML:     out.HTML,
		Target:   out.Target,
		Ref:      ref,
		Patch:    patch,
		Navigate: nav,
		Ops:      ops,
		Pushes:   pushes,
	})
}

func (c *conn) write(ctx context.Context, msg outMsg) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	c.writeM.Lock()
	defer c.writeM.Unlock()
	wctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_ = c.ws.Write(wctx, websocket.MessageText, data)
}

func csrfMatch(r *http.Request, submitted string) bool {
	got := csrf.TokenFromRequest(r)
	if got == "" || submitted == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(got), []byte(submitted)) == 1
}
