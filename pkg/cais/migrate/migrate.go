package migrate

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"sort"
	"strings"
	"time"
)

// schema_migrations records applied migration filenames — ApplyDir is safe to call on every boot.
const schemaTable = `CREATE TABLE IF NOT EXISTS schema_migrations (
  version TEXT PRIMARY KEY NOT NULL,
  applied_at TEXT NOT NULL DEFAULT (datetime('now'))
);`

// Entry describes a migration file and whether it has been applied.
type Entry struct {
	Version string
	Applied bool
}

// ApplyDir runs pending SQL migrations from a filesystem directory.
func ApplyDir(db *sql.DB, dir string) error {
	return Apply(db, os.DirFS(dir), ".")
}

// StatusDir returns migration status for SQL files in a filesystem directory.
func StatusDir(db *sql.DB, dir string) ([]Entry, error) {
	return Status(db, os.DirFS(dir), ".")
}

// RollbackResult describes the outcome of rolling back a migration.
type RollbackResult struct {
	Version    string
	RanDownSQL bool
}

// RollbackLastDir removes the last applied migration record from a filesystem directory.
func RollbackLastDir(db *sql.DB, dir string) (RollbackResult, error) {
	return RollbackLast(db, os.DirFS(dir), ".")
}

// Apply runs pending SQL migrations from dir inside migrations in sorted order.
func Apply(db *sql.DB, migrations fs.FS, dir string) error {
	if err := ensureSchemaTable(db); err != nil {
		return err
	}

	files, err := listSQL(migrations, dir)
	if err != nil {
		return err
	}

	for _, name := range files {
		version := strings.TrimSuffix(name, ".sql")
		applied, err := isApplied(db, version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		sqlPath := path.Join(dir, name)
		sqlBytes, err := fs.ReadFile(migrations, sqlPath)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		upSQL, _ := parseMigrationSQL(string(sqlBytes))
		if err := applyMigration(db, version, upSQL); err != nil {
			if errors.Is(err, errMigrationClaimed) {
				continue
			}
			return err
		}
	}

	return nil
}

// errMigrationClaimed means another process applied the migration first (#123).
var errMigrationClaimed = errors.New("migration already applied")

// ensureSchemaTable creates schema_migrations, retrying SQLITE_BUSY: several
// processes booting together on a fresh file fight over the WAL switch and the
// first DDL, and the loser used to fail before any migration ran (#123).
func ensureSchemaTable(db *sql.DB) error {
	var err error
	for attempt := 0; attempt < schemaTableRetries; attempt++ {
		if _, err = db.Exec(schemaTable); err == nil {
			return nil
		}
		if !isSQLiteBusy(err) {
			break
		}
		time.Sleep(schemaTableRetryDelay)
	}
	return fmt.Errorf("ensure schema_migrations: %w", err)
}

const (
	schemaTableRetries    = 50
	schemaTableRetryDelay = 100 * time.Millisecond
)

// isSQLiteBusy matches the driver's lock errors. modernc exposes no typed code
// through database/sql, and retrying a busy lock is the only recovery.
func isSQLiteBusy(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "SQLITE_BUSY") || strings.Contains(msg, "database is locked")
}

// Status returns migration files in order with applied state.
func Status(db *sql.DB, migrations fs.FS, dir string) ([]Entry, error) {
	if err := ensureSchemaTable(db); err != nil {
		return nil, err
	}

	files, err := listSQL(migrations, dir)
	if err != nil {
		return nil, err
	}

	var entries []Entry
	for _, name := range files {
		version := strings.TrimSuffix(name, ".sql")
		applied, err := isApplied(db, version)
		if err != nil {
			return nil, err
		}
		entries = append(entries, Entry{Version: version, Applied: applied})
	}
	return entries, nil
}

// RollbackLast rolls back the last applied migration. When the migration file
// contains a -- down section, that SQL is executed before removing the
// schema_migrations record. Without a down section, only the record is removed.
func RollbackLast(db *sql.DB, migrations fs.FS, dir string) (RollbackResult, error) {
	entries, err := Status(db, migrations, dir)
	if err != nil {
		return RollbackResult{}, err
	}

	var lastApplied string
	for _, e := range entries {
		if e.Applied {
			lastApplied = e.Version
		}
	}
	if lastApplied == "" {
		return RollbackResult{}, fmt.Errorf("no applied migrations to roll back")
	}

	sqlPath := path.Join(dir, lastApplied+".sql")
	sqlBytes, err := fs.ReadFile(migrations, sqlPath)
	if err != nil {
		return RollbackResult{}, fmt.Errorf("read migration %s: %w", lastApplied, err)
	}
	_, downSQL := parseMigrationSQL(string(sqlBytes))
	ranDown := downSQL != ""

	tx, err := db.Begin()
	if err != nil {
		return RollbackResult{}, fmt.Errorf("begin rollback %s: %w", lastApplied, err)
	}

	if ranDown {
		if _, err := tx.Exec(downSQL); err != nil {
			_ = tx.Rollback()
			return RollbackResult{}, fmt.Errorf("rollback migration %s: %w", lastApplied, err)
		}
	}

	if _, err := tx.Exec("DELETE FROM schema_migrations WHERE version = ?", lastApplied); err != nil {
		_ = tx.Rollback()
		return RollbackResult{}, fmt.Errorf("remove migration %s: %w", lastApplied, err)
	}
	if err := tx.Commit(); err != nil {
		return RollbackResult{}, fmt.Errorf("commit rollback %s: %w", lastApplied, err)
	}

	return RollbackResult{Version: lastApplied, RanDownSQL: ranDown}, nil
}

func listSQL(migrations fs.FS, dir string) ([]string, error) {
	entries, err := fs.ReadDir(migrations, dir)
	if err != nil {
		return nil, fmt.Errorf("read migrations dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)
	return files, nil
}

const (
	markerUp   = "-- up"
	markerDown = "-- down"
)

// parseMigrationSQL splits a migration file into up and down SQL sections.
// When no -- up or -- down markers are present, the entire file is treated as up SQL.
func parseMigrationSQL(content string) (up, down string) {
	lines := strings.Split(content, "\n")

	hasUp, hasDown := false, false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == markerUp {
			hasUp = true
		}
		if trimmed == markerDown {
			hasDown = true
		}
	}
	if !hasUp && !hasDown {
		return strings.TrimSpace(content), ""
	}

	section := 0 // 0=before markers, 1=up, 2=down
	var upLines, downLines, preUpLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch trimmed {
		case markerUp:
			section = 1
			continue
		case markerDown:
			section = 2
			continue
		}

		switch section {
		case 0:
			preUpLines = append(preUpLines, line)
		case 1:
			upLines = append(upLines, line)
		case 2:
			downLines = append(downLines, line)
		}
	}

	if hasUp {
		up = strings.TrimSpace(strings.Join(upLines, "\n"))
	} else {
		up = strings.TrimSpace(strings.Join(preUpLines, "\n"))
	}
	down = strings.TrimSpace(strings.Join(downLines, "\n"))
	return up, down
}

func applyMigration(db *sql.DB, version, sql string) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", version, err)
	}

	// Claim the version first: the INSERT takes the write lock, so a second
	// process waits (busy_timeout) and then sees 0 rows instead of running the
	// DDL twice (#123). A failure before COMMIT rolls the claim back.
	res, err := tx.Exec("INSERT OR IGNORE INTO schema_migrations (version) VALUES (?)", version)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("claim migration %s: %w", version, err)
	}
	claimed, err := res.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("claim migration %s: %w", version, err)
	}
	if claimed == 0 {
		_ = tx.Rollback()
		return errMigrationClaimed
	}

	if _, err := tx.Exec(sql); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("apply migration %s: %w", version, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s: %w", version, err)
	}
	return nil
}

func isApplied(db *sql.DB, version string) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", version).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check migration %s: %w", version, err)
	}
	return count > 0, nil
}
