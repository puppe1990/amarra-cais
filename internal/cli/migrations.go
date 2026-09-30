package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func nextMigrationFile(dir, slug string, dryRun bool) (relPath, num string, err error) {
	migrationsDir := filepath.Join(dir, "internal/store/migrations")
	if !dryRun {
		if err := os.MkdirAll(migrationsDir, 0o755); err != nil {
			return "", "", err
		}
	}
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return "", "", err
	}
	maxNum := 0
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			var n int
			if _, scanErr := fmt.Sscanf(e.Name(), "%03d_", &n); scanErr == nil && n > maxNum {
				maxNum = n
			}
		}
	}
	num = fmt.Sprintf("%03d", maxNum+1)
	relPath = filepath.Join("internal/store/migrations", num+"_"+slug+".sql")
	return relPath, num, nil
}

// ensureMigrationPlaceholder writes the scaffold's init migration back when a
// destroy emptied the migrations dir: migrations.go embeds
// //go:embed migrations/*.sql, which fails to match on an empty dir (#254, #74).
func ensureMigrationPlaceholder(dir string, dryRun bool) error {
	root := filepath.Join(dir, "internal/store/migrations")
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read migrations dir: %w", err)
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			return nil
		}
	}
	rel := filepath.Join("internal/store/migrations", "001_init.sql")
	if dryRun {
		printfScaffold("create", rel)
		return nil
	}
	return writeScaffoldFile(filepath.Join(dir, rel), []byte(tplMigrationInit), 0o644, rel, false)
}
