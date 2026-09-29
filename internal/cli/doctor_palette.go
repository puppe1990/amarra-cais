package cli

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// #207: Tailwind drops unknown colour classes with no error. Kit markup still
// uses the scaffold palette, so a re-theme that drops ink/foam/copper/tide
// leaves components unstyled unless doctor warns.
func checkPalette(dir string) doctorCheck {
	const name = "tailwind palette"
	cfgPath := filepath.Join(dir, "tailwind.config.js")
	body, err := os.ReadFile(cfgPath)
	if err != nil {
		return doctorCheck{Name: name, OK: true, Detail: "skipped (no tailwind.config.js)"}
	}
	defined := parseTailwindColorKeys(string(body))
	markup, err := readTemplateMarkup(dir)
	if err != nil {
		return doctorCheck{Name: name, OK: true, Detail: err.Error()}
	}
	missing := missingPaletteTokens(markup, defined)
	if len(missing) == 0 {
		return doctorCheck{Name: name, OK: true}
	}
	return doctorCheck{
		Name:     name,
		Optional: true,
		Detail:   fmt.Sprintf("templates use colour tokens missing from tailwind.config.js: %s — kit components will have no CSS", strings.Join(missing, ", ")),
		FixHint:  "add the tokens to theme.extend.colors, or restyle the kit classes in web/templates/",
	}
}

var (
	tailwindColorKey = regexp.MustCompile(`([A-Za-z][\w-]*)\s*:`)
	colorClassToken  = regexp.MustCompile(`(?:^|[\s"'` + "`" + `])(?:text|bg|border|ring|from|to|via|outline|accent|fill|stroke|divide|decoration|caret)-(?:\[.+?\]|([a-z]+)(?:-\d+)?(?:/\d+)?)`)
)

func parseTailwindColorKeys(cfg string) map[string]bool {
	idx := strings.Index(cfg, "colors:")
	if idx < 0 {
		return map[string]bool{}
	}
	rest := cfg[idx+len("colors:"):]
	start := strings.Index(rest, "{")
	if start < 0 {
		return map[string]bool{}
	}
	block, ok := braceBlock(rest[start:])
	if !ok {
		return map[string]bool{}
	}
	out := map[string]bool{}
	for _, m := range tailwindColorKey.FindAllStringSubmatch(block, -1) {
		if len(m) > 1 {
			out[m[1]] = true
		}
	}
	return out
}

func braceBlock(s string) (string, bool) {
	if s == "" || s[0] != '{' {
		return "", false
	}
	depth := 0
	for i, r := range s {
		switch r {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return s[:i+1], true
			}
		}
	}
	return "", false
}

func missingPaletteTokens(markup string, defined map[string]bool) []string {
	seen := map[string]struct{}{}
	for _, m := range colorClassToken.FindAllStringSubmatch(markup, -1) {
		if len(m) < 2 || m[1] == "" {
			continue
		}
		token := m[1]
		if skipPaletteToken(token) || defined[token] {
			continue
		}
		seen[token] = struct{}{}
	}
	out := make([]string, 0, len(seen))
	for token := range seen {
		out = append(out, token)
	}
	sort.Strings(out)
	return out
}

func skipPaletteToken(token string) bool {
	if knownTailwindColor[token] || tailwindNonColorToken[token] {
		return true
	}
	if len(token) <= 1 {
		return true
	}
	return false
}

func readTemplateMarkup(dir string) (string, error) {
	root := filepath.Join(dir, "web", "templates")
	var b strings.Builder
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if !strings.HasSuffix(path, ".html") {
			return nil
		}
		body, rerr := os.ReadFile(path)
		if rerr != nil {
			return rerr
		}
		b.Write(body)
		b.WriteByte('\n')
		return nil
	})
	return b.String(), err
}

var knownTailwindColor = map[string]bool{
	"black": true, "white": true, "transparent": true, "current": true, "inherit": true,
	"slate": true, "gray": true, "zinc": true, "neutral": true, "stone": true,
	"red": true, "orange": true, "amber": true, "yellow": true, "lime": true,
	"green": true, "emerald": true, "teal": true, "cyan": true, "sky": true,
	"blue": true, "indigo": true, "violet": true, "purple": true, "fuchsia": true,
	"pink": true, "rose": true,
}

var tailwindNonColorToken = map[string]bool{
	"xs": true, "sm": true, "md": true, "lg": true, "xl": true, "base": true,
	"none": true, "solid": true, "dashed": true, "dotted": true, "double": true,
	"hidden": true, "collapse": true, "separate": true, "start": true, "end": true,
}
