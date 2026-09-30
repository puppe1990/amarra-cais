package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeGit puts stub git/pre-commit binaries first on PATH and returns the call
// log. FAKE_GIT_INSIDE=1 makes `git rev-parse --is-inside-work-tree` succeed,
// which is how an app created inside an existing repo looks.
func fakeGit(t *testing.T, withPreCommit bool) string {
	t.Helper()
	binDir := t.TempDir()
	logPath := filepath.Join(binDir, "calls.log")
	scripts := map[string]string{
		"git": "#!/bin/sh\n" +
			"echo \"git $*\" >> \"$FAKE_CALLS\"\n" +
			"if [ \"$1\" = \"rev-parse\" ]; then\n" +
			"  if [ \"$FAKE_GIT_INSIDE\" = \"1\" ]; then exit 0; fi\n" +
			"  exit 1\n" +
			"fi\n" +
			"exit 0\n",
	}
	if withPreCommit {
		scripts["pre-commit"] = "#!/bin/sh\necho \"pre-commit $*\" >> \"$FAKE_CALLS\"\nexit 0\n"
	}
	for name, body := range scripts {
		if err := os.WriteFile(filepath.Join(binDir, name), []byte(body), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("FAKE_CALLS", logPath)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return logPath
}

func readCalls(t *testing.T, logPath string) string {
	t.Helper()
	calls, err := os.ReadFile(logPath)
	if err != nil {
		return ""
	}
	return string(calls)
}

func runNewProbe(t *testing.T, extra ...string) string {
	t.Helper()
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := filepath.Join(t.TempDir(), "probe")
	args := append([]string{"new", dir, "--module", "github.com/example/probe"}, extra...)
	var out bytes.Buffer
	if err := (&CLI{Out: &out}).Run(args); err != nil {
		t.Fatal(err)
	}
	return out.String()
}

// A fresh app must be a git work tree, otherwise the shipped CI never runs and
// the pre-commit hooks cannot be installed.
func TestCLI_New_initsGitAndInstallsPreCommit(t *testing.T) {
	logPath := fakeGit(t, true)
	out := runNewProbe(t)

	calls := readCalls(t, logPath)
	if !strings.Contains(calls, "git init") {
		t.Errorf("expected git init, calls:\n%s", calls)
	}
	if !strings.Contains(calls, "pre-commit install") {
		t.Errorf("expected pre-commit install, calls:\n%s", calls)
	}
	if !strings.Contains(out, "git init") {
		t.Errorf("output should report the repo bootstrap:\n%s", out)
	}
}

// Inside an existing work tree the app must not get a nested repo, and the
// outer repo's hooks must stay untouched.
func TestCLI_New_skipsGitInsideWorkTree(t *testing.T) {
	logPath := fakeGit(t, true)
	t.Setenv("FAKE_GIT_INSIDE", "1")
	out := runNewProbe(t)

	calls := readCalls(t, logPath)
	if strings.Contains(calls, "git init") {
		t.Errorf("must not create a nested repo, calls:\n%s", calls)
	}
	if strings.Contains(calls, "pre-commit install") {
		t.Errorf("must not touch the outer repo hooks, calls:\n%s", calls)
	}
	if !strings.Contains(out, "existing git work tree") {
		t.Errorf("output should explain the skip:\n%s", out)
	}
}

func TestCLI_New_noGitFlagSkipsRepo(t *testing.T) {
	logPath := fakeGit(t, true)
	out := runNewProbe(t, "--no-git")

	if calls := readCalls(t, logPath); strings.Contains(calls, "git ") {
		t.Errorf("--no-git must not touch git, calls:\n%s", calls)
	}
	if strings.Contains(out, "git init") {
		t.Errorf("--no-git should stay quiet about git:\n%s", out)
	}
}

func TestCLI_New_pointsAtPreCommitInstallWhenToolMissing(t *testing.T) {
	logPath := fakeGit(t, false)
	out := runNewProbe(t)

	calls := readCalls(t, logPath)
	if !strings.Contains(calls, "git init") {
		t.Errorf("git init must still run, calls:\n%s", calls)
	}
	if !strings.Contains(out, "make pre-commit-install") {
		t.Errorf("output should point at make pre-commit-install:\n%s", out)
	}
}
