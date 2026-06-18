package sqlite

import (
	"context"
	"testing"

	"github.com/rluisb/lazyai/packages/orchestrator/internal/db"
)

func TestActivityStoreCountsOnlyActiveStates(t *testing.T) {
	store, database := newActivityStoreTestAdapter(t)

	mustExecActivity(t, database, `INSERT INTO chain_runs (id, definition_name, state, project_root, state_json, created_at, updated_at) VALUES ('chain-active', 'demo', 'running', '/', '{}', 'now', 'now')`)
	mustExecActivity(t, database, `INSERT INTO chain_runs (id, definition_name, state, project_root, state_json, created_at, updated_at) VALUES ('chain-done', 'demo', 'completed', '/', '{}', 'now', 'now')`)
	mustExecActivity(t, database, `INSERT INTO team_runs (id, definition_name, state, project_root, state_json, created_at, updated_at) VALUES ('team-active', 'demo', 'synthesizing', '/', '{}', 'now', 'now')`)
	mustExecActivity(t, database, `INSERT INTO workflow_runs (id, definition_name, state, project_root, state_json, created_at, updated_at) VALUES ('workflow-active', 'demo', 'waiting_on_child', '/', '{}', 'now', 'now')`)
	mustExecActivity(t, database, `INSERT INTO queue_jobs (id, job_type, status, payload_json, created_at) VALUES ('queue-active', 'demo', 'pending', '{}', 'now')`)
	mustExecActivity(t, database, `INSERT INTO queue_jobs (id, job_type, status, payload_json, created_at) VALUES ('queue-done', 'demo', 'completed', '{}', 'now')`)

	counts, err := store.ActiveRunCounts(context.Background())
	if err != nil {
		t.Fatalf("active counts: %v", err)
	}
	if counts.Chains != 1 || counts.Teams != 1 || counts.Workflows != 1 || counts.QueueJobs != 1 || counts.Total != 4 {
		t.Fatalf("unexpected active counts: %+v", counts)
	}
}

func newActivityStoreTestAdapter(t *testing.T) (*ActivityStore, *db.DB) {
	t.Helper()
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := database.RunMigrations(); err != nil {
		t.Fatalf("migrate db: %v", err)
	}
	return NewActivityStore(database), database
}

func mustExecActivity(t *testing.T, database *db.DB, query string) {
	t.Helper()
	if _, err := database.Exec(query); err != nil {
		t.Fatalf("exec %q: %v", query, err)
	}
}
