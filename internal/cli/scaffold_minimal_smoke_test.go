package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// The minimal and blank scaffolds must build: they ship migrations.go with
// //go:embed migrations/*.sql, which fails when the directory has no .sql (#74).
func TestScaffoldMinimalAndBlank_compile(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	caisDir := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	t.Setenv("CAIS_REPLACE", caisDir)

	for _, tc := range []struct {
		name           string
		minimal, blank bool
	}{
		{"minimal", true, false},
		{"blank", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			appDir := filepath.Join(t.TempDir(), tc.name)
			if err := scaffoldNewApp(appDir, scaffoldData{
				AppName:    tc.name,
				ModulePath: "github.com/puppe1990/" + tc.name + "app",
			}, tc.minimal, tc.blank); err != nil {
				t.Fatal(err)
			}

			tidy := exec.Command("go", "mod", "tidy")
			tidy.Dir = appDir
			if out, err := tidy.CombinedOutput(); err != nil {
				t.Fatalf("go mod tidy: %v\n%s", err, out)
			}

			build := exec.Command("go", "build", "./...")
			build.Dir = appDir
			build.Env = append(os.Environ(), "CGO_ENABLED=0")
			if out, err := build.CombinedOutput(); err != nil {
				t.Fatalf("go build ./... failed: %v\n%s", err, out)
			}
		})
	}
}
