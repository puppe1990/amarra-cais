package cli

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// #245: destroy model deleted the model file and migration but left the store
// methods, the RunSeeds insert and the model's tests pointing at the missing
// type. `go build` broke with no hint about the command that caused it, so the
// removal now covers every own reference and refuses to run while another file
// (a handler, say) still calls a method it would delete.

// resourceStoreMethodNames lists the store/interface methods `amarra-cais g
// resource` generates for a model.
func resourceStoreMethodNames(data scaffoldData) map[string]bool {
	return nameSet(
		"Insert"+data.Pascal,
		"Update"+data.Pascal,
		"Delete"+data.Pascal,
		"Find"+data.Pascal+"ByID",
		"ListAll"+data.PluralPascal,
		"List"+data.PluralPascal,
		"SeedDemo"+data.PluralPascal,
		"count"+data.PluralPascal,
	)
}

// modelStoreMethodNames adds the names the scaffold's built-in contact model
// uses — FindContact/CountContacts instead of FindContactByID/countContacts —
// plus List<Model>Options when the model is a reference target.
func modelStoreMethodNames(data scaffoldData) map[string]bool {
	names := resourceStoreMethodNames(data)
	names["Find"+data.Pascal] = true
	names["Count"+data.PluralPascal] = true
	names["List"+data.Pascal+"Options"] = true
	return names
}

func nameSet(names ...string) map[string]bool {
	out := make(map[string]bool, len(names))
	for _, name := range names {
		out[name] = true
	}
	return out
}

// modelRef is a surviving reference to a symbol `destroy model` removes.
type modelRef struct {
	file string
	// symbol is either the store method name or "models.<Pascal>".
	symbol string
	call   bool
}

// modelDestroySkip lists the files destroy model rewrites itself: their
// references are cleaned, not counted as external (#245).
func modelDestroySkip(removed []string) map[string]bool {
	skip := map[string]bool{
		"internal/store/store.go":      true,
		"internal/store/store_test.go": true,
		"internal/db/seeds.go":         true,
		"cmd/server/main.go":           true,
	}
	for _, rel := range removed {
		skip[filepath.ToSlash(rel)] = true
	}
	return skip
}

// externalModelRefs lists the go files (outside the ones destroy model
// rewrites) that still use the model or one of its store methods.
func externalModelRefs(dir, pascal string, methodNames, skip map[string]bool) ([]modelRef, error) {
	rels, err := goFilesUnder(dir, []string{"internal", "cmd"}, skip)
	if err != nil {
		return nil, err
	}
	var refs []modelRef
	for _, rel := range rels {
		body, readErr := os.ReadFile(filepath.Join(dir, rel))
		if readErr != nil {
			return nil, readErr
		}
		symbol, call := modelSymbolIn(string(body), pascal, methodNames)
		if symbol != "" {
			refs = append(refs, modelRef{file: rel, symbol: symbol, call: call})
		}
	}
	return refs, nil
}

// modelStillReferenced names the file that made destroy model refuse, with the
// command that clears it (#245).
func modelStillReferenced(data scaffoldData, ref modelRef) error {
	verb := "calls"
	if !ref.call {
		verb = "references"
	}
	if handler, ok := handlerNameForRef(ref.file); ok {
		return fmt.Errorf(
			"cannot destroy model %s: %s still %s %s — run: amarra-cais destroy handler %s",
			data.Snake, ref.file, verb, ref.symbol, handler,
		)
	}
	return fmt.Errorf(
		"cannot destroy model %s: %s still %s %s — remove the call first, or pass --force",
		data.Snake, ref.file, verb, ref.symbol,
	)
}

// handlerNameForRef maps internal/handlers/<name>.go (or its _test.go) to the
// `destroy handler` target. admin_* handlers belong to a generated resource, so
// they get the generic hint instead of a destroy command that would not work.
func handlerNameForRef(rel string) (string, bool) {
	rel = strings.TrimSuffix(rel, "_test.go")
	name, ok := strings.CutPrefix(rel, "internal/handlers/")
	if !ok {
		return "", false
	}
	name = strings.TrimSuffix(name, ".go")
	if name == "" || strings.Contains(name, "/") || strings.HasPrefix(name, "admin_") {
		return "", false
	}
	return name, true
}

// modelSymbolIn returns the first symbol in src that only exists because of the
// model: a models.<Pascal> selector or one of its store methods. call reports
// whether that symbol was used as a method rather than as the bare type.
func modelSymbolIn(src, pascal string, methodNames map[string]bool) (string, bool) {
	_, file, err := parseGo(src, "scan.go")
	if err != nil {
		return "", false
	}
	return modelSymbolInNode(file, pascal, methodNames)
}

func modelSymbolInNode(n ast.Node, pascal string, methodNames map[string]bool) (string, bool) {
	symbol := ""
	call := false
	ast.Inspect(n, func(node ast.Node) bool {
		if symbol != "" {
			return false
		}
		sel, ok := node.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if id, ok := sel.X.(*ast.Ident); ok && id.Name == "models" && sel.Sel.Name == pascal {
			symbol = "models." + pascal
			return false
		}
		if methodNames[sel.Sel.Name] {
			symbol = sel.Sel.Name
			call = true
			return false
		}
		return true
	})
	return symbol, call
}

func unpatchStoreForModel(dir string, data scaffoldData, methodNames map[string]bool, dryRun bool) error {
	path := filepath.Join(dir, "internal/store/store.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content, err := removeStoreMethodsNamed(string(body), methodNames)
	if err != nil {
		return err
	}
	return updateScaffoldFile(path, []byte(content), "internal/store/store.go", dryRun)
}

// unpatchStoreTestForModel drops the tests that exercise the model's store
// methods: they cannot compile once the methods and the type are gone (#245).
func unpatchStoreTestForModel(dir string, data scaffoldData, methodNames map[string]bool, dryRun bool) error {
	path := filepath.Join(dir, "internal/store/store_test.go")
	body, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read generated file: %w", err)
	}
	names, err := modelTestDecls(string(body), data, methodNames)
	if err != nil {
		return err
	}
	if len(names) == 0 {
		return nil
	}
	content, err := removeDeclsByName(string(body), names)
	if err != nil {
		return err
	}
	content = dropModelsImport(content)
	return updateScaffoldFile(path, []byte(content), "internal/store/store_test.go", dryRun)
}

// modelTestDecls names the top-level funcs of store_test.go that use the model
// type, one of its store methods, or — for the generated TestStore_* schema
// checks — the table its migration creates.
func modelTestDecls(src string, data scaffoldData, methodNames map[string]bool) (map[string]bool, error) {
	_, file, err := parseGo(src, "store_test.go")
	if err != nil {
		return nil, err
	}
	tableRef := tableRefRegexp(data.Plural)
	names := map[string]bool{}
	for _, decl := range file.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if symbol, _ := modelSymbolInNode(fd, data.Pascal, methodNames); symbol != "" {
			names[fd.Name.Name] = true
			continue
		}
		if strings.HasPrefix(fd.Name.Name, "TestStore_") && nodeMatchesString(fd, tableRef) {
			names[fd.Name.Name] = true
		}
	}
	return names, nil
}

// tableRefRegexp matches the table as a SQL identifier — quoted
// (`name='contacts'`) or right after a keyword (`FROM contacts`) — never inside
// a longer word like contacts_log.
func tableRefRegexp(table string) *regexp.Regexp {
	quoted := regexp.QuoteMeta(table)
	return regexp.MustCompile(`(?i)(?:['"` + "`" + `]` + quoted + `['"` + "`" + `]|\b(?:from|into|update|join|table)\s+` + quoted + `\b)`)
}

func nodeMatchesString(n ast.Node, re *regexp.Regexp) bool {
	found := false
	ast.Inspect(n, func(node ast.Node) bool {
		lit, ok := node.(*ast.BasicLit)
		if ok && lit.Kind == token.STRING && re.MatchString(strings.Trim(lit.Value, "`\"")) {
			found = true
			return false
		}
		return true
	})
	return found
}

// unpatchSeedsForModel drops the RunSeeds statements that insert the model's
// demo row. The block match used by resources misses the scaffold's
// `if _, err := s.InsertContact(models.Contact{...})` shape (#245).
func unpatchSeedsForModel(dir string, data scaffoldData, methodNames map[string]bool, dryRun bool) error {
	path := filepath.Join(dir, "internal/db/seeds.go")
	body, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read generated file: %w", err)
	}
	dropped := false
	content, err := removeStmtsFromFunc(string(body), "RunSeeds", func(st ast.Stmt) bool {
		symbol, _ := modelSymbolInNode(st, data.Pascal, methodNames)
		if symbol == "" {
			return false
		}
		dropped = true
		return true
	})
	if err != nil {
		return err
	}
	if !dropped {
		return nil
	}
	content = dropModelsImport(content)
	return updateScaffoldFile(path, []byte(content), "internal/db/seeds.go", dryRun)
}
