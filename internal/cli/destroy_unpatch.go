package cli

import (
	"go/ast"
	"os"
	"path/filepath"
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
	content, err := removeStmtsFromFunc(content, "registerRoutes", drop)
	if err != nil {
		return "", err
	}
	// The generator's admin limiter/token vars and the middleware/http imports
	// are only used by the routes just removed; leaving them breaks the build on
	// an app whose remaining routes (a dashboard, say) need neither (#248).
	if content, err = pruneUnusedAssigns(content, "registerRoutes"); err != nil {
		return "", err
	}
	content = dropUnusedImport(content, "http")
	content = dropUnusedImport(content, "middleware")
	return content, nil
}

func unpatchRoutesForHandler(dir string, data scaffoldData, dryRun bool) error {
	path := filepath.Join(dir, "internal/app/routes.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content, err := unpatchHandlerRoutes(string(body), data)
	if err != nil {
		return err
	}
	return updateScaffoldFile(path, []byte(content), "internal/app/routes.go", dryRun)
}

// unpatchHandlerRoutes drops the handler assignment and every route that
// calls it, including nested paths such as POST /dashboard/tap (#397).
func unpatchHandlerRoutes(content string, data scaffoldData) (string, error) {
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
			return stmtCallsIdent(s, camel)
		}
		return false
	}
	content, err := removeStmtsFromFunc(content, "registerRoutes", drop)
	if err != nil {
		return "", err
	}
	content, err = pruneUnusedAssigns(content, "registerRoutes")
	if err != nil {
		return "", err
	}
	content = dropUnusedImport(content, "http")
	return dropUnusedImport(content, "middleware"), nil
}
