// Package cmd provides CLI commands for the LazyAI toolkit.
package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rluisb/lazyai/packages/cli/internal/runtime"
)

const runtimeSchemaVersionV2 = 2

// getRuntimeDBPath returns the path to the runtime database.
// Uses .specify/session.db in the current working directory.
func getRuntimeDBPath() string {
	dir, _ := os.Getwd()
	return filepath.Join(dir, ".specify", "session.db")
}

// openRuntimeDB opens the runtime database, creating it if necessary,
// and upgrades legacy runtime databases to SchemaV2 before returning.
func openRuntimeDB() (*runtime.DB, error) {
	dbPath := getRuntimeDBPath()

	db, err := runtime.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("open runtime database: %w", err)
	}

	schemaVersion, err := detectRuntimeSchemaVersion(db)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("detect runtime schema: %w", err)
	}

	switch schemaVersion {
	case 0, runtimeSchemaVersionV2:
		if err := applyRuntimeSchema(db); err != nil {
			_ = db.Close()
			return nil, err
		}
		return db, nil
	case 1:
		backupPath := runtimeBackupPath(dbPath)
		if err := createRuntimeBackup(db, backupPath); err != nil {
			_ = db.Close()
			return nil, err
		}
		if err := db.Close(); err != nil {
			return nil, fmt.Errorf("close runtime database before migration: %w", err)
		}
		if err := migrateRuntimeDBFromBackup(dbPath, backupPath); err != nil {
			return nil, err
		}

		migratedDB, err := runtime.Open(dbPath)
		if err != nil {
			return nil, fmt.Errorf("reopen migrated runtime database: %w", err)
		}
		if err := applyRuntimeSchema(migratedDB); err != nil {
			_ = migratedDB.Close()
			return nil, err
		}
		return migratedDB, nil
	default:
		_ = db.Close()
		return nil, fmt.Errorf("unsupported runtime schema version: %d", schemaVersion)
	}
}

func applyRuntimeSchema(db *runtime.DB) error {
	if _, err := db.Exec(runtime.SchemaCurrent); err != nil {
		return fmt.Errorf("apply runtime schema: %w", err)
	}
	return nil
}

func detectRuntimeSchemaVersion(db *runtime.DB) (int, error) {
	tableCount, err := countRuntimeUserTables(db)
	if err != nil {
		return 0, err
	}
	if tableCount == 0 {
		return 0, nil
	}

	recordedVersion, hasRecordedVersion, err := currentRecordedRuntimeSchemaVersion(db)
	if err != nil {
		return 0, err
	}
	if hasRecordedVersion && recordedVersion >= runtimeSchemaVersionV2 {
		return runtimeSchemaVersionV2, nil
	}

	handoffExists, err := runtimeTableExists(db, "handoff")
	if err != nil {
		return 0, err
	}
	agentDefaultsExists, err := runtimeTableExists(db, "agent_defaults")
	if err != nil {
		return 0, err
	}
	if handoffExists && agentDefaultsExists {
		return runtimeSchemaVersionV2, nil
	}

	for _, marker := range []string{"task_queue", "task_claims", "workflow_instances", "parallel_tasks", "messages", "model_calls"} {
		exists, err := runtimeTableExists(db, marker)
		if err != nil {
			return 0, err
		}
		if exists {
			return 1, nil
		}
	}

	sessionsExists, err := runtimeTableExists(db, "sessions")
	if err != nil {
		return 0, err
	}
	dispatchesExists, err := runtimeTableExists(db, "dispatches")
	if err != nil {
		return 0, err
	}
	if sessionsExists && dispatchesExists {
		return 1, nil
	}

	return 0, nil
}

func countRuntimeUserTables(db *runtime.DB) (int, error) {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%'").Scan(&count); err != nil {
		return 0, fmt.Errorf("count runtime tables: %w", err)
	}
	return count, nil
}

func currentRecordedRuntimeSchemaVersion(db *runtime.DB) (int, bool, error) {
	exists, err := runtimeTableExists(db, "schema_migrations")
	if err != nil {
		return 0, false, err
	}
	if !exists {
		return 0, false, nil
	}

	var version sql.NullInt64
	if err := db.QueryRow("SELECT MAX(version) FROM schema_migrations").Scan(&version); err != nil {
		return 0, false, fmt.Errorf("query runtime schema version: %w", err)
	}
	if !version.Valid {
		return 0, false, nil
	}
	return int(version.Int64), true, nil
}

func runtimeTableExists(db *runtime.DB, tableName string) (bool, error) {
	var exists int
	err := db.QueryRow(
		"SELECT 1 FROM sqlite_master WHERE type = 'table' AND name = ?",
		tableName,
	).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("check table %s: %w", tableName, err)
	}
	return true, nil
}

func runtimeBackupPath(dbPath string) string {
	return dbPath + ".backup"
}

func runtimeMigrationTempPath(dbPath string) string {
	return dbPath + ".migrating"
}

func createRuntimeBackup(db *runtime.DB, backupPath string) error {
	tempBackupPath := backupPath + ".tmp"
	if err := removeRuntimeArtifacts(tempBackupPath); err != nil {
		return err
	}
	if err := db.BackupTo(tempBackupPath); err != nil {
		return fmt.Errorf("create runtime backup: %w", err)
	}
	if err := os.Rename(tempBackupPath, backupPath); err != nil {
		if err := os.Remove(backupPath); err != nil && !os.IsNotExist(err) {
			_ = removeRuntimeArtifacts(tempBackupPath)
			return fmt.Errorf("replace existing runtime backup: %w", err)
		}
		if err := os.Rename(tempBackupPath, backupPath); err != nil {
			_ = removeRuntimeArtifacts(tempBackupPath)
			return fmt.Errorf("finalize runtime backup: %w", err)
		}
	}
	return nil
}

func migrateRuntimeDBFromBackup(dbPath string, backupPath string) error {
	tempPath := runtimeMigrationTempPath(dbPath)
	if err := removeRuntimeArtifacts(tempPath); err != nil {
		return err
	}

	migratedDB, err := runtime.Open(tempPath)
	if err != nil {
		return fmt.Errorf("open migration database: %w", err)
	}
	success := false
	defer func() {
		_ = migratedDB.Close()
		if !success {
			_ = removeRuntimeArtifacts(tempPath)
		}
	}()

	if err := applyRuntimeSchema(migratedDB); err != nil {
		return err
	}
	if err := copyRuntimeV1Data(migratedDB, backupPath); err != nil {
		return err
	}
	if err := validateRuntimeSchema(migratedDB); err != nil {
		return err
	}
	if err := migratedDB.Close(); err != nil {
		return fmt.Errorf("close migrated runtime database: %w", err)
	}
	if err := removeRuntimeArtifacts(dbPath + "-wal"); err != nil {
		return err
	}
	if err := removeRuntimeArtifacts(dbPath + "-shm"); err != nil {
		return err
	}
	if err := os.Rename(tempPath, dbPath); err != nil {
		return fmt.Errorf("swap migrated runtime database: %w", err)
	}
	success = true
	return nil
}

func copyRuntimeV1Data(db *runtime.DB, backupPath string) error {
	quotedBackupPath := strings.ReplaceAll(backupPath, "'", "''")
	if _, err := db.Exec(fmt.Sprintf("ATTACH DATABASE '%s' AS legacy", quotedBackupPath)); err != nil {
		return fmt.Errorf("attach runtime backup: %w", err)
	}
	defer func() {
		_, _ = db.Exec("DETACH DATABASE legacy")
	}()

	if err := db.WithTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(`
			INSERT INTO sessions (id, started_at, ended_at, agent, model, goal, repo, worktree, status, token_total, summary, tags)
			SELECT
				id,
				started_at,
				ended_at,
				CASE
					WHEN agent IS NULL OR TRIM(agent) = '' OR agent IN ('loop-driver', 'orchestrator') THEN 'primary-agent'
					ELSE agent
				END,
				model,
				goal,
				repo,
				worktree,
				status,
				token_total,
				summary,
				tags
			FROM legacy.sessions
		`); err != nil {
			return fmt.Errorf("migrate sessions: %w", err)
		}

		if _, err := tx.Exec(`
			INSERT INTO dispatches (id, session_id, seq, parent_id, agent, model, task, phase, workflow, mode, started_at, ended_at, result, token_used, error_message, summary, files_touched)
			SELECT
				id,
				session_id,
				seq,
				parent_id,
				CASE
					WHEN agent IS NULL OR TRIM(agent) = '' OR agent IN ('loop-driver', 'orchestrator') THEN 'primary-agent'
					ELSE agent
				END,
				model,
				task,
				phase,
				workflow,
				mode,
				started_at,
				ended_at,
				result,
				token_used,
				error_message,
				summary,
				files_touched
			FROM legacy.dispatches
		`); err != nil {
			return fmt.Errorf("migrate dispatches: %w", err)
		}

		if _, err := tx.Exec(`
			INSERT INTO ledger_refs (id, session_id, event_type, metadata, created_at)
			SELECT
				seq,
				session_id,
				event_type,
				'workflow_run_id=' || COALESCE(workflow_run_id, '') ||
				';event_hash=' || COALESCE(event_hash, '') ||
				';prev_hash=' || COALESCE(prev_hash, ''),
				created_at
			FROM legacy.ledger_refs
		`); err != nil {
			return fmt.Errorf("migrate ledger refs: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("copy V1 runtime data: %w", err)
	}

	return nil
}

func validateRuntimeSchema(db *runtime.DB) error {
	for _, tableName := range []string{"schema_migrations", "sessions", "dispatches", "handoff", "agent_defaults", "ledger_refs"} {
		exists, err := runtimeTableExists(db, tableName)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("runtime schema missing required table %q", tableName)
		}
	}

	version, ok, err := currentRecordedRuntimeSchemaVersion(db)
	if err != nil {
		return err
	}
	if !ok || version < runtimeSchemaVersionV2 {
		return fmt.Errorf("runtime schema version not recorded as V2")
	}

	return nil
}

func removeRuntimeArtifacts(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove %s: %w", path, err)
	}
	return nil
}
