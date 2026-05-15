package agentmemory

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DB wraps a SQLite database handle for agent memory storage.
type DB struct {
	*sql.DB
	path string
}

// Open opens or creates a SQLite database and applies all pending migrations.
func Open(path string) (*DB, error) {
	if path != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			return nil, fmt.Errorf("create database directory: %w", err)
		}
	}

	sqlDB, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if _, err := sqlDB.Exec("PRAGMA foreign_keys=ON"); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}
	if path != ":memory:" {
		if _, err := sqlDB.Exec("PRAGMA journal_mode=WAL"); err != nil {
			sqlDB.Close()
			return nil, fmt.Errorf("enable WAL: %w", err)
		}
	}

	if err := RunMigrations(sqlDB); err != nil {
		sqlDB.Close()
		return nil, err
	}

	return &DB{DB: sqlDB, path: path}, nil
}

// Close closes the underlying database handle.
func (db *DB) Close() error {
	return db.DB.Close()
}

// Path returns the database path used by Open.
func (db *DB) Path() string {
	return db.path
}
