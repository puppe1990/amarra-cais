package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// #102: "not tracked in the manifest" was treated as "unmodified", so destroy
// deleted hand-written files (and every file from `cais new`, which never
// recorded the manifest) without warning.
func TestScaffoldNewApp_recordsManifest(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "manifestnew")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "manifestnew",
		ModulePath: "github.com/puppe1990/manifestnew",
	}, true, false); err != nil {
		t.Fatal(err)
	}

	entries := readManifest(t, appDir)
	for _, want := range []string{
		"go.mod",
		"internal/handlers/home.go",
		"web/templates/pages/home.html",
		"web/templates/layouts/app.html",
		"AGENTS.md",
	} {
		if _, ok := entries[want]; !ok {
			t.Errorf("manifest missing %q", want)
		}
	}
	if _, ok := entries[generatedManifestRel]; ok {
		t.Error("manifest should not track itself")
	}
}

func TestScaffoldHandler_recordsManifest(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "manifesthandler")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "manifesthandler",
		ModulePath: "github.com/puppe1990/manifesthandler",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldHandler(appDir, "settings", false); err != nil {
		t.Fatal(err)
	}

	entries := readManifest(t, appDir)
	for _, want := range []string{
		"internal/handlers/settings.go",
		"internal/handlers/settings_test.go",
		"web/templates/pages/settings.html",
	} {
		if _, ok := entries[want]; !ok {
			t.Errorf("manifest missing %q", want)
		}
	}
}

func TestDestroyHandler_skipsUntrackedFileWithoutForce(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "ghosthandler")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "ghosthandler",
		ModulePath: "github.com/puppe1990/ghosthandler",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	ghost := filepath.Join(appDir, "internal/handlers/ghost.go")
	if err := os.WriteFile(ghost, []byte("package handlers\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	out := captureScaffoldOut(t)
	if err := destroyHandler(appDir, "ghost", false, false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(ghost); err != nil {
		t.Error("hand-written file deleted without --force")
	}
	if !strings.Contains(out.String(), "not recorded as generated") {
		t.Errorf("skip warning should explain the untracked file, got:\n%s", out.String())
	}

	if err := destroyHandler(appDir, "ghost", false, true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(ghost); !os.IsNotExist(err) {
		t.Error("--force should remove the untracked target")
	}
}

func TestDestroyHandler_skipsModifiedTrackedFile(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "modhandler")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "modhandler",
		ModulePath: "github.com/puppe1990/modhandler",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldHandler(appDir, "settings", false); err != nil {
		t.Fatal(err)
	}
	handlerPath := filepath.Join(appDir, "internal/handlers/settings.go")
	f, err := os.OpenFile(handlerPath, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString("\n// hand-written tweak\n"); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	out := captureScaffoldOut(t)
	if err := destroyHandler(appDir, "settings", false, false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(handlerPath); err != nil {
		t.Error("modified handler deleted without --force")
	}
	if !strings.Contains(out.String(), "modified since generation") {
		t.Errorf("skip warning should point at --force, got:\n%s", out.String())
	}
	if _, err := os.Stat(filepath.Join(appDir, "internal/handlers/settings_test.go")); !os.IsNotExist(err) {
		t.Error("unmodified tracked test file should be removed")
	}

	if err := destroyHandler(appDir, "settings", false, true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(handlerPath); !os.IsNotExist(err) {
		t.Error("--force should remove the modified handler")
	}
}

func readManifest(t *testing.T, dir string) map[string]string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(dir, generatedManifestRel))
	if err != nil {
		t.Fatalf("manifest missing: %v", err)
	}
	var entries map[string]string
	if err := json.Unmarshal(raw, &entries); err != nil {
		t.Fatalf("manifest not valid JSON: %v", err)
	}
	return entries
}

func captureScaffoldOut(t *testing.T) *bytes.Buffer {
	t.Helper()
	out := &bytes.Buffer{}
	old := scaffoldOut
	setScaffoldOut(out)
	t.Cleanup(func() { scaffoldOut = old })
	return out
}
