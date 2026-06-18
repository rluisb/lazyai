package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/rluisb/lazyai/packages/orchestrator/internal/db"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
	"github.com/rluisb/lazyai/packages/orchestrator/ports"
)

var _ ports.WorkflowStateStore = (*WorkflowStateStore)(nil)

// WorkflowStateStore persists workflow lifecycle run state in SQLite.
type WorkflowStateStore struct {
	database *db.DB
}

// NewWorkflowStateStore creates a SQLite-backed workflow state store adapter.
func NewWorkflowStateStore(database *db.DB) *WorkflowStateStore {
	return &WorkflowStateStore{database: database}
}

// SaveWorkflowState creates or updates a workflow lifecycle run state snapshot.
func (s *WorkflowStateStore) SaveWorkflowState(projectRoot string, workflowState *types.WorkflowState) error {
	encoded, err := json.Marshal(workflowState)
	if err != nil {
		return fmt.Errorf("marshal workflow state: %w", err)
	}
	_, err = s.database.Exec(`
		INSERT INTO workflow_runs (id, definition_name, definition_version, state, current_phase_id, project_root, state_json, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			state = excluded.state,
			current_phase_id = excluded.current_phase_id,
			state_json = excluded.state_json,
			updated_at = excluded.updated_at
	`, workflowState.WorkflowID, workflowState.DefinitionName, workflowState.DefinitionVersion, workflowState.State, nullableString(workflowState.CurrentPhaseID), projectRoot, string(encoded), workflowState.CreatedAt, workflowState.UpdatedAt)
	if err != nil {
		return fmt.Errorf("save workflow state %s: %w", workflowState.WorkflowID, err)
	}
	return nil
}

// LoadWorkflowState returns a workflow lifecycle run state snapshot by ID.
func (s *WorkflowStateStore) LoadWorkflowState(id string) (*types.WorkflowState, error) {
	var stateJSON string
	err := s.database.QueryRow(`SELECT state_json FROM workflow_runs WHERE id = ?`, id).Scan(&stateJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("workflow state not found: %s", id)
		}
		return nil, err
	}
	var workflowState types.WorkflowState
	if err := json.Unmarshal([]byte(stateJSON), &workflowState); err != nil {
		return nil, fmt.Errorf("decode workflow state %s: %w", id, err)
	}
	return &workflowState, nil
}
