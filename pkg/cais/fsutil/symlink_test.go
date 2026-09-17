package fsutil

import (
	"os"
	"path/filepath"
	"testing"
)

// #134: writes followed symlinks, so a planted link (internal/, store.go,
// web/static/js/amarra.js) made the CLI overwrite files outside the app.
func TestRefuseSymlinkWrite(t *testing.T) {
	dir := t.TempDir()
	real := filepath.Join(dir, "real.txt")
	if err := os.WriteFile(real, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := RefuseSymlinkWrite(real); err != nil {
		t.Errorf("real file rejected: %v", err)
	}
	if err := RefuseSymlinkWrite(filepath.Join(dir, "missing", "new.txt")); err != nil {
		t.Errorf("missing path under a real dir rejected: %v", err)
	}

	link := filepath.Join(dir, "link.txt")
	if err := os.Symlink(real, link); err != nil {
		t.Fatal(err)
	}
	if err := RefuseSymlinkWrite(link); err == nil {
		t.Error("symlink target accepted")
	}

	linkDir := filepath.Join(dir, "linked-dir")
	if err := os.Symlink(dir, linkDir); err != nil {
		t.Fatal(err)
	}
	if err := RefuseSymlinkWrite(filepath.Join(linkDir, "sub", "file.txt")); err == nil {
		t.Error("symlinked ancestor accepted")
	}
}
