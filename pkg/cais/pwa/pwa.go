package pwa

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/fs"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed assets/*
var assets embed.FS

const ThemeColor = "#c9893a"

type Config struct {
	Name        string
	ShortName   string
	Description string
	StartURL    string
	Display     string
	ThemeColor  string
	IconPath    string
}

func DefaultConfig(name string) Config {
	short := name
	if len(short) > 12 {
		short = short[:12]
	}
	return Config{
		Name:        name,
		ShortName:   short,
		Description: name + " — powered by Cais",
		StartURL:    "/",
		Display:     "fullscreen",
		ThemeColor:  ThemeColor,
	}
}

// HeadHTML returns meta tags and links to include in layout <head>.
func HeadHTML() string {
	return `<link rel="manifest" href="/static/manifest.webmanifest" />
    <meta name="theme-color" content="#c9893a" />
    <meta name="mobile-web-app-capable" content="yes" />
    <meta name="apple-mobile-web-app-capable" content="yes" />
    <meta name="apple-mobile-web-app-status-bar-style" content="black-translucent" />
    <meta name="apple-mobile-web-app-title" content="Cais" />
    <link rel="apple-touch-icon" href="/static/icons/icon.png" />
    <link rel="icon" href="/static/icons/icon.png" type="image/png" />`
}

// RegisterScript returns inline JS to register the service worker.
func RegisterScript() string {
	return RegisterScriptForEnv("production")
}

// RegisterScriptForEnv skips the service worker in development so static assets are not cached during hot reload.
func RegisterScriptForEnv(env string) string {
	if env == "development" {
		return `<script>
      if ("serviceWorker" in navigator) {
        navigator.serviceWorker.getRegistrations().then(function (regs) {
          regs.forEach(function (r) { r.unregister(); });
        });
        if ("caches" in window) {
          caches.keys().then(function (keys) {
            keys.forEach(function (k) { caches.delete(k); });
          });
        }
      }
    </script>`
	}
	return `<script>
      if ("serviceWorker" in navigator) {
        navigator.serviceWorker.register("/static/js/sw.js");
      }
    </script>`
}

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
		{"assets/offline.html", "offline.html"},
		{"assets/go-on-cais.jpg", "img/go-on-cais.jpg"},
	} {
		if err := copyAsset(pair.src, filepath.Join(staticDir, pair.dst)); err != nil {
			return err
		}
	}
	// SW via SyncServiceWorker so CACHE_VERSION is preserved on upgrades.
	if _, _, err := SyncServiceWorker(appDir); err != nil {
		return err
	}

	if err := writeOGImage(filepath.Join(staticDir, "og.png")); err != nil {
		return err
	}
	if err := writeAppIcons(filepath.Join(staticDir, "icons"), cfg.IconPath); err != nil {
		return err
	}

	return nil
}

// InstallTo writes PWA assets using DefaultConfig(name).
func InstallTo(appDir, name string) error {
	return WriteStatic(appDir, DefaultConfig(name))
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
		{"assets/offline.html", "offline.html"},
		{"assets/go-on-cais.jpg", "img/go-on-cais.jpg"},
	} {
		if err := copyAsset(pair.src, filepath.Join(staticDir, pair.dst)); err != nil {
			return err
		}
	}
	if _, _, err := SyncServiceWorker(appDir); err != nil {
		return err
	}

	if err := writeOGImage(filepath.Join(staticDir, "og.png")); err != nil {
		return err
	}
	if err := writeAppIcons(filepath.Join(staticDir, "icons"), cfg.IconPath); err != nil {
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
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

func copyAsset(src, dst string) error {
	data, err := assets.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}

// writeOGImage copies the neutral 1200x630 preview shipped with the framework.
// Apps replace web/static/og.png with their own art (#64).
func writeOGImage(path string) error {
	return copyAsset("assets/og.png", path)
}

func fill(img *image.RGBA, c color.RGBA) {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			img.Set(x, y, c)
		}
	}
}

func encodePNG(path string, img image.Image) error {
	body, err := encodePNGBytes(img)
	if err != nil {
		return err
	}
	return os.WriteFile(path, body, 0o644)
}

func encodePNGBytes(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// iconSource returns the icon an app scaffolds with: the caller's brand file
// when configured, otherwise the framework's neutral placeholder tile.
func iconSource(iconPath string) ([]byte, error) {
	if iconPath != "" {
		return os.ReadFile(iconPath)
	}
	return assets.ReadFile("assets/icon.png")
}

func writeAppIcons(dir string, iconPath string) error {
	data, err := iconSource(iconPath)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "icon.png"), data, 0o644); err != nil {
		return err
	}
	src, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return err
	}
	for _, size := range []int{192, 512} {
		dst := resizeNearest(src, size, size)
		if err := encodePNG(filepath.Join(dir, fmt.Sprintf("icon-%d.png", size)), dst); err != nil {
			return err
		}
	}
	// Separate maskable file: platforms crop `any` icons, so keep a version whose
	// content sits inside the safe zone (#64).
	if err := encodePNG(filepath.Join(dir, "icon-512-maskable.png"), maskableIcon(src, 512)); err != nil {
		return err
	}
	return nil
}

// maskableIcon insets src to 80% of the canvas over its own corner color, the
// Android safe zone for maskable icons.
func maskableIcon(src image.Image, size int) *image.RGBA {
	inner := resizeNearest(src, size*4/5, size*4/5)
	bounds := src.Bounds()
	bg := color.RGBAModel.Convert(src.At(bounds.Min.X, bounds.Min.Y)).(color.RGBA)
	dst := image.NewRGBA(image.Rect(0, 0, size, size))
	fill(dst, bg)
	offset := (size - inner.Bounds().Dx()) / 2
	for y := 0; y < inner.Bounds().Dy(); y++ {
		for x := 0; x < inner.Bounds().Dx(); x++ {
			dst.Set(offset+x, offset+y, inner.At(x, y))
		}
	}
	return dst
}

// DefaultBrandAssets returns the brand files a scaffold writes, keyed by their
// path under web/static. doctor compares app files against them to warn while
// the app still ships the placeholder (#64).
func DefaultBrandAssets() map[string][]byte {
	out := map[string][]byte{}
	data, err := assets.ReadFile("assets/icon.png")
	if err != nil {
		return out
	}
	src, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return out
	}
	out["icons/icon.png"] = data
	encoded := map[string]image.Image{
		"icons/icon-192.png":          resizeNearest(src, 192, 192),
		"icons/icon-512.png":          resizeNearest(src, 512, 512),
		"icons/icon-512-maskable.png": maskableIcon(src, 512),
	}
	for rel, img := range encoded {
		body, err := encodePNGBytes(img)
		if err != nil {
			return map[string][]byte{}
		}
		out[rel] = body
	}
	if og, err := assets.ReadFile("assets/og.png"); err == nil {
		out["og.png"] = og
	}
	return out
}

func resizeNearest(src image.Image, w, h int) *image.RGBA {
	bounds := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	sw, sh := bounds.Dx(), bounds.Dy()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			sx := bounds.Min.X + x*sw/w
			sy := bounds.Min.Y + y*sh/h
			dst.Set(x, y, src.At(sx, sy))
		}
	}
	return dst
}

// FS returns embedded PWA assets (for tests).
func FS() (fs.FS, error) {
	return fs.Sub(assets, "assets")
}
