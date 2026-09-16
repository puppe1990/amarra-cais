package view

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/puppe1990/amarra-cais/pkg/amarra"
	"github.com/puppe1990/amarra-cais/pkg/cais"
)

type Page struct {
	Layout string // default "app"
	Name   string // "home", "items/index"
	Frame  string // frame id (e.g. "cart"); fragment only when Amarra-Frame is set
	Data   any
	Status int
	// CacheControl overrides the default "no-store" (#80). Empty means
	// default; a handler can also preset the header (e.g. ETag list pages).
	CacheControl string
}

// Write renders a full layout+page, a Drive document (layout kept so JS can
// morph #amarra-main), or a Frame fragment (frame:<id> only — no shell).
// Status is committed only after the template executes so a missing page/frame
// cannot stick a 422 on a 500 error body.
func Write(w http.ResponseWriter, r *http.Request, rec *Renderer, p Page, cfg cais.Config) {
	tmpl, err := rec.lookupPage(p.Name)
	if err != nil {
		writeRenderError(w, err, cfg)
		return
	}
	name := writeTemplateName(r, p)
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, p.Data); err != nil {
		writeRenderError(w, err, cfg)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Dynamic pages default to no-store so browsers never heuristically
	// cache HTML + inline scripts (#80). Handlers opt into caching via
	// Page.CacheControl, a preset Cache-Control, or a preset ETag
	// (httpx.SetETag list flow keeps its 304 behavior).
	if p.CacheControl != "" {
		w.Header().Set("Cache-Control", p.CacheControl)
	} else if w.Header().Get("Cache-Control") == "" && w.Header().Get("ETag") == "" {
		w.Header().Set("Cache-Control", "no-store")
	}
	if p.Status != 0 {
		w.WriteHeader(p.Status)
	}
	if _, err := w.Write(buf.Bytes()); err != nil {
		log.Printf("write html: %v", err)
	}
}

func (rec *Renderer) lookupPage(name string) (*template.Template, error) {
	if rec == nil {
		return nil, fmt.Errorf("page %q not found", name)
	}
	rec.mu.RLock()
	tmpl, ok := rec.pages[name]
	rec.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("page %q not found", name)
	}
	return tmpl, nil
}

// writeTemplateName uses Amarra-Frame when present; otherwise the layout.
// html/template rejects nested define inside content — frames are sibling
// {{ define "frame:<id>" }} in the page file.
func writeTemplateName(r *http.Request, p Page) string {
	if id := amarra.FrameID(r); id != "" {
		return "frame:" + id
	}
	if p.Layout == "" {
		return "app"
	}
	return p.Layout
}

func writeRenderError(w http.ResponseWriter, err error, cfg cais.Config) {
	if cfg.SanitizeErrors() {
		log.Printf("render error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
