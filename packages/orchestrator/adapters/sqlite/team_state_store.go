package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/rluisb/lazyai/packages/orchestrator/internal/db"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
	"github.com/rluisb/lazyai/packages/orchestrator/ports"
)

var _ ports.TeamStateStore = (*TeamStateStore)(nil)

// TeamStateStore persists team lifecycle run state in SQLite.
type TeamStateStore struct {
	database *db.DB
}

// NewTeamStateStore creates a SQLite-backed team state store adapter.
func NewTeamStateStore(database *db.DB) *TeamStateStore {
	return &TeamStateStore{database: database}
}

// SaveTeamState creates or updates a team lifecycle run state snapshot.
func (s *TeamStateStore) SaveTeamState(projectRoot string, teamState *types.TeamState) error {
	encoded, err := json.Marshal(teamState)
	if err != nil {
		return fmt.Errorf("marshal team state: %w", err)
	}
	_, err = s.database.Exec(`
		INSERT INTO team_runs (id, definition_name, definition_version, state, project_root, state_json, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			state = excluded.state,
			state_json = excluded.state_json,
			updated_at = excluded.updated_at
	`, teamState.TeamID, teamState.DefinitionName, teamState.DefinitionVersion, teamState.State, projectRoot, string(encoded), teamState.CreatedAt, teamState.UpdatedAt)
	if err != nil {
		return fmt.Errorf("save team state %s: %w", teamState.TeamID, err)
	}
	return nil
}

// LoadTeamState returns a team lifecycle run state snapshot by ID.
func (s *TeamStateStore) LoadTeamState(id string) (*types.TeamState, error) {
	var stateJSON string
	err := s.database.QueryRow(`SELECT state_json FROM team_runs WHERE id = ?`, id).Scan(&stateJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("team state not found: %s", id)
		}
		return nil, err
	}
	var teamState types.TeamState
	if err := json.Unmarshal([]byte(stateJSON), &teamState); err != nil {
		return nil, fmt.Errorf("decode team state %s: %w", id, err)
	}
	return &teamState, nil
}
