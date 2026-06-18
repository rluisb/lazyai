package ports

import (
	"context"

	"github.com/rluisb/lazyai/packages/orchestrator/domain"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
)

// ErrorJournalStore is the query port for persisted run error journal entries.
type ErrorJournalStore interface {
	ListErrorJournalEntries(ctx context.Context, runID string, limit int) ([]domain.ErrorJournalEntry, error)
	CountErrorJournalEntry(ctx context.Context, kind types.RunKind, id string) (int, error)
	CountErrorJournalEntriesByRun(ctx context.Context, refs []domain.RunRef) (map[domain.RunRef]int, error)
}
