// Auth scaffold templates for cais g auth and cais new (full app).
package cli

const tplUserModel = `package models

import "time"

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}
`

const tplMigration002Auth = `CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sessions (
    token TEXT PRIMARY KEY NOT NULL,
    user_id INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME NOT NULL DEFAULT (datetime('now', '+7 days'))
);

CREATE TABLE IF NOT EXISTS password_reset_tokens (
    token_hash TEXT PRIMARY KEY NOT NULL,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`

const tplStorePasswordReset = `package store

import (
	"fmt"
	"time"

	"github.com/puppe1990/amarra-cais/pkg/cais/passwordreset"
)

// Reset tokens are bearer credentials: persist only passwordreset.Hash(token)
// and return the raw value to the notifier, so a DB, backup or SQL-log leak
// cannot be replayed (#222). Lookup/consume hash the presented token first.
func (s *SQLiteStore) CreatePasswordResetToken(userID int64) (string, error) {
	if err := s.ClearPasswordResetTokens(userID); err != nil {
		return "", err
	}

	token, err := passwordreset.NewToken()
	if err != nil {
		return "", err
	}
	expiresAt := time.Now().UTC().Add(passwordreset.DefaultTTL).Format("2006-01-02 15:04:05")
	if _, err := s.db.Exec(
		"INSERT INTO password_reset_tokens (token_hash, user_id, expires_at) VALUES (?, ?, ?)",
		passwordreset.Hash(token), userID, expiresAt,
	); err != nil {
		return "", fmt.Errorf("insert reset token: %w", err)
	}
	return token, nil
}

func (s *SQLiteStore) ClearPasswordResetTokens(userID int64) error {
	if _, err := s.db.Exec("DELETE FROM password_reset_tokens WHERE user_id = ?", userID); err != nil {
		return fmt.Errorf("clear reset tokens: %w", err)
	}
	return nil
}

func (s *SQLiteStore) FindPasswordResetUserID(token string) (int64, bool) {
	var userID int64
	err := s.db.QueryRow(
		"SELECT user_id FROM password_reset_tokens WHERE token_hash = ? AND expires_at > datetime('now')",
		passwordreset.Hash(token),
	).Scan(&userID)
	if err != nil {
		return 0, false
	}
	return userID, true
}

func (s *SQLiteStore) ResetPasswordWithToken(token, passwordHash string) error {
	userID, ok := s.FindPasswordResetUserID(token)
	if !ok {
		return fmt.Errorf("invalid or expired reset token")
	}

	tx, err := s.db.Raw().Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec("UPDATE users SET password_hash = ? WHERE id = ?", passwordHash, userID); err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	if _, err := tx.Exec("DELETE FROM password_reset_tokens WHERE token_hash = ?", passwordreset.Hash(token)); err != nil {
		return fmt.Errorf("delete reset token: %w", err)
	}
	if _, err := tx.Exec("DELETE FROM sessions WHERE user_id = ?", userID); err != nil {
		return fmt.Errorf("revoke sessions: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit reset: %w", err)
	}
	return nil
}
`

const tplStorePasswordResetTest = `package store

import (
	"testing"
	"time"

	"github.com/puppe1990/amarra-cais/pkg/cais/passwordreset"
	"github.com/puppe1990/amarra-cais/pkg/cais/session"
)

func newResetTestStore(t *testing.T) *SQLiteStore {
	t.Helper()
	s, err := NewSQLiteStore(":memory:", "development")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestCreatePasswordResetToken_storesDigestNotRawToken(t *testing.T) {
	s := newResetTestStore(t)
	user, err := s.FindUserByEmail("demo@example.com")
	if err != nil {
		t.Fatal(err)
	}

	token, err := s.CreatePasswordResetToken(user.ID)
	if err != nil {
		t.Fatal(err)
	}

	var stored string
	if err := s.db.QueryRow(
		"SELECT token_hash FROM password_reset_tokens WHERE user_id = ?", user.ID,
	).Scan(&stored); err != nil {
		t.Fatal(err)
	}
	if stored == token {
		t.Fatal("stored reset token must not equal the raw bearer token")
	}
	if stored != passwordreset.Hash(token) {
		t.Fatalf("stored = %q, want passwordreset.Hash(token)", stored)
	}
}

func TestFindPasswordResetUserID_ignoresExpired(t *testing.T) {
	s := newResetTestStore(t)
	user, err := s.FindUserByEmail("demo@example.com")
	if err != nil {
		t.Fatal(err)
	}

	expired := "expired-token"
	if _, err := s.db.Exec(
		"INSERT INTO password_reset_tokens (token_hash, user_id, expires_at) VALUES (?, ?, ?)",
		passwordreset.Hash(expired), user.ID, time.Now().UTC().Add(-time.Minute).Format("2006-01-02 15:04:05"),
	); err != nil {
		t.Fatal(err)
	}
	if _, ok := s.FindPasswordResetUserID(expired); ok {
		t.Fatal("expired token must not resolve")
	}
}

func TestResetPasswordWithToken_consumesToken(t *testing.T) {
	s := newResetTestStore(t)
	user, err := s.FindUserByEmail("demo@example.com")
	if err != nil {
		t.Fatal(err)
	}
	token, err := s.CreatePasswordResetToken(user.ID)
	if err != nil {
		t.Fatal(err)
	}

	passwordHash, err := session.HashPassword("new-password-123")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.ResetPasswordWithToken(token, passwordHash); err != nil {
		t.Fatalf("reset with valid token: %v", err)
	}
	if _, ok := s.FindPasswordResetUserID(token); ok {
		t.Fatal("token must be consumed after a successful reset")
	}

	updated, err := s.FindUserByEmail("demo@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if !session.VerifyPassword(updated.PasswordHash, "new-password-123") {
		t.Fatal("password was not updated")
	}
}
`
