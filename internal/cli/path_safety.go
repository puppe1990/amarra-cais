package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// #101: generator/destroy names come from argv and are joined into file
// paths. Only kebab/snake lowercase names are valid inputs; anything with a
// path separator, "..", or a leading dot could escape the app dir.
var generatedNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)

func validateGeneratedName(kind, name string) error {
	if generatedNamePattern.MatchString(name) {
		return nil
	}
	return fmt.Errorf(
		"invalid name %q for %s: use lowercase letters, digits, underscores or hyphens, starting with a letter (e.g. %q)",
		name, kind, "bookmark",
	)
}

// ensureRelPath refuses rel paths that resolve outside the app dir before any
// write or removal, independent of which caller built them.
func ensureRelPath(rel string) error {
	if rel == "" {
		return fmt.Errorf("invalid generated path: empty")
	}
	if filepath.IsAbs(rel) {
		return fmt.Errorf("refusing absolute generated path %q", rel)
	}
	clean := filepath.Clean(rel)
	if clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("refusing generated path %q outside the app directory", rel)
	}
	return nil
}
