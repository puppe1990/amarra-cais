package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// #248: destroy handler/auth removed files and some routes but left the
// statements those routes were the only user of (contactLimit, resetLimit), and
// store.go kept an import only the removed methods used — `go build ./...`
// failed in the generated app.
func scaffoldCompileProbe(t *testing.T, name string) string {
	t.Helper()
	t.Setenv("CAIS_SKIP_TIDY", "1")
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	caisDir := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	t.Setenv("CAIS_REPLACE", caisDir)

	appDir := filepath.Join(t.TempDir(), name)
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    name,
		ModulePath: "github.com/puppe1990/" + name,
	}, false, false); err != nil {
		t.Fatal(err)
	}
	return appDir
}

func buildGeneratedApp(t *testing.T, appDir string) {
	t.Helper()
	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = appDir
	if out, err := tidy.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy: %v\n%s", err, out)
	}
	build := exec.Command("go", "build", "./...")
	build.Dir = appDir
	build.Env = append(os.Environ(), "CGO_ENABLED=0")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build ./... after destroy failed: %v\n%s", err, out)
	}
}

func TestDestroyHandler_leavesCompilableApp(t *testing.T) {
	appDir := scaffoldCompileProbe(t, "destroyhandler")
	if err := destroyHandler(appDir, "contact", false, false); err != nil {
		t.Fatal(err)
	}
	routes := readDestroyFixture(t, filepath.Join(appDir, "internal/app/routes.go"))
	if strings.Contains(routes, "contactLimit") {
		t.Errorf("orphaned contactLimit survived:\n%s", routes)
	}
	if !strings.Contains(routes, "home.ServeHTTP") || !strings.Contains(routes, "dashboard.ServeHTTP") {
		t.Errorf("destroy handler dropped unrelated routes:\n%s", routes)
	}
	requireParses(t, "routes.go", routes)
	buildGeneratedApp(t, appDir)
}

func TestDestroyAuth_leavesCompilableApp(t *testing.T) {
	appDir := scaffoldCompileProbe(t, "destroyauth")
	if err := destroyAuth(appDir, false); err != nil {
		t.Fatal(err)
	}
	routes := readDestroyFixture(t, filepath.Join(appDir, "internal/app/routes.go"))
	for _, gone := range []string{
		`"/login"`, `"/signup"`, `"/logout"`, "/forgot-password", "/reset-password",
		"loginLimit", "resetLimit", "auth.",
	} {
		if strings.Contains(routes, gone) {
			t.Errorf("routes.go still has %s:\n%s", gone, routes)
		}
	}
	if !strings.Contains(routes, "/contact") || !strings.Contains(routes, "/dashboard") {
		t.Errorf("destroy auth dropped unrelated routes:\n%s", routes)
	}
	requireParses(t, "routes.go", routes)

	store := readDestroyFixture(t, filepath.Join(appDir, "internal/store/store.go"))
	if strings.Contains(store, `"strings"`) {
		t.Errorf("store.go keeps the import only CreateUser used:\n%s", store)
	}
	requireParses(t, "store.go", store)

	buildGeneratedApp(t, appDir)
}
