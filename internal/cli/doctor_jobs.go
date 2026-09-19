package cli

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func checkJobsUI(dir string) doctorCheck {
	data, err := os.ReadFile(filepath.Join(dir, "internal", "app", "app.go"))
	if err != nil {
		return doctorCheck{Name: "/jobs dashboard", OK: true, Detail: "skipped (no internal/app/app.go)"}
	}
	hasRegister := strings.Contains(string(data), "jobsui.Register")
	hasDB := storeExposesDB(dir)
	switch {
	case hasRegister && hasDB:
		return doctorCheck{Name: "/jobs dashboard", OK: true, Detail: "localhost queue viewer"}
	case hasRegister:
		// Register without DB() does not compile, so this is not a green state (#194).
		return doctorCheck{
			Name:     "/jobs dashboard",
			Optional: true,
			Detail:   "jobsui.Register present but the Store interface is missing DB() — app will not compile",
			FixHint:  "expose DB() *sql.DB on the Store interface (internal/store/store.go)",
		}
	default:
		return doctorCheck{
			Name:     "/jobs dashboard",
			Optional: true,
			Detail:   "missing jobsui.Register — queue viewer not mounted",
			FixHint:  "expose DB() *sql.DB on the Store interface, then add jobsui.Register(r, deps.Store.DB()) in app.New (after routes)",
		}
	}
}

// storeExposesDB parses internal/store/store.go and reports whether the Store
// interface declares DB(). A plain string search would also match the concrete
// *SQLiteStore method, hiding a Store interface that still lacks it (#194).
func storeExposesDB(dir string) bool {
	path := filepath.Join(dir, "internal", "store", "store.go")
	f, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return false
	}
	for _, decl := range f.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.TYPE {
			continue
		}
		for _, spec := range gd.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok || ts.Name.Name != "Store" {
				continue
			}
			iface, ok := ts.Type.(*ast.InterfaceType)
			if !ok {
				continue
			}
			for _, method := range iface.Methods.List {
				for _, name := range method.Names {
					if name.Name == "DB" {
						return true
					}
				}
			}
		}
	}
	return false
}
