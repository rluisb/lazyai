package sqlite

import (
	"context"

	"github.com/rluisb/lazyai/packages/orchestrator/domain"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/db"
	"github.com/rluisb/lazyai/packages/orchestrator/ports"
)

var _ ports.ActivityStore = (*ActivityStore)(nil)

// ActivityStore reads active persisted run/job counts from the orchestrator SQLite database.
type ActivityStore struct {
	database *db.DB
}

// NewActivityStore creates a SQLite-backed activity store adapter.
func NewActivityStore(database *db.DB) *ActivityStore {
	return &ActivityStore{database: database}
}

// ActiveRunCounts counts persisted runs/jobs that are still active or in progress.
func (s *ActivityStore) ActiveRunCounts(ctx context.Context) (domain.ActiveRunCounts, error) {
	counts := domain.ActiveRunCounts{}
	var err error

	counts.Chains, err = s.countRows(ctx, `SELECT COUNT(*) FROM chain_runs WHERE state IN ('created', 'running', 'gated', 'paused')`)
	if err != nil {
		return counts, err
	}
	counts.Teams, err = s.countRows(ctx, `SELECT COUNT(*) FROM team_runs WHERE state IN ('created', 'running', 'synthesizing', 'paused')`)
	if err != nil {
		return counts, err
	}
	counts.Workflows, err = s.countRows(ctx, `SELECT COUNT(*) FROM workflow_runs WHERE state IN ('created', 'running', 'waiting_on_child', 'gated', 'awaiting_recovery', 'paused')`)
	if err != nil {
		return counts, err
	}
	counts.QueueJobs, err = s.countRows(ctx, `SELECT COUNT(*) FROM queue_jobs WHERE status IN ('pending', 'claimed')`)
	if err != nil {
		return counts, err
	}
	counts.Total = counts.Chains + counts.Teams + counts.Workflows + counts.QueueJobs
	return counts, nil
}

func (s *ActivityStore) countRows(ctx context.Context, query string) (int, error) {
	var count int
	if err := s.database.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}
