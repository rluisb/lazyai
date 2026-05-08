package state

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
)

// CreateWorkflowState builds initial WorkflowState from a workflow definition.
func CreateWorkflowState(def *types.WorkflowDefinition, plan *types.ExecutionPlan) *types.WorkflowState {
	now := time.Now().UTC().Format(time.RFC3339)
	phases := make([]types.WorkflowPhaseState, 0, len(def.Phases))
	for _, p := range def.Phases {
		phases = append(phases, types.WorkflowPhaseState{
			PhaseID: p.ID,
			Kind:    p.Kind,
			State:   types.PhasePending,
			Ref:     p.Ref,
			Gate:    p.Gate,
			Prompt:  p.Prompt,
		})
	}

	// Mark entry phase as running
	if len(phases) > 0 {
		phases[0].State = types.PhaseRunning
		phases[0].StartedAt = now
	}

	return &types.WorkflowState{
		WorkflowID:        uuid.NewString(),
		DefinitionName:    def.Name,
		DefinitionVersion: def.Version,
		ExecutionPlanID:   plan.ID,
		State:             types.WorkflowRunning,
		Task:              plan.Task,
		EntryPhaseID:      def.Entry,
		CurrentPhaseID:    def.Entry,
		Phases:            phases,
		ChildRuns:         []types.WorkflowChildRun{},
		BudgetPolicy:      plan.BudgetPolicy,
		Budget:            createEmptyBudgetForWorkflow(plan),
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

// ApplyWorkflowChildLaunch records that a child run has been launched for a phase.
func ApplyWorkflowChildLaunch(state *types.WorkflowState, phaseID, runID, runKind, definitionName string) error {
	phase := requireWorkflowPhase(state, phaseID)
	now := time.Now().UTC().Format(time.RFC3339)

	phase.State = types.PhaseWaitingOnChild
	phase.ChildRun = &types.WorkflowChildRun{
		PhaseID:        phaseID,
		RunID:          runID,
		RunKind:        runKind,
		DefinitionName: definitionName,
		LaunchedAt:     now,
	}
	state.ChildRuns = append(state.ChildRuns, *phase.ChildRun)
	state.State = types.WorkflowWaitingOnChild
	state.UpdatedAt = now
	return nil
}

// AdvanceWorkflowState advances the workflow after a child completes or a gate decision.
func AdvanceWorkflowState(state *types.WorkflowState, outcome string, recovery *types.WorkflowRecoveryDecision) (*types.WorkflowState, error) {
	next := cloneWorkflowState(state)
	now := time.Now().UTC().Format(time.RFC3339)

	// Handle recovery decision
	if recovery != nil {
		switch recovery.Type {
		case "retry":
			return retryWorkflowPhase(next, recovery.TargetPhaseID, now)
		case "escalate":
			return escalateWorkflowPhase(next, recovery.TargetPhaseID, recovery.Recipient, now)
		case "handoff":
			next.State = types.WorkflowHandoff
			next.HandoffSummary = recovery.Summary
			next.UpdatedAt = now
			return next, nil
		}
	}

	// Find current phase
	phase := requireWorkflowPhase(next, next.CurrentPhaseID)

	// Handle gate phases
	if phase.Kind == "gate" {
		return handleGatePhase(next, phase, outcome, now)
	}

	// Handle terminal phases
	if phase.Kind == "terminal" {
		next.State = types.WorkflowCompleted
		next.UpdatedAt = now
		return next, nil
	}

	// Child run completed — advance to next phase
	phase.State = types.PhaseCompleted
	phase.CompletedAt = now
	phase.LastOutcome = outcome
	if phase.ChildRun != nil {
		phase.ChildRun.CompletedAt = now
		phase.ChildRun.Outcome = outcome
	}

	return advanceToNextPhase(next, phase, outcome, now)
}

// ──────────────────────── internal ───────────────────────────

func retryWorkflowPhase(state *types.WorkflowState, phaseID, now string) (*types.WorkflowState, error) {
	phase := requireWorkflowPhase(state, phaseID)
	phase.State = types.PhaseRunning
	phase.StartedAt = now
	phase.CompletedAt = ""
	phase.LastOutcome = ""
	phase.ChildRun = nil
	state.State = types.WorkflowRunning
	state.CurrentPhaseID = phaseID
	state.UpdatedAt = now
	return state, nil
}

func escalateWorkflowPhase(state *types.WorkflowState, phaseID, recipient, now string) (*types.WorkflowState, error) {
	state.State = types.WorkflowAwaitingRecovery
	state.HandoffSummary = fmt.Sprintf("Phase %s escalated to %s", phaseID, recipient)
	state.UpdatedAt = now
	return state, nil
}

func handleGatePhase(state *types.WorkflowState, phase *types.WorkflowPhaseState, outcome, now string) (*types.WorkflowState, error) {
	phase.State = types.PhaseCompleted
	phase.CompletedAt = now
	phase.LastOutcome = outcome
	return advanceToNextPhase(state, phase, outcome, now)
}

func advanceToNextPhase(state *types.WorkflowState, current *types.WorkflowPhaseState, outcome, now string) (*types.WorkflowState, error) {
	// Find the next phase based on outcome
	nextID := findNextPhase(state, current, outcome)
	if nextID == "" {
		state.State = types.WorkflowCompleted
		state.CurrentPhaseID = ""
		state.UpdatedAt = now
		return state, nil
	}

	nextPhase := requireWorkflowPhase(state, nextID)

	// Handle terminal
	if nextPhase.Kind == "terminal" {
		state.State = types.WorkflowCompleted
		state.CurrentPhaseID = nextID
		state.UpdatedAt = now
		return state, nil
	}

	// Advance to next phase
	nextPhase.State = types.PhaseRunning
	nextPhase.StartedAt = now
	state.State = types.WorkflowRunning
	state.CurrentPhaseID = nextID
	state.UpdatedAt = now
	return state, nil
}

func findNextPhase(state *types.WorkflowState, current *types.WorkflowPhaseState, outcome string) string {
	// Try to find a phase with an "on" mapping matching this outcome
	// Simple heuristic: look for the next phase by index
	for i, p := range state.Phases {
		if p.PhaseID == current.PhaseID && i+1 < len(state.Phases) {
			return state.Phases[i+1].PhaseID
		}
	}
	return ""
}

func requireWorkflowPhase(state *types.WorkflowState, phaseID string) *types.WorkflowPhaseState {
	for i := range state.Phases {
		if state.Phases[i].PhaseID == phaseID {
			return &state.Phases[i]
		}
	}
	panic(fmt.Sprintf("unknown workflow phase: %s", phaseID))
}

func createEmptyBudgetForWorkflow(plan *types.ExecutionPlan) types.BudgetState {
	return types.BudgetState{
		PolicyID:      plan.BudgetPolicy.ID,
		Scope:         "workflow",
		Tokens:        buildBudgetDimWf(plan.BudgetPolicy.Tokens),
		CostUsd:       buildBudgetDimWf(plan.BudgetPolicy.CostUsd),
		WallClockMs:   buildBudgetDimWf(plan.BudgetPolicy.WallClockMs),
		Retries:       buildBudgetDimWf(plan.BudgetPolicy.Retries),
		ByStep:        map[string]types.StepUsage{},
		LastUpdatedAt: plan.CreatedAt,
	}
}

func buildBudgetDimWf(threshold *types.BudgetThreshold) types.BudgetDimensionState {
	d := types.BudgetDimensionState{}
	if threshold != nil {
		d.Limit = threshold.Limit
		d.Remaining = threshold.Limit
	}
	return d
}

func cloneWorkflowState(s *types.WorkflowState) *types.WorkflowState {
	b, _ := json.Marshal(s)
	var c types.WorkflowState
	json.Unmarshal(b, &c)
	return &c
}
