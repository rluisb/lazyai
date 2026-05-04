package db

import "testing"

func TestActiveRunCountsCountsOnlyActiveStates(t *testing.T) {
	database, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := database.RunMigrations(); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	mustExec(t, database, `INSERT INTO chain_runs (id, definition_name, state, project_root, state_json, created_at, updated_at) VALUES ('chain-active', 'demo', 'running', '/', '{}', 'now', 'now')`)
	mustExec(t, database, `INSERT INTO chain_runs (id, definition_name, state, project_root, state_json, created_at, updated_at) VALUES ('chain-done', 'demo', 'completed', '/', '{}', 'now', 'now')`)
	mustExec(t, database, `INSERT INTO team_runs (id, definition_name, state, project_root, state_json, created_at, updated_at) VALUES ('team-active', 'demo', 'synthesizing', '/', '{}', 'now', 'now')`)
	mustExec(t, database, `INSERT INTO workflow_runs (id, definition_name, state, project_root, state_json, created_at, updated_at) VALUES ('workflow-active', 'demo', 'waiting_on_child', '/', '{}', 'now', 'now')`)
	mustExec(t, database, `INSERT INTO queue_jobs (id, job_type, status, payload_json, created_at) VALUES ('queue-active', 'demo', 'pending', '{}', 'now')`)
	mustExec(t, database, `INSERT INTO queue_jobs (id, job_type, status, payload_json, created_at) VALUES ('queue-done', 'demo', 'completed', '{}', 'now')`)

	counts, err := database.ActiveRunCounts()
	if err != nil {
		t.Fatalf("active counts: %v", err)
	}
	if counts.Chains != 1 || counts.Teams != 1 || counts.Workflows != 1 || counts.QueueJobs != 1 || counts.Total != 4 {
		t.Fatalf("unexpected active counts: %+v", counts)
	}
}

func mustExec(t *testing.T, database *DB, query string) {
	t.Helper()
	if _, err := database.Exec(query); err != nil {
		t.Fatalf("exec %q: %v", query, err)
	}
}
