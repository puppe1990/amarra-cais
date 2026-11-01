package cli

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// #385: missing catalog keys render as the raw key (Catalog.T). Doctor warns
// on catalog parity and on static t / .t / .T literals in templates and Go.
func checkI18n(dir string) []doctorCheck {
	catalogs := parseAppI18nCatalogs(filepath.Join(dir, "internal", "i18n"))
	if len(catalogs) == 0 {
		return []doctorCheck{{
			Name:   "i18n",
			OK:     true,
			Detail: "skipped (no internal/i18n)",
		}}
	}

	union := map[string]struct{}{}
	for _, keys := range catalogs {
		for k := range keys {
			union[k] = struct{}{}
		}
	}
	for k := range collectUsedI18nKeys(dir) {
		union[k] = struct{}{}
	}

	locales := make([]string, 0, len(catalogs))
	for loc := range catalogs {
		locales = append(locales, loc)
	}
	sort.Strings(locales)

	var warns []doctorCheck
	for _, loc := range locales {
		missing := missingI18nKeys(union, catalogs[loc])
		if len(missing) == 0 {
			continue
		}
		noun := "chaves ausentes"
		if len(missing) == 1 {
			noun = "chave ausente"
		}
		warns = append(warns, doctorCheck{
			Name:     "i18n " + loc,
			Optional: true,
			Detail:   fmt.Sprintf("%d %s: %s", len(missing), noun, strings.Join(missing, ", ")),
			FixHint:  "adicione ao catálogo " + loc + " ou corrija o literal",
		})
	}
	if len(warns) > 0 {
		return warns
	}
	return []doctorCheck{{
		Name:   "i18n",
		OK:     true,
		Detail: strings.Join(locales, "/") + " catalogs match",
	}}
}

func missingI18nKeys(union map[string]struct{}, have map[string]struct{}) []string {
	var missing []string
	for k := range union {
		if _, ok := have[k]; !ok {
			missing = append(missing, k)
		}
	}
	sort.Strings(missing)
	return missing
}

func parseAppI18nCatalogs(i18nDir string) map[string]map[string]struct{} {
	ents, err := os.ReadDir(i18nDir)
	if err != nil {
		return nil
	}
	fset := token.NewFileSet()
	namedMaps := map[string]map[string]struct{}{}
	localeIdents := map[string]string{}
	out := map[string]map[string]struct{}{}

	for _, ent := range ents {
		if ent.IsDir() || !strings.HasSuffix(ent.Name(), ".go") || strings.HasSuffix(ent.Name(), "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Join(i18nDir, ent.Name()), nil, 0)
		if err != nil {
			continue
		}
		ast.Inspect(f, func(n ast.Node) bool {
			vs, ok := n.(*ast.ValueSpec)
			if !ok {
				return true
			}
			for i, name := range vs.Names {
				if i >= len(vs.Values) {
					continue
				}
				lit, ok := vs.Values[i].(*ast.CompositeLit)
				if !ok {
					continue
				}
				switch i18nMapKind(lit) {
				case i18nMapStringString:
					namedMaps[name.Name] = compositeStringKeys(lit)
				case i18nMapStringMap:
					collectLocaleComposite(lit, localeIdents, out)
				}
			}
			return true
		})
	}

	for loc, ident := range localeIdents {
		if m, ok := namedMaps[ident]; ok {
			out[loc] = m
		}
	}
	if len(out) > 0 {
		return out
	}
	for name, keys := range namedMaps {
		if loc, ok := localeFromVar(name); ok {
			out[loc] = keys
		}
	}
	return out
}

const (
	i18nMapNone = iota
	i18nMapStringString
	i18nMapStringMap
)

func i18nMapKind(lit *ast.CompositeLit) int {
	mt, ok := lit.Type.(*ast.MapType)
	if !ok || identName(mt.Key) != "string" {
		return i18nMapNone
	}
	switch v := mt.Value.(type) {
	case *ast.Ident:
		if v.Name == "string" {
			return i18nMapStringString
		}
	case *ast.MapType:
		if identName(v.Key) == "string" && identName(v.Value) == "string" {
			return i18nMapStringMap
		}
	}
	return i18nMapNone
}

func identName(n ast.Expr) string {
	id, ok := n.(*ast.Ident)
	if !ok {
		return ""
	}
	return id.Name
}

func compositeStringKeys(lit *ast.CompositeLit) map[string]struct{} {
	out := map[string]struct{}{}
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		k, ok := astStringLit(kv.Key)
		if !ok {
			continue
		}
		out[k] = struct{}{}
	}
	return out
}

func collectLocaleComposite(lit *ast.CompositeLit, idents map[string]string, inline map[string]map[string]struct{}) {
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		loc, ok := astStringLit(kv.Key)
		if !ok {
			continue
		}
		switch v := kv.Value.(type) {
		case *ast.Ident:
			idents[loc] = v.Name
		case *ast.CompositeLit:
			inline[loc] = compositeStringKeys(v)
		}
	}
}

func localeFromVar(name string) (string, bool) {
	base := strings.ToLower(strings.TrimSuffix(name, "Messages"))
	switch base {
	case "en", "pt", "es", "zh", "fr", "de", "it", "ja":
		return base, true
	}
	return "", false
}

var tplI18nKey = regexp.MustCompile(`\bt\s+"([^"%\\]+)"`)

func collectUsedI18nKeys(dir string) map[string]struct{} {
	out := map[string]struct{}{}
	collectTemplateI18nKeys(filepath.Join(dir, "web", "templates"), out)
	collectGoI18nKeys(filepath.Join(dir, "internal"), out)
	return out
}

func collectTemplateI18nKeys(root string, out map[string]struct{}) {
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".html") {
			return nil
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		for _, m := range tplI18nKey.FindAllSubmatch(body, -1) {
			out[string(m[1])] = struct{}{}
		}
		return nil
	})
}

func collectGoI18nKeys(root string, out map[string]struct{}) {
	fset := token.NewFileSet()
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			switch d.Name() {
			case "vendor", "node_modules", "testdata":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return nil
		}
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || !isI18nTCall(call.Fun) {
				return true
			}
			key, ok := firstStringArg(call.Args)
			if ok && !strings.Contains(key, "%") {
				out[key] = struct{}{}
			}
			return true
		})
		return nil
	})
}

func isI18nTCall(fun ast.Expr) bool {
	switch x := fun.(type) {
	case *ast.Ident:
		return x.Name == "t" || x.Name == "T"
	case *ast.SelectorExpr:
		return x.Sel.Name == "t" || x.Sel.Name == "T"
	}
	return false
}

func firstStringArg(args []ast.Expr) (string, bool) {
	for _, a := range args {
		if s, ok := astStringLit(a); ok {
			return s, true
		}
	}
	return "", false
}

func astStringLit(n ast.Node) (string, bool) {
	lit, ok := n.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}
	s, err := strconv.Unquote(lit.Value)
	if err != nil {
		return "", false
	}
	return s, true
}
