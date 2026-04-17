package db

import (
	"fmt"
	"strings"
)

// migrations holds the ordered list of schema migrations.
// Each migration has an up and down SQL string.
var migrations = []struct {
	Version uint
	Up      string
	Down    string
}{
	{
		Version: 1,
		Up:      migrationUp001,
		Down:    migrationDown001,
	},
}

// migrationSQL holds the SQL for creating and interacting with the
// schema_migrations tracking table.
const (
	createMigrationsTableSQL = `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    dirty   INTEGER NOT NULL DEFAULT 0
);
`
	getCurrentVersionSQL = `SELECT COALESCE(MAX(version), 0) FROM schema_migrations WHERE dirty = 0`
	setVersionSQL        = `INSERT OR REPLACE INTO schema_migrations (version, dirty) VALUES (?, 0)`
)

// RunMigrations runs all pending schema migrations on the database.
// It uses a simple version-tracking approach instead of the golang-migrate
// library to avoid pulling in the CGO-dependent mattn/go-sqlite3 driver.
func RunMigrations(database *DB) error {
	db := database.DB
	// Ensure the migrations tracking table exists.
	if _, err := db.Exec(createMigrationsTableSQL); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	// Get current version.
	var currentVersion uint
	row := db.QueryRow(getCurrentVersionSQL)
	if err := row.Scan(&currentVersion); err != nil {
		return fmt.Errorf("read current migration version: %w", err)
	}

	// Run each pending migration in order within a transaction.
	for _, m := range migrations {
		if m.Version <= currentVersion {
			continue
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("begin migration %d transaction: %w", m.Version, err)
		}

		// Execute each statement in the migration.
		// We split on ";" to handle multiple statements since
		// SQLite's Exec doesn't support multiple statements in one call.
		for _, stmt := range splitStatements(m.Up) {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if _, err := tx.Exec(stmt); err != nil {
				tx.Rollback()
				return fmt.Errorf("migration %d statement %q: %w", m.Version, truncate(stmt, 80), err)
			}
		}

		// Record the new version.
		if _, err := tx.Exec(setVersionSQL, m.Version); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration version %d: %w", m.Version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %d: %w", m.Version, err)
		}
	}

	return nil
}

// splitStatements splits a SQL string into individual statements on ";".
// This is a simple splitter that does not handle semicolons inside string
// literals or comments, but is sufficient for our controlled migration SQL.
func splitStatements(sql string) []string {
	return strings.Split(sql, ";")
}

// truncate returns the first n characters of s, appending "..." if truncated.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// Migration version 1 SQL constants.

const migrationUp001 = `
CREATE TABLE IF NOT EXISTS meta (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    schema_version INTEGER NOT NULL DEFAULT 1,
    cli_version TEXT NOT NULL,
    installed_at TEXT NOT NULL,
    last_updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS config (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    scope TEXT NOT NULL,
    tools TEXT NOT NULL DEFAULT '[]',
    cli_tools TEXT NOT NULL DEFAULT '[]',
    enable_servers TEXT NOT NULL DEFAULT '[]',
    project_name TEXT NOT NULL DEFAULT '',
    workspace_name TEXT NOT NULL DEFAULT '',
    target_dir TEXT NOT NULL DEFAULT '',
    planning_dir TEXT NOT NULL DEFAULT 'specs',
    planning_repo_path TEXT NOT NULL DEFAULT '',
    repos TEXT NOT NULL DEFAULT '[]',
    global_ref TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS selections (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    templates TEXT NOT NULL DEFAULT '[]',
    rules TEXT NOT NULL DEFAULT '[]',
    agents TEXT NOT NULL DEFAULT '[]',
    skills TEXT NOT NULL DEFAULT '[]',
    prompts TEXT NOT NULL DEFAULT '[]',
    infra TEXT NOT NULL DEFAULT '[]',
    constitution TEXT NOT NULL DEFAULT '[]',
    features TEXT NOT NULL DEFAULT '{}',
    git_conventions TEXT NOT NULL DEFAULT '{}',
    preset TEXT NOT NULL DEFAULT 'standard'
);

CREATE TABLE IF NOT EXISTS tracked_files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    path TEXT NOT NULL UNIQUE,
    hash TEXT NOT NULL,
    source TEXT NOT NULL,
    owner TEXT NOT NULL DEFAULT 'library',
    status TEXT NOT NULL DEFAULT 'installed',
    installed_at TEXT NOT NULL,
    last_checked_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS operations (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    timestamp TEXT NOT NULL,
    files_affected TEXT NOT NULL DEFAULT '[]',
    result TEXT NOT NULL,
    backup_paths TEXT NOT NULL DEFAULT '[]',
    error TEXT DEFAULT ''
);

CREATE TABLE IF NOT EXISTS sync (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    last_sync_at TEXT NOT NULL,
    dirty INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS feature_flags (
    key TEXT PRIMARY KEY,
    value INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_tracked_files_path ON tracked_files(path);
CREATE INDEX IF NOT EXISTS idx_operations_timestamp ON operations(timestamp);
`

const migrationDown001 = `
DROP TABLE IF EXISTS feature_flags;
DROP TABLE IF EXISTS sync;
DROP TABLE IF EXISTS operations;
DROP TABLE IF EXISTS tracked_files;
DROP TABLE IF EXISTS selections;
DROP TABLE IF EXISTS config;
DROP TABLE IF EXISTS meta;
DROP TABLE IF EXISTS schema_migrations;
`
