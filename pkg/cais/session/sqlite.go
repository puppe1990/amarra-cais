package session

import (
	"database/sql"
	"fmt"
	"time"
)

const sessionTTL = 7 * 24 * time.Hour

const sqliteSchema = `CREATE TABLE IF NOT EXISTS sessions (
  token TEXT PRIMARY KEY NOT NULL,
  user_id INTEGER NOT NULL,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  expires_at TEXT NOT NULL
);`

// EnsureSQLiteSchema creates the sessions table when missing.
func EnsureSQLiteSchema(db *sql.DB) error {
	if _, err := db.Exec(sqliteSchema); err != nil {
		return fmt.Errorf("sessions schema: %w", err)
	}
	return migrateSQLiteSchema(db)
}

// migrateSQLiteSchema preserves each legacy session's original seven-day TTL.
// SQLite forbids computed defaults in ALTER on populated tables (#351).
func migrateSQLiteSchema(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("sessions schema begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := backfillSQLiteExpiry(tx); err != nil {
		return fmt.Errorf("sessions schema migrate: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("sessions schema commit: %w", err)
	}
	return nil
}

func backfillSQLiteExpiry(tx *sql.Tx) error {
	var count int
	if err := tx.QueryRow("SELECT COUNT(*) FROM pragma_table_info('sessions') WHERE name = 'expires_at'").Scan(&count); err != nil {
		return err
	}
	if count != 0 {
		return nil
	}
	if _, err := tx.Exec("ALTER TABLE sessions ADD COLUMN expires_at TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	// Invalid/missing creation dates expire immediately; never extend credentials
	// whose age cannot be established. Keep ALTER and backfill atomic.
	_, err := tx.Exec("UPDATE sessions SET expires_at = COALESCE(datetime(created_at, '+7 days'), datetime('now'))")
	return err
}

// SQLiteStore persists sessions in SQLite.
type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(db *sql.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

func (s *SQLiteStore) Create(userID int64) (string, error) {
	token, err := newToken()
	if err != nil {
		return "", err
	}
	expiresAt := time.Now().UTC().Add(sessionTTL).Format("2006-01-02 15:04:05")
	if _, err := s.db.Exec(
		"INSERT INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?)",
		token, userID, expiresAt,
	); err != nil {
		return "", fmt.Errorf("insert session: %w", err)
	}
	return token, nil
}

func (s *SQLiteStore) Get(token string) (int64, bool) {
	var id int64
	err := s.db.QueryRow(
		"SELECT user_id FROM sessions WHERE token = ? AND expires_at > datetime('now')",
		token,
	).Scan(&id)
	if err != nil {
		return 0, false
	}
	return id, true
}

func (s *SQLiteStore) Delete(token string) error {
	if _, err := s.db.Exec("DELETE FROM sessions WHERE token = ?", token); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func (s *SQLiteStore) PruneExpired() (int64, error) {
	res, err := s.db.Exec("DELETE FROM sessions WHERE expires_at <= datetime('now')")
	if err != nil {
		return 0, fmt.Errorf("prune expired sessions: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("prune expired sessions: %w", err)
	}
	return n, nil
}
