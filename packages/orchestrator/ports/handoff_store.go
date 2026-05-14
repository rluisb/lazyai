package ports

import (
	"context"

	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
)

// HandoffStore is the persistence port for resumable handoff documents.
type HandoffStore interface {
	SaveHandoffDocument(doc *types.HandoffDocument) error
}

// HandoffQueryStore is the read-side persistence port for resumable handoff documents.
type HandoffQueryStore interface {
	ListHandoffDocuments(ctx context.Context, kind types.RunKind, runID string) ([]types.HandoffDocument, error)
}
