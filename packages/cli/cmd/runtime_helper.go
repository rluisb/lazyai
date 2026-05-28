// Package cmd provides CLI commands for the LazyAI toolkit.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rluisb/lazyai/packages/cli/internal/runtime"
)

// getRuntimeDBPath returns the path to the runtime database.
// Uses .specify/session.db in the current working directory.
func getRuntimeDBPath() string {
	dir, _ := os.Getwd()
	return filepath.Join(dir, ".specify", "session.db")
}

// openRuntimeDB opens the runtime database, creating it if necessary,
// and applies the unified Go schema (SchemaV1).
func openRuntimeDB() (*runtime.DB, error) {
	dbPath := getRuntimeDBPath()

	db, err := runtime.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("open runtime database: %w", err)
	}

	// Apply schema if tables don't exist
	if _, err := db.Exec(runtime.SchemaV1); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("apply runtime schema: %w", err)
	}

	return db, nil
}

// ensureSession creates a session row if it does not already exist.
// This prevents FK constraint failures when task_queue or workflow_instances
// reference a session_id that has no corresponding sessions row.
func ensureSession(db *runtime.DB, sessionID string) error {
	var exists int
	err := db.QueryRow("SELECT 1 FROM sessions WHERE id = ?", sessionID).Scan(&exists)
	if err == nil {
		return nil // already exists
	}
	if err.Error() != "sql: no rows in result set" {
		return fmt.Errorf("check session: %w", err)
	}

	_, err = db.Exec(
		"INSERT INTO sessions (id, started_at, agent, status) VALUES (?, ?, ?, ?)",
		sessionID, time.Now().UTC().Format(time.RFC3339), "cli", "active",
	)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}
