// Patches generated apps when cais g resource runs (store, routes, layout, seeds).
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func patchStoreForResource(dir string, data scaffoldData, dryRun bool, force bool) error {
	path := filepath.Join(dir, "internal/store/store.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(body)
	if strings.Contains(content, "Insert"+data.Pascal) {
		if !force {
			return nil
		}
		var err error
		content, err = removeStoreResourceMethods(content, data)
		if err != nil {
			return err
		}
	}

	ifaceMarker := "\n\tClose() error"
	if !strings.Contains(content, ifaceMarker) {
		return fmt.Errorf("could not patch store interface")
	}
	// The reference-options impl check must run against the pre-patch body:
	// after the interface insert the method name is already present, which
	// used to skip generating the SQLiteStore impl and left the app broken.
	prePatch := content
	listMethod := fmt.Sprintf("\n\tListAll%s(search, sort, dir string) ([]models.%s, error)", data.PluralPascal, data.Pascal)
	if data.Paginate {
		listMethod = fmt.Sprintf(
			"\n\tList%s(search, sort, dir string, page, perPage int) ([]models.%s, int, error)%s",
			data.PluralPascal, data.Pascal, listMethod,
		)
	}
	ifaceInsert := fmt.Sprintf(
		"\n\tInsert%s(models.%s) (int64, error)\n\tUpdate%s(models.%s) error\n\tDelete%s(id int64) error\n\tFind%sByID(id int64) (models.%s, error)%s",
		data.Pascal, data.Pascal,
		data.Pascal, data.Pascal,
		data.Pascal,
		data.Pascal, data.Pascal,
		listMethod,
	)
	if data.Seed {
		ifaceInsert += fmt.Sprintf("\n\tSeedDemo%s() error", data.PluralPascal)
	}
	for _, f := range uniqueReferenceFields(data.Fields) {
		method := fmt.Sprintf("\n\tList%sOptions() ([]models.SelectOption, error)", f.RefPascal)
		if !strings.Contains(content, "List"+f.RefPascal+"Options()") {
			ifaceInsert += method
		}
	}
	content = strings.Replace(content, ifaceMarker, ifaceInsert+ifaceMarker, 1)

	implMarker := "\nfunc (s *SQLiteStore) Close()"
	implInsert := buildResourceStoreMethods(data)
	implInsert += buildReferenceStoreMethods(data.Fields, prePatch)
	if data.Paginate {
		implInsert += buildResourcePaginatedStoreMethod(data)
	}
	if data.Seed {
		implInsert += buildResourceSeed(data)
	}
	if hasBoolField(data.Fields) && !strings.Contains(content, "func boolInt(") {
		implInsert = "\nfunc boolInt(v bool) int {\n\tif v {\n\t\treturn 1\n\t}\n\treturn 0\n}\n" + implInsert
	}
	if data.Paginate && !strings.Contains(content, "pkg/cais/pagination") {
		content = strings.Replace(content,
			`"github.com/puppe1990/amarra-cais/pkg/cais/sqllog"`,
			`"github.com/puppe1990/amarra-cais/pkg/cais/pagination"
	"github.com/puppe1990/amarra-cais/pkg/cais/sqllog"`,
			1,
		)
	}
	content = strings.Replace(content, implMarker, implInsert+implMarker, 1)

	if !strings.Contains(content, data.ModulePath+"/internal/models") {
		content = strings.Replace(content,
			`_ "modernc.org/sqlite"`,
			`_ "modernc.org/sqlite"

	"`+data.ModulePath+`/internal/models"`,
			1,
		)
	}

	return updateScaffoldFile(path, []byte(content), "internal/store/store.go", dryRun)
}

func patchStoreTestForResource(dir string, data scaffoldData, dryRun bool) error {
	path := filepath.Join(dir, "internal/store/store_test.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(body)
	if strings.Contains(content, "TestStore_Insert"+data.Pascal) {
		return nil
	}

	insertArgs, insertSetup := buildInsertTestLiteral(data.Fields)
	insert := fmt.Sprintf(`
func TestStore_Insert%s(t *testing.T) {
	s := newTestStore(t)
%s	id, err := s.Insert%s(models.%s{%s})
	if err != nil {
		t.Fatal(err)
	}
	if id == 0 {
		t.Error("id = 0")
	}
}
`, data.Pascal, insertSetup, data.Pascal, data.Pascal, insertArgs)

	if !strings.Contains(content, data.ModulePath+"/internal/models") {
		content = strings.Replace(content,
			`import "testing"`,
			`import (
	"testing"

	"`+data.ModulePath+`/internal/models"
)`,
			1,
		)
	}
	content = strings.TrimRight(content, "\n") + "\n" + insert
	if dryRun {
		return nil
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

// buildInsertTestLiteral returns the Go statements that seed one parent row
// per required reference field (FK constraints fail otherwise) plus the
// struct literal for the Insert test.
func buildInsertTestLiteral(fields []FieldDef) (literal, setup string) {
	var parts []string
	for _, f := range fields {
		if f.RefTable == "" || !f.Required {
			continue
		}
		varName := strings.ToLower(f.RefPascal) + "ID"
		setup += fmt.Sprintf("%s, err := s.Insert%s(models.%s{Name: \"Sample\"})\n\tif err != nil {\n\t\tt.Fatal(err)\n\t}\n\t", varName, f.RefPascal, f.RefPascal)
		parts = append(parts, f.Pascal+": "+varName)
	}
	for _, f := range fields {
		if f.RefTable != "" || !f.Required {
			continue
		}
		parts = append(parts, f.Pascal+": "+seedValueForField(f))
	}
	if len(parts) == 0 && len(fields) > 0 {
		f := fields[0]
		if f.RefTable != "" {
			varName := strings.ToLower(f.RefPascal) + "ID"
			setup += fmt.Sprintf("%s, err := s.Insert%s(models.%s{Name: \"Sample\"})\n\tif err != nil {\n\t\tt.Fatal(err)\n\t}\n\t", varName, f.RefPascal, f.RefPascal)
			parts = append(parts, f.Pascal+": "+varName)
		} else {
			parts = append(parts, f.Pascal+": "+seedValueForField(f))
		}
	}
	return strings.Join(parts, ", "), setup
}

func patchRoutesForResource(dir string, data scaffoldData, dryRun bool, force bool) error {
	path := filepath.Join(dir, "internal/app/routes.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(body)
	if strings.Contains(content, "/admin/"+data.Plural) {
		if !force {
			return nil
		}
		upgraded := upgradeResourceRouteHandlers(content, data)
		if upgraded == content {
			return nil
		}
		return updateScaffoldFile(path, []byte(upgraded), "internal/app/routes.go", dryRun)
	}

	if !strings.Contains(content, frameworkModule+"/pkg/cais/middleware") {
		content = strings.Replace(content,
			`"`+frameworkModule+`/pkg/cais"`,
			`"`+frameworkModule+`/pkg/cais"
	"`+frameworkModule+`/pkg/cais/middleware"`,
			1,
		)
	}

	adminVar := "admin" + data.PluralPascal
	var insert strings.Builder
	if data.Public {
		pubVar := lowerFirst(data.PluralPascal)
		fmt.Fprintf(&insert, "\t%s := handlers.New%sHandler(deps.Views, deps.Store, deps.Site, cfg)\n", pubVar, data.PluralPascal)
		if firstBoolField(data.Fields) != nil {
			fmt.Fprintf(&insert, "\tr.Post(\"/%s/{id}/toggle\", cais.IntParam(\"id\", %s.Toggle))\n", data.Plural, pubVar)
		}
		fmt.Fprintf(&insert, "\tr.Get(\"/%s\", %s.List)\n", data.Plural, pubVar)
	}
	fmt.Fprintf(&insert, "\t%s := handlers.NewAdmin%sHandler(deps.Views, deps.Store, deps.Site, cfg)\n", adminVar, data.PluralPascal)
	if data.AdminAuth == "bearer" {
		fmt.Fprintf(&insert, "\tr.Group(middleware.AdminAuth(cfg), func(g *cais.Router) {\n")
	} else {
		fmt.Fprintf(&insert, "\tr.Group(middleware.RequireAuth(\"/login\"), func(g *cais.Router) {\n")
	}
	fmt.Fprintf(&insert, "\t\tg.Get(\"/admin/%s\", %s.Index)\n", data.Plural, adminVar)
	fmt.Fprintf(&insert, "\t\tg.Get(\"/admin/%s/{id}\", cais.IntParam(\"id\", %s.Show))\n", data.Plural, adminVar)
	fmt.Fprintf(&insert, "\t\tg.Get(\"/admin/%s/new\", %s.New)\n", data.Plural, adminVar)
	fmt.Fprintf(&insert, "\t\tg.Post(\"/admin/%s\", %s.Create)\n", data.Plural, adminVar)
	fmt.Fprintf(&insert, "\t\tg.Get(\"/admin/%s/{id}/edit\", cais.IntParam(\"id\", %s.Edit))\n", data.Plural, adminVar)
	fmt.Fprintf(&insert, "\t\tg.Post(\"/admin/%s/{id}\", cais.IntParam(\"id\", %s.Update))\n", data.Plural, adminVar)
	fmt.Fprintf(&insert, "\t\tg.Post(\"/admin/%s/{id}/delete\", cais.IntParam(\"id\", %s.Delete))\n", data.Plural, adminVar)
	fmt.Fprintf(&insert, "\t})\n")

	var err2 error
	content, err2 = insertBeforeFunctionEnd(content, "registerRoutes", insert.String())
	if err2 != nil {
		return fmt.Errorf("could not patch routes.go: %w", err2)
	}
	if err := updateScaffoldFile(path, []byte(content), "internal/app/routes.go", dryRun); err != nil {
		return err
	}
	return patchLayoutNav(dir, data, dryRun)
}

// upgradeResourceRouteHandlers rewrites legacy handler constructors when --force
// regenerates handlers but routes.go already exists (pre meta.Site wiring).
func upgradeResourceRouteHandlers(content string, data scaffoldData) string {
	replacements := [][2]string{
		{
			fmt.Sprintf("handlers.NewAdmin%sHandler(deps.Renderer, deps.Store, cfg)", data.PluralPascal),
			fmt.Sprintf("handlers.NewAdmin%sHandler(deps.Views, deps.Store, deps.Site, cfg)", data.PluralPascal),
		},
		{
			fmt.Sprintf("handlers.NewAdmin%sHandler(deps.Renderer, deps.Store, deps.Site, cfg)", data.PluralPascal),
			fmt.Sprintf("handlers.NewAdmin%sHandler(deps.Views, deps.Store, deps.Site, cfg)", data.PluralPascal),
		},
		{
			fmt.Sprintf("handlers.NewAdmin%sHandler(deps.Store, deps.Site, deps.Inertia)", data.PluralPascal),
			fmt.Sprintf("handlers.NewAdmin%sHandler(deps.Views, deps.Store, deps.Site, cfg)", data.PluralPascal),
		},
	}
	if data.Public {
		replacements = append(replacements,
			[2]string{
				fmt.Sprintf("handlers.New%sHandler(deps.Renderer, deps.Store, cfg)", data.PluralPascal),
				fmt.Sprintf("handlers.New%sHandler(deps.Views, deps.Store, deps.Site, cfg)", data.PluralPascal),
			},
			[2]string{
				fmt.Sprintf("handlers.New%sHandler(deps.Renderer, deps.Store, deps.Site, cfg)", data.PluralPascal),
				fmt.Sprintf("handlers.New%sHandler(deps.Views, deps.Store, deps.Site, cfg)", data.PluralPascal),
			},
			[2]string{
				fmt.Sprintf("handlers.New%sHandler(deps.Store, deps.Site, deps.Inertia)", data.PluralPascal),
				fmt.Sprintf("handlers.New%sHandler(deps.Views, deps.Store, deps.Site, cfg)", data.PluralPascal),
			},
		)
	}
	for _, pair := range replacements {
		content = strings.ReplaceAll(content, pair[0], pair[1])
	}
	return content
}

// layoutNavMarker is embedded in scaffold layouts (tplLayout*). Generators insert
// public resource links after it; destroy removes links by href. Do not remove from templates.
const layoutNavMarker = "<!-- cais:nav -->"

func layoutNavFile(dir string) (path string, inertia bool) {
	app := filepath.Join(dir, "web/templates/layouts/app.html")
	if _, err := os.Stat(app); err == nil {
		return app, false
	}
	return filepath.Join(dir, "web/templates/layouts/base.html"), false
}

func publicNavLink(data scaffoldData, inertia bool) string {
	_ = inertia
	return fmt.Sprintf(`          <a href="/%s" class="px-3 py-2 font-mono text-[10px] uppercase tracking-[0.22em] text-foam/50 hover:text-foam flex-shrink-0">%s</a>
`, data.Plural, toTitle(data.Plural))
}

func patchLayoutNav(dir string, data scaffoldData, dryRun bool) error {
	if !data.Public {
		return nil
	}
	path, inertia := layoutNavFile(dir)
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(body)
	linkHref := `href="/` + data.Plural + `"`
	if strings.Contains(content, linkHref) {
		return nil
	}
	link := publicNavLink(data, inertia)
	switch {
	case strings.Contains(content, layoutNavMarker):
		content = strings.Replace(content, layoutNavMarker, layoutNavMarker+"\n"+link, 1)
	case strings.Contains(content, "</nav>"):
		if inertia {
			content = strings.Replace(content, "</nav>", link+"  </nav>", 1)
		} else {
			content = strings.Replace(content, "</nav>", link+"        </nav>", 1)
		}
	default:
		return fmt.Errorf("%s: missing %s marker and </nav> element", path, layoutNavMarker)
	}
	if !inertia {
		content = patchLayoutLogoHref(dir, content, data)
	}
	if dryRun {
		return nil
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func patchLayoutLogoHref(dir, content string, data scaffoldData) string {
	routes, err := os.ReadFile(filepath.Join(dir, "internal/app/routes.go"))
	if err != nil {
		return content
	}
	if strings.Contains(string(routes), `r.Get("/", home`) {
		return content
	}
	return strings.Replace(content,
		`<a href="/" class="font-bold`,
		fmt.Sprintf(`<a href="/%s" class="font-bold`, data.Plural),
		1,
	)
}

func patchSeedsForResource(dir string, data scaffoldData, dryRun bool) error {
	path := filepath.Join(dir, "internal/db/seeds.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(body)
	if strings.Contains(content, "SeedDemo"+data.PluralPascal) {
		return nil
	}
	marker := "\t// cais:seeds\n"
	if !strings.Contains(content, marker) {
		return fmt.Errorf("could not patch seeds.go — missing cais:seeds marker")
	}
	insert := fmt.Sprintf(`%s	if err := s.SeedDemo%s(); err != nil {
		return err
	}
`, marker, data.PluralPascal)
	content = strings.Replace(content, marker, insert, 1)
	return updateScaffoldFile(path, []byte(content), "internal/db/seeds.go", dryRun)
}

func patchMainForSeed(dir string, data scaffoldData, dryRun bool) error {
	path := filepath.Join(dir, "cmd/server/main.go")
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(body)
	if strings.Contains(content, "SeedDemo"+data.PluralPascal) {
		return patchLayoutNav(dir, data, dryRun)
	}
	marker := "\n\tstaticDir, err := cais.ResolveWebDir(\"static\", cfg.StaticDir)"
	seed := fmt.Sprintf(`
	if err := s.SeedDemo%s(); err != nil {
		_ = s.Close()
		return nil, fmt.Errorf("seed: %%w", err)
	}
`, data.PluralPascal)
	if !strings.Contains(content, marker) {
		return fmt.Errorf("could not patch main.go for seed")
	}
	content = strings.Replace(content, marker, seed+marker, 1)
	if err := updateScaffoldFile(path, []byte(content), "cmd/server/main.go", dryRun); err != nil {
		return err
	}
	return patchLayoutNav(dir, data, dryRun)
}
