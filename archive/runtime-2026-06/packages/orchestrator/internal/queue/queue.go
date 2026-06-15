package queue

import (
	"github.com/rluisb/lazyai/packages/orchestrator/domain"
	"github.com/rluisb/lazyai/packages/orchestrator/ports"
)

// Status represents the lifecycle state of a job.
type Status = domain.QueueJobStatus

const (
	StatusPending   = domain.QueueJobStatusPending
	StatusClaimed   = domain.QueueJobStatusClaimed
	StatusCompleted = domain.QueueJobStatusCompleted
	StatusFailed    = domain.QueueJobStatusFailed
)

// Job represents a queued background job.
type Job = domain.QueueJob

// EnqueueInput describes a job to enqueue.
type EnqueueInput = domain.EnqueueJobInput

// Queue is a durable job queue service backed by a persistence port.
type Queue struct {
	store ports.JobQueueStore
}

// New creates a new queue.
func New(store ports.JobQueueStore) *Queue {
	return &Queue{store: store}
}

// Enqueue adds a job to the queue.
func (q *Queue) Enqueue(input EnqueueInput) (*Job, error) {
	return q.store.EnqueueJob(input)
}

// Dequeue atomically claims the next pending job.
func (q *Queue) Dequeue(jobType string) (*Job, error) {
	return q.store.DequeueJob(jobType)
}

// Complete marks a job as completed.
func (q *Queue) Complete(id string) error {
	return q.store.CompleteJob(id)
}

// Fail marks a job as failed; re-queues if attempts remain.
func (q *Queue) Fail(id string, errData map[string]any) error {
	return q.store.FailJob(id, errData)
}

// Reclaim resets jobs claimed longer than timeoutMs ago.
func (q *Queue) Reclaim(timeoutMs int) (int, error) {
	return q.store.ReclaimJobs(timeoutMs)
}

// Get returns a job by ID.
func (q *Queue) Get(id string) (*Job, error) {
	return q.store.GetJob(id)
}

// List returns jobs optionally filtered by status.
func (q *Queue) List(status string, limit int) ([]Job, error) {
	return q.store.ListJobs(status, limit)
}
