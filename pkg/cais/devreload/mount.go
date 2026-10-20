// Package devreload wires `amarra-cais dev` reloads: templates are re-parsed
// in-process on edit and every open tab is pinged over an SSE route, so an
// HTML/CSS change never needs an air rebuild or a killed WebSocket hub (#340).
package devreload

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/puppe1990/amarra-cais/pkg/cais"
	"github.com/puppe1990/amarra-cais/pkg/cais/netutil"
)

// Enabled reports whether the reload route and watcher are active — development
// only, like the /logs viewer.
func Enabled(env string) bool {
	return env == "development"
}

// Register mounts the SSE route. Production never gets it, and loopback-only
// keeps it off the LAN URL used for phone testing (mirrors devlog.Register).
func (h *Hub) Register(r *cais.Router, env string) {
	if h == nil || r == nil || !Enabled(env) {
		return
	}
	r.Get(ReloadPath, localOnly(h).ServeHTTP)
}

func localOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !netutil.IsLoopbackRequest(r) {
			http.Error(w, "dev reload only available on localhost", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Watch starts the dev watcher over the app's templates directory and built
// stylesheet: a template edit reloads the renderer in-process (via reloadHTML,
// which may be nil), and every change pings the browser. Returns nil when the
// environment is not development.
func Watch(ctx context.Context, cfg cais.Config, reloadHTML func(dir string) error) *Hub {
	if !Enabled(cfg.Env) {
		return nil
	}
	templatesDir, err := cais.ResolveWebDir("templates", cfg.TemplatesDir)
	if err != nil {
		log.Printf("dev reload: %v", err)
		return nil
	}

	hub := New()
	targets := append([]string{templatesDir}, builtStylesheet(cfg))
	watcher, err := NewWatcher(targets, func(path string) {
		if strings.HasSuffix(path, ".html") {
			if reloadHTML == nil {
				return
			}
			if err := reloadHTML(templatesDir); err != nil {
				// A template mid-edit fails to parse: keep the last good renderer
				// and let the next save try again.
				log.Printf("dev reload: %s: %v", path, err)
				return
			}
		}
		log.Printf("dev reload: %s", path)
		hub.Notify()
	})
	if err != nil {
		log.Printf("dev reload: %v", err)
		return hub
	}
	go watcher.Run(ctx, Config{})
	return hub
}

// builtStylesheet is the Tailwind output the `dev` command watches; empty when
// the app has no built stylesheet yet.
func builtStylesheet(cfg cais.Config) string {
	staticDir, err := cais.ResolveWebDir("static", cfg.StaticDir)
	if err != nil {
		return ""
	}
	path := filepath.Join(staticDir, "css", "styles.css")
	if _, err := os.Stat(path); err != nil {
		return ""
	}
	return path
}
