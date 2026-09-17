package pwa

import (
	"os"
	"path/filepath"
	"testing"
)

// #134: PWA asset writes followed symlinks out of the app tree.
func TestCopyAsset_refusesSymlinkedDestination(t *testing.T) {
	outside := filepath.Join(t.TempDir(), "victim.js")
	if err := os.WriteFile(outside, []byte("original\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	link := filepath.Join(dir, "amarra.js")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}

	if err := copyAsset("assets/amarra.js", link); err == nil {
		t.Fatal("copyAsset followed a symlink")
	}
	body, err := os.ReadFile(outside)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "original\n" {
		t.Errorf("outside file was overwritten: %q", body)
	}
}
