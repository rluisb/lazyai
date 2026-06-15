package budget

import (
	"testing"

	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
	"github.com/rluisb/lazyai/packages/orchestrator/ports"
)

func TestTrackerImplementsBudgetTrackerPort(t *testing.T) {
	tracker := NewTracker()
	var _ ports.BudgetTracker = tracker

	state := &types.BudgetState{
		Tokens:  types.BudgetDimensionState{Limit: 100, Remaining: 100},
		Retries: types.BudgetDimensionState{Limit: 2, Remaining: 2},
	}
	tracker.Update(state, "step-1", &types.StepUsage{TotalTokens: 25})
	tracker.IncrementRetries(state, "step-1")
	evaluation := tracker.Evaluate(state, &types.BudgetPolicy{Tokens: &types.BudgetThreshold{Limit: 100}})

	if state.Tokens.Consumed != 25 || state.Retries.Consumed != 1 || evaluation.Overall != types.HealthOK {
		t.Fatalf("tracker did not preserve budget behavior: state=%+v evaluation=%+v", state, evaluation)
	}
}

func TestUpdate_AccumulatesUsage(t *testing.T) {
	state := &types.BudgetState{
		PolicyID:    "policy-1",
		Scope:       "team",
		Tokens:      types.BudgetDimensionState{Limit: 100000, Consumed: 0, Remaining: 100000},
		CostUsd:     types.BudgetDimensionState{Limit: 500, Consumed: 0, Remaining: 500},
		WallClockMs: types.BudgetDimensionState{Limit: 300000, Consumed: 0, Remaining: 300000},
		Retries:     types.BudgetDimensionState{Limit: 3, Consumed: 0, Remaining: 3},
		ByStep:      map[string]types.StepUsage{},
	}

	usage := &types.StepUsage{
		InputTokens:  10000,
		OutputTokens: 5000,
		TotalTokens:  15000,
		CostUsd:      0.30,
		WallClockMs:  5000,
	}

	Update(state, "step-1", usage)

	if state.Tokens.Consumed != 15000 {
		t.Errorf("tokens consumed: expected 15000, got %d", state.Tokens.Consumed)
	}
	if state.Tokens.Remaining != 85000 {
		t.Errorf("tokens remaining: expected 85000, got %d", state.Tokens.Remaining)
	}
	if state.CostUsd.Consumed != 30 {
		t.Errorf("cost consumed: expected 30 (cents), got %d", state.CostUsd.Consumed)
	}
	if state.WallClockMs.Consumed != 5000 {
		t.Errorf("wall clock ms consumed: expected 5000, got %d", state.WallClockMs.Consumed)
	}

	// Check by-step accumulation
	stepUsage := state.ByStep["step-1"]
	if stepUsage.TotalTokens != 15000 {
		t.Errorf("byStep[step-1].totalTokens: expected 15000, got %d", stepUsage.TotalTokens)
	}
}

func TestUpdate_WarningTriggered(t *testing.T) {
	state := &types.BudgetState{
		PolicyID: "policy-1",
		Scope:    "team",
		Tokens:   types.BudgetDimensionState{Limit: 100000, Consumed: 0, Remaining: 100000},
		ByStep:   map[string]types.StepUsage{},
	}

	// 80% of limit
	usage := &types.StepUsage{TotalTokens: 80000}
	Update(state, "step-1", usage)

	if !state.Tokens.WarningTriggered {
		t.Error("expected warningTriggered to be true at 80%")
	}
	if state.Tokens.PausedAtLimit {
		t.Error("expected pausedAtLimit to be false before 100%")
	}
}

func TestUpdate_PausedAtLimit(t *testing.T) {
	state := &types.BudgetState{
		PolicyID: "policy-1",
		Scope:    "team",
		Tokens:   types.BudgetDimensionState{Limit: 100000, Consumed: 0, Remaining: 100000},
		ByStep:   map[string]types.StepUsage{},
	}

	// At or above limit
	usage := &types.StepUsage{TotalTokens: 100000}
	Update(state, "step-1", usage)

	if !state.Tokens.PausedAtLimit {
		t.Error("expected pausedAtLimit to be true at 100%")
	}
}

func TestUpdate_MergesMultipleSteps(t *testing.T) {
	state := &types.BudgetState{
		PolicyID: "policy-1",
		Scope:    "chain",
		Tokens:   types.BudgetDimensionState{Limit: 100000, Consumed: 0, Remaining: 100000},
		ByStep:   map[string]types.StepUsage{},
	}

	Update(state, "step-1", &types.StepUsage{TotalTokens: 10000, CostUsd: 0.10})
	Update(state, "step-2", &types.StepUsage{TotalTokens: 20000, CostUsd: 0.20})
	Update(state, "step-1", &types.StepUsage{TotalTokens: 5000, CostUsd: 0.05})

	if state.Tokens.Consumed != 35000 {
		t.Errorf("tokens consumed: expected 35000, got %d", state.Tokens.Consumed)
	}
	if state.CostUsd.Consumed != 35 {
		t.Errorf("cost consumed: expected 35 (cents), got %d", state.CostUsd.Consumed)
	}

	// step-1 should have accumulated tokens
	step1 := state.ByStep["step-1"]
	if step1.TotalTokens != 15000 {
		t.Errorf("byStep[step-1].totalTokens: expected 15000, got %d", step1.TotalTokens)
	}
}

func TestIncrementRetries(t *testing.T) {
	state := &types.BudgetState{
		PolicyID: "policy-1",
		Scope:    "team",
		Retries:  types.BudgetDimensionState{Limit: 3, Consumed: 0, Remaining: 3},
		ByStep:   map[string]types.StepUsage{},
	}

	IncrementRetries(state, "step-1")
	if state.Retries.Consumed != 1 {
		t.Errorf("retries consumed: expected 1, got %d", state.Retries.Consumed)
	}
	if state.Retries.Remaining != 2 {
		t.Errorf("retries remaining: expected 2, got %d", state.Retries.Remaining)
	}

	// Increment again
	IncrementRetries(state, "step-1")
	if state.Retries.Consumed != 2 {
		t.Errorf("retries consumed: expected 2, got %d", state.Retries.Consumed)
	}

	// At limit
	state.Retries.Consumed = 3
	state.Retries.Limit = 3
	IncrementRetries(state, "step-1")
	if !state.Retries.PausedAtLimit {
		t.Error("expected pausedAtLimit to be true when at limit")
	}
}

func TestEvaluate_Ok(t *testing.T) {
	state := &types.BudgetState{
		Tokens:      types.BudgetDimensionState{Limit: 100000, Consumed: 10000},
		CostUsd:     types.BudgetDimensionState{Limit: 100, Consumed: 10},
		WallClockMs: types.BudgetDimensionState{Limit: 100000, Consumed: 10000},
		Retries:     types.BudgetDimensionState{Limit: 3, Consumed: 0},
	}
	policy := &types.BudgetPolicy{
		Tokens:      &types.BudgetThreshold{Limit: 100000},
		CostUsd:     &types.BudgetThreshold{Limit: 100},
		WallClockMs: &types.BudgetThreshold{Limit: 100000},
		Retries:     &types.BudgetThreshold{Limit: 3},
	}

	ev := Evaluate(state, policy)

	if ev.Overall != types.HealthOK {
		t.Errorf("expected overall=ok, got %s", ev.Overall)
	}
	if ev.RecommendedAction != "continue" {
		t.Errorf("expected recommendedAction=continue, got %s", ev.RecommendedAction)
	}
	if ev.ShouldPause {
		t.Error("expected shouldPause=false")
	}
}

func TestEvaluate_Warning(t *testing.T) {
	state := &types.BudgetState{
		Tokens:      types.BudgetDimensionState{Limit: 100000, Consumed: 75000}, // 75% — warning
		CostUsd:     types.BudgetDimensionState{Limit: 100, Consumed: 10},
		WallClockMs: types.BudgetDimensionState{Limit: 100000, Consumed: 10000},
		Retries:     types.BudgetDimensionState{Limit: 3, Consumed: 0},
	}
	policy := &types.BudgetPolicy{
		Tokens:      &types.BudgetThreshold{Limit: 100000},
		CostUsd:     &types.BudgetThreshold{Limit: 100},
		WallClockMs: &types.BudgetThreshold{Limit: 100000},
		Retries:     &types.BudgetThreshold{Limit: 3},
	}

	ev := Evaluate(state, policy)

	if ev.Overall != types.HealthWarning {
		t.Errorf("expected overall=warning, got %s", ev.Overall)
	}
	if ev.RecommendedAction != "warn" {
		t.Errorf("expected recommendedAction=warn, got %s", ev.RecommendedAction)
	}
}

func TestEvaluate_LimitReached(t *testing.T) {
	state := &types.BudgetState{
		Tokens:      types.BudgetDimensionState{Limit: 100000, Consumed: 95000}, // 95% — limit reached
		CostUsd:     types.BudgetDimensionState{Limit: 100, Consumed: 10},
		WallClockMs: types.BudgetDimensionState{Limit: 100000, Consumed: 10000},
		Retries:     types.BudgetDimensionState{Limit: 3, Consumed: 0},
	}
	policy := &types.BudgetPolicy{
		Tokens:      &types.BudgetThreshold{Limit: 100000},
		CostUsd:     &types.BudgetThreshold{Limit: 100},
		WallClockMs: &types.BudgetThreshold{Limit: 100000},
		Retries:     &types.BudgetThreshold{Limit: 3},
	}

	ev := Evaluate(state, policy)

	if ev.Overall != types.HealthLimitReached {
		t.Errorf("expected overall=limit_reached, got %s", ev.Overall)
	}
	if ev.RecommendedAction != "pause" {
		t.Errorf("expected recommendedAction=pause, got %s", ev.RecommendedAction)
	}
	if !ev.ShouldPause {
		t.Error("expected shouldPause=true at limit")
	}
}

func TestEvaluate_UsesPolicyDefaultAction(t *testing.T) {
	state := &types.BudgetState{
		Tokens:      types.BudgetDimensionState{Limit: 100000, Consumed: 100000, PausedAtLimit: true},
		CostUsd:     types.BudgetDimensionState{Limit: 100, Consumed: 0},
		WallClockMs: types.BudgetDimensionState{Limit: 100000, Consumed: 0},
		Retries:     types.BudgetDimensionState{Limit: 3, Consumed: 0},
	}
	policy := &types.BudgetPolicy{
		Tokens:               &types.BudgetThreshold{Limit: 100000},
		CostUsd:              &types.BudgetThreshold{Limit: 100},
		WallClockMs:          &types.BudgetThreshold{Limit: 100000},
		Retries:              &types.BudgetThreshold{Limit: 3},
		DefaultActionOnLimit: "abort",
	}

	ev := Evaluate(state, policy)

	if ev.RecommendedAction != "abort" {
		t.Errorf("expected recommendedAction=abort from policy, got %s", ev.RecommendedAction)
	}
}
