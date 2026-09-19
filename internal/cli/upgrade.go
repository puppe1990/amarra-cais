package cli

import (
	"fmt"
	"io"
	"strings"
)

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
