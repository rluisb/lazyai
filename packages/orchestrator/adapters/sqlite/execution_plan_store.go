package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/rluisb/lazyai/packages/orchestrator/internal/db"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
	"github.com/rluisb/lazyai/packages/orchestrator/ports"
)

var _ ports.ExecutionPlanStore = (*ExecutionPlanStore)(nil)

// ExecutionPlanStore persists compiled lifecycle execution plans in SQLite.
type ExecutionPlanStore struct {
	database *db.DB
}

// NewExecutionPlanStore creates a SQLite-backed execution plan store adapter.
func NewExecutionPlanStore(database *db.DB) *ExecutionPlanStore {
	return &ExecutionPlanStore{database: database}
}

// SaveExecutionPlan creates or updates a compiled execution plan.
func (s *ExecutionPlanStore) SaveExecutionPlan(plan *types.ExecutionPlan) error {
	encoded, err := json.Marshal(plan)
	if err != nil {
		return fmt.Errorf("marshal execution plan: %w", err)
	}
	_, err = s.database.Exec(`
		INSERT INTO execution_plans (id, kind, definition_name, definition_version, project_root, plan_json, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			kind = excluded.kind,
			definition_name = excluded.definition_name,
			definition_version = excluded.definition_version,
			project_root = excluded.project_root,
			plan_json = excluded.plan_json
	`, plan.ID, plan.Kind, plan.Definition.Name, plan.Definition.Version, plan.Project.RootPath, string(encoded), plan.CreatedAt)
	if err != nil {
		return fmt.Errorf("save execution plan %s: %w", plan.ID, err)
	}
	return nil
}

// LoadExecutionPlan returns a compiled execution plan by ID.
func (s *ExecutionPlanStore) LoadExecutionPlan(id string) (*types.ExecutionPlan, error) {
	var planJSON string
	err := s.database.QueryRow(`SELECT plan_json FROM execution_plans WHERE id = ?`, id).Scan(&planJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("execution plan not found: %s", id)
		}
		return nil, err
	}
	var plan types.ExecutionPlan
	if err := json.Unmarshal([]byte(planJSON), &plan); err != nil {
		return nil, fmt.Errorf("decode execution plan %s: %w", id, err)
	}
	return &plan, nil
}
