package cli

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateLive_writesViewAndRegisters(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := filepath.Join(t.TempDir(), "app")
	if err := scaffoldNewApp(dir, scaffoldData{AppName: "app", ModulePath: "example.com/app"}, true, false); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	if err := (&CLI{Out: io.Discard}).Run([]string{"g", "live", "counter"}); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{
		"internal/handlers/counter_live.go",
		"internal/handlers/counter_live_test.go",
		"web/templates/pages/counter.html",
	} {
		if _, err := os.Stat(filepath.Join(dir, path)); err != nil {
			t.Fatalf("missing %s: %v", path, err)
		}
	}
	src, err := os.ReadFile(filepath.Join(dir, "internal/handlers/counter_live.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(src), "live.View") && !strings.Contains(string(src), "func (v *CounterLive) Handle") {
		t.Fatal("live view missing Handle")
	}
	page, _ := os.ReadFile(filepath.Join(dir, "web/templates/pages/counter.html"))
	if !strings.Contains(string(page), `amarra-live="counter"`) || !strings.Contains(string(page), `amarra-click="inc"`) {
		t.Fatalf("page missing live bindings: %s", page)
	}
	routes, _ := os.ReadFile(filepath.Join(dir, "internal/app/routes.go"))
	body := string(routes)
	if !strings.Contains(body, `r.Get("/live/counter"`) {
		t.Fatal("routes missing GET /live/counter")
	}
	if !strings.Contains(body, `hub.Register("counter"`) {
		t.Fatal("routes missing hub.Register counter")
	}
}

func TestGenerateStreamChat_liveFlag(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := filepath.Join(t.TempDir(), "app")
	if err := scaffoldNewApp(dir, scaffoldData{AppName: "app", ModulePath: "example.com/app"}, true, false); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	if err := (&CLI{Out: io.Discard}).Run([]string{"g", "stream", "chat", "--live"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "internal/handlers/chat_live.go")); err != nil {
		t.Fatal(err)
	}
	page, _ := os.ReadFile(filepath.Join(dir, "web/templates/pages/chat.html"))
	if !strings.Contains(string(page), `amarra-live="chat"`) || !strings.Contains(string(page), `amarra-submit="send"`) {
		t.Fatalf("chat page missing live bindings: %s", page)
	}
	routes, _ := os.ReadFile(filepath.Join(dir, "internal/app/routes.go"))
	if !strings.Contains(string(routes), `hub.Register("chat"`) {
		t.Fatal("missing chat live register")
	}
	if !strings.Contains(string(routes), "NewChatLive(deps.Store)") {
		t.Fatal("chat live factory should take store")
	}
}
