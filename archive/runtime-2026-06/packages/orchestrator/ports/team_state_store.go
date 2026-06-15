package ports

import "github.com/rluisb/lazyai/packages/orchestrator/internal/types"

// TeamStateStore is the persistence port for team lifecycle run state.
type TeamStateStore interface {
	SaveTeamState(projectRoot string, teamState *types.TeamState) error
	LoadTeamState(id string) (*types.TeamState, error)
}
