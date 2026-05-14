package ports

import "github.com/rluisb/lazyai/packages/orchestrator/internal/types"

// ExecutionPlanStore is the persistence port for compiled lifecycle execution plans.
type ExecutionPlanStore interface {
	SaveExecutionPlan(plan *types.ExecutionPlan) error
	LoadExecutionPlan(id string) (*types.ExecutionPlan, error)
}
