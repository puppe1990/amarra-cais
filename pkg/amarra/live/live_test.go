package live

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/puppe1990/amarra-cais/pkg/cais/csrf"
)

type counterView struct {
	n int
}

func (v *counterView) Mount(context.Context, Socket) error { return nil }

func (v *counterView) Handle(_ context.Context, ev Event) error {
	switch ev.Name {
	case "inc", "nudge":
		v.n++
		return nil
	default:
		return fmt.Errorf("unknown event %s", ev.Name)
	}
}

func (v *counterView) Render() Rendered {
	return Rendered{Target: "count", HTML: fmt.Sprintf("<span>%d</span>", v.n)}
}

type panicView struct{ counterView }

func (v *panicView) Handle(context.Context, Event) error {
	panic("boom")
}

func testHub() *Hub {
	h := NewHub(Config{OriginPatterns: []string{"*"}})
	h.Register("counter", func() View { return &counterView{} })
	return h
}

func TestHandler_rejectsPlainGET(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/amarra/live", nil)
	Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusUpgradeRequired {
		t.Fatalf("code %d, want 426", rr.Code)
	}
}

func TestHandler_missingView(t *testing.T) {
	s := httptest.NewServer(testHub().Handler())
	t.Cleanup(s.Close)
	req, _ := http.NewRequest(http.MethodGet, s.URL+"/", nil)
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("code %d, want 400", resp.StatusCode)
	}
}

func TestHandler_unknownView(t *testing.T) {
	s := httptest.NewServer(testHub().Handler())
	t.Cleanup(s.Close)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, _, err := websocket.Dial(ctx, wsURL(s)+"?view=nope", nil)
	if err == nil {
		t.Fatal("expected dial error for unknown view")
	}
}

func TestLive_joinAndInc(t *testing.T) {
	s := httptest.NewServer(testHub().Handler())
	t.Cleanup(s.Close)
	c := dialLive(t, s, "counter", "tok")
	defer func() { _ = c.Close(websocket.StatusNormalClosure, "") }()

	writeJSON(t, c, inMsg{Type: typeJoin, CSRF: "tok"})
	ok := readJSON(t, c)
	if ok.Type != typeOK || ok.HTML != "<span>0</span>" || ok.Target != "count" {
		t.Fatalf("join = %+v", ok)
	}

	writeJSON(t, c, inMsg{Type: typeEvent, Event: "inc", Ref: "1"})
	morph := readJSON(t, c)
	if morph.Type != typeMorph || morph.HTML != "<span>1</span>" {
		t.Fatalf("morph = %+v", morph)
	}
	ack := readJSON(t, c)
	if ack.Type != typeAck || ack.Ref != "1" {
		t.Fatalf("ack = %+v", ack)
	}
}

func TestLive_csrfRejects(t *testing.T) {
	s := httptest.NewServer(testHub().Handler())
	t.Cleanup(s.Close)
	c := dialLive(t, s, "counter", "tok")
	defer func() { _ = c.Close(websocket.StatusNormalClosure, "") }()
	writeJSON(t, c, inMsg{Type: typeJoin, CSRF: "nope"})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, _, err := c.Read(ctx)
	if err == nil {
		t.Fatal("expected close after bad csrf")
	}
}

func TestLive_unknownEventStaysUp(t *testing.T) {
	s := httptest.NewServer(testHub().Handler())
	t.Cleanup(s.Close)
	c := dialLive(t, s, "counter", "tok")
	defer func() { _ = c.Close(websocket.StatusNormalClosure, "") }()
	writeJSON(t, c, inMsg{Type: typeJoin, CSRF: "tok"})
	_ = readJSON(t, c)
	writeJSON(t, c, inMsg{Type: typeEvent, Event: "nope", Ref: "9"})
	errMsg := readJSON(t, c)
	if errMsg.Type != typeError {
		t.Fatalf("want error, got %+v", errMsg)
	}
	writeJSON(t, c, inMsg{Type: typeEvent, Event: "inc", Ref: "10"})
	morph := readJSON(t, c)
	if morph.Type != typeMorph {
		t.Fatalf("connection died after unknown event: %+v", morph)
	}
}

func TestLive_maxConns(t *testing.T) {
	h := NewHub(Config{MaxConns: 1, OriginPatterns: []string{"*"}})
	h.Register("counter", func() View { return &counterView{} })
	s := httptest.NewServer(h.Handler())
	t.Cleanup(s.Close)
	c1 := dialLive(t, s, "counter", "tok")
	defer func() { _ = c1.Close(websocket.StatusNormalClosure, "") }()
	writeJSON(t, c1, inMsg{Type: typeJoin, CSRF: "tok"})
	_ = readJSON(t, c1)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, resp, err := websocket.Dial(ctx, wsURL(s)+"?view=counter", &websocket.DialOptions{
		HTTPHeader: cookieHeader("tok"),
	})
	if err == nil {
		t.Fatal("expected second dial to fail")
	}
	if resp != nil && resp.StatusCode != http.StatusServiceUnavailable {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status %d body %s", resp.StatusCode, body)
	}
}

func TestLive_broadcast(t *testing.T) {
	h := testHub()
	s := httptest.NewServer(h.Handler())
	t.Cleanup(s.Close)
	c1 := dialLive(t, s, "counter", "tok")
	c2 := dialLive(t, s, "counter", "tok")
	defer func() { _ = c1.Close(websocket.StatusNormalClosure, "") }()
	defer func() { _ = c2.Close(websocket.StatusNormalClosure, "") }()
	writeJSON(t, c1, inMsg{Type: typeJoin, CSRF: "tok"})
	writeJSON(t, c2, inMsg{Type: typeJoin, CSRF: "tok"})
	_ = readJSON(t, c1)
	_ = readJSON(t, c2)

	h.Broadcast("counter", Event{Name: "nudge"})
	m1 := readJSON(t, c1)
	m2 := readJSON(t, c2)
	if m1.Type != typeMorph || m2.Type != typeMorph {
		t.Fatalf("broadcast morphs: %+v %+v", m1, m2)
	}
}

func TestLive_panicCloses(t *testing.T) {
	h := NewHub(Config{OriginPatterns: []string{"*"}})
	h.Register("boom", func() View { return &panicView{} })
	s := httptest.NewServer(h.Handler())
	t.Cleanup(s.Close)
	c := dialLive(t, s, "boom", "tok")
	defer func() { _ = c.Close(websocket.StatusNormalClosure, "") }()
	writeJSON(t, c, inMsg{Type: typeJoin, CSRF: "tok"})
	_ = readJSON(t, c)
	writeJSON(t, c, inMsg{Type: typeEvent, Event: "inc", Ref: "1"})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, _, err := c.Read(ctx)
	if err == nil {
		t.Fatal("expected close after panic")
	}
}

type patchStreamView struct {
	sock Socket
}

func (v *patchStreamView) Mount(_ context.Context, sock Socket) error {
	v.sock = sock
	return nil
}

func (v *patchStreamView) Handle(_ context.Context, ev Event) error {
	if ev.Name != "inc" {
		return fmt.Errorf("unknown event %s", ev.Name)
	}
	v.sock.Patch("/counter?n=1")
	v.sock.Stream("append", "log", "<li>1</li>")
	v.sock.Push("tick", map[string]int{"n": 1})
	return nil
}

func (v *patchStreamView) Render() Rendered {
	return Rendered{Target: "count", HTML: "<span>1</span>"}
}

func TestLive_patchStreamPush(t *testing.T) {
	h := NewHub(Config{OriginPatterns: []string{"*"}})
	h.Register("fancy", func() View { return &patchStreamView{} })
	s := httptest.NewServer(h.Handler())
	t.Cleanup(s.Close)
	c := dialLive(t, s, "fancy", "tok")
	defer func() { _ = c.Close(websocket.StatusNormalClosure, "") }()
	writeJSON(t, c, inMsg{Type: typeJoin, CSRF: "tok"})
	_ = readJSON(t, c)
	writeJSON(t, c, inMsg{Type: typeEvent, Event: "inc", Ref: "1"})
	morph := readJSON(t, c)
	if morph.Type != typeMorph {
		t.Fatalf("morph = %+v", morph)
	}
	if morph.Patch != "/counter?n=1" {
		t.Fatalf("patch = %q", morph.Patch)
	}
	if len(morph.Ops) != 1 || morph.Ops[0].Kind != "append" || morph.Ops[0].Target != "log" {
		t.Fatalf("ops = %+v", morph.Ops)
	}
	if len(morph.Pushes) != 1 || morph.Pushes[0].Event != "tick" {
		t.Fatalf("pushes = %+v", morph.Pushes)
	}
}

func TestLive_handlerEmptyHubUnknownView(t *testing.T) {
	var hits atomic.Int32
	h := Handler()
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		h.ServeHTTP(w, r)
	}))
	t.Cleanup(s.Close)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, _, err := websocket.Dial(ctx, wsURL(s)+"?view=counter", nil)
	if err == nil {
		t.Fatal("empty hub should reject unknown view")
	}
	if hits.Load() == 0 {
		t.Fatal("handler was not hit")
	}
}

func wsURL(s *httptest.Server) string {
	return "ws" + strings.TrimPrefix(s.URL, "http")
}

func cookieHeader(tok string) http.Header {
	h := make(http.Header)
	h.Set("Cookie", csrf.CookieName+"="+tok)
	return h
}

func dialLive(t *testing.T, s *httptest.Server, view, tok string) *websocket.Conn {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	t.Cleanup(cancel)
	c, _, err := websocket.Dial(ctx, wsURL(s)+"?view="+view, &websocket.DialOptions{
		HTTPHeader: cookieHeader(tok),
	})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func writeJSON(t *testing.T, c *websocket.Conn, msg inMsg) {
	t.Helper()
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := c.Write(ctx, websocket.MessageText, data); err != nil {
		t.Fatal(err)
	}
}

func readJSON(t *testing.T, c *websocket.Conn) outMsg {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, data, err := c.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var msg outMsg
	if err := json.Unmarshal(data, &msg); err != nil {
		t.Fatalf("json %s: %v", data, err)
	}
	return msg
}

type mountOrderView struct {
	mu      sync.Mutex
	mounts  int
	handles int
}

func (v *mountOrderView) Mount(context.Context, Socket) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.mounts++
	return nil
}

func (v *mountOrderView) Handle(context.Context, Event) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.handles++
	return nil
}

func (v *mountOrderView) Render() Rendered {
	return Rendered{Target: "order", HTML: "<span>0</span>"}
}

func (v *mountOrderView) counts() (mounts, handles int) {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.mounts, v.handles
}

// #112: Broadcast can deliver an event before the join/Mount handshake
// completes. Handle must not run with zero state (e.g. inserting into
// conversation 0 in the chat scaffold).
func TestLive_ignoresEventsBeforeJoin(t *testing.T) {
	h := NewHub(Config{OriginPatterns: []string{"*"}})
	view := &mountOrderView{}
	h.Register("order", func() View { return view })
	s := httptest.NewServer(h.Handler())
	t.Cleanup(s.Close)

	c := dialLive(t, s, "order", "tok")
	defer func() { _ = c.Close(websocket.StatusNormalClosure, "") }()

	h.Broadcast("order", Event{Name: "inc"})
	time.Sleep(50 * time.Millisecond)

	writeJSON(t, c, inMsg{Type: typeJoin, CSRF: "tok"})
	ok := readJSON(t, c)
	if ok.Type != typeOK {
		t.Fatalf("join = %+v", ok)
	}
	mounts, handles := view.counts()
	if mounts != 1 {
		t.Fatalf("mounts = %d, want 1", mounts)
	}
	if handles != 0 {
		t.Fatalf("handled %d event(s) before join", handles)
	}

	// After the join, events flow normally.
	writeJSON(t, c, inMsg{Type: typeEvent, Event: "inc", Ref: "1"})
	_ = readJSON(t, c)
	_, handles = view.counts()
	if handles != 1 {
		t.Fatalf("handles after join = %d, want 1", handles)
	}
}

type failingView struct {
	msg string
}

func (v *failingView) Mount(context.Context, Socket) error { return nil }

func (v *failingView) Handle(context.Context, Event) error { return errors.New(v.msg) }

func (v *failingView) Render() Rendered {
	return Rendered{Target: "err", HTML: "<b></b>"}
}

func liveErrorMsg(t *testing.T, env, msg string) outMsg {
	t.Helper()
	h := NewHub(Config{Env: env, OriginPatterns: []string{"*"}})
	h.Register("failing", func() View { return &failingView{msg: msg} })
	s := httptest.NewServer(h.Handler())
	t.Cleanup(s.Close)
	c := dialLive(t, s, "failing", "tok")
	defer func() { _ = c.Close(websocket.StatusNormalClosure, "") }()
	writeJSON(t, c, inMsg{Type: typeJoin, CSRF: "tok"})
	_ = readJSON(t, c)
	writeJSON(t, c, inMsg{Type: typeEvent, Event: "boom", Ref: "1"})
	return readJSON(t, c)
}

// #98: Handle errors went verbatim to the websocket — SQL/path details leaked
// to any client. Outside development the client sees a generic message.
func TestLive_handleErrorHiddenOutsideDevelopment(t *testing.T) {
	errMsg := liveErrorMsg(t, "production", "sql: no such table: secret_widgets")
	if errMsg.Type != typeError {
		t.Fatalf("expected error message, got %+v", errMsg)
	}
	if strings.Contains(errMsg.Message, "secret_widgets") {
		t.Fatalf("internal error leaked to the client: %q", errMsg.Message)
	}
}

func TestLive_handleErrorShownInDevelopment(t *testing.T) {
	errMsg := liveErrorMsg(t, "development", "sql: no such table: secret_widgets")
	if !strings.Contains(errMsg.Message, "secret_widgets") {
		t.Fatalf("development should keep the detail, got %q", errMsg.Message)
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
