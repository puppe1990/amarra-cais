package cli

import (
	"fmt"
	"go/ast"
	"os"
	"path/filepath"
	"strings"
)

func destroyAuth(dir string, dryRun bool) error {
	files := []string{
		"internal/models/user.go",
		"internal/handlers/auth.go",
		"internal/handlers/auth_test.go",
		"internal/handlers/auth_signup_test.go",
		"internal/handlers/auth_reset_test.go",
		"internal/store/password_reset.go",
		"internal/store/password_reset_test.go",
		"web/src/pages/Login.svelte",
		"web/src/pages/Signup.svelte",
		"web/src/pages/ForgotPassword.svelte",
		"web/src/pages/ResetPassword.svelte",
		"web/src/components/AuthLayout.svelte",
		// Legacy HTMX pages (pre-Inertia auth generator)
		"web/templates/pages/login.html",
		"web/templates/pages/signup.html",
		"web/templates/pages/forgot_password.html",
		"web/templates/pages/reset_password.html",
	}

	migrationsDir := filepath.Join(dir, "internal/store/migrations")
	entries, _ := os.ReadDir(migrationsDir)
	for _, e := range entries {
		if !e.IsDir() && strings.Contains(e.Name(), "_auth.sql") {
			files = append(files, filepath.Join("internal/store/migrations", e.Name()))
		}
	}

	for _, rel := range files {
		full := filepath.Join(dir, rel)
		if _, err := os.Stat(full); err != nil {
			continue
		}
		if dryRun {
			printfScaffold("remove", rel)
			continue
		}
		if err := os.Remove(full); err != nil {
			return fmt.Errorf("remove %s: %w", rel, err)
		}
	}

	if err := unpatchStoreForAuth(dir, dryRun); err != nil {
		return err
	}
	if err := unpatchAppForAuth(dir, dryRun); err != nil {
		return err
	}
	return unpatchRoutesForAuthDestroy(dir, dryRun)
}

func unpatchStoreForAuth(dir string, dryRun bool) error {
	path := filepath.Join(dir, "internal/store/store.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content, err := removeStoreAuthMethods(string(body))
	if err != nil {
		return err
	}
	return updateScaffoldFile(path, []byte(content), "internal/store/store.go", dryRun)
}

func removeStoreAuthMethods(content string) (string, error) {
	patterns := []string{
		"FindUserByEmail",
		"CreateUser",
		"CreatePasswordResetToken",
		"ClearPasswordResetTokens",
		"FindPasswordResetUserID",
		"ResetPasswordWithToken",
	}
	names := make(map[string]bool, len(patterns))
	for _, p := range patterns {
		names[p] = true
	}
	content, err := removeDeclsByName(content, names)
	if err != nil {
		return "", err
	}
	content, err = removeInterfaceMethods(content, "Store", names)
	if err != nil {
		return "", err
	}
	content = cleanupStoreImports(content)
	content = dropUnusedImport(content, "session")
	// CreateUser's UNIQUE check was the only strings user (#248).
	content = dropUnusedImport(content, "strings")
	return content, nil
}

func unpatchAppForAuth(dir string, dryRun bool) error {
	path := filepath.Join(dir, "internal/app/app.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := unpatchAuthMiddleware(string(body))
	return updateScaffoldFile(path, []byte(content), "internal/app/app.go", dryRun)
}

// unpatchAuthMiddleware is a no-op: LoadSession/Flash are baseline middleware in full and blank apps.
func unpatchAuthMiddleware(content string) string {
	return content
}

func unpatchRoutesForAuthDestroy(dir string, dryRun bool) error {
	path := filepath.Join(dir, "internal/app/routes.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content, err := unpatchAuthRoutes(string(body))
	if err != nil {
		return err
	}
	return updateScaffoldFile(path, []byte(content), "internal/app/routes.go", dryRun)
}

// unpatchAuthRoutes drops the auth handler assignment and every route that
// calls it (#248); login/reset limiters were only used by those routes, so the
// unused-declaration prune takes their statements too.
func unpatchAuthRoutes(content string) (string, error) {
	drop := func(st ast.Stmt) bool {
		if assign, ok := st.(*ast.AssignStmt); ok {
			for _, lhs := range assign.Lhs {
				if id, ok := lhs.(*ast.Ident); ok && id.Name == "auth" {
					return true
				}
			}
		}
		if expr, ok := st.(*ast.ExprStmt); ok {
			return stmtCallsIdent(expr, "auth")
		}
		return false
	}
	content, err := removeStmtsFromFunc(content, "registerRoutes", drop)
	if err != nil {
		return "", err
	}
	if content, err = pruneUnusedAssigns(content, "registerRoutes"); err != nil {
		return "", err
	}
	// The dashboard keeps its route once the auth redirect target is gone.
	content = strings.Replace(content,
		`middleware.RequireAuthFunc("/login", dashboard.ServeHTTP)`,
		`dashboard.ServeHTTP`,
		1,
	)
	content = dropUnusedImport(content, "middleware")
	return dropUnusedImport(content, "http"), nil
}

func destroyMigration(dir, name string, dryRun bool) error {
	data := dataForHandler(name)
	suffix := "_" + data.Snake + ".sql"

	migrationsDir := filepath.Join(dir, "internal/store/migrations")
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return err
	}

	var removed int
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), suffix) {
			continue
		}
		rel := filepath.Join("internal/store/migrations", e.Name())
		if dryRun {
			printfScaffold("remove", rel)
			removed++
			continue
		}
		if err := os.Remove(filepath.Join(dir, rel)); err != nil {
			return fmt.Errorf("remove %s: %w", rel, err)
		}
		removed++
	}

	if removed == 0 {
		return fmt.Errorf("no migration matching %q", suffix)
	}
	return nil
}
