package pwa

import (
	"embed"
	"io/fs"
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
	// Force rewrites app-owned brand assets (manifest, offline.html, og.png, icons)
	// that already exist. Framework runtime (amarra.js, sw.js) is always refreshed;
	// without Force `pwa` can upgrade the runtime without clobbering branding (#186).
	Force bool
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
    <link rel="icon" href="/static/favicon.svg" type="image/svg+xml" />`
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
        navigator.serviceWorker.register("/static/js/sw.js", { scope: "/" });
      }
    </script>`
}

// FS returns embedded PWA assets (for tests).
func FS() (fs.FS, error) {
	return fs.Sub(assets, "assets")
}
