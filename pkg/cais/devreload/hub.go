package devreload

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"sync"
	"time"
)

// ReloadPath is the dev-only SSE endpoint the scaffold layout subscribes to.
const ReloadPath = "/_dev/reload"

// heartbeat keeps proxies and idle timeouts from closing a quiet stream.
const heartbeat = 25 * time.Second

// Hub fans "something changed on disk" out to every open tab. Subscribers are
// in-process only: `amarra-cais dev` runs a single replica.
type Hub struct {
	mu   sync.Mutex
	subs map[chan struct{}]struct{}
	// bootID identifies this process. A reconnect that carries a different id
	// means an air rebuild restarted the server (a .go edit), and the open tab
	// reloads once to pick the new binary up (#340).
	bootID string
}

func New() *Hub {
	return &Hub{subs: map[chan struct{}]struct{}{}, bootID: newBootID()}
}

func newBootID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "boot"
	}
	return hex.EncodeToString(buf)
}

// Notify wakes every subscriber. A slow client is skipped instead of blocking
// the writer — a missed ping only costs one manual reload.
func (h *Hub) Notify() {
	if h == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.subs {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

func (h *Hub) subscribe() chan struct{} {
	ch := make(chan struct{}, 1)
	h.mu.Lock()
	h.subs[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *Hub) unsubscribe(ch chan struct{}) {
	h.mu.Lock()
	delete(h.subs, ch)
	h.mu.Unlock()
}

// ServeHTTP streams `event: reload` on every Notify until the tab goes away.
func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "dev reload needs a flushing response writer", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, ": connected\n\n")
	_, _ = io.WriteString(w, "event: boot\ndata: "+h.bootID+"\n\n")
	flusher.Flush()

	ch := h.subscribe()
	defer h.unsubscribe(ch)
	ticker := time.NewTicker(heartbeat)
	defer ticker.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-ch:
			if _, err := io.WriteString(w, "event: reload\ndata: 1\n\n"); err != nil {
				return
			}
			flusher.Flush()
		case <-ticker.C:
			if _, err := io.WriteString(w, ": ping\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}
