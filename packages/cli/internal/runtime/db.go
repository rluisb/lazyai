// Package runtime provides the Go-native runtime for LazyAI multi-agent execution.
//
// This package replaces the bash script runtime (session-db.sh, task-queue.sh,
// workflow-run.sh, ledger.sh) with a unified SQLite-backed Go implementation.
//
// Architecture:
//   - runtime/db: Connection management, transactions, WAL mode
//   - runtime/migrate: Versioned schema migrations
//   - runtime/session: Session lifecycle, dispatches, parallel tasks
//   - runtime/taskqueue: Atomic claiming, DLQ, zombie sweep
//   - runtime/workflow: YAML execution, phase dispatch
//   - runtime/ledger: Hash chain, locking, redaction
//   - runtime/dispatch: Agent execution interface
//
// All tables use TEXT for timestamps (ISO-8601 UTC) and INTEGER for booleans (0/1).
package runtime

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
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

	// Open with busy timeout and foreign keys
	dsn := fmt.Sprintf("%s?_fk=1&_busy_timeout=5000&_journal_mode=WAL", path)
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

// Path returns the database file path.
func (db *DB) Path() string {
	return db.path
}

// BackupTo creates a backup copy of the database.
func (db *DB) BackupTo(destPath string) error {
	// Use SQLite backup API for online backup
	_, err := db.Exec(fmt.Sprintf(
		"VACUUM INTO '%s'",
		destPath,
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
