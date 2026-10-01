package cli

import (
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func unpatchRoutesForResource(dir string, data scaffoldData, dryRun bool) error {
	path := filepath.Join(dir, "internal/app/routes.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content, err := unpatchResourceRoutes(string(body), data)
	if err != nil {
		return err
	}
	return updateScaffoldFile(path, []byte(content), "internal/app/routes.go", dryRun)
}

// unpatchResourceRoutes removes only the statements the generator added,
// matched via go/ast (#168). User comments and custom routes that merely
// mention /admin/<plural> survive; braces inside string literals can't corrupt
// block detection because the r.Group statement is removed as one AST node.

func unpatchResourceRoutes(content string, data scaffoldData) (string, error) {
	adminVar := "admin" + data.PluralPascal
	pubVar := lowerFirst(data.PluralPascal)
	adminPaths := map[string]bool{
		"/admin/" + data.Plural:                  true,
		"/admin/" + data.Plural + "/{id}":        true, // GET show, PUT update, DELETE destroy (#269)
		"/admin/" + data.Plural + "/new":         true,
		"/admin/" + data.Plural + "/{id}/edit":   true,
		"/admin/" + data.Plural + "/{id}/delete": true, // pre-#269 POST destroy
		"/admin/" + data.Plural + "/bulk-delete": true,
	}
	pubPaths := map[string]bool{
		"/" + data.Plural: true,
		// Pre-#261 generators registered this anonymous POST; destroy still
		// removes it from apps that were generated with the toggle.
		"/" + data.Plural + "/{id}/toggle": true,
	}
	drop := func(st ast.Stmt) bool {
		return isGeneratedRouteStmt(st, pubVar, adminVar, pubPaths, adminPaths)
	}
	return removeStmtsFromFunc(content, "registerRoutes", drop)
}

func unpatchRoutesForHandler(dir string, data scaffoldData, dryRun bool) error {
	path := filepath.Join(dir, "internal/app/routes.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	// Statement-exact removal (#168): the handler's var assignment and its
	// r.Get/r.Post on "/<snake>" calling <camel>.ServeHTTP.
	camel := data.Camel
	drop := func(st ast.Stmt) bool {
		switch s := st.(type) {
		case *ast.AssignStmt:
			for _, lhs := range s.Lhs {
				if id, ok := lhs.(*ast.Ident); ok && id.Name == camel {
					return true
				}
			}
		case *ast.ExprStmt:
			call, ok := s.X.(*ast.CallExpr)
			if !ok || len(call.Args) == 0 {
				return false
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return false
			}
			if recv, ok := sel.X.(*ast.Ident); !ok || recv.Name != "r" {
				return false
			}
			lit, ok := call.Args[0].(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING || strings.Trim(lit.Value, "`\"") != "/"+data.Snake {
				return false
			}
			matches := false
			ast.Inspect(call, func(n ast.Node) bool {
				if inner, ok := n.(*ast.SelectorExpr); ok {
					if id, ok := inner.X.(*ast.Ident); ok && id.Name == camel {
						matches = true
					}
				}
				return true
			})
			return matches
		}
		return false
	}
	content, err := removeStmtsFromFunc(string(body), "registerRoutes", drop)
	if err != nil {
		return err
	}
	content, err = pruneUnusedAssigns(content, "registerRoutes")
	if err != nil {
		return err
	}
	content = dropUnusedImport(content, "http")
	content = dropUnusedImport(content, "middleware")
	return updateScaffoldFile(path, []byte(content), "internal/app/routes.go", dryRun)
}
