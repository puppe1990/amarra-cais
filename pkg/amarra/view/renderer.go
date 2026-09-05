package view

import (
	"fmt"
	"html/template"
	"io/fs"
	"path"
	"sort"
	"strings"
	"sync"

	"github.com/puppe1990/amarra-cais/pkg/cais/forms"
	"github.com/puppe1990/amarra-cais/pkg/cais/htmxattrs"
	"github.com/puppe1990/amarra-cais/pkg/cais/i18n"
	"github.com/puppe1990/amarra-cais/pkg/cais/meta"
	"github.com/puppe1990/amarra-cais/pkg/cais/money"
)

type Renderer struct {
	mu    sync.RWMutex
	pages map[string]*template.Template
}

type namedTemplateSrc struct {
	name string
	src  string
}

// Load parses layouts, pages, and components once at boot.
// Unknown <.component> tags fail here so apps do not serve broken views.
func Load(fsys fs.FS, catalog *i18n.Catalog) (*Renderer, error) {
	if catalog == nil {
		catalog = i18n.DefaultCatalog()
	}

	components, err := loadComponentSources(fsys)
	if err != nil {
		return nil, err
	}

	layouts, err := loadExpandedLayouts(fsys, components)
	if err != nil {
		return nil, err
	}
	if len(layouts) == 0 {
		return nil, fmt.Errorf("no layout templates found")
	}

	pagePaths, err := listPagePaths(fsys)
	if err != nil {
		return nil, err
	}

	pages := make(map[string]*template.Template, len(pagePaths))
	for _, pagePath := range pagePaths {
		raw, err := fs.ReadFile(fsys, pagePath)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", pagePath, err)
		}
		expanded, err := ExpandAll(string(raw), components)
		if err != nil {
			return nil, fmt.Errorf("expand %s: %w", pagePath, err)
		}
		name := pageNameFromPath(pagePath)
		tmpl, err := parsePageTemplates(layouts, name, expanded, catalog)
		if err != nil {
			return nil, fmt.Errorf("parse page %s: %w", name, err)
		}
		pages[name] = tmpl
	}

	return &Renderer{pages: pages}, nil
}

func loadComponentSources(fsys fs.FS) (map[string]string, error) {
	paths, err := fs.Glob(fsys, "components/*.html")
	if err != nil {
		return nil, err
	}
	components := make(map[string]string, len(paths))
	for _, p := range paths {
		raw, err := fs.ReadFile(fsys, p)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", p, err)
		}
		stem := strings.TrimSuffix(path.Base(p), ".html")
		components[stem] = string(raw)
	}
	return components, nil
}

func loadExpandedLayouts(fsys fs.FS, components map[string]string) ([]namedTemplateSrc, error) {
	paths, err := fs.Glob(fsys, "layouts/*.html")
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	layouts := make([]namedTemplateSrc, 0, len(paths))
	for _, p := range paths {
		raw, err := fs.ReadFile(fsys, p)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", p, err)
		}
		expanded, err := ExpandAll(string(raw), components)
		if err != nil {
			return nil, fmt.Errorf("expand %s: %w", p, err)
		}
		layouts = append(layouts, namedTemplateSrc{name: path.Base(p), src: expanded})
	}
	return layouts, nil
}

func listPagePaths(fsys fs.FS) ([]string, error) {
	top, err := fs.Glob(fsys, "pages/*.html")
	if err != nil {
		return nil, err
	}
	nested, err := fs.Glob(fsys, "pages/*/*.html")
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0, len(top)+len(nested))
	paths = append(paths, top...)
	paths = append(paths, nested...)
	sort.Strings(paths)
	return paths, nil
}

func pageNameFromPath(pagePath string) string {
	name := strings.TrimSuffix(pagePath, ".html")
	return strings.TrimPrefix(name, "pages/")
}

func parsePageTemplates(layouts []namedTemplateSrc, pageName, pageSrc string, catalog *i18n.Catalog) (*template.Template, error) {
	root := template.New("_amarra").Funcs(templateFuncs(catalog))
	for _, layout := range layouts {
		if _, err := root.New("layout:" + layout.name).Parse(layout.src); err != nil {
			return nil, fmt.Errorf("layout %s: %w", layout.name, err)
		}
	}
	if _, err := root.New("page:" + pageName).Parse(pageSrc); err != nil {
		return nil, err
	}
	return root, nil
}

// templateFuncs copies the merge in pkg/cais/render.go so jobsui/devlog can
// keep cais.Renderer until those UIs move.
func templateFuncs(catalog *i18n.Catalog) template.FuncMap {
	extra := meta.TemplateFuncs()
	for k, v := range forms.Funcs() {
		extra[k] = v
	}
	for k, v := range htmxattrs.Funcs() {
		extra[k] = v
	}
	extra["formatMoney"] = money.FormatBRL
	return i18n.MergeFuncs(catalog, extra)
}
