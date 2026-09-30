package sqllog

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/puppe1990/amarra-cais/pkg/cais/logentry"
	"github.com/puppe1990/amarra-cais/pkg/cais/logtime"
)

type Config struct {
	Enabled bool
	Writer  io.Writer
	JSON    bool // structured JSON lines for /logs and agent tooling
}

type DB struct {
	inner  *sql.DB
	config Config
}

type Tx struct {
	inner  *sql.Tx
	config Config
}

// Wrap instruments *sql.DB without replacing the driver — store code keeps using familiar Exec/Query APIs.
func Wrap(db *sql.DB, cfg Config) *DB {
	if cfg.Writer == nil {
		cfg.Writer = os.Stdout
	}
	return &DB{inner: db, config: cfg}
}

func (d *DB) Raw() *sql.DB {
	return d.inner
}

func (d *DB) Close() error {
	return d.inner.Close()
}

func (d *DB) Exec(query string, args ...any) (sql.Result, error) {
	return d.ExecContext(context.Background(), query, args...)
}

func (d *DB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	start := time.Now()
	result, err := d.inner.ExecContext(ctx, query, args...)
	d.log(query, args, start, time.Since(start), err)
	return result, err
}

func (d *DB) Query(query string, args ...any) (*sql.Rows, error) {
	return d.QueryContext(context.Background(), query, args...)
}

func (d *DB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	start := time.Now()
	rows, err := d.inner.QueryContext(ctx, query, args...)
	d.log(query, args, start, time.Since(start), err)
	return rows, err
}

func (d *DB) QueryRow(query string, args ...any) *sql.Row {
	return d.QueryRowContext(context.Background(), query, args...)
}

func (d *DB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	start := time.Now()
	row := d.inner.QueryRowContext(ctx, query, args...)
	// *sql.Row defers the error until Scan, including context.Canceled.
	d.log(query, args, start, time.Since(start), nil)
	return row
}

func (d *DB) Begin() (*Tx, error) {
	tx, err := d.inner.Begin()
	if err != nil {
		return nil, err
	}
	return &Tx{inner: tx, config: d.config}, nil
}

func (t *Tx) Exec(query string, args ...any) (sql.Result, error) {
	return t.ExecContext(context.Background(), query, args...)
}

func (t *Tx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	start := time.Now()
	result, err := t.inner.ExecContext(ctx, query, args...)
	t.log(query, args, start, time.Since(start), err)
	return result, err
}

func (t *Tx) Query(query string, args ...any) (*sql.Rows, error) {
	return t.QueryContext(context.Background(), query, args...)
}

func (t *Tx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	start := time.Now()
	rows, err := t.inner.QueryContext(ctx, query, args...)
	t.log(query, args, start, time.Since(start), err)
	return rows, err
}

func (t *Tx) QueryRow(query string, args ...any) *sql.Row {
	return t.QueryRowContext(context.Background(), query, args...)
}

func (t *Tx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	start := time.Now()
	row := t.inner.QueryRowContext(ctx, query, args...)
	t.log(query, args, start, time.Since(start), nil)
	return row
}

func (t *Tx) Commit() error {
	return t.inner.Commit()
}

func (t *Tx) Rollback() error {
	return t.inner.Rollback()
}

func (d *DB) log(query string, args []any, at time.Time, elapsed time.Duration, err error) {
	if !d.config.Enabled {
		return
	}
	writeLog(d.config, query, args, at, elapsed, err)
}

func (t *Tx) log(query string, args []any, at time.Time, elapsed time.Duration, err error) {
	if !t.config.Enabled {
		return
	}
	writeLog(t.config, query, args, at, elapsed, err)
}

func writeLog(cfg Config, query string, args []any, at time.Time, elapsed time.Duration, err error) {
	query = strings.Join(strings.Fields(query), " ")
	w := cfg.Writer
	if w == nil {
		w = os.Stdout
	}
	if cfg.JSON {
		entry := logentry.Entry{
			Kind:       "sql",
			At:         at.UTC(),
			Operation:  operationLabel(query),
			Query:      query,
			Args:       args,
			DurationMS: float64(elapsed.Microseconds()) / 1000,
		}
		if err != nil {
			entry.Error = err.Error()
		}
		_ = logentry.Write(w, entry)
		return
	}
	label := operationLabel(query)
	line := fmt.Sprintf("  %s (%s)  %s  %s  at %s", label, formatDuration(elapsed), query, formatArgs(args), logtime.Format(at))
	if err != nil {
		line += fmt.Sprintf("  ERROR: %v", err)
	}
	_, _ = fmt.Fprintln(w, line)
}
