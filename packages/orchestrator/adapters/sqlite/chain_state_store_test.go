package sqlite

import (
	"testing"

	"github.com/rluisb/lazyai/packages/orchestrator/internal/db"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
)

func TestChainStateStorePersistsLoadsAndUpdatesChainState(t *testing.T) {
	store := newChainStateStoreTestAdapter(t)

	chainState := &types.ChainState{
		ChainID:           "chain-1",
		DefinitionName:    "hexagonal-phase",
		DefinitionVersion: "1.0.0",
		ExecutionPlanID:   "plan-1",
		State:             types.ChainRunning,
		Task:              "extract a chain state seam",
		CurrentStepID:     "research",
		EntryStepID:       "research",
		Steps: []types.StepState{
			{StepID: "research", State: types.StepRunning},
		},
		Budget:    types.BudgetState{PolicyID: "policy-1", Scope: "chain"},
		CreatedAt: "2026-05-13T10:00:00Z",
		UpdatedAt: "2026-05-13T10:00:00Z",
	}

	if err := store.SaveChainState("/workspace/project", chainState); err != nil {
		t.Fatalf("save chain state: %v", err)
	}

	loaded, err := store.LoadChainState(chainState.ChainID)
	if err != nil {
		t.Fatalf("load chain state: %v", err)
	}
	if loaded.ChainID != chainState.ChainID || loaded.DefinitionName != chainState.DefinitionName || loaded.CurrentStepID != "research" {
		t.Fatalf("loaded chain state did not round-trip: %+v", loaded)
	}
	if len(loaded.Steps) != 1 || loaded.Steps[0].StepID != "research" {
		t.Fatalf("loaded steps did not round-trip: %+v", loaded.Steps)
	}

	chainState.CurrentStepID = "implement"
	chainState.CompletedStepIDs = []string{"research"}
	chainState.UpdatedAt = "2026-05-13T10:01:00Z"
	if err := store.SaveChainState("/workspace/project", chainState); err != nil {
		t.Fatalf("update chain state: %v", err)
	}

	updated, err := store.LoadChainState(chainState.ChainID)
	if err != nil {
		t.Fatalf("load updated chain state: %v", err)
	}
	if updated.CurrentStepID != "implement" || len(updated.CompletedStepIDs) != 1 {
		t.Fatalf("updated chain state did not round-trip: %+v", updated)
	}
}

func TestChainStateStoreReturnsNotFound(t *testing.T) {
	store := newChainStateStoreTestAdapter(t)

	if _, err := store.LoadChainState("missing-chain"); err == nil {
		t.Fatal("expected missing chain state to return an error")
	}
}

func newChainStateStoreTestAdapter(t *testing.T) *ChainStateStore {
	t.Helper()
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := database.RunMigrations(); err != nil {
		t.Fatalf("migrate db: %v", err)
	}
	return NewChainStateStore(database)
}
