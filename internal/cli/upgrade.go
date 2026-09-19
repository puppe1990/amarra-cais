package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func (c *CLI) cmdUpgrade(args []string) error {
	dir, err := c.appDir()
	if err != nil {
		return err
	}
	dryRun, target, err := parseUpgradeArgs(args)
	if err != nil {
		return err
	}

	body, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return err
	}
	content := string(body)
	if strings.Contains(content, "replace "+frameworkModule) {
		return fmt.Errorf("go.mod has a local replace for %s — run: amarra-cais link --unlink, then retry", frameworkModule)
	}

	fromRaw := extractCaisVersion(content)
	from := parseSemverCore(fromRaw)
	to := parseUpgradeTargetCore(target)
	if from.OK && to.OK && compareSemverCore(to, from) < 0 {
		return fmt.Errorf("target v%s is older than the current v%s", formatSemver(to), formatSemver(from))
	}

	_, _ = fmt.Fprintf(c.Out, "→ upgrade %s → %s\n", displayUpgradeVersion(fromRaw), target)
	printMigrationChecklist(c.Out, from, to, fromRaw)

	if dryRun {
		_, _ = fmt.Fprintln(c.Out, "  --dry-run: no files changed")
		return nil
	}

	_, _ = fmt.Fprintf(c.Out, "→ go get %s@%s\n", frameworkModule, target)
	if err := runCmd(dir, "go", "get", frameworkModule+"@"+target); err != nil {
		return fmt.Errorf("go get: %w", err)
	}
	if err := installAppDeps(c.Out, dir); err != nil {
		return err
	}
	if err := runDoctor(c.Out, dir, doctorOptions{}); err != nil {
		_, _ = fmt.Fprintf(c.Out, "⚠ doctor reported issues: %v — fix the checklist items above, then re-run doctor\n", err)
	}
	return nil
}

// parseUpgradeTargetCore returns semverCore{} for `latest` (unknown target).
func parseUpgradeTargetCore(target string) semverCore {
	if target == "latest" {
		return semverCore{}
	}
	return parseSemverCore(target)
}

func displayUpgradeVersion(raw string) string {
	if raw == "" || raw == "?" {
		return "unknown"
	}
	return "v" + strings.TrimPrefix(raw, "v")
}

// parseUpgradeArgs separates flags from the optional target version.
func parseUpgradeArgs(args []string) (dryRun bool, target string, err error) {
	target = "latest"
	hasTarget := false
	for _, arg := range args {
		switch {
		case arg == "--dry-run":
			dryRun = true
		case strings.HasPrefix(arg, "-"):
			return false, "", fmt.Errorf("unknown flag %q (usage: amarra-cais upgrade [version] [--dry-run])", arg)
		case hasTarget:
			return false, "", fmt.Errorf("usage: amarra-cais upgrade [version] [--dry-run]")
		default:
			target = arg
			hasTarget = true
		}
	}
	target, err = normalizeUpgradeTarget(target)
	return dryRun, target, err
}

// normalizeUpgradeTarget keeps `latest` or canonicalizes a version to vX.Y.Z.
func normalizeUpgradeTarget(arg string) (string, error) {
	arg = strings.TrimSpace(arg)
	if arg == "" || arg == "latest" {
		return "latest", nil
	}
	v := parseSemverCore(arg)
	if !v.OK {
		return "", fmt.Errorf("invalid version %q — use vX.Y.Z or omit for latest", arg)
	}
	return "v" + formatSemver(v), nil
}

// printMigrationChecklist writes the curated steps for the from→to range.
func printMigrationChecklist(w io.Writer, from, to semverCore, fromRaw string) {
	steps := migrationsBetween(from, to)
	if len(steps) == 0 {
		_, _ = fmt.Fprintln(w, "  no known breaking changes in this range — see CHANGELOG.md")
		return
	}
	if !from.OK {
		_, _ = fmt.Fprintf(w, "  could not detect the current version (%s); showing every step up to the target:\n", fromRaw)
	}
	_, _ = fmt.Fprintln(w, "  migration checklist:")
	for _, s := range steps {
		_, _ = fmt.Fprintf(w, "  - [%s] %s: %s\n", s.Version, s.Title, s.Action)
	}
}
