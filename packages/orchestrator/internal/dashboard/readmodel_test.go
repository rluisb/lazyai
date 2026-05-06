package dashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/rluisb/lazyai/packages/orchestrator/internal/db"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
)

func TestReadModelOverviewIncludesHealthRunsErrorsAndCatalogCounts(t *testing.T) {
	database := newDashboardTestDB(t)
	seedRun(t, database, types.RunKindChain, "chain-running", "release", "2", "running", "implement", chainStateJSON(t, "chain-running", "release", "2", "running", "implement"), "2026-05-05T10:00:00Z")
	seedRun(t, database, types.RunKindTeam, "team-completed", "launch", "1", "completed", "", `{}`, "2026-05-05T10:01:00Z")
	seedRun(t, database, types.RunKindWorkflow, "workflow-waiting", "ship", "3", "waiting_on_child", "phase-1", `{}`, "2026-05-05T10:02:00Z")
	seedError(t, database, "err-1", "chain-running", types.RunKindChain, "release", "implement", "transient", "dispatch_failed", "dispatch failed", "2026-05-05T10:03:00Z")

	model := NewReadModel(database)
	overview, err := model.Overview(context.Background(), HealthView{Status: "ok", Name: "lazyai-orchestrator"}, CatalogCounts{Total: 2, ByKind: map[string]int{"chain": 1, "team": 1}})
	if err != nil {
		t.Fatalf("overview: %v", err)
	}

	if overview.Health.Status != "ok" {
		t.Fatalf("health not preserved: %+v", overview.Health)
	}
	if overview.ActiveRuns.Total != 2 || overview.ActiveRuns.Chains != 1 || overview.ActiveRuns.Workflows != 1 {
		t.Fatalf("active run counts mismatch: %+v", overview.ActiveRuns)
	}
	if overview.RunCountsByState["running"] != 1 || overview.RunCountsByState["completed"] != 1 || overview.RunCountsByState["waiting_on_child"] != 1 {
		t.Fatalf("run counts by state mismatch: %+v", overview.RunCountsByState)
	}
	if len(overview.RecentRuns) != 3 || overview.RecentRuns[0].ID != "workflow-waiting" {
		t.Fatalf("recent runs not ordered newest first: %+v", overview.RecentRuns)
	}
	if len(overview.RecentErrors) != 1 || overview.RecentErrors[0].Code != "dispatch_failed" {
		t.Fatalf("recent errors mismatch: %+v", overview.RecentErrors)
	}
	if overview.CatalogCounts.Total != 2 || overview.GeneratedAt == "" {
		t.Fatalf("catalog/generated fields mismatch: %+v", overview)
	}
}

func TestReadModelRunListAppliesFiltersBoundsAndCursor(t *testing.T) {
	database := newDashboardTestDB(t)
	for i := 0; i < 205; i++ {
		id := fmt.Sprintf("chain-%03d", i)
		seedRun(t, database, types.RunKindChain, id, "release", "1", "running", "step", `{}`, fmt.Sprintf("2026-05-05T10:%02d:%02dZ", i/60, i%60))
	}
	seedRun(t, database, types.RunKindTeam, "team-done", "launch", "1", "completed", "", `{}`, "2026-05-05T11:00:00Z")

	model := NewReadModel(database)
	page, err := model.ListRuns(context.Background(), RunListOptions{Kind: types.RunKindChain, State: "running"})
	if err != nil {
		t.Fatalf("list runs: %v", err)
	}
	if len(page.Items) != DefaultRunLimit || page.NextCursor != "50" {
		t.Fatalf("default limit/cursor mismatch: len=%d cursor=%q", len(page.Items), page.NextCursor)
	}
	if page.Items[0].ID != "chain-204" {
		t.Fatalf("runs not ordered by updated_at desc: first=%s", page.Items[0].ID)
	}

	page, err = model.ListRuns(context.Background(), RunListOptions{Kind: types.RunKindChain, State: "running", Limit: 500})
	if err != nil {
		t.Fatalf("list runs max: %v", err)
	}
	if len(page.Items) != MaxRunLimit || page.NextCursor != "200" {
		t.Fatalf("max limit/cursor mismatch: len=%d cursor=%q", len(page.Items), page.NextCursor)
	}

	page, err = model.ListRuns(context.Background(), RunListOptions{Kind: types.RunKindChain, State: "running", Limit: 10, Cursor: "10"})
	if err != nil {
		t.Fatalf("list runs cursor: %v", err)
	}
	if len(page.Items) != 10 || page.Items[0].ID != "chain-194" || page.NextCursor != "20" {
		t.Fatalf("cursor page mismatch: %+v", page)
	}
}

func TestReadModelRunListAttentionSearchAndHasErrors(t *testing.T) {
	database := newDashboardTestDB(t)
	seedRun(t, database, types.RunKindChain, "chain-running", "release", "1", "running", "build", `{}`, "2026-05-05T10:05:00Z")
	seedRun(t, database, types.RunKindChain, "chain-failed", "release", "1", "failed", "build", `{}`, "2026-05-05T10:04:00Z")
	seedRun(t, database, types.RunKindChain, "chain-gated", "release", "1", "gated", "build", `{}`, "2026-05-05T10:03:00Z")
	seedRun(t, database, types.RunKindChain, "chain-paused", "release", "1", "paused", "build", `{}`, "2026-05-05T10:02:00Z")
	seedRun(t, database, types.RunKindChain, "chain-completed", "launch", "1", "completed", "build", `{}`, "2026-05-05T10:01:00Z")
	seedRun(t, database, types.RunKindWorkflow, "workflow-await", "ship", "1", "awaiting_recovery", "phase-1", `{}`, "2026-05-05T10:00:00Z")
	seedError(t, database, "err-failed-1", "chain-failed", types.RunKindChain, "release", "build", "fatal", "boom", "boom", "2026-05-05T10:04:30Z")

	// "recent" attention should match runs updated within the last hour.
	recentTime := time.Now().UTC().Add(-30 * time.Minute).Format(time.RFC3339)
	seedRun(t, database, types.RunKindChain, "chain-recent", "release", "1", "running", "build", `{}`, recentTime)

	model := NewReadModel(database)

	// Attention=running matches only running-state runs.
	page, err := model.ListRuns(context.Background(), RunListOptions{Attention: "running"})
	if err != nil {
		t.Fatalf("attention=running: %v", err)
	}
	for _, item := range page.Items {
		if item.State != "running" {
			t.Fatalf("attention=running returned non-running run %q (state=%s)", item.ID, item.State)
		}
	}

	// Attention=failed matches failed only.
	page, err = model.ListRuns(context.Background(), RunListOptions{Attention: "failed"})
	if err != nil {
		t.Fatalf("attention=failed: %v", err)
	}
	if len(page.Items) != 1 || page.Items[0].ID != "chain-failed" {
		t.Fatalf("attention=failed mismatch: %+v", page.Items)
	}

	// Attention=gated matches gated/paused/awaiting_recovery/waiting_on_child.
	page, err = model.ListRuns(context.Background(), RunListOptions{Attention: "gated"})
	if err != nil {
		t.Fatalf("attention=gated: %v", err)
	}
	got := map[string]bool{}
	for _, item := range page.Items {
		got[item.ID] = true
	}
	for _, want := range []string{"chain-gated", "chain-paused", "workflow-await"} {
		if !got[want] {
			t.Fatalf("attention=gated missing %s in %v", want, got)
		}
	}

	// Attention=recent only matches runs updated within the last hour.
	page, err = model.ListRuns(context.Background(), RunListOptions{Attention: "recent"})
	if err != nil {
		t.Fatalf("attention=recent: %v", err)
	}
	if len(page.Items) != 1 || page.Items[0].ID != "chain-recent" {
		t.Fatalf("attention=recent mismatch: %+v", page.Items)
	}

	// Search matches against id and definition name (case-insensitive).
	page, err = model.ListRuns(context.Background(), RunListOptions{Search: "LAUNCH"})
	if err != nil {
		t.Fatalf("search by definition name: %v", err)
	}
	if len(page.Items) != 1 || page.Items[0].ID != "chain-completed" {
		t.Fatalf("search by definition name mismatch: %+v", page.Items)
	}
	page, err = model.ListRuns(context.Background(), RunListOptions{Search: "GATED"})
	if err != nil {
		t.Fatalf("search by id: %v", err)
	}
	if len(page.Items) != 1 || page.Items[0].ID != "chain-gated" {
		t.Fatalf("search by id mismatch: %+v", page.Items)
	}

	// HasErrors limits to runs with at least one error journal entry.
	page, err = model.ListRuns(context.Background(), RunListOptions{HasErrors: true})
	if err != nil {
		t.Fatalf("has_errors: %v", err)
	}
	if len(page.Items) != 1 || page.Items[0].ID != "chain-failed" || page.Items[0].ErrorCount != 1 {
		t.Fatalf("has_errors mismatch: %+v", page.Items)
	}
}

func TestReadModelRunListSummariesIncludeBulkLoadedErrorsAndBudgets(t *testing.T) {
	database := newDashboardTestDB(t)
	seedExecutionPlan(t, database, "plan-warning", types.BudgetPolicy{ID: "policy-warning", Scope: "chain", Tokens: &types.BudgetThreshold{Limit: 100, WarnAt: 50}, DefaultActionOnLimit: "pause"})
	seedExecutionPlan(t, database, "plan-ok", types.BudgetPolicy{ID: "policy-ok", Scope: "chain", Tokens: &types.BudgetThreshold{Limit: 100, WarnAt: 50}, DefaultActionOnLimit: "pause"})
	seedRun(t, database, types.RunKindChain, "chain-warning", "release", "1", "running", "build", chainBudgetStateJSON(t, "chain-warning", "plan-warning", 75), "2026-05-05T10:03:00Z")
	seedRun(t, database, types.RunKindChain, "chain-ok", "release", "1", "running", "test", chainBudgetStateJSON(t, "chain-ok", "plan-ok", 10), "2026-05-05T10:02:00Z")
	seedRun(t, database, types.RunKindTeam, "team-no-errors", "launch", "1", "running", "", teamBudgetStateJSON(t, "team-no-errors", 25), "2026-05-05T10:01:00Z")
	seedError(t, database, "err-warning-1", "chain-warning", types.RunKindChain, "release", "build", "transient", "one", "one", "2026-05-05T10:04:00Z")
	seedError(t, database, "err-warning-2", "chain-warning", types.RunKindChain, "release", "build", "fatal", "two", "two", "2026-05-05T10:05:00Z")
	seedError(t, database, "err-ok", "chain-ok", types.RunKindChain, "release", "test", "transient", "three", "three", "2026-05-05T10:06:00Z")

	model := NewReadModel(database)
	page, err := model.ListRuns(context.Background(), RunListOptions{Limit: 10})
	if err != nil {
		t.Fatalf("list runs: %v", err)
	}

	itemsByID := map[string]RunSummary{}
	for _, item := range page.Items {
		itemsByID[item.ID] = item
	}
	if itemsByID["chain-warning"].ErrorCount != 2 || itemsByID["chain-warning"].BudgetHealth != string(types.HealthWarning) {
		t.Fatalf("warning chain summary mismatch: %+v", itemsByID["chain-warning"])
	}
	if itemsByID["chain-ok"].ErrorCount != 1 || itemsByID["chain-ok"].BudgetHealth != string(types.HealthOK) {
		t.Fatalf("ok chain summary mismatch: %+v", itemsByID["chain-ok"])
	}
	if itemsByID["team-no-errors"].ErrorCount != 0 || itemsByID["team-no-errors"].BudgetHealth != string(types.HealthOK) {
		t.Fatalf("team summary mismatch: %+v", itemsByID["team-no-errors"])
	}
}

func TestReadModelRunDetailBudgetAndErrorsTolerateMalformedState(t *testing.T) {
	database := newDashboardTestDB(t)
	seedRun(t, database, types.RunKindChain, "chain-bad", "broken", "1", "running", "step", `{`, "2026-05-05T10:00:00Z")
	seedEvent(t, database, "chain-bad", "step_started", `{"stepId":"step"}`, "2026-05-05T10:01:00Z")
	seedError(t, database, "err-bad", "chain-bad", types.RunKindChain, "broken", "step", "fatal", "bad_state", "bad state", "2026-05-05T10:02:00Z")

	model := NewReadModel(database)
	detail, err := model.GetRunDetail(context.Background(), types.RunKindChain, "chain-bad")
	if err != nil {
		t.Fatalf("detail should not fail on malformed state_json: %v", err)
	}
	if detail.Summary.ID != "chain-bad" || detail.StateDecodeError == "" {
		t.Fatalf("detail did not expose malformed state safely: %+v", detail)
	}
	if len(detail.Events) != 1 || len(detail.Errors) != 1 {
		t.Fatalf("detail did not include events/errors: %+v", detail)
	}

	budget, err := model.GetBudget(context.Background(), types.RunKindChain, "chain-bad")
	if err != nil {
		t.Fatalf("budget should not fail on malformed state_json: %v", err)
	}
	if budget.State != nil || budget.DecodeError == "" {
		t.Fatalf("budget did not degrade gracefully: %+v", budget)
	}
}

func TestReadModelBudgetUsesDecodedStateAndPolicyWhenAvailable(t *testing.T) {
	database := newDashboardTestDB(t)
	state := types.WorkflowState{
		WorkflowID:        "workflow-budget",
		DefinitionName:    "ship",
		DefinitionVersion: "1",
		State:             types.WorkflowRunning,
		BudgetPolicy: types.BudgetPolicy{
			ID: "workflow-policy", Scope: "workflow", Tokens: &types.BudgetThreshold{Limit: 100, WarnAt: 50}, DefaultActionOnLimit: "pause",
		},
		Budget: types.BudgetState{PolicyID: "workflow-policy", Scope: "workflow", Tokens: types.BudgetDimensionState{Limit: 100, Consumed: 75}, ByStep: map[string]types.StepUsage{"phase-1": {TotalTokens: 75}}, LastUpdatedAt: "2026-05-05T10:00:00Z"},
	}
	encoded, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("marshal workflow state: %v", err)
	}
	seedRun(t, database, types.RunKindWorkflow, "workflow-budget", "ship", "1", "running", "phase-1", string(encoded), "2026-05-05T10:00:00Z")

	model := NewReadModel(database)
	budget, err := model.GetBudget(context.Background(), types.RunKindWorkflow, "workflow-budget")
	if err != nil {
		t.Fatalf("budget: %v", err)
	}
	if budget.State == nil || budget.State.PolicyID != "workflow-policy" {
		t.Fatalf("budget state mismatch: %+v", budget)
	}
	if budget.Evaluation == nil || budget.Evaluation.Overall != types.HealthWarning {
		t.Fatalf("budget evaluation mismatch: %+v", budget.Evaluation)
	}
	if budget.ByStep["phase-1"].TotalTokens != 75 || budget.LastUpdatedAt != "2026-05-05T10:00:00Z" {
		t.Fatalf("budget byStep/lastUpdated mismatch: %+v", budget)
	}
}

func TestReadModelErrorsAreBoundedAndFilterable(t *testing.T) {
	database := newDashboardTestDB(t)
	for i := 0; i < 105; i++ {
		seedError(t, database, fmt.Sprintf("err-%03d", i), "chain-errors", types.RunKindChain, "release", "step", "transient", "code", "message", fmt.Sprintf("2026-05-05T10:%02d:%02dZ", i/60, i%60))
	}
	seedError(t, database, "err-other", "team-errors", types.RunKindTeam, "team", "", "fatal", "other", "other", "2026-05-05T12:00:00Z")

	model := NewReadModel(database)
	entries, err := model.ListErrors(context.Background(), ErrorListOptions{Limit: 500})
	if err != nil {
		t.Fatalf("list errors: %v", err)
	}
	if len(entries) != MaxErrorLimit || entries[0].ID != "err-other" {
		t.Fatalf("max errors/order mismatch: len=%d first=%s", len(entries), entries[0].ID)
	}

	entries, err = model.ListErrors(context.Background(), ErrorListOptions{RunID: "chain-errors", Limit: 10})
	if err != nil {
		t.Fatalf("list filtered errors: %v", err)
	}
	if len(entries) != 10 {
		t.Fatalf("filtered errors length = %d, want 10", len(entries))
	}
	for _, entry := range entries {
		if entry.RunID != "chain-errors" {
			t.Fatalf("unexpected filtered entry: %+v", entry)
		}
	}
}

func newDashboardTestDB(t *testing.T) *db.DB {
	t.Helper()
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := database.RunMigrations(); err != nil {
		t.Fatalf("migrate db: %v", err)
	}
	return database
}

func seedRun(t *testing.T, database *db.DB, kind types.RunKind, id, name, version, state, current, stateJSON, updatedAt string) {
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

func seedEvent(t *testing.T, database *db.DB, runID, eventType, eventJSON, createdAt string) {
	t.Helper()
	if _, err := database.Exec(`INSERT INTO run_events (run_id, event_type, event_json, created_at) VALUES (?, ?, ?, ?)`, runID, eventType, eventJSON, createdAt); err != nil {
		t.Fatalf("seed event: %v", err)
	}
}

func seedError(t *testing.T, database *db.DB, id, runID string, kind types.RunKind, definitionName, stepID, category, code, message, createdAt string) {
	t.Helper()
	entry := map[string]any{"id": id, "error": map[string]any{"category": category, "code": code, "message": message}, "timestamp": createdAt}
	entryJSON, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("marshal error entry: %v", err)
	}
	if _, err := database.Exec(`INSERT INTO error_journal (id, run_id, run_kind, definition_name, step_id, category, code, message, entry_json, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, id, runID, string(kind), definitionName, stepID, category, code, message, string(entryJSON), createdAt); err != nil {
		t.Fatalf("seed error: %v", err)
	}
}

func seedExecutionPlan(t *testing.T, database *db.DB, id string, policy types.BudgetPolicy) {
	t.Helper()
	planJSON, err := json.Marshal(types.ExecutionPlan{ID: id, Kind: "chain", BudgetPolicy: policy})
	if err != nil {
		t.Fatalf("marshal execution plan: %v", err)
	}
	if _, err := database.Exec(`INSERT INTO execution_plans (id, kind, definition_name, definition_version, project_root, plan_json, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, id, "chain", "release", "1", "/repo", string(planJSON), "2026-05-05T09:30:00Z"); err != nil {
		t.Fatalf("seed execution plan: %v", err)
	}
}

func chainStateJSON(t *testing.T, id, name, version, state, current string) string {
	t.Helper()
	encoded, err := json.Marshal(types.ChainState{
		ChainID:           id,
		DefinitionName:    name,
		DefinitionVersion: version,
		State:             types.ChainLifecycleState(state),
		CurrentStepID:     current,
		Steps:             []types.StepState{{StepID: current, State: types.StepRunning}},
		Budget:            types.BudgetState{PolicyID: "default", Scope: "chain", ByStep: map[string]types.StepUsage{}},
	})
	if err != nil {
		t.Fatalf("marshal chain state: %v", err)
	}
	return string(encoded)
}

func chainBudgetStateJSON(t *testing.T, id, planID string, consumedTokens int) string {
	t.Helper()
	encoded, err := json.Marshal(types.ChainState{
		ChainID:           id,
		DefinitionName:    "release",
		DefinitionVersion: "1",
		ExecutionPlanID:   planID,
		State:             types.ChainRunning,
		CurrentStepID:     "build",
		Budget:            types.BudgetState{PolicyID: "policy", Scope: "chain", Tokens: types.BudgetDimensionState{Limit: 100, Consumed: consumedTokens}, ByStep: map[string]types.StepUsage{}},
	})
	if err != nil {
		t.Fatalf("marshal chain budget state: %v", err)
	}
	return string(encoded)
}

func teamBudgetStateJSON(t *testing.T, id string, consumedTokens int) string {
	t.Helper()
	encoded, err := json.Marshal(types.TeamState{
		TeamID: id,
		State:  types.TeamRunning,
		BudgetPolicy: types.BudgetPolicy{
			ID: "team-policy", Scope: "team", Tokens: &types.BudgetThreshold{Limit: 100, WarnAt: 50}, DefaultActionOnLimit: "pause",
		},
		Budget: types.BudgetState{PolicyID: "team-policy", Scope: "team", Tokens: types.BudgetDimensionState{Limit: 100, Consumed: consumedTokens}, ByStep: map[string]types.StepUsage{}},
	})
	if err != nil {
		t.Fatalf("marshal team budget state: %v", err)
	}
	return string(encoded)
}
