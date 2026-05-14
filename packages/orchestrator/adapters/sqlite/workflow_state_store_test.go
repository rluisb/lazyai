package sqlite

import (
	"testing"

	"github.com/rluisb/lazyai/packages/orchestrator/internal/db"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
)

func TestWorkflowStateStorePersistsLoadsAndUpdatesWorkflowState(t *testing.T) {
	store := newWorkflowStateStoreTestAdapter(t)

	workflowState := &types.WorkflowState{
		WorkflowID:        "workflow-1",
		DefinitionName:    "hexagonal-phase",
		DefinitionVersion: "1.0.0",
		ExecutionPlanID:   "plan-1",
		State:             types.WorkflowRunning,
		Task:              "extract a workflow state seam",
		EntryPhaseID:      "research",
		CurrentPhaseID:    "research",
		Phases: []types.WorkflowPhaseState{
			{PhaseID: "research", Kind: "chain", Ref: "research-chain", State: types.PhaseRunning},
			{PhaseID: "ship", Kind: "terminal", State: types.PhasePending},
		},
		ChildRuns: []types.WorkflowChildRun{
			{PhaseID: "research", RunID: "chain-1", RunKind: "chain", DefinitionName: "research-chain", LaunchedAt: "2026-05-13T10:00:30Z"},
		},
		Budget:    types.BudgetState{PolicyID: "policy-1", Scope: "workflow"},
		CreatedAt: "2026-05-13T10:00:00Z",
		UpdatedAt: "2026-05-13T10:00:00Z",
	}

	if err := store.SaveWorkflowState("/workspace/project", workflowState); err != nil {
		t.Fatalf("save workflow state: %v", err)
	}

	loaded, err := store.LoadWorkflowState(workflowState.WorkflowID)
	if err != nil {
		t.Fatalf("load workflow state: %v", err)
	}
	if loaded.WorkflowID != workflowState.WorkflowID || loaded.DefinitionName != workflowState.DefinitionName || loaded.CurrentPhaseID != "research" {
		t.Fatalf("loaded workflow state did not round-trip: %+v", loaded)
	}
	if len(loaded.Phases) != 2 || loaded.Phases[0].PhaseID != "research" {
		t.Fatalf("loaded phases did not round-trip: %+v", loaded.Phases)
	}
	if len(loaded.ChildRuns) != 1 || loaded.ChildRuns[0].RunID != "chain-1" {
		t.Fatalf("loaded child runs did not round-trip: %+v", loaded.ChildRuns)
	}

	workflowState.State = types.WorkflowCompleted
	workflowState.CurrentPhaseID = "ship"
	workflowState.Phases[0].State = types.PhaseCompleted
	workflowState.Phases[1].State = types.PhaseCompleted
	workflowState.UpdatedAt = "2026-05-13T10:01:00Z"
	if err := store.SaveWorkflowState("/workspace/project", workflowState); err != nil {
		t.Fatalf("update workflow state: %v", err)
	}

	updated, err := store.LoadWorkflowState(workflowState.WorkflowID)
	if err != nil {
		t.Fatalf("load updated workflow state: %v", err)
	}
	if updated.State != types.WorkflowCompleted || updated.CurrentPhaseID != "ship" || updated.Phases[1].State != types.PhaseCompleted {
		t.Fatalf("updated workflow state did not round-trip: %+v", updated)
	}
}

func TestWorkflowStateStoreReturnsNotFound(t *testing.T) {
	store := newWorkflowStateStoreTestAdapter(t)

	if _, err := store.LoadWorkflowState("missing-workflow"); err == nil {
		t.Fatal("expected missing workflow state to return an error")
	}
}

func newWorkflowStateStoreTestAdapter(t *testing.T) *WorkflowStateStore {
	t.Helper()
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := database.RunMigrations(); err != nil {
		t.Fatalf("migrate db: %v", err)
	}
	return NewWorkflowStateStore(database)
}
