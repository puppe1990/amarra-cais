package view

import (
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
	Frame  string // optional define name "frame:cart"
	Data   any
	Status int
}

// Write renders a full layout+page, a Drive document (layout kept so JS can
// morph #amarra-main), or a Frame fragment (frame:<id> only — no shell).
func Write(w http.ResponseWriter, r *http.Request, rec *Renderer, p Page, cfg cais.Config) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if p.Status != 0 {
		w.WriteHeader(p.Status)
	}
	tmpl, err := rec.lookupPage(p.Name)
	if err != nil {
		writeRenderError(w, err, cfg)
		return
	}
	name := writeTemplateName(r, p)
	if err := tmpl.ExecuteTemplate(w, name, p.Data); err != nil {
		writeRenderError(w, err, cfg)
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

// writeTemplateName prefers Amarra-Frame, then Page.Frame, else the layout.
// Frame names are sibling {{ define "frame:<id>" }} in the page file —
// html/template rejects nested define inside content.
func writeTemplateName(r *http.Request, p Page) string {
	frameID := amarra.FrameID(r)
	if frameID == "" {
		frameID = p.Frame
	}
	if frameID != "" {
		return "frame:" + frameID
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
