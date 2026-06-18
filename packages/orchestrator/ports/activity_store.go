package ports

import (
	"context"

	"github.com/rluisb/lazyai/packages/orchestrator/domain"
)

// ActivityStore is the persistence port for active persisted run/job counts.
type ActivityStore interface {
	ActiveRunCounts(ctx context.Context) (domain.ActiveRunCounts, error)
}
