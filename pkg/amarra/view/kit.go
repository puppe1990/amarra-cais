package view

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"strings"
)

//go:embed components/*.html
var shippedKit embed.FS

// ShippedComponents maps filename stems to framework-owned kit HTML.
// App files in components/*.html override the same stem.
func ShippedComponents() map[string]string {
	out, err := readShippedComponents()
	if err != nil {
		return map[string]string{}
	}
	return out
}

func readShippedComponents() (map[string]string, error) {
	paths, err := fs.Glob(shippedKit, "components/*.html")
	if err != nil {
		return nil, fmt.Errorf("amarra shipped kit: %w", err)
	}
	out := make(map[string]string, len(paths))
	for _, p := range paths {
		raw, err := fs.ReadFile(shippedKit, p)
		if err != nil {
			return nil, fmt.Errorf("amarra shipped kit %s: %w", p, err)
		}
		stem := strings.TrimSuffix(path.Base(p), ".html")
		out[stem] = string(raw)
	}
	return out, nil
}
