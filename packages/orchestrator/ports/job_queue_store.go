package ports

import "github.com/rluisb/lazyai/packages/orchestrator/domain"

// JobQueueStore is the persistence port for durable background jobs.
type JobQueueStore interface {
	EnqueueJob(input domain.EnqueueJobInput) (*domain.QueueJob, error)
	DequeueJob(jobType string) (*domain.QueueJob, error)
	CompleteJob(id string) error
	FailJob(id string, errData map[string]any) error
	ReclaimJobs(timeoutMs int) (int, error)
	GetJob(id string) (*domain.QueueJob, error)
	ListJobs(status string, limit int) ([]domain.QueueJob, error)
}
