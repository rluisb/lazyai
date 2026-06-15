package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/rluisb/lazyai/packages/cli/internal/runtime"
	runtimesession "github.com/rluisb/lazyai/packages/cli/internal/runtime/session"
)

// withTempDir changes to a temporary directory and restores the original
// working directory on test cleanup. Returns the temp directory path.
func withTempDir(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get wd: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origWd)
	})
	return tmpDir
}

func TestGetRuntimeDBPath(t *testing.T) {
	path := getRuntimeDBPath()

	if !strings.Contains(path, ".specify") {
		t.Errorf("expected path to contain '.specify', got: %s", path)
	}
	if !strings.HasSuffix(path, "session.db") {
		t.Errorf("expected path to end with 'session.db', got: %s", path)
	}
}

func TestOpenRuntimeDBCreatesV2SchemaOnEmptyDatabase(t *testing.T) {
	tmpDir := withTempDir(t)

	db, err := openRuntimeDB()
	if err != nil {
		t.Fatalf("openRuntimeDB failed: %v", err)
	}
	defer db.Close()

	dbPath := filepath.Join(tmpDir, ".specify", "session.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatalf("expected database file at %s", dbPath)
	}

	for _, tableName := range []string{"schema_migrations", "sessions", "dispatches", "handoff", "agent_defaults", "ledger_refs"} {
		assertRuntimeTableExists(t, db, tableName)
	}
	for _, tableName := range []string{"task_queue", "workflow_instances", "parallel_tasks", "messages", "model_calls"} {
		assertRuntimeTableMissing(t, db, tableName)
	}

	assertRuntimeSchemaVersion(t, db, runtimeSchemaVersionV2)
	assertAgentDefault(t, db, "opencode", "primary-agent", ".opencode/AGENTS.md")
	assertAgentDefault(t, db, "claude-code", "primary-agent", "CLAUDE.md")
	assertAgentDefault(t, db, "copilot", "primary-agent", ".github/copilot-instructions.md")
}

func TestOpenRuntimeDBMigratesFKSaturatedV1Database(t *testing.T) {
	withTempDir(t)
	legacyPath := seedLegacyRuntimeDB(t)

	db, err := openRuntimeDB()
	if err != nil {
		t.Fatalf("openRuntimeDB migration failed: %v", err)
	}
	defer db.Close()

	backupPath := runtimeBackupPath(legacyPath)
	if _, err := os.Stat(backupPath); err != nil {
		t.Fatalf("expected migration backup at %s: %v", backupPath, err)
	}

	for _, tableName := range []string{"schema_migrations", "sessions", "dispatches", "handoff", "agent_defaults", "ledger_refs"} {
		assertRuntimeTableExists(t, db, tableName)
	}
	for _, tableName := range []string{"task_queue", "workflow_instances", "parallel_tasks", "messages", "model_calls"} {
		assertRuntimeTableMissing(t, db, tableName)
	}

	assertRuntimeSchemaVersion(t, db, runtimeSchemaVersionV2)

	var agent, goal, summary string
	if err := db.QueryRow("SELECT agent, goal, summary FROM sessions WHERE id = ?", "ses_legacy").Scan(&agent, &goal, &summary); err != nil {
		t.Fatalf("query migrated session: %v", err)
	}
	if agent != "primary-agent" {
		t.Fatalf("session agent = %q, want primary-agent", agent)
	}
	if goal != "migrate runtime schema" {
		t.Fatalf("session goal = %q", goal)
	}
	if summary != "legacy summary" {
		t.Fatalf("session summary = %q", summary)
	}

	var dispatchAgent, dispatchTask, metadata string
	if err := db.QueryRow("SELECT agent, task FROM dispatches WHERE session_id = ?", "ses_legacy").Scan(&dispatchAgent, &dispatchTask); err != nil {
		t.Fatalf("query migrated dispatch: %v", err)
	}
	if dispatchAgent != "primary-agent" {
		t.Fatalf("dispatch agent = %q, want primary-agent", dispatchAgent)
	}
	if dispatchTask != "legacy task" {
		t.Fatalf("dispatch task = %q", dispatchTask)
	}
	if err := db.QueryRow("SELECT metadata FROM ledger_refs WHERE id = 1").Scan(&metadata); err != nil {
		t.Fatalf("query migrated ledger ref: %v", err)
	}
	if !strings.Contains(metadata, "event_hash=hash_1") || !strings.Contains(metadata, "prev_hash=prev_0") {
		t.Fatalf("ledger metadata = %q", metadata)
	}

	assertCount(t, db, "SELECT COUNT(*) FROM sessions", 1)
	assertCount(t, db, "SELECT COUNT(*) FROM dispatches", 1)
	assertCount(t, db, "SELECT COUNT(*) FROM ledger_refs", 1)
	assertAgentDefault(t, db, "opencode", "primary-agent", ".opencode/AGENTS.md")
}

func TestOpenRuntimeDBMigrationFailureKeepsBackupAndLegacyDB(t *testing.T) {
	withTempDir(t)
	legacyPath := seedLegacyRuntimeDB(t)

	migrationPath := runtimeMigrationTempPath(legacyPath)
	if err := os.MkdirAll(migrationPath, 0o755); err != nil {
		t.Fatalf("mkdir migration path: %v", err)
	}
	if err := os.WriteFile(filepath.Join(migrationPath, "blocker"), []byte("keep"), 0o644); err != nil {
		t.Fatalf("seed migration blocker: %v", err)
	}

	_, err := openRuntimeDB()
	if err == nil {
		t.Fatal("expected migration failure when migration path is a directory")
	}

	backupPath := runtimeBackupPath(legacyPath)
	if _, statErr := os.Stat(backupPath); statErr != nil {
		t.Fatalf("expected backup to exist after migration failure: %v", statErr)
	}

	legacyDB, openErr := runtime.Open(legacyPath)
	if openErr != nil {
		t.Fatalf("open legacy database after failure: %v", openErr)
	}
	defer legacyDB.Close()

	assertRuntimeTableExists(t, legacyDB, "task_queue")
	assertRuntimeTableMissing(t, legacyDB, "handoff")

	var agent string
	if err := legacyDB.QueryRow("SELECT agent FROM sessions WHERE id = ?", "ses_legacy").Scan(&agent); err != nil {
		t.Fatalf("query legacy session agent: %v", err)
	}
	if agent != "loop-driver" {
		t.Fatalf("legacy session agent = %q, want loop-driver", agent)
	}
}

func TestRuntimeDBRestoreRoundTripFromBackup(t *testing.T) {
	withTempDir(t)
	legacyPath := seedLegacyRuntimeDB(t)

	migratedDB, err := openRuntimeDB()
	if err != nil {
		t.Fatalf("openRuntimeDB migration failed: %v", err)
	}
	migratedDB.Close()

	backupPath := runtimeBackupPath(legacyPath)
	cmd := &cobra.Command{}
	cmd.Flags().Bool("force", false, "")
	if err := cmd.Flags().Set("force", "true"); err != nil {
		t.Fatalf("set force flag: %v", err)
	}

	stdout, _ := captureOutput(t, func() {
		err = runRestoreRuntimeDB(cmd, []string{backupPath})
	})
	if err != nil {
		t.Fatalf("runRestoreRuntimeDB failed: %v", err)
	}
	if !strings.Contains(stdout, "Runtime database restored") {
		t.Fatalf("stdout = %q, want restore confirmation", stdout)
	}

	restoredDB, err := runtime.Open(legacyPath)
	if err != nil {
		t.Fatalf("open restored database: %v", err)
	}
	defer restoredDB.Close()
	assertRuntimeTableExists(t, restoredDB, "task_queue")
	assertRuntimeTableMissing(t, restoredDB, "handoff")

	var restoredAgent string
	if err := restoredDB.QueryRow("SELECT agent FROM sessions WHERE id = ?", "ses_legacy").Scan(&restoredAgent); err != nil {
		t.Fatalf("query restored session: %v", err)
	}
	if restoredAgent != "loop-driver" {
		t.Fatalf("restored session agent = %q, want loop-driver", restoredAgent)
	}

	preRestoreDB, err := runtime.Open(legacyPath + ".pre-restore")
	if err != nil {
		t.Fatalf("open pre-restore database: %v", err)
	}
	defer preRestoreDB.Close()
	assertRuntimeTableExists(t, preRestoreDB, "handoff")
	assertRuntimeTableMissing(t, preRestoreDB, "task_queue")
}

func TestRuntimeV2SessionDispatchHandoffRoundTrip(t *testing.T) {
	withTempDir(t)

	db, err := openRuntimeDB()
	if err != nil {
		t.Fatalf("openRuntimeDB failed: %v", err)
	}
	defer db.Close()

	mgr := runtimesession.NewManager(db)
	s, err := mgr.Start("ship phase three", runtimesession.StartOptions{Agent: "primary-agent", Tags: []string{"phase3", "migration"}})
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if err := mgr.UpdateSummary(s.ID, "schema migrated"); err != nil {
		t.Fatalf("UpdateSummary failed: %v", err)
	}

	dispatch, err := mgr.Dispatch(s.ID, runtimesession.DispatchOptions{
		Agent:        "builder",
		Task:         "verify runtime",
		Phase:        "verification",
		FilesTouched: []string{"packages/cli/cmd/runtime_helper.go"},
	})
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}
	if err := mgr.CompleteDispatch(dispatch.ID, "done", 42); err != nil {
		t.Fatalf("CompleteDispatch failed: %v", err)
	}
	if err := mgr.End(s.ID); err != nil {
		t.Fatalf("End failed: %v", err)
	}

	handoffPath := "specs/memory/handoffs/test-phase3.md"
	if _, err := db.Exec(
		"INSERT INTO handoff (session_id, path, goal, status, created_at) VALUES (?, ?, ?, ?, ?)",
		s.ID,
		handoffPath,
		"ship phase three",
		"done",
		runtime.Now(),
	); err != nil {
		t.Fatalf("insert handoff: %v", err)
	}

	sessionRow, err := mgr.Get(s.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if sessionRow.Summary != "schema migrated" {
		t.Fatalf("summary = %q, want schema migrated", sessionRow.Summary)
	}
	if sessionRow.EndedAt == nil {
		t.Fatal("expected EndedAt to be populated")
	}
	if len(sessionRow.Tags) != 2 {
		t.Fatalf("tags = %v, want 2 tags", sessionRow.Tags)
	}

	dispatches, err := mgr.ListDispatches(s.ID)
	if err != nil {
		t.Fatalf("ListDispatches failed: %v", err)
	}
	if len(dispatches) != 1 {
		t.Fatalf("len(dispatches) = %d, want 1", len(dispatches))
	}
	if dispatches[0].Result != "done" || dispatches[0].TokenUsed != 42 {
		t.Fatalf("dispatch result = %+v", dispatches[0])
	}
	if len(dispatches[0].FilesTouched) != 1 || dispatches[0].FilesTouched[0] != "packages/cli/cmd/runtime_helper.go" {
		t.Fatalf("files_touched = %v", dispatches[0].FilesTouched)
	}

	var path, goal, status string
	if err := db.QueryRow("SELECT path, goal, status FROM handoff WHERE session_id = ?", s.ID).Scan(&path, &goal, &status); err != nil {
		t.Fatalf("query handoff: %v", err)
	}
	if path != handoffPath || goal != "ship phase three" || status != "done" {
		t.Fatalf("handoff row = (%q, %q, %q)", path, goal, status)
	}
}

func TestEnsureSession(t *testing.T) {
	withTempDir(t)

	db, err := openRuntimeDB()
	if err != nil {
		t.Fatalf("openRuntimeDB failed: %v", err)
	}
	defer db.Close()

	sessionID := "test-session"
	if err := ensureSession(db, sessionID); err != nil {
		t.Fatalf("ensureSession first call failed: %v", err)
	}
	assertCountArgs(t, db, "SELECT COUNT(*) FROM sessions WHERE id = ?", 1, sessionID)

	var status, agent string
	if err := db.QueryRow("SELECT status, agent FROM sessions WHERE id = ?", sessionID).Scan(&status, &agent); err != nil {
		t.Fatalf("query ensured session: %v", err)
	}
	if status != "active" {
		t.Fatalf("status = %q, want active", status)
	}
	if agent != "cli" {
		t.Fatalf("agent = %q, want cli", agent)
	}

	if err := ensureSession(db, sessionID); err != nil {
		t.Fatalf("ensureSession second call failed: %v", err)
	}
	assertCountArgs(t, db, "SELECT COUNT(*) FROM sessions WHERE id = ?", 1, sessionID)
}

func seedLegacyRuntimeDB(t *testing.T) string {
	t.Helper()
	dbPath := getRuntimeDBPath()
	db, err := runtime.Open(dbPath)
	if err != nil {
		t.Fatalf("runtime.Open legacy DB: %v", err)
	}

	if _, err := db.Exec(runtime.SchemaV1); err != nil {
		_ = db.Close()
		t.Fatalf("apply SchemaV1: %v", err)
	}

	seedLegacyRuntimeData(t, db)
	if err := db.Close(); err != nil {
		t.Fatalf("close legacy DB: %v", err)
	}
	return dbPath
}

func seedLegacyRuntimeData(t *testing.T, db *runtime.DB) {
	t.Helper()

	startedAt := time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC).Format(time.RFC3339)
	endedAt := time.Date(2026, 6, 14, 12, 10, 0, 0, time.UTC).Format(time.RFC3339)

	mustExec(t, db,
		"INSERT INTO sessions (id, started_at, ended_at, agent, model, goal, repo, worktree, status, token_total, summary, tags) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"ses_legacy", startedAt, endedAt, "loop-driver", "gpt-5", "migrate runtime schema", "lazyai", "worktrees/phase3", "ended", 128, "legacy summary", "phase3,legacy",
	)

	dispatchID := mustExecLastInsertID(t, db,
		"INSERT INTO dispatches (session_id, seq, agent, model, task, phase, workflow, mode, started_at, ended_at, result, token_used, error_message, summary, files_touched) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"ses_legacy", 1, "orchestrator", "gpt-5", "legacy task", "implement", "legacy-workflow", "serial", startedAt, endedAt, "ok", 64, "", "dispatch summary", "runtime_helper.go,session.go",
	)

	mustExec(t, db, "INSERT INTO decisions (session_id, dispatch_id, title, context, decision, rationale, alternatives, created_at, tags) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)", "ses_legacy", dispatchID, "keep boring migration", "phase3", "temp-db swap", "safer rollback", "in-place rewrite", startedAt, "migration")
	mustExec(t, db, "INSERT INTO artifacts (session_id, dispatch_id, path, action, created_at) VALUES (?, ?, ?, ?, ?)", "ses_legacy", dispatchID, "packages/cli/cmd/runtime_helper.go", "modified", startedAt)
	mustExec(t, db, "INSERT INTO memories (session_id, title, content, tags, importance, verified, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)", "ses_legacy", "phase3", "remember migration", "runtime", "high", 1, startedAt)
	mustExec(t, db, "INSERT INTO token_log (session_id, dispatch_id, token_count, context_pct, recorded_at) VALUES (?, ?, ?, ?, ?)", "ses_legacy", dispatchID, 64, 0.5, startedAt)
	mustExec(t, db, "INSERT INTO parallel_tasks (session_id, parent_dispatch_id, wave_id, agent, task, status, started_at, completed_at, result, output_path, error_message, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", "ses_legacy", dispatchID, "wave_1", "reviewer", "review schema", "completed", startedAt, endedAt, "clean", "report.txt", "", startedAt)
	mustExec(t, db, "INSERT INTO messages (session_id, from_agent, to_agent, subject, body, priority, status, created_at, read_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)", "ses_legacy", "planner", "builder", "phase3", "ship it", "normal", "read", startedAt, endedAt)
	mustExec(t, db, "INSERT INTO barriers (session_id, barrier_id, expected_count, arrived_count, status, created_at, resolved_at) VALUES (?, ?, ?, ?, ?, ?, ?)", "ses_legacy", "barrier_1", 2, 2, "resolved", startedAt, endedAt)
	mustExec(t, db, "INSERT INTO locks (session_id, lock_name, held_by, acquired_at, released_at, status) VALUES (?, ?, ?, ?, ?, ?)", "ses_legacy", "schema", "builder", startedAt, endedAt, "released")
	mustExec(t, db, "INSERT INTO teams (name, description, agents, created_at, updated_at) VALUES (?, ?, ?, ?, ?)", "phase3-team", "legacy team", "builder,reviewer", startedAt, endedAt)
	mustExec(t, db, "INSERT INTO workflows (id, name, description, trigger_cmd, config_json, version, created_at, updated_at, team, steps) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", "wf_legacy", "legacy workflow", "phase3 workflow", "lazyai workflow", "{}", 1, startedAt, endedAt, "phase3-team", "[]")
	workflowInstanceID := mustExecLastInsertID(t, db, "INSERT INTO workflow_instances (workflow_name, session_id, status, current_step, total_steps, result, on_failure, started_at, completed_at, error_message) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", "legacy workflow", "ses_legacy", "completed", 1, 1, "ok", "stop", startedAt, endedAt, "")
	mustExec(t, db, "INSERT INTO workflow_steps (instance_id, step_order, agent, task, mode, status, started_at, completed_at, result, output_path, error_message) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", workflowInstanceID, 1, "builder", "legacy task", "serial", "completed", startedAt, endedAt, "ok", "report.txt", "")
	mustExec(t, db, "INSERT INTO model_calls (session_id, dispatch_id, provider, model, tokens_input, tokens_output, cost, latency_ms, called_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)", "ses_legacy", dispatchID, "openai", "gpt-5", 32, 32, 0.12, 250, startedAt)
	mustExec(t, db, "INSERT INTO tool_calls (session_id, dispatch_id, tool_name, command, exit_code, output, called_at) VALUES (?, ?, ?, ?, ?, ?, ?)", "ses_legacy", dispatchID, "bash", "go test ./packages/cli/...", 0, "ok", startedAt)
	mustExec(t, db, "INSERT INTO gate_results (session_id, workflow_instance_id, phase_name, gate_type, status, human_required, human_decision, evaluated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", "ses_legacy", workflowInstanceID, "phase2", "approval", "approved", 1, "go ahead", endedAt)
	mustExec(t, db, "INSERT INTO ledger_refs (seq, session_id, workflow_run_id, event_type, event_hash, prev_hash, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)", 1, "ses_legacy", "run_1", "session_end", "hash_1", "prev_0", endedAt)
	mustExec(t, db, "INSERT INTO cost_snapshots (session_id, workflow_instance_id, total_cost, token_count, recorded_at) VALUES (?, ?, ?, ?, ?)", "ses_legacy", workflowInstanceID, 0.12, 64, endedAt)
	mustExec(t, db, "INSERT INTO checkpoints (session_id, checkpoint_name, state_json, created_at) VALUES (?, ?, ?, ?)", "ses_legacy", "before-handoff", "{}", endedAt)
	evalRunID := mustExecLastInsertID(t, db, "INSERT INTO eval_runs (session_id, suite_name, dataset_path, started_at, completed_at, status, score) VALUES (?, ?, ?, ?, ?, ?, ?)", "ses_legacy", "phase3", "dataset.json", startedAt, endedAt, "passed", 1.0)
	mustExec(t, db, "INSERT INTO eval_results (eval_run_id, assertion_name, passed, expected, actual, latency_ms, recorded_at) VALUES (?, ?, ?, ?, ?, ?, ?)", evalRunID, "schema", 1, "v2", "v2", 5, endedAt)
	taskID := mustExecLastInsertID(t, db, "INSERT INTO task_queue (session_id, topic, task, status, max_agents, dedupe_key, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)", "ses_legacy", "phase3", "migrate", "claimed", 1, "dedupe-1", startedAt)
	mustExec(t, db, "INSERT INTO task_claims (task_id, agent, claimed_at) VALUES (?, ?, ?)", taskID, "builder", startedAt)
	mustExec(t, db, "INSERT INTO task_dlq (task_id, failed_agent, error_message, context_dump, failed_at) VALUES (?, ?, ?, ?, ?)", taskID, "builder", "none", "{}", endedAt)
	mustExec(t, db, "INSERT INTO task_messages (task_id, from_agent, to_agent, body, created_at) VALUES (?, ?, ?, ?, ?)", taskID, "planner", "builder", "finish migration", startedAt)
}

func mustExec(t *testing.T, db *runtime.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.Exec(query, args...); err != nil {
		t.Fatalf("exec %q failed: %v", query, err)
	}
}

func mustExecLastInsertID(t *testing.T, db *runtime.DB, query string, args ...any) int64 {
	t.Helper()
	result, err := db.Exec(query, args...)
	if err != nil {
		t.Fatalf("exec %q failed: %v", query, err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("LastInsertId for %q failed: %v", query, err)
	}
	return id
}

func assertRuntimeTableExists(t *testing.T, db *runtime.DB, tableName string) {
	t.Helper()
	exists, err := runtimeTableExists(db, tableName)
	if err != nil {
		t.Fatalf("runtimeTableExists(%s): %v", tableName, err)
	}
	if !exists {
		t.Fatalf("expected table %q to exist", tableName)
	}
}

func assertRuntimeTableMissing(t *testing.T, db *runtime.DB, tableName string) {
	t.Helper()
	exists, err := runtimeTableExists(db, tableName)
	if err != nil {
		t.Fatalf("runtimeTableExists(%s): %v", tableName, err)
	}
	if exists {
		t.Fatalf("expected table %q to be absent", tableName)
	}
}

func assertRuntimeSchemaVersion(t *testing.T, db *runtime.DB, want int) {
	t.Helper()
	got, ok, err := currentRecordedRuntimeSchemaVersion(db)
	if err != nil {
		t.Fatalf("currentRecordedRuntimeSchemaVersion: %v", err)
	}
	if !ok {
		t.Fatal("expected schema version to be recorded")
	}
	if got != want {
		t.Fatalf("schema version = %d, want %d", got, want)
	}
}

func assertAgentDefault(t *testing.T, db *runtime.DB, toolID, wantAgent, wantInstructions string) {
	t.Helper()
	var agent, instructions string
	if err := db.QueryRow("SELECT default_agent, instructions FROM agent_defaults WHERE tool_id = ?", toolID).Scan(&agent, &instructions); err != nil {
		t.Fatalf("query agent_defaults for %s: %v", toolID, err)
	}
	if agent != wantAgent || instructions != wantInstructions {
		t.Fatalf("agent_defaults[%s] = (%q, %q), want (%q, %q)", toolID, agent, instructions, wantAgent, wantInstructions)
	}
}

func assertCount(t *testing.T, db *runtime.DB, query string, want int) {
	t.Helper()
	assertCountArgs(t, db, query, want)
}

func assertCountArgs(t *testing.T, db *runtime.DB, query string, want int, args ...any) {
	t.Helper()
	var got int
	if err := db.QueryRow(query, args...).Scan(&got); err != nil {
		t.Fatalf("count query %q failed: %v", query, err)
	}
	if got != want {
		t.Fatalf("count query %q = %d, want %d", query, got, want)
	}
}
