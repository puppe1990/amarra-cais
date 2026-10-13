package devreload

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/puppe1990/amarra-cais/pkg/cais"
)

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// #340: the edit that starts a dev reload is a template write; the watcher has to
// name that file so the caller knows whether to re-parse the renderer.
func TestWatcher_Poll_reportsChangedFile(t *testing.T) {
	dir := t.TempDir()
	page := filepath.Join(dir, "pages", "home.html")
	writeFile(t, page, "home v1")

	w, err := NewWatcher([]string{dir}, func(string) {})
	if err != nil {
		t.Fatal(err)
	}
	if changed, ok := w.Poll(); ok {
		t.Fatalf("first poll after New must be quiet, got %q", changed)
	}

	writeFile(t, page, "home v2")
	changed, ok := w.Poll()
	if !ok || changed != page {
		t.Fatalf("Poll = %q, %v; want %q, true", changed, ok, page)
	}
	if _, ok := w.Poll(); ok {
		t.Fatal("a settled file must not report twice")
	}
}

func TestWatcher_Poll_ignoresEditorBreadcrumbs(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "pages", "home.html"), "home")

	w, err := NewWatcher([]string{dir}, func(string) {})
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dir, ".DS_Store"), "junk")
	writeFile(t, filepath.Join(dir, "pages", "home.html~"), "backup")
	writeFile(t, filepath.Join(dir, ".git", "HEAD"), "ref: refs/heads/main")

	if changed, ok := w.Poll(); ok {
		t.Fatalf("editor breadcrumbs must not trigger a reload, got %q", changed)
	}
}

func TestWatcher_Run_reportsChangeUntilContextIsDone(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "home.html"), "v1")

	changes := make(chan string, 4)
	w, err := NewWatcher([]string{dir}, func(path string) { changes <- path })
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go w.Run(ctx, Config{Poll: 20 * time.Millisecond, Settle: 5 * time.Millisecond})

	writeFile(t, filepath.Join(dir, "home.html"), "v2")

	select {
	case got := <-changes:
		if !strings.HasSuffix(got, "home.html") {
			t.Fatalf("change = %q", got)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("watcher never reported the edit")
	}
	cancel()
}

// A missing target (no built stylesheet yet) must not kill the watcher.
func TestWatcher_New_toleratesMissingTargets(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "static", "css", "styles.css")
	w, err := NewWatcher([]string{dir, missing}, func(string) {})
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, missing, "body{}")
	if changed, ok := w.Poll(); !ok || changed != missing {
		t.Fatalf("Poll = %q, %v; want %q, true", changed, ok, missing)
	}
}

func TestHub_ServeHTTP_streamsReloadEvents(t *testing.T) {
	hub := New()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, ReloadPath, nil)
	ctx, cancel := context.WithCancel(req.Context())
	req = req.WithContext(ctx)

	done := make(chan struct{})
	go func() {
		hub.ServeHTTP(rr, req)
		close(done)
	}()

	// The subscription is registered before the first flush, so Notify lands.
	waitForSubscriber(t, hub)
	hub.Notify()
	cancel()
	<-done

	body := rr.Body.String()
	if !strings.Contains(body, ": connected") {
		t.Fatalf("stream preamble missing: %q", body)
	}
	if !strings.Contains(body, "event: boot\ndata: "+hub.bootID) {
		t.Fatalf("boot id missing: %q", body)
	}
	if !strings.Contains(body, "event: reload") {
		t.Fatalf("reload event missing: %q", body)
	}
	if got := rr.Header().Get("Content-Type"); got != "text/event-stream" {
		t.Fatalf("content-type = %q", got)
	}
}

func waitForSubscriber(t *testing.T, hub *Hub) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		hub.mu.Lock()
		n := len(hub.subs)
		hub.mu.Unlock()
		if n > 0 {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("no subscriber registered")
}

func TestRegister_devOnlyAndLocalhostOnly(t *testing.T) {
	dev := cais.NewRouter()
	New().Register(dev, "development")

	local := httptest.NewRequest(http.MethodGet, ReloadPath, nil)
	local.RemoteAddr = "127.0.0.1:5555"
	rr := httptest.NewRecorder()
	dev.ServeHTTP(rr, local.WithContext(cancelAfterFirstWrite(local.Context())))
	if rr.Code != http.StatusOK {
		t.Fatalf("loopback dev request = %d", rr.Code)
	}

	remote := httptest.NewRequest(http.MethodGet, ReloadPath, nil)
	remote.RemoteAddr = "203.0.113.7:5555"
	rr = httptest.NewRecorder()
	dev.ServeHTTP(rr, remote)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("non-loopback dev request = %d, want 403", rr.Code)
	}

	prod := cais.NewRouter()
	New().Register(prod, "production")
	rr = httptest.NewRecorder()
	prod.ServeHTTP(rr, local)
	if rr.Code == http.StatusOK {
		t.Fatal("production must not serve the dev reload route")
	}
}

func cancelAfterFirstWrite(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()
	return ctx
}
