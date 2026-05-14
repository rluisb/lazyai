package sqlite

import (
	"context"
	"errors"
	"testing"

	"github.com/rluisb/lazyai/packages/orchestrator/domain"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/db"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
)

func TestRunReadStoreListRunsFiltersSearchPaginationAndHasErrors(t *testing.T) {
	store, database := newRunReadStoreTestAdapter(t)
	seedRunReadRun(t, database, types.RunKindChain, "chain-running", "release_100%", "1", "running", "build", `{}`, "2026-05-05T10:05:00Z")
	seedRunReadRun(t, database, types.RunKindChain, "chain-failed", "release", "1", "failed", "test", `{}`, "2026-05-05T10:04:00Z")
	seedRunReadRun(t, database, types.RunKindChain, "chain-gated", "deploy", "1", "gated", "gate", `{}`, "2026-05-05T10:03:00Z")
	seedRunReadRun(t, database, types.RunKindTeam, "team-running", "launch", "1", "running", "", `{}`, "2026-05-05T10:02:00Z")
	seedRunReadRun(t, database, types.RunKindWorkflow, "workflow-await", "ship", "1", "awaiting_recovery", "phase-1", `{}`, "2026-05-05T10:01:00Z")
	seedRunReadRun(t, database, types.RunKindChain, "chain-wildcard-decoy", "releaseX100Y", "1", "running", "build", `{}`, "2026-05-05T10:00:00Z")
	seedRunReadRun(t, database, types.RunKindChain, "chain-running-old", "legacy", "1", "running", "build", `{}`, "2026-05-05T09:59:00Z")
	seedRunReadError(t, database, "err-failed", "chain-failed", types.RunKindChain)

	page, err := store.ListRuns(context.Background(), domain.RunListFilter{Kind: types.RunKindChain, State: "running", Limit: 2})
	if err != nil {
		t.Fatalf("list filtered runs: %v", err)
	}
	if len(page.Items) != 2 || page.NextCursor != "2" {
		t.Fatalf("page/cursor mismatch: %+v", page)
	}
	if page.Items[0].ID != "chain-running" || page.Items[1].ID != "chain-wildcard-decoy" {
		t.Fatalf("filtered order mismatch: %+v", page.Items)
	}

	page, err = store.ListRuns(context.Background(), domain.RunListFilter{Kind: types.RunKindChain, State: "running", Limit: 2, Cursor: "2"})
	if err != nil {
		t.Fatalf("list cursor page: %v", err)
	}
	if len(page.Items) != 1 || page.Items[0].ID != "chain-running-old" || page.NextCursor != "" {
		t.Fatalf("second page mismatch: %+v", page)
	}

	page, err = store.ListRuns(context.Background(), domain.RunListFilter{Search: "RELEASE_100%", Limit: 10})
	if err != nil {
		t.Fatalf("search escaped wildcard: %v", err)
	}
	if len(page.Items) != 1 || page.Items[0].ID != "chain-running" {
		t.Fatalf("search should treat wildcards literally: %+v", page.Items)
	}

	page, err = store.ListRuns(context.Background(), domain.RunListFilter{Attention: "gated", Limit: 10})
	if err != nil {
		t.Fatalf("attention gated: %v", err)
	}
	got := map[string]bool{}
	for _, item := range page.Items {
		got[item.ID] = true
	}
	if !got["chain-gated"] || !got["workflow-await"] || len(page.Items) != 2 {
		t.Fatalf("attention gated mismatch: %+v", page.Items)
	}

	page, err = store.ListRuns(context.Background(), domain.RunListFilter{HasErrors: true, Limit: 10})
	if err != nil {
		t.Fatalf("has errors: %v", err)
	}
	if len(page.Items) != 1 || page.Items[0].ID != "chain-failed" {
		t.Fatalf("has errors mismatch: %+v", page.Items)
	}

	if _, err := store.ListRuns(context.Background(), domain.RunListFilter{Attention: "bogus"}); err == nil {
		t.Fatalf("expected invalid attention error")
	}
}

func TestRunReadStoreCountRunsByState(t *testing.T) {
	store, database := newRunReadStoreTestAdapter(t)
	seedRunReadRun(t, database, types.RunKindChain, "chain-running", "release", "1", "running", "build", `{}`, "2026-05-05T10:00:00Z")
	seedRunReadRun(t, database, types.RunKindTeam, "team-running", "launch", "1", "running", "", `{}`, "2026-05-05T10:01:00Z")
	seedRunReadRun(t, database, types.RunKindWorkflow, "workflow-completed", "ship", "1", "completed", "", `{}`, "2026-05-05T10:02:00Z")

	counts, err := store.CountRunsByState(context.Background())
	if err != nil {
		t.Fatalf("count runs by state: %v", err)
	}
	if counts["running"] != 2 || counts["completed"] != 1 {
		t.Fatalf("counts mismatch: %+v", counts)
	}
}

func TestRunReadStoreFindRunRowAndNotFound(t *testing.T) {
	store, database := newRunReadStoreTestAdapter(t)
	seedRunReadRun(t, database, types.RunKindWorkflow, "workflow-detail", "ship", "3", "waiting_on_child", "phase-1", `{"ok":true}`, "2026-05-05T10:00:00Z")

	row, err := store.FindRunRow(context.Background(), types.RunKindWorkflow, "workflow-detail")
	if err != nil {
		t.Fatalf("find run row: %v", err)
	}
	if row.Kind != types.RunKindWorkflow || row.ID != "workflow-detail" || row.Current != "phase-1" || row.StateJSON != `{"ok":true}` {
		t.Fatalf("row mismatch: %+v", row)
	}

	_, err = store.FindRunRow(context.Background(), types.RunKindWorkflow, "missing")
	if err == nil {
		t.Fatalf("expected missing run error")
	}
	var notFound domain.RunReadNotFoundError
	if !errors.As(err, &notFound) || notFound.Resource != "run" || notFound.ID != "workflow/missing" {
		t.Fatalf("missing run error mismatch: %T %v", err, err)
	}

	_, err = store.FindRunRow(context.Background(), types.RunKind("bogus"), "missing")
	if err == nil {
		t.Fatalf("expected invalid kind error")
	}
	if !errors.As(err, &notFound) || notFound.Resource != "run kind" || notFound.ID != "bogus" {
		t.Fatalf("invalid kind error mismatch: %T %v", err, err)
	}
}

func newRunReadStoreTestAdapter(t *testing.T) (*RunReadStore, *db.DB) {
	t.Helper()
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := database.RunMigrations(); err != nil {
		t.Fatalf("migrate db: %v", err)
	}
	return NewRunReadStore(database), database
}

func seedRunReadRun(t *testing.T, database *db.DB, kind types.RunKind, id, name, version, state, current, stateJSON, updatedAt string) {
	t.Helper()
	createdAt := "2026-05-05T09:00:00Z"
	var query string
	var args []any
	switch kind {
	case types.RunKindChain:
		query = `INSERT INTO chain_runs (id, definition_name, definition_version, state, current_step_id, project_root, state_json, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
		args = []any{id, name, version, state, current, "/repo", stateJSON, createdAt, updatedAt}
	case types.RunKindTeam:
		query = `INSERT INTO team_runs (id, definition_name, definition_version, state, project_root, state_json, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
		args = []any{id, name, version, state, "/repo", stateJSON, createdAt, updatedAt}
	case types.RunKindWorkflow:
		query = `INSERT INTO workflow_runs (id, definition_name, definition_version, state, current_phase_id, project_root, state_json, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
		args = []any{id, name, version, state, current, "/repo", stateJSON, createdAt, updatedAt}
	default:
		t.Fatalf("unsupported kind %q", kind)
	}
	if _, err := database.Exec(query, args...); err != nil {
		t.Fatalf("seed run %s: %v", id, err)
	}
}

func seedRunReadError(t *testing.T, database *db.DB, id, runID string, kind types.RunKind) {
	t.Helper()
	if _, err := database.Exec(`INSERT INTO error_journal (id, run_id, run_kind, definition_name, category, code, message, entry_json, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, id, runID, string(kind), "release", "fatal", "boom", "boom", `{}`, "2026-05-05T10:06:00Z"); err != nil {
		t.Fatalf("seed error: %v", err)
	}
}
