// PWA asset installation (WriteStatic / Install*), split from brand/icon
// helpers so pwa.go stays under the line cap (#288).
package pwa

import (
	"bytes"
	"os"
	"path/filepath"
	"text/template"

	"github.com/puppe1990/amarra-cais/pkg/cais/fsutil"
)

// WriteStatic writes default PWA assets into web/static for an app.
// Includes amarra.js so the shared SW PRECACHE cache.addAll does not 404.
func WriteStatic(appDir string, cfg Config) error {
	if cfg.ThemeColor == "" {
		cfg.ThemeColor = ThemeColor
	}
	if cfg.StartURL == "" {
		cfg.StartURL = "/"
	}

	staticDir := filepath.Join(appDir, "web", "static")
	if err := os.MkdirAll(filepath.Join(staticDir, "icons"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(staticDir, "js"), 0o755); err != nil {
		return err
	}

	if err := writeManifest(filepath.Join(staticDir, "manifest.webmanifest"), cfg); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(staticDir, "img"), 0o755); err != nil {
		return err
	}

	for _, pair := range []struct{ src, dst string }{
		{"assets/amarra.js", "js/amarra.js"},
		{"assets/htmx.min.js", "js/htmx.min.js"},
		{"assets/idiomorph-ext.min.js", "js/idiomorph-ext.min.js"},
		{"assets/sse-ext.min.js", "js/sse-ext.min.js"},
		{"assets/cais-core.js", "js/cais-core.js"},
		{"assets/cais-chat.js", "js/cais-chat.js"},
		{"assets/cais-chat-logic.mjs", "js/cais-chat-logic.mjs"},
	} {
		if err := copyAsset(pair.src, filepath.Join(staticDir, pair.dst)); err != nil {
			return err
		}
	}
	if err := writeUserAsset("assets/offline.html", filepath.Join(staticDir, "offline.html"), cfg.Force); err != nil {
		return err
	}
	if err := writeUserAsset("assets/go-on-cais.jpg", filepath.Join(staticDir, "img", "go-on-cais.jpg"), cfg.Force); err != nil {
		return err
	}
	// SW via SyncServiceWorker so CACHE_VERSION is preserved on upgrades.
	if _, _, err := SyncServiceWorker(appDir); err != nil {
		return err
	}

	if err := writeOGImage(filepath.Join(staticDir, "og.png"), cfg.Force); err != nil {
		return err
	}
	if err := writeUserAsset("assets/favicon.svg", filepath.Join(staticDir, "favicon.svg"), cfg.Force); err != nil {
		return err
	}
	if err := writeAppIcons(filepath.Join(staticDir, "icons"), cfg.IconPath, cfg.Force); err != nil {
		return err
	}

	return nil
}

// InstallTo writes PWA assets using DefaultConfig(name), overwriting existing
// app-owned assets (this is the explicit "write defaults" entry point).
func InstallTo(appDir, name string) error {
	cfg := DefaultConfig(name)
	cfg.Force = true
	return WriteStatic(appDir, cfg)
}

// WriteStaticInertia writes PWA assets for Inertia+Svelte apps (no HTMX JS bundles).
// Copies amarra.js so the shared SW PRECACHE cache.addAll does not 404.
func WriteStaticInertia(appDir string, cfg Config) error {
	if cfg.ThemeColor == "" {
		cfg.ThemeColor = ThemeColor
	}
	if cfg.StartURL == "" {
		cfg.StartURL = "/"
	}

	staticDir := filepath.Join(appDir, "web", "static")
	if err := os.MkdirAll(filepath.Join(staticDir, "icons"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(staticDir, "js"), 0o755); err != nil {
		return err
	}
	if err := writeManifest(filepath.Join(staticDir, "manifest.webmanifest"), cfg); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(staticDir, "img"), 0o755); err != nil {
		return err
	}

	for _, pair := range []struct{ src, dst string }{
		{"assets/amarra.js", "js/amarra.js"},
	} {
		if err := copyAsset(pair.src, filepath.Join(staticDir, pair.dst)); err != nil {
			return err
		}
	}
	if err := writeUserAsset("assets/offline.html", filepath.Join(staticDir, "offline.html"), cfg.Force); err != nil {
		return err
	}
	if err := writeUserAsset("assets/go-on-cais.jpg", filepath.Join(staticDir, "img", "go-on-cais.jpg"), cfg.Force); err != nil {
		return err
	}
	if _, _, err := SyncServiceWorker(appDir); err != nil {
		return err
	}

	if err := writeOGImage(filepath.Join(staticDir, "og.png"), cfg.Force); err != nil {
		return err
	}
	if err := writeUserAsset("assets/favicon.svg", filepath.Join(staticDir, "favicon.svg"), cfg.Force); err != nil {
		return err
	}
	if err := writeAppIcons(filepath.Join(staticDir, "icons"), cfg.IconPath, cfg.Force); err != nil {
		return err
	}

	return nil
}

// InstallForInertia writes PWA assets for Inertia scaffolds (no HTMX).
func InstallForInertia(appDir, name string) error {
	return WriteStaticInertia(appDir, DefaultConfig(name))
}

// WriteStaticAmarra writes PWA assets for HTML-first Amarra apps.
// Ships amarra.js only — no HTMX, sse-ext, idiomorph-ext, or cais.js.
func WriteStaticAmarra(appDir string, cfg Config) error {
	if err := WriteStaticInertia(appDir, cfg); err != nil {
		return err
	}
	return copyAsset("assets/amarra.js", filepath.Join(appDir, "web", "static", "js", "amarra.js"))
}

// InstallForAmarra writes PWA assets for Amarra scaffolds (no HTMX).
func InstallForAmarra(appDir, name string) error {
	return WriteStaticAmarra(appDir, DefaultConfig(name))
}

func writeManifest(path string, cfg Config) error {
	if !cfg.Force {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
	}
	display := cfg.Display
	if display == "" {
		display = "fullscreen"
	}
	type manifestData struct {
		Config
		Display string
	}
	const tpl = `{
  "name": {{printf "%q" .Name}},
  "short_name": {{printf "%q" .ShortName}},
  "description": {{printf "%q" .Description}},
  "start_url": {{printf "%q" .StartURL}},
  "display": {{printf "%q" .Display}},
  "background_color": "#081014",
  "theme_color": {{printf "%q" .ThemeColor}},
  "orientation": "portrait-primary",
  "icons": [
    {
      "src": "/static/icons/icon-192.png",
      "sizes": "192x192",
      "type": "image/png",
      "purpose": "any"
    },
    {
      "src": "/static/icons/icon-512.png",
      "sizes": "512x512",
      "type": "image/png",
      "purpose": "any"
    },
    {
      "src": "/static/icons/icon-512-maskable.png",
      "sizes": "512x512",
      "type": "image/png",
      "purpose": "maskable"
    }
  ]
}
`
	t, err := template.New("manifest").Parse(tpl)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, manifestData{Config: cfg, Display: display}); err != nil {
		return err
	}
	return writeFileSafe(path, buf.Bytes())
}

func copyAsset(src, dst string) error {
	data, err := assets.ReadFile(src)
	if err != nil {
		return err
	}
	return writeFileSafe(dst, data)
}

// writeUserAsset writes an app-owned asset (offline page, brand image) only when
// missing unless force, so `pwa` can refresh the runtime without losing branding (#186).
func writeUserAsset(src, dst string, force bool) error {
	if !force && fileExists(dst) {
		return nil
	}
	return copyAsset(src, dst)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// writeFileSafe refuses to follow a planted symlink at the destination (#134).
func writeFileSafe(path string, data []byte) error {
	if err := fsutil.RefuseSymlinkWrite(path); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
