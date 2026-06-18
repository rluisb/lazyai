package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/rluisb/lazyai/packages/orchestrator/internal/db"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
	"github.com/rluisb/lazyai/packages/orchestrator/ports"
)

var _ ports.ChainStateStore = (*ChainStateStore)(nil)

// ChainStateStore persists chain lifecycle run state in SQLite.
type ChainStateStore struct {
	database *db.DB
}

// NewChainStateStore creates a SQLite-backed chain state store adapter.
func NewChainStateStore(database *db.DB) *ChainStateStore {
	return &ChainStateStore{database: database}
}

// SaveChainState creates or updates a chain lifecycle run state snapshot.
func (s *ChainStateStore) SaveChainState(projectRoot string, chainState *types.ChainState) error {
	encoded, err := json.Marshal(chainState)
	if err != nil {
		return fmt.Errorf("marshal chain state: %w", err)
	}
	_, err = s.database.Exec(`
		INSERT INTO chain_runs (id, definition_name, definition_version, state, current_step_id, project_root, state_json, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			state = excluded.state,
			current_step_id = excluded.current_step_id,
			state_json = excluded.state_json,
			updated_at = excluded.updated_at
	`, chainState.ChainID, chainState.DefinitionName, chainState.DefinitionVersion, chainState.State, nullableString(chainState.CurrentStepID), projectRoot, string(encoded), chainState.CreatedAt, chainState.UpdatedAt)
	if err != nil {
		return fmt.Errorf("save chain state %s: %w", chainState.ChainID, err)
	}
	return nil
}

// LoadChainState returns a chain lifecycle run state snapshot by ID.
func (s *ChainStateStore) LoadChainState(id string) (*types.ChainState, error) {
	var stateJSON string
	err := s.database.QueryRow(`SELECT state_json FROM chain_runs WHERE id = ?`, id).Scan(&stateJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("chain state not found: %s", id)
		}
		return nil, err
	}
	var chainState types.ChainState
	if err := json.Unmarshal([]byte(stateJSON), &chainState); err != nil {
		return nil, fmt.Errorf("decode chain state %s: %w", id, err)
	}
	return &chainState, nil
}
