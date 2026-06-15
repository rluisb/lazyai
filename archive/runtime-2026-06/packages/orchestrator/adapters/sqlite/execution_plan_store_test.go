package sqlite

import (
	"testing"

	"github.com/rluisb/lazyai/packages/orchestrator/internal/db"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
)

func TestExecutionPlanStorePersistsLoadsAndUpdatesPlans(t *testing.T) {
	store := newExecutionPlanStoreTestAdapter(t)

	plan := &types.ExecutionPlan{
		ID:        "plan-1",
		Kind:      string(types.RunKindChain),
		Task:      "extract a seam",
		CreatedAt: "2026-05-13T10:00:00Z",
		Definition: types.DefinitionRef{
			Kind:    string(types.KindChain),
			Name:    "hexagonal-phase",
			Version: "1.0.0",
		},
		Project: types.ProjectStackContext{RootPath: "/workspace/project"},
		CompiledSteps: []types.CompiledStepPlan{
			{ID: "research", Agent: "implementor", Instructions: "research first"},
		},
	}

	if err := store.SaveExecutionPlan(plan); err != nil {
		t.Fatalf("save execution plan: %v", err)
	}

	loaded, err := store.LoadExecutionPlan(plan.ID)
	if err != nil {
		t.Fatalf("load execution plan: %v", err)
	}
	if loaded.ID != plan.ID || loaded.Definition.Name != plan.Definition.Name || loaded.Project.RootPath != plan.Project.RootPath {
		t.Fatalf("loaded plan did not round-trip: %+v", loaded)
	}
	if len(loaded.CompiledSteps) != 1 || loaded.CompiledSteps[0].ID != "research" {
		t.Fatalf("loaded compiled steps did not round-trip: %+v", loaded.CompiledSteps)
	}

	plan.Task = "extract an execution plan seam"
	plan.CompiledSteps = append(plan.CompiledSteps, types.CompiledStepPlan{ID: "implement", Agent: "implementor", Instructions: "implement seam"})
	if err := store.SaveExecutionPlan(plan); err != nil {
		t.Fatalf("update execution plan: %v", err)
	}

	updated, err := store.LoadExecutionPlan(plan.ID)
	if err != nil {
		t.Fatalf("load updated execution plan: %v", err)
	}
	if updated.Task != "extract an execution plan seam" || len(updated.CompiledSteps) != 2 {
		t.Fatalf("updated plan did not round-trip: %+v", updated)
	}
}

func TestExecutionPlanStoreReturnsNotFound(t *testing.T) {
	store := newExecutionPlanStoreTestAdapter(t)

	if _, err := store.LoadExecutionPlan("missing-plan"); err == nil {
		t.Fatal("expected missing execution plan to return an error")
	}
}

func newExecutionPlanStoreTestAdapter(t *testing.T) *ExecutionPlanStore {
	t.Helper()
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := database.RunMigrations(); err != nil {
		t.Fatalf("migrate db: %v", err)
	}
	return NewExecutionPlanStore(database)
}
