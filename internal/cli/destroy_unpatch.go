package cli

import (
	"fmt"
	"go/ast"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// removeGeneratedFiles deletes rel paths under dir. Files whose content no
// longer matches the generation manifest are skipped with a warning unless
// force is set (#169). Deleted entries leave the manifest.

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
		"/admin/" + data.Plural + "/{id}":        true,
		"/admin/" + data.Plural + "/new":         true,
		"/admin/" + data.Plural + "/{id}/edit":   true,
		"/admin/" + data.Plural + "/{id}/delete": true,
		"/admin/" + data.Plural + "/bulk-delete": true,
	}
	pubPaths := map[string]bool{
		"/" + data.Plural:                  true,
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
	return updateScaffoldFile(path, []byte(content), "internal/app/routes.go", dryRun)
}

func unpatchStoreForResource(dir string, data scaffoldData, removed []string, dryRun bool) error {
	path := filepath.Join(dir, "internal/store/store.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content, err := removeStoreResourceMethods(string(body), data)
	if err != nil {
		return err
	}
	orphans, err := orphanOptionMethods(dir, removed, string(body))
	if err != nil {
		return err
	}
	if len(orphans) > 0 {
		if content, err = removeDeclsByName(content, orphans); err != nil {
			return err
		}
		if content, err = removeInterfaceMethods(content, "Store", orphans); err != nil {
			return err
		}
	}
	return updateScaffoldFile(path, []byte(content), "internal/store/store.go", dryRun)
}

// parentReferences lists surviving files that still use generated symbols of
// data after its own store methods/tests are removed (#108). store.go and
// store_test.go are evaluated post-unpatch so the parent's own generated code
// does not count as a reference; anything left (a child's SeedDemo*, tests)
// does.

func parentReferences(dir string, data scaffoldData, removed []string) ([]string, error) {
	storeRel := filepath.ToSlash(filepath.Join("internal", "store", "store.go"))
	testRel := filepath.ToSlash(filepath.Join("internal", "store", "store_test.go"))
	symbols := []string{
		"models." + data.Pascal,
		"Insert" + data.Pascal + "(",
		"Update" + data.Pascal + "(",
		"Delete" + data.Pascal + "(",
		"Find" + data.Pascal + "ByID(",
		"List" + data.Pascal + "Options(",
	}
	var refs []string

	storeBody, err := os.ReadFile(filepath.Join(dir, storeRel))
	if err == nil {
		content, rmErr := removeStoreResourceMethods(string(storeBody), data)
		if rmErr != nil {
			return nil, rmErr
		}
		if containsAny(content, symbols) {
			refs = append(refs, storeRel)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("read %s: %w", storeRel, err)
	}

	testBody, err := os.ReadFile(filepath.Join(dir, testRel))
	if err == nil {
		content, rmErr := removeDeclsByName(string(testBody), map[string]bool{"TestStore_Insert" + data.Pascal: true})
		if rmErr != nil {
			return nil, rmErr
		}
		if containsAny(content, symbols) {
			refs = append(refs, testRel)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("read %s: %w", testRel, err)
	}

	skip := make(map[string]bool, len(removed)+2)
	for _, rel := range removed {
		skip[filepath.ToSlash(rel)] = true
	}
	skip[storeRel] = true
	skip[testRel] = true
	rels, err := survivingGoFiles(dir, skip)
	if err != nil {
		return nil, err
	}
	for _, rel := range rels {
		body, readErr := os.ReadFile(filepath.Join(dir, rel))
		if readErr != nil {
			return nil, readErr
		}
		if containsAny(string(body), symbols) {
			refs = append(refs, rel)
		}
	}
	return refs, nil
}

// orphanOptionMethods names the List<X>Options methods declared in store.go
// that no surviving file calls anymore (#108). They are generated when a
// resource references another; destroying the last child must drop them, or a
// later `destroy resource <parent>` would be refused by parentReferences.

func orphanOptionMethods(dir string, removed []string, storeBody string) (map[string]bool, error) {
	names := listOptionMethodNames(storeBody)
	if len(names) == 0 {
		return nil, nil
	}
	skip := make(map[string]bool, len(removed)+1)
	for _, rel := range removed {
		skip[filepath.ToSlash(rel)] = true
	}
	skip["internal/store/store.go"] = true
	rels, err := survivingGoFiles(dir, skip)
	if err != nil {
		return nil, err
	}
	orphans := map[string]bool{}
	for name := range names {
		used := false
		for _, rel := range rels {
			body, readErr := os.ReadFile(filepath.Join(dir, rel))
			if readErr != nil {
				return nil, readErr
			}
			if strings.Contains(string(body), name+"(") {
				used = true
				break
			}
		}
		if !used {
			orphans[name] = true
		}
	}
	return orphans, nil
}

// listOptionMethodNames returns the List<X>Options methods declared in src.

func listOptionMethodNames(src string) map[string]bool {
	_, file, err := parseGo(src, "store.go")
	if err != nil {
		return nil
	}
	names := map[string]bool{}
	for _, decl := range file.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Recv == nil {
			continue
		}
		if strings.HasPrefix(fd.Name.Name, "List") && strings.HasSuffix(fd.Name.Name, "Options") {
			names[fd.Name.Name] = true
		}
	}
	return names
}

// survivingGoFiles lists internal/**/*.go rel paths, skipping the given ones.

func survivingGoFiles(dir string, skip map[string]bool) ([]string, error) {
	var rels []string
	err := filepath.WalkDir(filepath.Join(dir, "internal"), func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".go") {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		relSlash := filepath.ToSlash(rel)
		if !skip[relSlash] {
			rels = append(rels, relSlash)
		}
		return nil
	})
	return rels, err
}

func containsAny(content string, needles []string) bool {
	for _, needle := range needles {
		if strings.Contains(content, needle) {
			return true
		}
	}
	return false
}

// removeStoreResourceMethods removes generated methods and interface entries by
// exact name (#168): a hand-written countBookmarksByUser survives destroying
// resource bookmark even though countBookmarks does not.

func removeStoreResourceMethods(content string, data scaffoldData) (string, error) {
	patterns := []string{
		"Insert" + data.Pascal,
		"Update" + data.Pascal,
		"Delete" + data.Pascal,
		"Find" + data.Pascal + "ByID",
		"ListAll" + data.PluralPascal,
		"List" + data.PluralPascal,
		"SeedDemo" + data.PluralPascal,
		"count" + data.PluralPascal,
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
	content = regexp.MustCompile(`\nfunc strPtr\(s string\) \*string \{ return &s \}\n`).ReplaceAllString(content, "\n")
	content = regexp.MustCompile(`\nfunc int64Ptr\(n int64\) \*int64 \{ return &n \}\n`).ReplaceAllString(content, "\n")
	// boolInt spans multiple lines (nested braces); a regex stops at the inner
	// `}` and leaves `return 0` behind (#105). Remove the whole declaration.
	content, err = removeDeclsByName(content, map[string]bool{"boolInt": true})
	if err != nil {
		return "", err
	}
	return content, nil
}

func cleanupStoreImports(content string) string {
	if !strings.Contains(content, "models.") {
		re := regexp.MustCompile(`(?m)^\s*"[^"]+/internal/models"\s*\n`)
		content = re.ReplaceAllString(content, "")
		content = regexp.MustCompile(`import \(\n\n`).ReplaceAllString(content, "import (\n")
	}
	if !strings.Contains(content, "pagination.") {
		content = strings.Replace(content, "\t\""+frameworkModule+"/pkg/cais/pagination\"\n", "", 1)
	}
	return content
}

func unpatchStoreTestForResource(dir string, data scaffoldData, dryRun bool) error {
	path := filepath.Join(dir, "internal/store/store_test.go")
	body, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read generated file: %w", err)
	}
	content, err := removeDeclsByName(string(body), map[string]bool{"TestStore_Insert" + data.Pascal: true})
	if err != nil {
		return err
	}
	return updateScaffoldFile(path, []byte(content), "internal/store/store_test.go", dryRun)
}

func unpatchSeedsForResource(dir string, data scaffoldData, dryRun bool) error {
	path := filepath.Join(dir, "internal/db/seeds.go")
	body, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read generated file: %w", err)
	}
	block := fmt.Sprintf("\tif err := s.SeedDemo%s(); err != nil {\n\t\treturn err\n\t}\n", data.PluralPascal)
	content := strings.Replace(string(body), block, "", 1)
	return updateScaffoldFile(path, []byte(content), "internal/db/seeds.go", dryRun)
}

func unpatchMainForSeed(dir string, data scaffoldData, dryRun bool) error {
	path := filepath.Join(dir, "cmd/server/main.go")
	body, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read generated file: %w", err)
	}
	block := fmt.Sprintf(`
	if err := s.SeedDemo%s(); err != nil {
		_ = s.Close()
		return nil, fmt.Errorf("seed: %%w", err)
	}
`, data.PluralPascal)
	content := strings.Replace(string(body), block, "", 1)
	return updateScaffoldFile(path, []byte(content), "cmd/server/main.go", dryRun)
}

func unpatchLayoutNavForResource(dir string, data scaffoldData, dryRun bool) error {
	path, inertia := layoutNavFile(dir)
	body, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read layout: %w", err)
	}
	link := publicNavLink(data, inertia)
	content := strings.Replace(string(body), link, "", 1)
	rel := strings.TrimPrefix(path, dir+string(os.PathSeparator))
	return updateScaffoldFile(path, []byte(content), rel, dryRun)
}
