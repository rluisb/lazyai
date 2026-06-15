package ports

import "github.com/rluisb/lazyai/packages/orchestrator/internal/types"

// BudgetTracker is the service port for budget usage accounting and evaluation.
type BudgetTracker interface {
	Update(state *types.BudgetState, stepID string, usage *types.StepUsage)
	IncrementRetries(state *types.BudgetState, stepID string)
	Evaluate(state *types.BudgetState, policy *types.BudgetPolicy) types.BudgetEvaluation
}
