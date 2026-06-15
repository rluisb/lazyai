package ports

import (
	"context"

	"github.com/rluisb/lazyai/packages/orchestrator/domain"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
)

// RunReadStore is the query port for chain/team/workflow dashboard run reads.
type RunReadStore interface {
	ListRuns(ctx context.Context, filter domain.RunListFilter) (domain.RunListPage, error)
	CountRunsByState(ctx context.Context) (map[string]int, error)
	FindRunRow(ctx context.Context, kind types.RunKind, id string) (domain.RunRow, error)
}
