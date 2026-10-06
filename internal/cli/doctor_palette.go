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
// #246: only class attributes count — stroke-linecap="round" is an SVG
// attribute — and typography/alignment utilities (text-center,
// bg-gradient-to-r, fontSize keys like text-headline-xl) are not colour keys.
func checkPalette(dir string) doctorCheck {
	const name = "tailwind palette"
	cfgPath := filepath.Join(dir, "tailwind.config.js")
	body, err := os.ReadFile(cfgPath)
	if err != nil {
		return doctorCheck{Name: name, OK: true, Detail: "skipped (no tailwind.config.js)"}
	}
	theme := parseTailwindThemeKeys(string(body))
	markup, err := readTemplateMarkup(dir)
	if err != nil {
		return doctorCheck{Name: name, OK: true, Detail: err.Error()}
	}
	missing := missingPaletteTokens(markup, theme)
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
	// Hyphenated keys (headline-xl, code-sm) must be quoted in a JS object, so
	// both forms count (#246).
	tailwindKey    = regexp.MustCompile(`(?:"([\w-]+)"|([A-Za-z][\w-]*))\s*:`)
	classAttrValue = regexp.MustCompile(`(?s)\bclass\s*=\s*"([^"]*)"|\bclass\s*=\s*'([^']*)'`)
	// The captured utility is the whole suffix (red-500, headline-xl, ink/50) so
	// typography keys can be matched before the colour token is derived.
	colorClassToken = regexp.MustCompile(`(?:^|[\s"'` + "`" + `])(?:text|bg|border|ring|from|to|via|outline|accent|fill|stroke|divide|decoration|caret)-(\[[^\]]+\]|[a-z][a-z0-9-]*)`)
)

// tailwindThemeKeys is the slice of tailwind.config.js the palette check reads:
// the colours, plus the scales whose keys look like colour tokens
// (fontSize.headline-xl → text-headline-xl).
type tailwindThemeKeys struct {
	colors     map[string]bool
	fontSize   map[string]bool
	fontFamily map[string]bool
	boxShadow  map[string]bool
}

func (k tailwindThemeKeys) skipsScale(key string) bool {
	return k.fontSize[key] || k.fontFamily[key] || k.boxShadow[key]
}

func parseTailwindThemeKeys(cfg string) tailwindThemeKeys {
	return tailwindThemeKeys{
		colors:     keysAfterSection(cfg, "colors:"),
		fontSize:   keysAfterSection(cfg, "fontSize:"),
		fontFamily: keysAfterSection(cfg, "fontFamily:"),
		boxShadow:  keysAfterSection(cfg, "boxShadow:"),
	}
}

func parseTailwindColorKeys(cfg string) map[string]bool {
	return keysAfterSection(cfg, "colors:")
}

func keysAfterSection(cfg, section string) map[string]bool {
	idx := strings.Index(cfg, section)
	if idx < 0 {
		return map[string]bool{}
	}
	rest := cfg[idx+len(section):]
	start := strings.Index(rest, "{")
	if start < 0 {
		return map[string]bool{}
	}
	block, ok := braceBlock(rest[start:])
	if !ok {
		return map[string]bool{}
	}
	out := map[string]bool{}
	for _, m := range tailwindKey.FindAllStringSubmatch(block, -1) {
		for _, key := range m[1:] {
			if key != "" {
				out[key] = true
				break
			}
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

// classAttrValues returns the body of every class attribute. Scanning the whole
// file made SVG attributes such as stroke-linecap="round" look like classes (#246).
func classAttrValues(markup string) []string {
	matches := classAttrValue.FindAllStringSubmatch(markup, -1)
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		for _, group := range m[1:] {
			if group != "" {
				out = append(out, group)
				break
			}
		}
	}
	return out
}

func missingPaletteTokens(markup string, theme tailwindThemeKeys) []string {
	seen := map[string]struct{}{}
	for _, classes := range classAttrValues(markup) {
		for _, m := range colorClassToken.FindAllStringSubmatch(classes, -1) {
			if len(m) < 2 || m[1] == "" {
				continue
			}
			utility := strings.SplitN(m[1], "/", 2)[0]
			if strings.HasPrefix(utility, "[") || theme.skipsScale(utility) {
				continue
			}
			// #318: composed tokens (on-surface, border-subtle) — match the longest
			// hyphen prefix present in the theme instead of cutting at the first '-'.
			if paletteTokenDefined(utility, theme) {
				continue
			}
			token := utility
			if i := strings.IndexByte(token, '-'); i >= 0 {
				token = token[:i]
			}
			if skipPaletteToken(token) {
				continue
			}
			seen[token] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for token := range seen {
		out = append(out, token)
	}
	sort.Strings(out)
	return out
}

func paletteTokenDefined(utility string, theme tailwindThemeKeys) bool {
	for candidate := utility; candidate != ""; {
		if theme.colors[candidate] || theme.skipsScale(candidate) || skipPaletteToken(candidate) {
			return true
		}
		i := strings.LastIndexByte(candidate, '-')
		if i < 0 {
			return false
		}
		candidate = candidate[:i]
	}
	return false
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
	// #246: alignment/decoration utilities share the text-*/bg-* prefixes.
	"center": true, "left": true, "right": true, "justify": true, "gradient": true,
}
