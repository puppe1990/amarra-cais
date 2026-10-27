package live

import (
	"context"
	"fmt"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
)

type slowCounterView struct {
	mu sync.Mutex
	n  int
}

func (v *slowCounterView) Mount(context.Context, Socket) error { return nil }

func (v *slowCounterView) Handle(context.Context, Event) error {
	time.Sleep(2 * time.Millisecond)
	v.mu.Lock()
	v.n++
	v.mu.Unlock()
	return nil
}

func (v *slowCounterView) Render() Rendered {
	v.mu.Lock()
	defer v.mu.Unlock()
	return Rendered{Target: "count", HTML: fmt.Sprintf("<span>%d</span>", v.n)}
}

// #376: a full inbox drops events, so the client would keep a stale view.
// After draining the queue the connection sends one resync with the current
// Render, and Lost names how many events never ran.
func TestLive_overflowResyncsToCurrentRender(t *testing.T) {
	h := NewHub(Config{OriginPatterns: []string{"*"}})
	view := &slowCounterView{}
	h.Register("slow", func() View { return view })
	s := httptest.NewServer(h.Handler())
	t.Cleanup(s.Close)

	c := dialLive(t, s, "slow", "tok")
	defer func() { _ = c.Close(websocket.StatusNormalClosure, "") }()
	writeJSON(t, c, inMsg{Type: typeJoin, CSRF: "tok"})
	_ = readJSON(t, c)

	const events = 40
	for i := 0; i < events; i++ {
		h.Broadcast("slow", Event{Name: "inc"})
	}

	var resync outMsg
	for i := 0; i <= events; i++ {
		msg := readJSON(t, c)
		if msg.Type == typeResync {
			resync = msg
			break
		}
	}
	if resync.Type != typeResync {
		t.Fatal("no resync after overflow")
	}
	if resync.Lost <= 0 {
		t.Fatalf("resync.Lost = %d, want the dropped count", resync.Lost)
	}
	want := fmt.Sprintf("<span>%d</span>", events-int(resync.Lost))
	if resync.HTML != want {
		t.Errorf("resync HTML = %q, want %q (the current Render)", resync.HTML, want)
	}
	if got := h.Dropped(); got != resync.Lost {
		t.Errorf("Hub.Dropped() = %d, resync.Lost = %d", got, resync.Lost)
	}
}

// #96: push() dropped events with select/default and no counter or log — a
// slow client lost morphs silently. The hub now counts drops and logs them
// (first drop, then every 100).
func TestHub_countsDroppedEvents(t *testing.T) {
	h := NewHub(Config{OriginPatterns: []string{"*"}})
	c := &conn{hub: h, topic: "counter", inbox: make(chan Event, 1)}

	c.push(Event{Name: "kept"})
	c.push(Event{Name: "dropped-1"})
	c.push(Event{Name: "dropped-2"})

	if got := h.Dropped(); got != 2 {
		t.Fatalf("Dropped() = %d, want 2", got)
	}
	select {
	case ev := <-c.inbox:
		if ev.Name != "kept" {
			t.Fatalf("inbox head = %q, want the oldest event", ev.Name)
		}
	default:
		t.Fatal("inbox should still hold the first event")
	}
}
