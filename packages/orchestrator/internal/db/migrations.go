package db

import (
	"fmt"
	"sort"
	"time"
)

// Migration represents a versioned schema change.
type Migration struct {
	ID  string
	SQL string
}

var migrations = []Migration{
	{ID: "001_schema_migrations", SQL: `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id TEXT PRIMARY KEY,
			applied_at TEXT NOT NULL
		)
	`},
	{ID: "002_chain_runs", SQL: `
		CREATE TABLE IF NOT EXISTS chain_runs (
			id TEXT PRIMARY KEY,
			definition_name TEXT NOT NULL,
			definition_version TEXT,
			state TEXT NOT NULL,
			current_step_id TEXT,
			project_root TEXT NOT NULL,
			state_json TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)
	`},
	{ID: "003_team_runs", SQL: `
		CREATE TABLE IF NOT EXISTS team_runs (
			id TEXT PRIMARY KEY,
			definition_name TEXT NOT NULL,
			definition_version TEXT,
			state TEXT NOT NULL,
			project_root TEXT NOT NULL,
			state_json TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)
	`},
	{ID: "004_workflow_runs", SQL: `
		CREATE TABLE IF NOT EXISTS workflow_runs (
			id TEXT PRIMARY KEY,
			definition_name TEXT NOT NULL,
			definition_version TEXT,
			state TEXT NOT NULL,
			current_phase_id TEXT,
			project_root TEXT NOT NULL,
			state_json TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)
	`},
	{ID: "005_execution_plans", SQL: `
		CREATE TABLE IF NOT EXISTS execution_plans (
			id TEXT PRIMARY KEY,
			kind TEXT NOT NULL,
			definition_name TEXT NOT NULL,
			definition_version TEXT,
			project_root TEXT NOT NULL,
			plan_json TEXT NOT NULL,
			created_at TEXT NOT NULL
		)
	`},
	{ID: "006_handoffs", SQL: `
		CREATE TABLE IF NOT EXISTS handoffs (
			id TEXT PRIMARY KEY,
			run_id TEXT NOT NULL,
			run_kind TEXT NOT NULL,
			doc_json TEXT NOT NULL,
			created_at TEXT NOT NULL
		)
	`},
	{ID: "007_error_journal", SQL: `
		CREATE TABLE IF NOT EXISTS error_journal (
			id TEXT PRIMARY KEY,
			run_id TEXT,
			run_kind TEXT,
			definition_name TEXT NOT NULL,
			step_id TEXT,
			category TEXT NOT NULL,
			code TEXT NOT NULL,
			message TEXT NOT NULL,
			entry_json TEXT NOT NULL,
			created_at TEXT NOT NULL
		)
	`},
	{ID: "008_definitions", SQL: `
		CREATE TABLE IF NOT EXISTS definitions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			kind TEXT NOT NULL,
			name TEXT NOT NULL,
			active_version INTEGER,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			UNIQUE(kind, name)
		)
	`},
	{ID: "009_definition_versions", SQL: `
		CREATE TABLE IF NOT EXISTS definition_versions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			definition_id INTEGER NOT NULL,
			kind TEXT NOT NULL,
			name TEXT NOT NULL,
			version INTEGER NOT NULL,
			frontmatter_json TEXT NOT NULL,
			body TEXT NOT NULL,
			checksum TEXT NOT NULL,
			created_at TEXT NOT NULL,
			created_by TEXT,
			FOREIGN KEY (definition_id) REFERENCES definitions(id)
		)
	`},
	{ID: "010_queue_jobs", SQL: `
		CREATE TABLE IF NOT EXISTS queue_jobs (
			id TEXT PRIMARY KEY,
			job_type TEXT NOT NULL,
			payload_json TEXT NOT NULL DEFAULT '{}',
			status TEXT NOT NULL DEFAULT 'pending',
			priority INTEGER NOT NULL DEFAULT 0,
			attempts INTEGER NOT NULL DEFAULT 0,
			max_attempts INTEGER NOT NULL DEFAULT 3,
			error_json TEXT,
			created_at TEXT NOT NULL,
			claimed_at TEXT,
			completed_at TEXT
		)
	`},
	{ID: "011_run_events", SQL: `
		CREATE TABLE IF NOT EXISTS run_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			run_id TEXT NOT NULL,
			event_type TEXT NOT NULL,
			event_json TEXT NOT NULL,
			created_at TEXT NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_run_events_run_id ON run_events(run_id)
	`},
	{ID: "012_definitions_unique_version", SQL: `
		CREATE UNIQUE INDEX IF NOT EXISTS idx_def_version_unique
		ON definition_versions(definition_id, version)
	`},
}

// Run applies all pending migrations.
func (db *DB) RunMigrations() error {
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].ID < migrations[j].ID
	})

	for _, m := range migrations {
		var present int
		err := db.QueryRow("SELECT 1 FROM schema_migrations WHERE id = ?", m.ID).Scan(&present)
		if err == nil {
			continue
		}
		if _, err := db.Exec(m.SQL); err != nil {
			return fmt.Errorf("migration %s: %w", m.ID, err)
		}
		if _, err := db.Exec("INSERT INTO schema_migrations (id, applied_at) VALUES (?, ?)",
			m.ID, time.Now().UTC().Format(time.RFC3339)); err != nil {
			return fmt.Errorf("record migration %s: %w", m.ID, err)
		}
	}
	return nil
}
