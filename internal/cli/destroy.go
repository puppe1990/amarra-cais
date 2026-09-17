package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// removeGeneratedFiles deletes rel paths under dir. Files whose content no
// longer matches the generation manifest are skipped with a warning unless
// force is set (#169). Deleted entries leave the manifest.

func removeGeneratedFiles(dir string, rels []string, dryRun, force bool) error {
	var removed []string
	for _, rel := range rels {
		if err := ensureRelPath(rel); err != nil {
			return err
		}
		full := filepath.Join(dir, rel)
		if _, err := os.Stat(full); err != nil {
			continue
		}
		relSlash := filepath.ToSlash(rel)
		if !force {
			if !manifestHas(dir, relSlash) {
				printfScaffold("skip", rel+" (not recorded as generated; use --force to remove)")
				continue
			}
			if fileDiffersFromManifest(dir, relSlash) {
				printfScaffold("skip", rel+" (modified since generation; use --force to remove)")
				continue
			}
		}
		if dryRun {
			printfScaffold("remove", rel)
			continue
		}
		if err := os.Remove(full); err != nil {
			return fmt.Errorf("remove %s: %w", rel, err)
		}
		removed = append(removed, filepath.ToSlash(rel))
	}
	if len(removed) > 0 && !dryRun {
		if err := dropManifestEntries(dir, removed); err != nil {
			return fmt.Errorf("update manifest: %w", err)
		}
	}
	return nil
}

func destroyResource(dir, name string, dryRun, force bool) error {
	data := dataForResource(name)

	files := []string{
		filepath.Join("internal/models", data.Snake+".go"),
		filepath.Join("internal/handlers", "admin_"+data.Plural+".go"),
		filepath.Join("internal/handlers", "admin_"+data.Plural+"_test.go"),
		filepath.Join("web/templates/pages", "admin_"+data.Plural+".html"),
		filepath.Join("web/templates/pages", "admin_"+data.Snake+"_show.html"),
		filepath.Join("web/templates/pages", "admin_"+data.Snake+"_form.html"),
		filepath.Join("web/templates/partials", "admin_"+data.Snake+"_form_errors.html"),
		filepath.Join("web/templates/partials", "admin_"+data.Plural+"_index.html"),
		filepath.Join("web/src/pages", "Admin"+data.PluralPascal+".svelte"),
		filepath.Join("web/src/pages", "Admin"+data.Pascal+"Form.svelte"),
		filepath.Join("web/src/pages", "Admin"+data.Pascal+"Show.svelte"),
		filepath.Join("internal/handlers", data.Plural+".go"),
		filepath.Join("internal/handlers", data.Plural+"_test.go"),
		filepath.Join("web/templates/pages", data.Plural+".html"),
		filepath.Join("web/templates/partials", data.Plural+"_toggle.html"),
		filepath.Join("web/templates/partials", data.Plural+"_list.html"),
		filepath.Join("web/src/pages", data.PluralPascal+".svelte"),
	}

	migrationsDir := filepath.Join(dir, "internal/store/migrations")
	entries, err := os.ReadDir(migrationsDir)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read migrations dir: %w", err)
	}
	for _, e := range entries {
		if !e.IsDir() && strings.Contains(e.Name(), "_"+data.Plural+".sql") {
			files = append(files, filepath.Join("internal/store/migrations", e.Name()))
		}
	}

	if !force {
		refs, err := parentReferences(dir, data, files)
		if err != nil {
			return err
		}
		if len(refs) > 0 {
			return fmt.Errorf(
				"cannot destroy %s: still referenced by %s — destroy the referencing resource(s) first, or pass --force",
				data.Snake, strings.Join(refs, ", "),
			)
		}
	}

	if err := removeGeneratedFiles(dir, files, dryRun, force); err != nil {
		return err
	}

	if err := unpatchRoutesForResource(dir, data, dryRun); err != nil {
		return err
	}
	if err := unpatchStoreForResource(dir, data, files, dryRun); err != nil {
		return err
	}
	if err := unpatchStoreTestForResource(dir, data, dryRun); err != nil {
		return err
	}
	if err := unpatchSeedsForResource(dir, data, dryRun); err != nil {
		return err
	}
	if err := unpatchMainForSeed(dir, data, dryRun); err != nil {
		return err
	}
	return unpatchLayoutNavForResource(dir, data, dryRun)
}

func destroyModel(dir, name string, dryRun, force bool) error {
	data := dataForResource(name)

	files := []string{
		filepath.Join("internal/models", data.Snake+".go"),
	}

	migrationsDir := filepath.Join(dir, "internal/store/migrations")
	entries, err := os.ReadDir(migrationsDir)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read migrations dir: %w", err)
	}
	for _, e := range entries {
		if !e.IsDir() && strings.Contains(e.Name(), "_"+data.Plural+".sql") {
			files = append(files, filepath.Join("internal/store/migrations", e.Name()))
		}
	}

	if err := removeGeneratedFiles(dir, files, dryRun, force); err != nil {
		return err
	}
	return unpatchStoreForResource(dir, data, files, dryRun)
}

func destroyHandler(dir, name string, dryRun, force bool) error {
	data := dataForHandler(name)
	files := []string{
		filepath.Join("internal/handlers", data.Snake+".go"),
		filepath.Join("internal/handlers", data.Snake+"_test.go"),
		filepath.Join("web/templates/pages", data.Snake+".html"),
		filepath.Join("web/src/pages", data.Pascal+".svelte"),
	}
	if err := removeGeneratedFiles(dir, files, dryRun, force); err != nil {
		return err
	}
	return unpatchRoutesForHandler(dir, data, dryRun)
}

func (c *CLI) cmdDestroy(args []string) error {
	dryRun := false
	force := false
	filtered := make([]string, 0, len(args))
	for _, arg := range args {
		if arg == "--dry-run" {
			dryRun = true
			continue
		}
		if arg == "--force" {
			force = true
			continue
		}
		filtered = append(filtered, arg)
	}
	args = filtered
	setScaffoldOut(c.Out)

	if len(args) < 1 {
		return fmt.Errorf("usage: amarra-cais destroy [--dry-run] <resource|handler|model|component|auth|migration> [name]")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	if !isCaisApp(cwd) {
		return fmt.Errorf("not a Cais app")
	}

	kind := args[0]
	var genErr error
	switch kind {
	case "resource", "handler", "model", "migration", "component":
		if len(args) < 2 {
			return fmt.Errorf("usage: amarra-cais destroy [--dry-run] %s <name>", kind)
		}
		name := args[1]
		if err := validateGeneratedName(kind, name); err != nil {
			return err
		}
		switch kind {
		case "resource":
			genErr = destroyResource(cwd, name, dryRun, force)
		case "handler":
			genErr = destroyHandler(cwd, name, dryRun, force)
		case "model":
			genErr = destroyModel(cwd, name, dryRun, force)
		case "migration":
			genErr = destroyMigration(cwd, name, dryRun)
		case "component":
			genErr = destroyComponent(cwd, name, dryRun, force)
		}
		if genErr != nil {
			return genErr
		}
		if !dryRun {
			_, _ = fmt.Fprintf(c.Out, "=> Removed %s %s\n", kind, name)
		}
	case "auth":
		if len(args) > 1 {
			return fmt.Errorf("usage: amarra-cais destroy [--dry-run] auth")
		}
		genErr = destroyAuth(cwd, dryRun)
		if genErr != nil {
			return genErr
		}
		if !dryRun {
			_, _ = fmt.Fprintln(c.Out, "=> Removed auth")
		}
	default:
		return fmt.Errorf("unknown destroy target %q (use resource, handler, model, component, auth, or migration)", kind)
	}
	return nil
}
