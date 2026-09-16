// cmd/server/main.go templates (full and blank cais new).
package cli

const tplMain = `package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/puppe1990/amarra-cais/pkg/amarra/view"
	"github.com/puppe1990/amarra-cais/pkg/cais"
	"github.com/puppe1990/amarra-cais/pkg/cais/boot"
	"github.com/puppe1990/amarra-cais/pkg/cais/meta"

	"{{.ModulePath}}/internal/app"
	appi18n "{{.ModulePath}}/internal/i18n"
	"{{.ModulePath}}/internal/store"
	"{{.ModulePath}}/web"
)

func main() {
	cfg := cais.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}
	preferredPort := cfg.Port
	port, shifted, err := cais.ResolvePort(cfg.Port, cfg.Env)
	if err != nil {
		log.Fatal(err)
	}
	cfg.Port = port

	a, err := bootstrapWithConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}

	shiftedFrom := ""
	if shifted {
		shiftedFrom = preferredPort
	}
	boot.Print(os.Stdout, boot.Options{
		AppName:         "{{.AppName}}",
		Config:          cfg,
		Version:         boot.CaisVersion(),
		PortShiftedFrom: shiftedFrom,
	})
	// Graceful shutdown on SIGINT/SIGTERM (air sends SIGINT before each
	// rebuild when send_interrupt is set) so the listener and sqlite close
	// instead of lingering as a zombie holding :8080 and data/app.db (#77).
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := a.RunContext(ctx); err != nil {
		log.Fatal(err)
	}
}

func bootstrapWithConfig(cfg cais.Config) (*app.App, error) {
	tmplFS, err := fs.Sub(web.Templates, "templates")
	if err != nil {
		return nil, fmt.Errorf("templates: %w", err)
	}

	catalog := appi18n.NewCatalog(cfg.Locale)
	views, err := view.Load(tmplFS, catalog)
	if err != nil {
		return nil, fmt.Errorf("views: %w", err)
	}

	s, err := store.NewSQLiteStore(cfg.DBPath, cfg.Env)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	staticDir, err := cais.ResolveWebDir("static", cfg.StaticDir)
	if err != nil {
		_ = s.Close()
		return nil, err
	}

	return app.New(cfg, app.Deps{
		Views:     views,
		Store:     s,
		StaticDir: staticDir,
		Site:      meta.SiteFrom("{{.AppName}}", cfg.AppURL),
		Catalog:   catalog,
	})
}
`

const tplMainBlank = tplMain
