// Package fsutil guards framework/CLI writes against planted symlinks.
package fsutil

import (
	"fmt"
	"os"
	"path/filepath"
)

// RefuseSymlinkWrite errors when writing path would follow a symlink: the
// target itself or the deepest existing ancestor (e.g. an `internal/` or
// `web/static/js/amarra.js` link planted in a shared repo), which would make
// the CLI overwrite files outside the app tree (#134).
//
// Non-existent components are walked up to the first existing one and the walk
// stops there, so real directories above the app root are not inspected.
func RefuseSymlinkWrite(path string) error {
	p := path
	for {
		info, err := os.Lstat(p)
		if err == nil {
			if info.Mode()&os.ModeSymlink != 0 {
				return fmt.Errorf("refusing to write through symlink %s", p)
			}
			return nil
		}
		if !os.IsNotExist(err) {
			return fmt.Errorf("lstat %s: %w", p, err)
		}
		parent := filepath.Dir(p)
		if parent == p {
			return nil
		}
		p = parent
	}
}
