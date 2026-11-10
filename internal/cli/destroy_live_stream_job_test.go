package cli

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func runAppCLI(t *testing.T, appDir string, args ...string) {
	t.Helper()
	t.Chdir(appDir)
	if err := (&CLI{Out: io.Discard}).Run(args); err != nil {
		t.Fatalf("amarra-cais %s: %v", strings.Join(args, " "), err)
	}
}

func testGeneratedApp(t *testing.T, appDir string) {
	t.Helper()
	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = appDir
	if out, err := tidy.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy: %v\n%s", err, out)
	}
	cmd := exec.Command("go", "test", "./...")
	cmd.Dir = appDir
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go test ./... after destroy: %v\n%s", err, out)
	}
}

// #404: destroy live undoes g live, including the // cais:live-views patch.
func TestDestroyLive_unregistersView(t *testing.T) {
	appDir := scaffoldCompileProbe(t, "destlive")
	runAppCLI(t, appDir, "g", "live", "counter")
	runAppCLI(t, appDir, "destroy", "live", "counter")

	routes := readDestroyFixture(t, filepath.Join(appDir, "internal/app/routes.go"))
	if strings.Contains(routes, `hub.Register("counter"`) {
		t.Errorf("live register left behind:\n%s", routes)
	}
	if !strings.Contains(routes, liveViewsMarker) {
		t.Errorf("cais:live-views marker missing:\n%s", routes)
	}
	if strings.Contains(routes, `/live/counter`) {
		t.Errorf("live route left behind:\n%s", routes)
	}
	for _, rel := range []string{
		"internal/handlers/counter_live.go",
		"internal/handlers/counter_live_test.go",
		"web/templates/pages/counter.html",
	} {
		if _, err := os.Stat(filepath.Join(appDir, rel)); !os.IsNotExist(err) {
			t.Errorf("expected %s removed", rel)
		}
	}
	testGeneratedApp(t, appDir)
}

func TestDestroyStreamChat_leavesCompilableApp(t *testing.T) {
	appDir := scaffoldCompileProbe(t, "deststream")
	runAppCLI(t, appDir, "g", "stream", "chat")
	runAppCLI(t, appDir, "destroy", "stream", "chat")

	routes := readDestroyFixture(t, filepath.Join(appDir, "internal/app/routes.go"))
	if strings.Contains(routes, "NewChatHandler") || strings.Contains(routes, `"/chat"`) {
		t.Errorf("chat routes left behind:\n%s", routes)
	}
	store := readDestroyFixture(t, filepath.Join(appDir, "internal/store/store.go"))
	if strings.Contains(store, "ListConversations") {
		t.Errorf("store still has ListConversations:\n%s", store)
	}
	layout := readDestroyFixture(t, filepath.Join(appDir, "web/templates/layouts/app.html"))
	if strings.Contains(layout, `href="/chat"`) {
		t.Errorf("nav still has /chat:\n%s", layout)
	}
	buildGeneratedApp(t, appDir)
}

func TestDestroyJob_keepsWorker(t *testing.T) {
	appDir := scaffoldCompileProbe(t, "destjob")
	runAppCLI(t, appDir, "g", "job", "send_welcome")
	runAppCLI(t, appDir, "destroy", "job", "send_welcome")

	if _, err := os.Stat(filepath.Join(appDir, "internal/jobs/send_welcome.go")); !os.IsNotExist(err) {
		t.Error("expected send_welcome.go removed")
	}
	reg := readDestroyFixture(t, filepath.Join(appDir, "internal/jobs/registry.go"))
	if strings.Contains(reg, "PerformSendWelcome") {
		t.Errorf("registry still registers SendWelcome:\n%s", reg)
	}
	if !strings.Contains(reg, "KindPruneSessions") {
		t.Errorf("registry lost framework jobs:\n%s", reg)
	}
	if _, err := os.Stat(filepath.Join(appDir, "cmd/worker/main.go")); err != nil {
		t.Errorf("worker should stay while other jobs remain: %v", err)
	}
	buildGeneratedApp(t, appDir)
}
