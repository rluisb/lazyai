package budget

import (
	"time"

	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
)

// Tracker applies budget accounting rules behind the BudgetTracker service port.
type Tracker struct{}

// NewTracker creates the default budget tracking service.
func NewTracker() *Tracker {
	return &Tracker{}
}

// Evaluate returns a budget health assessment for a given state.
func (Tracker) Evaluate(state *types.BudgetState, policy *types.BudgetPolicy) types.BudgetEvaluation {
	return Evaluate(state, policy)
}

// Update adds step usage to the budget state.
func (Tracker) Update(state *types.BudgetState, stepID string, usage *types.StepUsage) {
	Update(state, stepID, usage)
}

// IncrementRetries records a consumed retry against the retry budget dimension.
func (Tracker) IncrementRetries(state *types.BudgetState, stepID string) {
	IncrementRetries(state, stepID)
}

// Evaluate returns a budget health assessment for a given state.
func Evaluate(state *types.BudgetState, policy *types.BudgetPolicy) types.BudgetEvaluation {
	ev := types.BudgetEvaluation{
		Overall: types.HealthOK,
		Dimensions: struct {
			Tokens      types.BudgetHealth `json:"tokens"`
			CostUsd     types.BudgetHealth `json:"costUsd"`
			WallClockMs types.BudgetHealth `json:"wallClockMs"`
			Retries     types.BudgetHealth `json:"retries"`
		}{
			Tokens:      dimHealth(policy.Tokens, &state.Tokens),
			CostUsd:     dimHealth(policy.CostUsd, &state.CostUsd),
			WallClockMs: dimHealth(policy.WallClockMs, &state.WallClockMs),
			Retries:     dimHealth(policy.Retries, &state.Retries),
		},
		RecommendedAction: "continue",
		ShouldPause:       false,
	}

	worst := types.HealthOK
	for _, h := range []types.BudgetHealth{ev.Dimensions.Tokens, ev.Dimensions.CostUsd, ev.Dimensions.WallClockMs, ev.Dimensions.Retries} {
		if healthWorse(h, worst) {
			worst = h
		}
	}
	ev.Overall = worst

	switch worst {
	case types.HealthWarning:
		ev.RecommendedAction = "warn"
	case types.HealthLimitReached:
		if policy != nil && policy.DefaultActionOnLimit != "" {
			ev.RecommendedAction = policy.DefaultActionOnLimit
		} else {
			ev.RecommendedAction = "pause"
		}
		ev.ShouldPause = ev.RecommendedAction == "pause" || ev.RecommendedAction == "abort"
	}

	return ev
}

// Update adds step usage to the budget state.
func Update(state *types.BudgetState, stepID string, usage *types.StepUsage) {
	if usage == nil {
		return
	}

	totalTokens := usage.TotalTokens
	if totalTokens == 0 {
		totalTokens = usage.InputTokens + usage.OutputTokens
	}

	state.Tokens.Consumed += totalTokens
	if state.Tokens.Limit > 0 && state.Tokens.Consumed >= state.Tokens.Limit {
		state.Tokens.PausedAtLimit = true
	}
	if state.Tokens.Limit > 0 && state.Tokens.Consumed >= int(float64(state.Tokens.Limit)*0.8) {
		state.Tokens.WarningTriggered = true
	}
	updateRemaining(&state.Tokens)

	state.CostUsd.Consumed += int(usage.CostUsd * 100)
	if state.CostUsd.Limit > 0 && state.CostUsd.Consumed >= state.CostUsd.Limit {
		state.CostUsd.PausedAtLimit = true
	}
	updateRemaining(&state.CostUsd)

	state.WallClockMs.Consumed += usage.WallClockMs
	if state.WallClockMs.Limit > 0 && state.WallClockMs.Consumed >= state.WallClockMs.Limit {
		state.WallClockMs.PausedAtLimit = true
	}
	updateRemaining(&state.WallClockMs)

	if state.ByStep == nil {
		state.ByStep = map[string]types.StepUsage{}
	}
	merged := state.ByStep[stepID]
	merged.InputTokens += usage.InputTokens
	merged.OutputTokens += usage.OutputTokens
	merged.TotalTokens += totalTokens
	merged.CostUsd += usage.CostUsd
	merged.WallClockMs += usage.WallClockMs
	state.ByStep[stepID] = merged
	state.LastUpdatedAt = time.Now().UTC().Format(time.RFC3339)
}

// IncrementRetries records a consumed retry against the retry budget dimension.
func IncrementRetries(state *types.BudgetState, stepID string) {
	state.Retries.Consumed++
	if state.Retries.Limit > 0 && state.Retries.Consumed >= state.Retries.Limit {
		state.Retries.PausedAtLimit = true
	}
	updateRemaining(&state.Retries)
	if state.ByStep == nil {
		state.ByStep = map[string]types.StepUsage{}
	}
	if _, ok := state.ByStep[stepID]; !ok {
		state.ByStep[stepID] = types.StepUsage{}
	}
	state.LastUpdatedAt = time.Now().UTC().Format(time.RFC3339)
}

// ──────────────────────── internal ───────────────────────────

func dimHealth(threshold *types.BudgetThreshold, dim *types.BudgetDimensionState) types.BudgetHealth {
	if dim.PausedAtLimit {
		return types.HealthLimitReached
	}
	if dim.WarningTriggered {
		return types.HealthWarning
	}
	if threshold != nil && threshold.Limit > 0 && dim.Consumed > 0 {
		pct := float64(dim.Consumed) / float64(threshold.Limit)
		if pct >= 0.9 {
			return types.HealthLimitReached
		}
		if pct >= 0.7 {
			return types.HealthWarning
		}
	}
	return types.HealthOK
}

func healthWorse(a, b types.BudgetHealth) bool {
	order := map[types.BudgetHealth]int{
		types.HealthOK:           0,
		types.HealthWarning:      1,
		types.HealthLimitReached: 2,
	}
	return order[a] > order[b]
}

func updateRemaining(dim *types.BudgetDimensionState) {
	if dim.Limit <= 0 {
		return
	}
	dim.Remaining = max(dim.Limit-dim.Consumed, 0)
}
