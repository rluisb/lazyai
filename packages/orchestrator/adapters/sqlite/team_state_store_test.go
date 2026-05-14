package sqlite

import (
	"testing"

	"github.com/rluisb/lazyai/packages/orchestrator/internal/db"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
)

func TestTeamStateStorePersistsLoadsAndUpdatesTeamState(t *testing.T) {
	store := newTeamStateStoreTestAdapter(t)

	teamState := &types.TeamState{
		TeamID:            "team-1",
		DefinitionName:    "hexagonal-phase",
		DefinitionVersion: "1.0.0",
		ExecutionPlanID:   "plan-1",
		State:             types.TeamRunning,
		Task:              "extract a team state seam",
		Tasks: []types.TeamTaskState{
			{TaskID: "researcher-0", Kind: "member", Role: "researcher", Agent: "researcher", State: types.TaskPending},
			{TaskID: "synthesize", Kind: "synthesize", Role: "synthesizer", Agent: "builder", State: types.TaskBlocked, DependsOn: []string{"researcher-0"}},
		},
		ReadyTaskIDs:    []string{"researcher-0"},
		SynthesisTaskID: "synthesize",
		Budget:          types.BudgetState{PolicyID: "policy-1", Scope: "team"},
		CreatedAt:       "2026-05-13T10:00:00Z",
		UpdatedAt:       "2026-05-13T10:00:00Z",
	}

	if err := store.SaveTeamState("/workspace/project", teamState); err != nil {
		t.Fatalf("save team state: %v", err)
	}

	loaded, err := store.LoadTeamState(teamState.TeamID)
	if err != nil {
		t.Fatalf("load team state: %v", err)
	}
	if loaded.TeamID != teamState.TeamID || loaded.DefinitionName != teamState.DefinitionName || loaded.SynthesisTaskID != "synthesize" {
		t.Fatalf("loaded team state did not round-trip: %+v", loaded)
	}
	if len(loaded.Tasks) != 2 || loaded.Tasks[0].TaskID != "researcher-0" {
		t.Fatalf("loaded tasks did not round-trip: %+v", loaded.Tasks)
	}

	teamState.State = types.TeamSynthesizing
	teamState.ReadyTaskIDs = []string{"synthesize"}
	teamState.UpdatedAt = "2026-05-13T10:01:00Z"
	if err := store.SaveTeamState("/workspace/project", teamState); err != nil {
		t.Fatalf("update team state: %v", err)
	}

	updated, err := store.LoadTeamState(teamState.TeamID)
	if err != nil {
		t.Fatalf("load updated team state: %v", err)
	}
	if updated.State != types.TeamSynthesizing || len(updated.ReadyTaskIDs) != 1 || updated.ReadyTaskIDs[0] != "synthesize" {
		t.Fatalf("updated team state did not round-trip: %+v", updated)
	}
}

func TestTeamStateStoreReturnsNotFound(t *testing.T) {
	store := newTeamStateStoreTestAdapter(t)

	if _, err := store.LoadTeamState("missing-team"); err == nil {
		t.Fatal("expected missing team state to return an error")
	}
}

func newTeamStateStoreTestAdapter(t *testing.T) *TeamStateStore {
	t.Helper()
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := database.RunMigrations(); err != nil {
		t.Fatalf("migrate db: %v", err)
	}
	return NewTeamStateStore(database)
}
