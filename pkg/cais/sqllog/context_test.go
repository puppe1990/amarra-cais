package sqllog

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

// sqlc DBTX minus PrepareContext — out of #264.
type contextQuerier interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

var (
	_ contextQuerier = (*DB)(nil)
	_ contextQuerier = (*Tx)(nil)
)

func wrapUsers(t *testing.T) (*DB, *bytes.Buffer) {
	t.Helper()
	raw, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = raw.Close() })
	if _, err := raw.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY, email TEXT)`); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	return Wrap(raw, Config{Enabled: true, Writer: &buf}), &buf
}

func canceled() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func assertCanceledLogged(t *testing.T, err error, buf *bytes.Buffer, query string) {
	t.Helper()
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	if !strings.Contains(buf.String(), query) {
		t.Fatalf("missing %q in log:\n%s", query, buf.String())
	}
}

func TestContextMethods_canceledReturnsCanceledAndLogs(t *testing.T) {
	for _, name := range []string{"DB", "Tx"} {
		t.Run(name, func(t *testing.T) {
			db, buf := wrapUsers(t)
			var q contextQuerier
			if name == "Tx" {
				tx, err := db.Begin()
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = tx.Rollback() })
				q = tx
			} else {
				q = db
			}

			t.Run("ExecContext", func(t *testing.T) {
				buf.Reset()
				_, err := q.ExecContext(canceled(), `INSERT INTO users (email) VALUES (?)`, "demo@example.com")
				assertCanceledLogged(t, err, buf, `INSERT INTO users`)
			})
			t.Run("QueryContext", func(t *testing.T) {
				buf.Reset()
				rows, err := q.QueryContext(canceled(), `SELECT email FROM users`)
				if rows != nil {
					_ = rows.Close()
				}
				assertCanceledLogged(t, err, buf, `SELECT email FROM users`)
			})
			t.Run("QueryRowContext", func(t *testing.T) {
				buf.Reset()
				var email string
				err := q.QueryRowContext(canceled(), `SELECT email FROM users`).Scan(&email)
				assertCanceledLogged(t, err, buf, `SELECT email FROM users`)
			})
		})
	}
}

func TestDB_ExecContext_logsSuccessfulQuery(t *testing.T) {
	db, buf := wrapUsers(t)
	if _, err := db.ExecContext(context.Background(), `INSERT INTO users (email) VALUES (?)`, "demo@example.com"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), `INSERT INTO users`) {
		t.Fatalf("missing SQL in log:\n%s", buf.String())
	}
}

func TestDB_shortMethodsStillWork(t *testing.T) {
	db, buf := wrapUsers(t)
	if _, err := db.Exec(`INSERT INTO users (email) VALUES (?)`, "demo@example.com"); err != nil {
		t.Fatal(err)
	}
	rows, err := db.Query(`SELECT email FROM users`)
	if err != nil {
		t.Fatal(err)
	}
	_ = rows.Close()
	var email string
	if err := db.QueryRow(`SELECT email FROM users`).Scan(&email); err != nil {
		t.Fatal(err)
	}
	if email != "demo@example.com" {
		t.Fatalf("email = %q", email)
	}
	out := buf.String()
	if !strings.Contains(out, `INSERT INTO users`) || !strings.Contains(out, `SELECT email FROM users`) {
		t.Fatalf("short methods missing from log:\n%s", out)
	}
}

func TestTx_shortMethodsStillWork(t *testing.T) {
	db, buf := wrapUsers(t)
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = tx.Rollback() })
	if _, err := tx.Exec(`INSERT INTO users (email) VALUES (?)`, "demo@example.com"); err != nil {
		t.Fatal(err)
	}
	rows, err := tx.Query(`SELECT email FROM users`)
	if err != nil {
		t.Fatal(err)
	}
	_ = rows.Close()
	var email string
	if err := tx.QueryRow(`SELECT email FROM users`).Scan(&email); err != nil {
		t.Fatal(err)
	}
	if email != "demo@example.com" {
		t.Fatalf("email = %q", email)
	}
	out := buf.String()
	if !strings.Contains(out, `INSERT INTO users`) || !strings.Contains(out, `SELECT email FROM users`) {
		t.Fatalf("short methods missing from log:\n%s", out)
	}
}
