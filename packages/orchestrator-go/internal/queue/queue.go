package queue

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ricardoborges-teachable/ai-setup/packages/orchestrator-go/internal/db"
)

// Status represents the lifecycle state of a job.
type Status string

const (
	StatusPending   Status = "pending"
	StatusClaimed   Status = "claimed"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
)

// Job represents a queued background job.
type Job struct {
	ID           string         `json:"id"`
	JobType      string         `json:"jobType"`
	Payload      map[string]any `json:"payload"`
	Status       Status         `json:"status"`
	Priority     int            `json:"priority"`
	Attempts     int            `json:"attempts"`
	MaxAttempts  int            `json:"maxAttempts"`
	Error        map[string]any `json:"error,omitempty"`
	CreatedAt    string         `json:"createdAt"`
	ClaimedAt    string         `json:"claimedAt,omitempty"`
	CompletedAt  string         `json:"completedAt,omitempty"`
}

// EnqueueInput describes a job to enqueue.
type EnqueueInput struct {
	JobType     string
	Payload     map[string]any
	Priority    int
	MaxAttempts int
}

// Queue is a durable job queue backed by SQLite.
type Queue struct {
	db *db.DB
}

// New creates a new queue.
func New(database *db.DB) *Queue {
	return &Queue{db: database}
}

// Enqueue adds a job to the queue.
func (q *Queue) Enqueue(input EnqueueInput) (*Job, error) {
	id := uuid.NewString()
	now := time.Now().UTC().Format(time.RFC3339)
	payloadJSON, _ := json.Marshal(input.Payload)
	maxAttempts := input.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 3
	}

	_, err := q.db.Exec(
		`INSERT INTO queue_jobs (id, job_type, payload_json, priority, max_attempts, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		id, input.JobType, string(payloadJSON), input.Priority, maxAttempts, now)
	if err != nil {
		return nil, err
	}
	return q.Get(id)
}

// Dequeue atomically claims the next pending job.
func (q *Queue) Dequeue(jobType string) (*Job, error) {
	now := time.Now().UTC().Format(time.RFC3339)

	var query string
	var args []any
	if jobType != "" {
		query = `SELECT id FROM queue_jobs WHERE status = 'pending' AND job_type = ? ORDER BY priority DESC, created_at ASC LIMIT 1`
		args = []any{jobType}
	} else {
		query = `SELECT id FROM queue_jobs WHERE status = 'pending' ORDER BY priority DESC, created_at ASC LIMIT 1`
		args = nil
	}

	var jobID string
	err := q.db.QueryRow(query, args...).Scan(&jobID)
	if err != nil {
		return nil, nil // empty queue
	}

	q.db.Exec(`UPDATE queue_jobs SET status = 'claimed', claimed_at = ?, attempts = attempts + 1 WHERE id = ?`, now, jobID)
	return q.Get(jobID)
}

// Complete marks a job as completed.
func (q *Queue) Complete(id string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := q.db.Exec(`UPDATE queue_jobs SET status = 'completed', completed_at = ? WHERE id = ?`, now, id)
	return err
}

// Fail marks a job as failed; re-queues if attempts remain.
func (q *Queue) Fail(id string, errData map[string]any) error {
	errJSON, _ := json.Marshal(errData)
	now := time.Now().UTC().Format(time.RFC3339)

	var attempts, maxAttempts int
	q.db.QueryRow(`SELECT attempts, max_attempts FROM queue_jobs WHERE id = ?`, id).Scan(&attempts, &maxAttempts)

	var nextStatus string
	var completedAt string
	if attempts < maxAttempts {
		nextStatus = "pending"
	} else {
		nextStatus = "failed"
		completedAt = now
	}

	if completedAt != "" {
		_, err := q.db.Exec(
			`UPDATE queue_jobs SET status = ?, error_json = ?, completed_at = ? WHERE id = ?`,
			nextStatus, string(errJSON), completedAt, id)
		return err
	}
	_, err := q.db.Exec(
		`UPDATE queue_jobs SET status = ?, error_json = ?, claimed_at = NULL WHERE id = ?`,
		nextStatus, string(errJSON), id)
	return err
}

// Reclaim resets jobs claimed longer than timeoutMs ago.
func (q *Queue) Reclaim(timeoutMs int) (int, error) {
	cutoff := time.Now().Add(-time.Duration(timeoutMs) * time.Millisecond).UTC().Format(time.RFC3339)
	result, err := q.db.Exec(
		`UPDATE queue_jobs SET status = 'pending', claimed_at = NULL WHERE status = 'claimed' AND claimed_at < ?`,
		cutoff)
	if err != nil {
		return 0, err
	}
	n, _ := result.RowsAffected()
	return int(n), nil
}

// Get returns a job by ID.
func (q *Queue) Get(id string) (*Job, error) {
	var j Job
	var payloadJSON, errorJSON string

	err := q.db.QueryRow(
		`SELECT id, job_type, payload_json, status, priority, attempts, max_attempts, COALESCE(error_json,''), created_at, COALESCE(claimed_at,''), COALESCE(completed_at,'') FROM queue_jobs WHERE id = ?`, id).
		Scan(&j.ID, &j.JobType, &payloadJSON, &j.Status, &j.Priority, &j.Attempts, &j.MaxAttempts, &errorJSON, &j.CreatedAt, &j.ClaimedAt, &j.CompletedAt)
	if err != nil {
		return nil, fmt.Errorf("job not found: %s", id)
	}
	json.Unmarshal([]byte(payloadJSON), &j.Payload)
	if errorJSON != "" {
		json.Unmarshal([]byte(errorJSON), &j.Error)
	}
	return &j, nil
}

// List returns jobs optionally filtered by status.
func (q *Queue) List(status string, limit int) ([]Job, error) {
	var query string
	var args []any
	if status != "" {
		query = `SELECT id, job_type, payload_json, status, created_at FROM queue_jobs WHERE status = ? ORDER BY created_at DESC LIMIT ?`
		args = []any{status, limit}
	} else {
		query = `SELECT id, job_type, payload_json, status, created_at FROM queue_jobs ORDER BY created_at DESC LIMIT ?`
		args = []any{limit}
	}

	rows, err := q.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []Job
	for rows.Next() {
		var j Job
		var payloadJSON string
		if err := rows.Scan(&j.ID, &j.JobType, &payloadJSON, &j.Status, &j.CreatedAt); err != nil {
			continue
		}
		json.Unmarshal([]byte(payloadJSON), &j.Payload)
		jobs = append(jobs, j)
	}
	return jobs, rows.Err()
}
