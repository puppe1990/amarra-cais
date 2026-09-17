package cli

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// #101: the CLI accepted argv names like "../../../victim" and joined them
// into file paths, deleting or creating .go/.html/.sql outside the app.
func TestValidateGeneratedName(t *testing.T) {
	valid := []string{"bookmark", "post", "post_comment", "user-settings", "category2", "add_tags"}
	for _, name := range valid {
		if err := validateGeneratedName("handler", name); err != nil {
			t.Errorf("validateGeneratedName(%q) = %v, want nil", name, err)
		}
	}
	invalid := []string{
		"../../../victim",
		"../victim",
		"..",
		"",
		"a/b",
		"\\evil",
		"/etc/passwd",
		"Post",
		"2fast",
		".hidden",
		"post name",
		"post;rm",
	}
	for _, name := range invalid {
		err := validateGeneratedName("model", name)
		if err == nil {
			t.Errorf("validateGeneratedName(%q) = nil, want error", name)
			continue
		}
		if !strings.Contains(err.Error(), "invalid name") {
			t.Errorf("validateGeneratedName(%q) error should explain the name: %v", name, err)
		}
	}
}

func newCLITestApp(t *testing.T, name string) (root, appDir string) {
	t.Helper()
	t.Setenv("CAIS_SKIP_TIDY", "1")
	root = t.TempDir()
	appDir = filepath.Join(root, name)
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    name,
		ModulePath: "github.com/puppe1990/" + name,
	}, true, false); err != nil {
		t.Fatal(err)
	}
	return root, appDir
}

func TestCmdGenerate_rejectsNameEscapingAppDir(t *testing.T) {
	root, appDir := newCLITestApp(t, "genescape")
	t.Chdir(appDir)

	escaped := filepath.Join(root, "escaped.go")
	err := (&CLI{Out: io.Discard}).Run([]string{"g", "handler", "../../../escaped"})
	if err == nil {
		t.Fatal("g handler accepted a traversal name")
	}
	if !strings.Contains(err.Error(), "invalid name") {
		t.Errorf("error should name the invalid input: %v", err)
	}
	if _, statErr := os.Stat(escaped); !os.IsNotExist(statErr) {
		t.Error("generator wrote escaped.go outside the app dir")
	}
}

func TestCmdDestroy_rejectsNameEscapingAppDir(t *testing.T) {
	root, appDir := newCLITestApp(t, "destescape")
	t.Chdir(appDir)

	victim := filepath.Join(root, "victim.go")
	if err := os.WriteFile(victim, []byte("package victim\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := (&CLI{Out: io.Discard}).Run([]string{"destroy", "model", "../../../victim"})
	if err == nil {
		t.Fatal("destroy accepted a traversal name")
	}
	if !strings.Contains(err.Error(), "invalid name") {
		t.Errorf("error should name the invalid input: %v", err)
	}
	if _, statErr := os.Stat(victim); statErr != nil {
		t.Error("destroy removed a file outside the app dir")
	}
}

func TestWriteScaffoldFile_rejectsEscapingRel(t *testing.T) {
	dir := t.TempDir()
	outside := filepath.Join(dir, "..", "sneaky.go")

	err := writeScaffoldFile(outside, []byte("package sneaky\n"), 0o644, "../sneaky.go", false)
	if err == nil {
		t.Fatal("writeScaffoldFile accepted a rel path escaping the app dir")
	}
	if _, statErr := os.Stat(outside); !os.IsNotExist(statErr) {
		t.Error("writeScaffoldFile wrote outside the app dir")
	}
}

func TestRemoveGeneratedFiles_rejectsEscapingRel(t *testing.T) {
	dir := t.TempDir()
	victim := filepath.Join(dir, "..", "victim.go")
	if err := os.WriteFile(victim, []byte("package victim\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(victim) })

	if err := removeGeneratedFiles(dir, []string{"../victim.go"}, false, true); err == nil {
		t.Fatal("removeGeneratedFiles accepted a rel path escaping the app dir")
	}
	if _, statErr := os.Stat(victim); statErr != nil {
		t.Error("removeGeneratedFiles deleted outside the app dir")
	}
}
