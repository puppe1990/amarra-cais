package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/puppe1990/amarra-cais/pkg/cais"
	"github.com/puppe1990/amarra-cais/pkg/cais/pwa"
)

func (c *CLI) cmdPWA(args []string) error {
	dir, err := c.appDir()
	if err != nil {
		if isCaisFramework(mustCwd()) {
			dir = mustCwd()
		} else {
			return err
		}
	}

	bump, force := false, false
	for _, arg := range args {
		switch arg {
		case "--bump":
			bump = true
		case "--force":
			force = true
		}
	}

	name := filepath.Base(dir)
	if name == "." || name == "/" || name == string(os.PathSeparator) {
		name = "Cais"
	}
	// Detect legacy cache-first SW before Install overwrites (for messaging).
	legacySW := false
	if data, rerr := os.ReadFile(filepath.Join(dir, "web", "static", "js", "sw.js")); rerr == nil {
		legacySW = !pwa.HasNetworkFirstSPA(string(data))
	}

	_, _ = fmt.Fprintln(c.Out, "→ writing PWA assets")
	// HTML Amarra apps: amarra.js + network-first SW. Never dump HTMX/Inertia bundles.
	// Existing brand assets (manifest, offline.html, og.png, icons) are preserved
	// unless --force; framework runtime is always refreshed (#186).
	cfg := pwa.DefaultConfig(name)
	cfg.Force = force
	err = pwa.WriteStaticAmarra(dir, cfg)
	if err != nil {
		return err
	}
	// Sync again is cheap/idempotent; returns preserved CACHE_VERSION for the log line.
	_, ver, err := pwa.SyncServiceWorker(dir)
	if err != nil {
		return err
	}
	if legacySW {
		_, _ = fmt.Fprintf(c.Out, "→ sw.js updated (network-first /static/js/amarra.js + /static/css/, CACHE_VERSION=%d)\n", ver)
		_, _ = fmt.Fprintln(c.Out, "  tip: amarra-cais pwa --bump after deploy so installed PWAs drop old caches")
	} else {
		_, _ = fmt.Fprintf(c.Out, "→ sw.js network-first for Amarra (CACHE_VERSION=%d)\n", ver)
	}
	// #187: --bump refreshes vendored assets first (above) and then invalidates the
	// cache, so a framework upgrade is one command instead of two.
	if bump {
		bumped, err := pwa.BumpCacheVersion(dir)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(c.Out, "→ PWA cache version bumped to %d (reinstall PWA on phone to pick up changes)\n", bumped)
	}
	return nil
}

func mustCwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return cwd
}

func preferredPort(dir string) string {
	if v := resolveEnvVar(dir, "PORT"); v != "" {
		return v
	}
	return ":8080"
}

func warnPortInUse(w io.Writer, dir string) {
	port := preferredPort(dir)
	if !strings.HasPrefix(port, ":") && !strings.Contains(port, ":") {
		port = ":" + port
	}
	if cais.PortBusy(port) {
		_, _ = fmt.Fprintf(w, "=> Warning: port %s already in use — another server may be running; phone/LAN clients can hit stale code\n", port)
	}
}
