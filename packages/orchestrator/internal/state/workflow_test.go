package state

import (
	"testing"

	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
)

func TestCreateWorkflowState(t *testing.T) {
	def := &types.WorkflowDefinition{
		DefinitionMetadata: types.DefinitionMetadata{Name: "test-wf", Version: "1.0"},
		Entry:               "phase-1",
		Phases: []types.WorkflowPhaseDefinition{
			{ID: "phase-1", Kind: "chain", Ref: "research-chain"},
			{ID: "phase-2", Kind: "team", Ref: "build-team"},
			{ID: "phase-3", Kind: "terminal"},
		},
	}
	plan := &types.ExecutionPlan{
		ID:          "plan-1",
		Task:        "build something",
		BudgetPolicy: types.BudgetPolicy{ID: "policy-1"},
		CreatedAt:   "2024-01-01T00:00:00Z",
	}

	state := CreateWorkflowState(def, plan)

	// Should have 3 phases
	if len(state.Phases) != 3 {
		t.Fatalf("expected 3 phases, got %d", len(state.Phases))
	}

	// Entry phase should be running
	if state.Phases[0].State != types.PhaseRunning {
		t.Errorf("phase-1: expected state=running, got %s", state.Phases[0].State)
	}
	if state.CurrentPhaseID != "phase-1" {
		t.Errorf("currentPhaseID: expected phase-1, got %s", state.CurrentPhaseID)
	}

	// Other phases should be pending
	if state.Phases[1].State != types.PhasePending {
		t.Errorf("phase-2: expected state=pending, got %s", state.Phases[1].State)
	}
	if state.Phases[2].State != types.PhasePending {
		t.Errorf("phase-3: expected state=pending, got %s", state.Phases[2].State)
	}

	// Workflow state should be running
	if state.State != types.WorkflowRunning {
		t.Errorf("expected workflow state=running, got %s", state.State)
	}

	// Child runs should be empty
	if len(state.ChildRuns) != 0 {
		t.Errorf("expected 0 child runs, got %d", len(state.ChildRuns))
	}
}

func TestApplyWorkflowChildLaunch(t *testing.T) {
	state := &types.WorkflowState{
		State:          types.WorkflowRunning,
		CurrentPhaseID: "phase-1",
		Phases: []types.WorkflowPhaseState{
			{PhaseID: "phase-1", Kind: "chain", State: types.PhaseRunning},
		},
		ChildRuns: []types.WorkflowChildRun{},
	}

	err := ApplyWorkflowChildLaunch(state, "phase-1", "run-123", "chain", "research-chain")
	if err != nil {
		t.Fatalf("apply child launch: %v", err)
	}

	if state.Phases[0].State != types.PhaseWaitingOnChild {
		t.Errorf("phase-1: expected state=waiting_on_child, got %s", state.Phases[0].State)
	}
	if state.Phases[0].ChildRun == nil {
		t.Fatal("phase-1: expected childRun to be set")
	}
	if state.Phases[0].ChildRun.RunID != "run-123" {
		t.Errorf("childRun.runID: expected run-123, got %s", state.Phases[0].ChildRun.RunID)
	}
	if state.State != types.WorkflowWaitingOnChild {
		t.Errorf("workflow state: expected waiting_on_child, got %s", state.State)
	}
	if len(state.ChildRuns) != 1 {
		t.Fatalf("expected 1 child run in state, got %d", len(state.ChildRuns))
	}
}

func TestAdvanceWorkflowState_SimplePhase(t *testing.T) {
	state := &types.WorkflowState{
		State:          types.WorkflowRunning,
		CurrentPhaseID: "phase-1",
		Phases: []types.WorkflowPhaseState{
			{PhaseID: "phase-1", Kind: "chain", State: types.PhaseRunning},
			{PhaseID: "phase-2", Kind: "chain", State: types.PhasePending}, // non-terminal so workflow continues
		},
	}

	// Advance from phase-1 with success outcome
	next, err := AdvanceWorkflowState(state, "success", nil)
	if err != nil {
		t.Fatalf("advance workflow: %v", err)
	}

	if next.Phases[0].State != types.PhaseCompleted {
		t.Errorf("phase-1: expected state=completed, got %s", next.Phases[0].State)
	}
	if next.CurrentPhaseID != "phase-2" {
		t.Errorf("currentPhaseID: expected phase-2, got %s", next.CurrentPhaseID)
	}
	if next.Phases[1].State != types.PhaseRunning {
		t.Errorf("phase-2: expected state=running, got %s", next.Phases[1].State)
	}
}

func TestAdvanceWorkflowState_TerminalPhase(t *testing.T) {
	state := &types.WorkflowState{
		State:          types.WorkflowRunning,
		CurrentPhaseID: "phase-1",
		Phases: []types.WorkflowPhaseState{
			{PhaseID: "phase-1", Kind: "terminal", State: types.PhaseRunning},
		},
	}

	next, err := AdvanceWorkflowState(state, "success", nil)
	if err != nil {
		t.Fatalf("advance workflow: %v", err)
	}

	if next.State != types.WorkflowCompleted {
		t.Errorf("expected workflow state=completed, got %s", next.State)
	}
}

func TestAdvanceWorkflowState_GatePhase(t *testing.T) {
	state := &types.WorkflowState{
		State:          types.WorkflowRunning,
		CurrentPhaseID: "gate-phase",
		Phases: []types.WorkflowPhaseState{
			{PhaseID: "gate-phase", Kind: "gate", State: types.PhaseRunning, Gate: "approval"},
			{PhaseID: "next-phase", Kind: "chain", State: types.PhasePending},
		},
	}

	// Gate phase should call findNextPhase on advance
	next, err := AdvanceWorkflowState(state, "approved", nil)
	if err != nil {
		t.Fatalf("advance gate phase: %v", err)
	}

	if next.Phases[0].State != types.PhaseCompleted {
		t.Errorf("gate-phase: expected state=completed, got %s", next.Phases[0].State)
	}
	if next.Phases[0].LastOutcome != "approved" {
		t.Errorf("gate-phase.lastOutcome: expected approved, got %s", next.Phases[0].LastOutcome)
	}
	if next.CurrentPhaseID != "next-phase" {
		t.Errorf("currentPhaseID: expected next-phase, got %s", next.CurrentPhaseID)
	}
}

func TestAdvanceWorkflowState_RecoveryRetry(t *testing.T) {
	state := &types.WorkflowState{
		State:          types.WorkflowRunning,
		CurrentPhaseID: "phase-1",
		Phases: []types.WorkflowPhaseState{
			{PhaseID: "phase-1", Kind: "chain", State: types.PhaseCompleted},
			{PhaseID: "phase-2", Kind: "chain", State: types.PhaseRunning},
		},
	}

	recovery := &types.WorkflowRecoveryDecision{
		Type:          "retry",
		TargetPhaseID: "phase-2",
		Reason:        "transient failure",
	}

	next, err := AdvanceWorkflowState(state, "", recovery)
	if err != nil {
		t.Fatalf("advance with retry recovery: %v", err)
	}

	if next.State != types.WorkflowRunning {
		t.Errorf("expected workflow state=running after retry, got %s", next.State)
	}
	if next.Phases[1].State != types.PhaseRunning {
		t.Errorf("phase-2: expected state=running after retry, got %s", next.Phases[1].State)
	}
	if next.Phases[1].CompletedAt != "" {
		t.Error("phase-2: expected completedAt to be cleared after retry")
	}
}

func TestAdvanceWorkflowState_RecoveryEscalate(t *testing.T) {
	state := &types.WorkflowState{
		State:          types.WorkflowRunning,
		CurrentPhaseID: "phase-1",
		Phases: []types.WorkflowPhaseState{
			{PhaseID: "phase-1", Kind: "chain", State: types.PhaseRunning},
		},
	}

	recovery := &types.WorkflowRecoveryDecision{
		Type:       "escalate",
		Reason:     "wrong approach",
		Recipient:  "senior-agent",
	}

	next, err := AdvanceWorkflowState(state, "", recovery)
	if err != nil {
		t.Fatalf("advance with escalate recovery: %v", err)
	}

	if next.State != types.WorkflowAwaitingRecovery {
		t.Errorf("expected workflow state=awaiting_recovery, got %s", next.State)
	}
	if next.HandoffSummary == "" {
		t.Error("expected handoffSummary to be set")
	}
}

func TestAdvanceWorkflowState_RecoveryHandoff(t *testing.T) {
	state := &types.WorkflowState{
		State:          types.WorkflowRunning,
		CurrentPhaseID: "phase-1",
		Phases: []types.WorkflowPhaseState{
			{PhaseID: "phase-1", Kind: "chain", State: types.PhaseRunning},
		},
	}

	recovery := &types.WorkflowRecoveryDecision{
		Type:    "handoff",
		Summary: "context exhausted, handing off",
	}

	next, err := AdvanceWorkflowState(state, "", recovery)
	if err != nil {
		t.Fatalf("advance with handoff recovery: %v", err)
	}

	if next.State != types.WorkflowHandoff {
		t.Errorf("expected workflow state=handoff, got %s", next.State)
	}
	if next.HandoffSummary != "context exhausted, handing off" {
		t.Errorf("handoffSummary: got %s", next.HandoffSummary)
	}
}