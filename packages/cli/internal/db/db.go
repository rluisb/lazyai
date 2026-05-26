// Package db provides a SQLite persistence layer for ai-setup using
// modernc.org/sqlite (pure Go, no CGO) and golang-migrate for schema migrations.
package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // register "sqlite" driver
)

// DB wraps *sql.DB with ai-setup specific defaults.
type DB struct {
	*sql.DB
	path string
}

// Open opens (or creates) a SQLite database at dbPath.
// It creates the parent directory if needed, enables WAL mode, and
// turns on foreign keys.
func Open(dbPath string) (*DB, error) {
	// Ensure parent directory exists.
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create db directory %s: %w", dir, err)
	}

	sqlDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite %s: %w", dbPath, err)
	}

	// Enable WAL mode and foreign keys.
	if _, err := sqlDB.Exec("PRAGMA journal_mode=WAL"); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("set WAL mode: %w", err)
	}
	if _, err := sqlDB.Exec("PRAGMA foreign_keys=ON"); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	// Set busy timeout to wait for locks instead of failing immediately (5 seconds)
	if _, err := sqlDB.Exec("PRAGMA busy_timeout=5000"); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("set busy timeout: %w", err)
	}

	// Limit connection pool to prevent "too many open connections"
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxLifetime(0)

	return &DB{DB: sqlDB, path: dbPath}, nil
}

// Close closes the underlying database connection.
func (db *DB) Close() error {
	return db.DB.Close()
}

// Path returns the filesystem path of the database file.
func (db *DB) Path() string {
	return db.path
}

// DefaultDBPath returns the default database path for a given target directory.
func DefaultDBPath(targetDir string) string {
	return filepath.Join(targetDir, ".ai-setup.db")
}
