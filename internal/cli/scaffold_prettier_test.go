package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// findRepoPrettier returns the prettier binary from the framework checkout
// (installed by npm ci). Empty when JS tooling is not installed.
func findRepoPrettier(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		candidate := filepath.Join(dir, "node_modules", ".bin", "prettier")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// TestScaffold_AppIsPrettierClean guards the prettier job shipped in the
// generated .github/workflows/ci.yml: a fresh app must pass npx prettier
// --check . before its first commit (#55).
func TestScaffold_AppIsPrettierClean(t *testing.T) {
	prettier := findRepoPrettier(t)
	if prettier == "" {
		t.Skip("prettier not installed — run npm ci")
	}

	cases := []struct {
		name    string
		minimal bool
		blank   bool
	}{
		{"full", false, false},
		{"minimal", true, false},
		{"blank", false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("CAIS_SKIP_TIDY", "1")
			appDir := filepath.Join(t.TempDir(), "probe")
			if err := scaffoldNewApp(appDir, scaffoldData{
				AppName:    "probe",
				ModulePath: "github.com/puppe1990/probe",
			}, tc.minimal, tc.blank); err != nil {
				t.Fatal(err)
			}

			cmd := exec.Command(prettier, "--check", ".")
			cmd.Dir = appDir
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Errorf("fresh scaffold fails its own prettier --check (#55):\n%s", out)
			}
		})
	}
}
