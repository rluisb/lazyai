package ports

import "github.com/rluisb/lazyai/packages/orchestrator/internal/types"

// WorkflowStateStore is the persistence port for workflow lifecycle run state.
type WorkflowStateStore interface {
	SaveWorkflowState(projectRoot string, workflowState *types.WorkflowState) error
	LoadWorkflowState(id string) (*types.WorkflowState, error)
}
