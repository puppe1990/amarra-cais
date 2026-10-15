package live

import (
	"context"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
)

type joinPanicView struct {
	counterView
	stage string
}

func (v *joinPanicView) Mount(context.Context, Socket) error {
	if v.stage == "mount" {
		panic("private mount probe")
	}
	return nil
}

func (v *joinPanicView) Render() Rendered { panic("private render probe") }

// A subprocess makes an unhandled goroutine panic a test failure, not a crash
// of the entire suite (#347).
func TestLive_joinPanicSubprocess(t *testing.T) {
	stage := os.Getenv("AMARRA_JOIN_PANIC_STAGE")
	if stage != "" {
		exerciseJoinPanic(t, stage)
		return
	}
	for _, stage := range []string{"mount", "render"} {
		t.Run(stage, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestLive_joinPanicSubprocess$", "-test.v")
			cmd.Env = append(os.Environ(), "AMARRA_JOIN_PANIC_STAGE="+stage)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("join panic killed server: %v\n%s", err, out)
			}
			if !strings.Contains(string(out), "private "+stage+" probe") || !strings.Contains(string(out), "runtime/debug.Stack") {
				t.Fatalf("panic detail and stack missing from server log: %s", out)
			}
		})
	}
}

func exerciseJoinPanic(t *testing.T, stage string) {
	t.Helper()
	h := testHub()
	h.cfg.Env = "production"
	h.Register("broken", func() View { return &joinPanicView{stage: stage} })
	s := httptest.NewServer(h.Handler())
	defer s.Close()
	healthy := dialLive(t, s, "counter", "tok")
	defer func() { _ = healthy.CloseNow() }()
	writeJSON(t, healthy, inMsg{Type: typeJoin, CSRF: "tok"})
	_ = readJSON(t, healthy)
	broken := dialLive(t, s, "broken", "tok")
	defer func() { _ = broken.CloseNow() }()
	writeJSON(t, broken, inMsg{Type: typeJoin, CSRF: "tok"})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, payload, err := broken.Read(ctx)
	if websocket.CloseStatus(err) != websocket.StatusInternalError {
		t.Fatalf("panic close = %v, payload %s", err, payload)
	}
	if strings.Contains(err.Error(), "private") {
		t.Fatalf("panic leaked: %v", err)
	}
	writeJSON(t, healthy, inMsg{Type: typeEvent, Event: "inc"})
	if got := readJSON(t, healthy); got.HTML != "<span>1</span>" {
		t.Fatalf("healthy connection: %+v", got)
	}
	fresh := dialLive(t, s, "counter", "tok")
	defer func() { _ = fresh.CloseNow() }()
	writeJSON(t, fresh, inMsg{Type: typeJoin, CSRF: "tok"})
	if got := readJSON(t, fresh); got.Type != typeOK {
		t.Fatalf("fresh join: %+v", got)
	}
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		h.mu.Lock()
		_, brokenTopic := h.topics["broken"]
		h.mu.Unlock()
		if !brokenTopic {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("broken connection remains in hub")
}
