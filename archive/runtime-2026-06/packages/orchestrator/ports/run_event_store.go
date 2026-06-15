package ports

import "github.com/rluisb/lazyai/packages/orchestrator/domain"

// RunEventStore is the persistence port for run lifecycle events.
type RunEventStore interface {
	AppendRunEvent(runID, eventType string, data map[string]any, createdAt string) error
	ReplayRunEvents(runID string, sinceID int) ([]domain.RunEvent, error)
	ReplayAllRunEvents(sinceID, limit int) ([]domain.RunEvent, error)
}
