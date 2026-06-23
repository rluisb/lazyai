// Package runtime provides the Go-native runtime for LazyAI multi-agent execution.
//
// This package replaces the bash script runtime (session-db.sh, task-queue.sh,
// workflow-run.sh, ledger.sh) with a unified SQLite-backed Go implementation.
//
// Architecture:
//   - runtime/db: Connection management, transactions, WAL mode
//   - runtime/migrate: Versioned schema migrations
//   - runtime/session: Session lifecycle, dispatches, parallel tasks
//   - runtime/ledger: Hash chain, locking, redaction
//
// All tables use TEXT for timestamps (ISO-8601 UTC) and INTEGER for booleans (0/1).
package runtime

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// DB wraps sql.DB with runtime-specific operations.
type DB struct {
	*sql.DB
	path string
}

// Open creates or opens a runtime database at the given path.
// Enables WAL mode, foreign keys, and busy timeout.
func Open(path string) (*DB, error) {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	// Open with busy timeout, foreign keys, and WAL mode. The
	// modernc.org/sqlite driver only honours pragmas passed via
	// _pragma=... query parameters; bare _busy_timeout/_fk/_journal_mode
	// params are silently ignored, which left busy_timeout at 0 and
	// caused BEGIN IMMEDIATE to fail with SQLITE_BUSY under concurrency.
	dsn := fmt.Sprintf("%s?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &DB{DB: db, path: path}, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.DB.Close()
}

// WithTx executes fn within a transaction.
// Automatically rolls back on error, commits on success.
func (db *DB) WithTx(fn func(*sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction failed: %v (rollback also failed: %v)", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// WithImmediateTx executes fn within an immediate transaction.
// BEGIN IMMEDIATE acquires the write lock up front, serializing
// read-modify-write sequences and preventing the lost-update race
// that a deferred BEGIN permits under concurrent access.
//
// The callback receives a pinned *sql.Conn; all queries/execs issued
// through it run on the same connection inside the transaction.
func (db *DB) WithImmediateTx(ctx context.Context, fn func(*sql.Conn) error) error {
	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
		return fmt.Errorf("begin immediate transaction: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			_, _ = conn.ExecContext(context.Background(), "ROLLBACK")
		}
	}()

	if err := fn(conn); err != nil {
		return err
	}

	if _, err := conn.ExecContext(ctx, "COMMIT"); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	committed = true
	return nil
}

// Path returns the database file path.
func (db *DB) Path() string {
	return db.path
}

// BackupTo creates a backup copy of the database.
func (db *DB) BackupTo(destPath string) error {
	// Escape single quotes to prevent SQL injection via the path.
	// VACUUM INTO accepts a string literal; doubling embedded single
	// quotes is the SQLite-compatible way to escape them.
	escapedPath := strings.ReplaceAll(destPath, "'", "''")
	_, err := db.Exec(fmt.Sprintf(
		"VACUUM INTO '%s'",
		escapedPath,
	))
	if err != nil {
		return fmt.Errorf("backup database: %w", err)
	}
	return nil
}

// Now returns the current time in ISO-8601 UTC format (the canonical
// timestamp format for all runtime tables).
func Now() string {
	return time.Now().UTC().Format(time.RFC3339)
}
