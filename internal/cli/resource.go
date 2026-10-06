// cais g resource orchestration: writes generated files then delegates patches to resource_patch.go.
package cli

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func scaffoldResource(dir, name string, opts resourceOpts) error {
	fields, err := parseFields(opts.Fields)
	if err != nil {
		return err
	}

	data := dataForResource(name)
	data.ModulePath = readModulePath(dir)
	data.AppDir = dir
	data.Fields = fields
	data.Public = opts.Public
	data.Seed = opts.Seed
	data.Paginate = opts.Paginate
	data.AdminAuth = opts.AdminAuth

	if err := validateReferenceParents(dir, data.Fields); err != nil {
		return err
	}

	migrationPath, migrationNum, err := nextMigrationFile(dir, data.Plural, opts.dryRun)
	if err != nil {
		return err
	}
	data.MigrationNum = migrationNum

	files := resourceFilesHTML(data, migrationPath)

	if hasReferenceFields(data.Fields) {
		selectPath := filepath.Join(dir, "internal/models/select_option.go")
		if _, err := os.Stat(selectPath); os.IsNotExist(err) {
			files["internal/models/select_option.go"] = tplSelectOptionModel
		}
	}

	// #130: validate every patch precondition and snapshot shared files before
	// writing, so a failure cannot leave a half-patched app behind.
	if err := preflightResourcePatches(dir, data, opts); err != nil {
		return err
	}
	snapshot, err := snapshotSharedFiles(dir, resourceSharedFiles(dir, opts))
	if err != nil {
		return err
	}
	var created []string
	fail := func(cause error) error {
		if !opts.dryRun {
			for _, rel := range created {
				_ = os.Remove(filepath.Join(dir, rel))
			}
			restoreSharedFiles(dir, snapshot)
		}
		return cause
	}

	for path, content := range files {
		full := filepath.Join(dir, path)
		if _, err := os.Stat(full); err == nil {
			if !opts.Force {
				return fmt.Errorf("%s already exists (use --force to overwrite)", path)
			}
			if opts.dryRun {
				printfScaffold("update", path)
				continue
			}
		} else {
			created = append(created, path)
		}
		if err := writeScaffoldFile(full, []byte(content), 0o644, path, opts.dryRun); err != nil {
			return fail(err)
		}
	}

	if err := patchStoreForResource(dir, data, opts.dryRun, opts.Force); err != nil {
		return fail(err)
	}
	if err := patchStoreTestForResource(dir, data, opts.dryRun); err != nil {
		return fail(err)
	}
	if err := patchRoutesForResource(dir, data, opts.dryRun, opts.Force); err != nil {
		return fail(err)
	}
	var finalErr error
	if data.Seed {
		if err := patchSeedsForResource(dir, data, opts.dryRun); err != nil {
			return fail(err)
		}
		finalErr = patchMainForSeed(dir, data, opts.dryRun)
	} else {
		finalErr = patchLayoutNav(dir, data, opts.dryRun)
	}
	if finalErr != nil {
		return fail(finalErr)
	}
	if opts.dryRun {
		return nil
	}
	if err := gofmtGoFiles(dir); err != nil {
		return fail(err)
	}
	// Record after gofmt: formatting rewrites the generated files (#169).
	rels := make([]string, 0, len(files))
	for rel := range files {
		rels = append(rels, filepath.ToSlash(rel))
	}
	if err := recordGeneratedFiles(dir, rels); err != nil {
		return fmt.Errorf("record generated manifest: %w", err)
	}
	return nil
}

// validateReferenceParents fails before any file is written when a references
// field points at a parent resource that was never generated: seeds and tests
// would call Insert<Parent> and reference a table that does not exist (#107).
func validateReferenceParents(dir string, fields []FieldDef) error {
	for _, f := range fields {
		if f.RefTable == "" {
			continue
		}
		parent := toSnake(f.RefPascal)
		modelPath := filepath.Join(dir, "internal/models", parent+".go")
		switch _, err := os.Stat(modelPath); {
		case err == nil:
			continue
		case errors.Is(err, fs.ErrNotExist):
			return fmt.Errorf(
				"field %q references table %q: internal/models/%s.go not found — run `amarra-cais g resource %s` first",
				f.Name, f.RefTable, parent, parent,
			)
		default:
			return fmt.Errorf("stat parent model %s: %w", modelPath, err)
		}
	}
	return nil
}

// patchMarker names a marker a patch needs in a shared file.
type patchMarker struct {
	rel    string
	marker string
	what   string
}

// preflightResourcePatches fails before any file is written when a shared file
// lacks the marker its patch needs — the previous flow mutated store.go and
// then aborted, leaving a half-patched app (#130).
func preflightResourcePatches(dir string, data scaffoldData, opts resourceOpts) error {
	checks := []patchMarker{
		{rel: "internal/store/store.go", marker: "\n\tClose() error", what: "Store interface"},
		{rel: "internal/store/store.go", marker: "\nfunc (s *SQLiteStore) Close()", what: "store implementation"},
	}
	if opts.Seed {
		checks = append(checks,
			patchMarker{rel: "internal/db/seeds.go", marker: "// cais:seeds", what: "seeds.go"},
			patchMarker{rel: "cmd/server/main.go", marker: "cais.ResolveWebDir", what: "main.go"},
		)
	}
	for _, check := range checks {
		body, err := os.ReadFile(filepath.Join(dir, check.rel))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("read %s: %w", check.rel, err)
		}
		if !strings.Contains(string(body), check.marker) {
			return fmt.Errorf("cannot patch %s: missing %s marker (%q)", check.rel, check.what, check.marker)
		}
	}

	// These patches return early when the resource is already present, so the
	// marker is only required when the file still needs patching.
	routes, err := os.ReadFile(filepath.Join(dir, "internal/app/routes.go"))
	if err != nil {
		return fmt.Errorf("read routes.go: %w", err)
	}
	alreadyRouted := strings.Contains(string(routes), `"/admin/`+data.Plural+`"`) || strings.Contains(string(routes), "`/admin/"+data.Plural+"`")
	if !alreadyRouted && !strings.Contains(string(routes), "func registerRoutes") {
		return fmt.Errorf("cannot patch internal/app/routes.go: missing registerRoutes")
	}

	if opts.Public {
		path, _ := layoutNavFile(dir)
		if body, err := os.ReadFile(path); err == nil {
			if !strings.Contains(string(body), layoutNavMarker) && !strings.Contains(string(body), "</nav>") {
				return fmt.Errorf("cannot patch %s: missing %s marker and </nav> element", path, layoutNavMarker)
			}
		}
	}
	return nil
}

// resourceSharedFiles lists the files generators patch in place.
func resourceSharedFiles(dir string, opts resourceOpts) []string {
	rels := []string{
		"internal/store/store.go",
		"internal/store/store_test.go",
		"internal/app/routes.go",
	}
	if opts.Seed {
		rels = append(rels, "internal/db/seeds.go", "cmd/server/main.go")
	}
	if opts.Public {
		if path, _ := layoutNavFile(dir); path != "" {
			if rel, err := filepath.Rel(dir, path); err == nil {
				rels = append(rels, rel)
			}
		}
	}
	return rels
}

// snapshotSharedFiles keeps the current bytes of existing shared files so a
// failed generation can restore them.
func snapshotSharedFiles(dir string, rels []string) (map[string][]byte, error) {
	snap := map[string][]byte{}
	for _, rel := range rels {
		body, err := os.ReadFile(filepath.Join(dir, rel))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("snapshot %s: %w", rel, err)
		}
		snap[rel] = body
	}
	return snap, nil
}

func restoreSharedFiles(dir string, snap map[string][]byte) {
	for rel, body := range snap {
		_ = os.WriteFile(filepath.Join(dir, rel), body, 0o644)
	}
}

func resourceFilesHTML(data scaffoldData, migrationPath string) map[string]string {
	files := map[string]string{
		filepath.Join("internal/models", data.Snake+".go"):                               buildResourceModel(data),
		filepath.Join("internal/handlers", "admin_"+data.Plural+".go"):                   buildResourceAdminHandler(data),
		filepath.Join("internal/handlers", "admin_"+data.Plural+"_test.go"):              buildResourceAdminTest(data),
		filepath.Join("web/templates/pages", "admin_"+data.Plural+".html"):               buildAdminIndexHTML(data),
		filepath.Join("web/templates/pages", "admin_"+data.Snake+"_show.html"):           buildAdminShowHTML(data),
		filepath.Join("web/templates/pages", "admin_"+data.Snake+"_form.html"):           buildAdminFormHTML(data),
		filepath.Join("web/templates/partials", "admin_"+data.Snake+"_form_errors.html"): buildAdminFormErrorsPartial(data),
		migrationPath: buildResourceMigration(data),
	}
	if data.Paginate {
		files[filepath.Join("web/templates/partials", "admin_"+data.Plural+"_index.html")] = buildAdminIndexPartial(data)
	}
	if data.Public {
		files[filepath.Join("internal/handlers", data.Plural+".go")] = buildResourcePublicHandler(data)
		if hasFieldNamed(data.Fields, "slug") {
			files[filepath.Join("web/templates/pages", data.Snake+".html")] = buildPublicShowHTML(data)
		}
		files[filepath.Join("internal/handlers", data.Plural+"_test.go")] = buildResourcePublicTest(data)
		files[filepath.Join("web/templates/pages", data.Plural+".html")] = buildPublicListHTML(data)
		if data.Paginate {
			files[filepath.Join("web/templates/partials", data.Plural+"_list.html")] = buildPublicListPartial(data)
		}
	}
	return files
}
