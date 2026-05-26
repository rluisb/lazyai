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
	{
		Version: 2,
		Up:      migrationUp002,
		Down:    migrationDown002,
	},
	{
		Version: 3,
		Up:      migrationUp003,
		Down:    migrationDown003,
	},
	{
		Version: 4,
		Up:      migrationUp004,
		Down:    migrationDown004,
	},
	{
		Version: 5,
		Up:      migrationUp005,
		Down:    migrationDown005,
	},
	{
		Version: 6,
		Up:      migrationUp006,
		Down:    migrationDown006,
	},
	{
		Version: 7,
		Up:      migrationUp007,
		Down:    migrationDown007,
	},
	{
		Version: 8,
		Up:      migrationUp008,
		Down:    migrationDown008,
	},
	{
		Version: 9,
		Up:      migrationUp009,
		Down:    migrationDown009,
	},
	{
		Version: 10,
		Up:      migrationUp010,
		Down:    migrationDown010,
	},
	{
		Version: 11,
		Up:      migrationUp011,
		Down:    migrationDown011,
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

// migration 002 — add commands + chatmodes columns to selections
// (spec 009 follow-up: persist Gemini commands and Copilot chatmodes so
// re-runs and ai-setup status see the full selection).
const migrationUp002 = `
ALTER TABLE selections ADD COLUMN commands TEXT NOT NULL DEFAULT '[]';
ALTER TABLE selections ADD COLUMN chatmodes TEXT NOT NULL DEFAULT '[]';
`

const migrationDown002 = `
ALTER TABLE selections DROP COLUMN chatmodes;
ALTER TABLE selections DROP COLUMN commands;
`

// migration 003 — add opencode_commands + opencode_modes columns
// (spec 011 phase 3: persist opencode slash commands and chat modes so
// re-runs and ai-setup status see the full selection).
const migrationUp003 = `
ALTER TABLE selections ADD COLUMN opencode_commands TEXT NOT NULL DEFAULT '[]';
ALTER TABLE selections ADD COLUMN opencode_modes TEXT NOT NULL DEFAULT '[]';
`

const migrationDown003 = `
ALTER TABLE selections DROP COLUMN opencode_modes;
ALTER TABLE selections DROP COLUMN opencode_commands;
`

// migration 004 — add opencode_plugins column
// (spec 011 phase 5: persist opencode plugin module names).
const migrationUp004 = `
ALTER TABLE selections ADD COLUMN opencode_plugins TEXT NOT NULL DEFAULT '[]';
`

const migrationDown004 = `
ALTER TABLE selections DROP COLUMN opencode_plugins;
`

// migration 005 — add kind + link_target columns to tracked_files
// (Go↔TS alignment: TS already has kind=file|symlink and linkTarget;
// Go needs to support the same fields for cross-compatibility.)
//
// Also add workspace_root + housekeeping columns to config table
// (Go↔TS alignment: Go types already define these but DB was missing them.)
const migrationUp005 = `
ALTER TABLE tracked_files ADD COLUMN kind TEXT NOT NULL DEFAULT 'file';
ALTER TABLE tracked_files ADD COLUMN link_target TEXT NOT NULL DEFAULT '';
ALTER TABLE config ADD COLUMN workspace_root TEXT NOT NULL DEFAULT '';
ALTER TABLE config ADD COLUMN housekeeping TEXT NOT NULL DEFAULT '{}';
`

const migrationDown005 = `
ALTER TABLE config DROP COLUMN housekeeping;
ALTER TABLE config DROP COLUMN workspace_root;
ALTER TABLE tracked_files DROP COLUMN link_target;
ALTER TABLE tracked_files DROP COLUMN kind;
`

const migrationUp006 = "ALTER TABLE config ADD COLUMN setup_type TEXT NOT NULL DEFAULT '';"

const migrationDown006 = "ALTER TABLE config DROP COLUMN setup_type;"

const migrationUp007 = `
ALTER TABLE config ADD COLUMN projectOverview TEXT NOT NULL DEFAULT '';
ALTER TABLE config ADD COLUMN namingConventions TEXT NOT NULL DEFAULT '';
ALTER TABLE config ADD COLUMN errorHandling TEXT NOT NULL DEFAULT '';
ALTER TABLE config ADD COLUMN apiConventions TEXT NOT NULL DEFAULT '';
ALTER TABLE config ADD COLUMN importOrder TEXT NOT NULL DEFAULT '';
ALTER TABLE config ADD COLUMN protectedBranch TEXT NOT NULL DEFAULT '';
ALTER TABLE config ADD COLUMN testCommand TEXT NOT NULL DEFAULT '';
ALTER TABLE config ADD COLUMN lintCommand TEXT NOT NULL DEFAULT '';
ALTER TABLE config ADD COLUMN buildCommand TEXT NOT NULL DEFAULT '';
ALTER TABLE config ADD COLUMN coverageThreshold INTEGER NOT NULL DEFAULT 80 CHECK(coverageThreshold >= 1 AND coverageThreshold <= 100);
`

const migrationDown007 = `
ALTER TABLE config DROP COLUMN coverageThreshold;
ALTER TABLE config DROP COLUMN buildCommand;
ALTER TABLE config DROP COLUMN lintCommand;
ALTER TABLE config DROP COLUMN testCommand;
ALTER TABLE config DROP COLUMN protectedBranch;
ALTER TABLE config DROP COLUMN importOrder;
ALTER TABLE config DROP COLUMN apiConventions;
ALTER TABLE config DROP COLUMN errorHandling;
ALTER TABLE config DROP COLUMN namingConventions;
ALTER TABLE config DROP COLUMN projectOverview;
`

// migration 008 — add opencode_providers column. Persists which provider
// auths the user wants OpenCode-side agents to draw models from
// (e.g., ["openai", "ollama-cloud"]). Empty list means "no preference set"
// — adapters fall back to live auth.DetectAll at install time.
const migrationUp008 = `
ALTER TABLE selections ADD COLUMN opencode_providers TEXT NOT NULL DEFAULT '[]';
`

const migrationDown008 = `
ALTER TABLE selections DROP COLUMN opencode_providers;
`

const migrationUp009 = `
-- Session tracking tables (ported from production AI agent runtime)

CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    goal TEXT,
    repo_name TEXT,
    status TEXT DEFAULT 'active',
    started_at TEXT NOT NULL,
    ended_at TEXT,
    model TEXT,
    cost_usd REAL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS agent_dispatches (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    seq INTEGER NOT NULL,
    agent TEXT NOT NULL,
    task TEXT,
    status TEXT DEFAULT 'pending',
    dispatched_at TEXT,
    completed_at TEXT,
    result TEXT,
    error TEXT,
    FOREIGN KEY (session_id) REFERENCES sessions(id)
);

CREATE TABLE IF NOT EXISTS quality_metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    workflow_id TEXT,
    agent TEXT NOT NULL,
    model TEXT,
    metric_name TEXT NOT NULL,
    metric_value REAL NOT NULL,
    recorded_at TEXT NOT NULL,
    FOREIGN KEY (session_id) REFERENCES sessions(id)
);

CREATE INDEX IF NOT EXISTS idx_dispatches_session ON agent_dispatches(session_id);
CREATE INDEX IF NOT EXISTS idx_dispatches_agent ON agent_dispatches(agent);
CREATE INDEX IF NOT EXISTS idx_metrics_session ON quality_metrics(session_id);
CREATE INDEX IF NOT EXISTS idx_metrics_name ON quality_metrics(metric_name);
`

const migrationDown009 = `
DROP TABLE IF EXISTS quality_metrics;
DROP TABLE IF EXISTS agent_dispatches;
DROP TABLE IF EXISTS sessions;
`

const migrationUp010 = `
-- Task queue tables (ported from production AI agent runtime)

CREATE TABLE IF NOT EXISTS tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id TEXT UNIQUE NOT NULL,
    description TEXT NOT NULL,
    agent TEXT,
    status TEXT DEFAULT 'pending',
    priority INTEGER DEFAULT 0,
    created_at TEXT NOT NULL,
    claimed_at TEXT,
    completed_at TEXT,
    claimed_by TEXT,
    result TEXT,
    error TEXT,
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    parent_task_id TEXT,
    metadata TEXT
);

CREATE TABLE IF NOT EXISTS task_queue (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id TEXT NOT NULL,
    queue_position INTEGER NOT NULL,
    enqueued_at TEXT NOT NULL,
    FOREIGN KEY (task_id) REFERENCES tasks(task_id)
);

CREATE TABLE IF NOT EXISTS dead_letter_queue (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id TEXT NOT NULL,
    failed_at TEXT NOT NULL,
    error TEXT,
    retry_count INTEGER,
    FOREIGN KEY (task_id) REFERENCES tasks(task_id)
);

CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_agent ON tasks(agent);
CREATE INDEX IF NOT EXISTS idx_tasks_priority ON tasks(priority DESC, created_at);
CREATE INDEX IF NOT EXISTS idx_task_queue_position ON task_queue(queue_position);
`

const migrationDown010 = `
DROP TABLE IF EXISTS dead_letter_queue;
DROP TABLE IF EXISTS task_queue;
DROP TABLE IF EXISTS tasks;
`

const migrationUp011 = `
-- Agent message bus tables

CREATE TABLE IF NOT EXISTS messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    message_id TEXT UNIQUE NOT NULL,
    from_agent TEXT NOT NULL,
    to_agent TEXT NOT NULL,
    subject TEXT,
    body TEXT,
    priority TEXT DEFAULT 'normal',
    status TEXT DEFAULT 'unread',
    created_at TEXT NOT NULL,
    read_at TEXT,
    session_id TEXT
);

CREATE INDEX IF NOT EXISTS idx_messages_to ON messages(to_agent);
CREATE INDEX IF NOT EXISTS idx_messages_from ON messages(from_agent);
CREATE INDEX IF NOT EXISTS idx_messages_status ON messages(status);
CREATE INDEX IF NOT EXISTS idx_messages_priority ON messages(priority);
`

const migrationDown011 = `
DROP TABLE IF EXISTS messages;
`
