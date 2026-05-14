package ports

import "github.com/rluisb/lazyai/packages/orchestrator/internal/types"

// ChainStateStore is the persistence port for chain lifecycle run state.
type ChainStateStore interface {
	SaveChainState(projectRoot string, chainState *types.ChainState) error
	LoadChainState(id string) (*types.ChainState, error)
}
