package cli

import (
	"fmt"
	"io"
	"os/exec"
)

// initAppRepo makes a fresh app a git work tree so the shipped GitHub CI and
// pre-commit hooks can actually run. An app created inside an existing work
// tree keeps the outer repo untouched (no nested repo, no hooks replaced), and
// a missing pre-commit binary only downgrades the hint — the config still
// ships with the app.
func initAppRepo(w io.Writer, dir string, skip bool) error {
	if skip {
		return nil
	}
	if insideWorkTree(dir) {
		_, _ = fmt.Fprintln(w, "→ inside an existing git work tree — skipping git init")
		return nil
	}

	_, _ = fmt.Fprintln(w, "→ git init")
	if err := runCmd(dir, "git", "init", "-q"); err != nil {
		return fmt.Errorf("git init: %w", err)
	}

	if _, err := exec.LookPath("pre-commit"); err != nil {
		_, _ = fmt.Fprintln(w, "  pre-commit not installed — run: make pre-commit-install to activate the shipped hooks")
		return nil
	}
	_, _ = fmt.Fprintln(w, "→ pre-commit install")
	if err := runCmd(dir, "pre-commit", "install"); err != nil {
		_, _ = fmt.Fprintf(w, "  pre-commit install failed (%v) — run: make pre-commit-install\n", err)
	}
	return nil
}

// insideWorkTree reports whether dir already belongs to a git repository.
func insideWorkTree(dir string) bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = dir
	return cmd.Run() == nil
}
