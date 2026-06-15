package domain

// QueueJobStatus represents the lifecycle state of a queued background job.
type QueueJobStatus string

const (
	QueueJobStatusPending   QueueJobStatus = "pending"
	QueueJobStatusClaimed   QueueJobStatus = "claimed"
	QueueJobStatusCompleted QueueJobStatus = "completed"
	QueueJobStatusFailed    QueueJobStatus = "failed"
)

// QueueJob represents a queued background job.
type QueueJob struct {
	ID          string         `json:"id"`
	JobType     string         `json:"jobType"`
	Payload     map[string]any `json:"payload"`
	Status      QueueJobStatus `json:"status"`
	Priority    int            `json:"priority"`
	Attempts    int            `json:"attempts"`
	MaxAttempts int            `json:"maxAttempts"`
	Error       map[string]any `json:"error,omitempty"`
	CreatedAt   string         `json:"createdAt"`
	ClaimedAt   string         `json:"claimedAt,omitempty"`
	CompletedAt string         `json:"completedAt,omitempty"`
}

// EnqueueJobInput describes a job to enqueue.
type EnqueueJobInput struct {
	JobType     string
	Payload     map[string]any
	Priority    int
	MaxAttempts int
}
