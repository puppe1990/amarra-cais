package cli

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"
)

// #106: with only bool fields, parseForm emitted `item := models.Flag{}` with
// empty validation/after blocks, so `return item, errs` landed on the `}` line
// (admin_flags.go: expected ';', found 'return').
func TestScaffoldResource_boolOnlyParses(t *testing.T) {
	for _, fields := range []string{"active:bool", "active:bool,verified:bool"} {
		t.Run(fields, func(t *testing.T) {
			t.Setenv("CAIS_SKIP_TIDY", "1")
			appDir := filepath.Join(t.TempDir(), "boolonly")
			if err := scaffoldNewApp(appDir, scaffoldData{
				AppName:    "boolonly",
				ModulePath: "github.com/puppe1990/boolonly",
			}, true, false); err != nil {
				t.Fatal(err)
			}
			if err := scaffoldResource(appDir, "flag", resourceOpts{Fields: fields, Seed: false}); err != nil {
				t.Fatalf("scaffoldResource(%q): %v", fields, err)
			}

			admin, err := os.ReadFile(filepath.Join(appDir, "internal/handlers/admin_flags.go"))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := parser.ParseFile(token.NewFileSet(), "admin_flags.go", admin, parser.SkipObjectResolution); err != nil {
				t.Errorf("admin_flags.go does not parse: %v", err)
			}
		})
	}
}
