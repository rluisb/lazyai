package sqlite

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rluisb/lazyai/packages/orchestrator/domain"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/db"
)

// JobQueueStore persists durable background jobs in the orchestrator SQLite database.
type JobQueueStore struct {
	database *db.DB
}

// NewJobQueueStore creates a SQLite-backed job queue store adapter.
func NewJobQueueStore(database *db.DB) *JobQueueStore {
	return &JobQueueStore{database: database}
}

// EnqueueJob adds a job to the queue.
func (s *JobQueueStore) EnqueueJob(input domain.EnqueueJobInput) (*domain.QueueJob, error) {
	id := uuid.NewString()
	now := time.Now().UTC().Format(time.RFC3339)
	payloadJSON, _ := json.Marshal(input.Payload)
	maxAttempts := input.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 3
	}

	_, err := s.database.Exec(
		`INSERT INTO queue_jobs (id, job_type, payload_json, priority, max_attempts, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		id, input.JobType, string(payloadJSON), input.Priority, maxAttempts, now)
	if err != nil {
		return nil, err
	}
	return s.GetJob(id)
}

// DequeueJob atomically claims the next pending job.
func (s *JobQueueStore) DequeueJob(jobType string) (*domain.QueueJob, error) {
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
	err := s.database.QueryRow(query, args...).Scan(&jobID)
	if err != nil {
		return nil, nil // empty queue
	}

	s.database.Exec(`UPDATE queue_jobs SET status = 'claimed', claimed_at = ?, attempts = attempts + 1 WHERE id = ?`, now, jobID)
	return s.GetJob(jobID)
}

// CompleteJob marks a job as completed.
func (s *JobQueueStore) CompleteJob(id string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.database.Exec(`UPDATE queue_jobs SET status = 'completed', completed_at = ? WHERE id = ?`, now, id)
	return err
}

// FailJob marks a job as failed; re-queues if attempts remain.
func (s *JobQueueStore) FailJob(id string, errData map[string]any) error {
	errJSON, _ := json.Marshal(errData)
	now := time.Now().UTC().Format(time.RFC3339)

	var attempts, maxAttempts int
	s.database.QueryRow(`SELECT attempts, max_attempts FROM queue_jobs WHERE id = ?`, id).Scan(&attempts, &maxAttempts)

	var nextStatus string
	var completedAt string
	if attempts < maxAttempts {
		nextStatus = string(domain.QueueJobStatusPending)
	} else {
		nextStatus = string(domain.QueueJobStatusFailed)
		completedAt = now
	}

	if completedAt != "" {
		_, err := s.database.Exec(
			`UPDATE queue_jobs SET status = ?, error_json = ?, completed_at = ? WHERE id = ?`,
			nextStatus, string(errJSON), completedAt, id)
		return err
	}
	_, err := s.database.Exec(
		`UPDATE queue_jobs SET status = ?, error_json = ?, claimed_at = NULL WHERE id = ?`,
		nextStatus, string(errJSON), id)
	return err
}

// ReclaimJobs resets jobs claimed longer than timeoutMs ago.
func (s *JobQueueStore) ReclaimJobs(timeoutMs int) (int, error) {
	cutoff := time.Now().Add(-time.Duration(timeoutMs) * time.Millisecond).UTC().Format(time.RFC3339)
	result, err := s.database.Exec(
		`UPDATE queue_jobs SET status = 'pending', claimed_at = NULL WHERE status = 'claimed' AND claimed_at < ?`,
		cutoff)
	if err != nil {
		return 0, err
	}
	n, _ := result.RowsAffected()
	return int(n), nil
}

// GetJob returns a job by ID.
func (s *JobQueueStore) GetJob(id string) (*domain.QueueJob, error) {
	var j domain.QueueJob
	var payloadJSON, errorJSON string

	err := s.database.QueryRow(
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

// ListJobs returns jobs optionally filtered by status.
func (s *JobQueueStore) ListJobs(status string, limit int) ([]domain.QueueJob, error) {
	var query string
	var args []any
	if status != "" {
		query = `SELECT id, job_type, payload_json, status, created_at FROM queue_jobs WHERE status = ? ORDER BY created_at DESC LIMIT ?`
		args = []any{status, limit}
	} else {
		query = `SELECT id, job_type, payload_json, status, created_at FROM queue_jobs ORDER BY created_at DESC LIMIT ?`
		args = []any{limit}
	}

	rows, err := s.database.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []domain.QueueJob
	for rows.Next() {
		var j domain.QueueJob
		var payloadJSON string
		if err := rows.Scan(&j.ID, &j.JobType, &payloadJSON, &j.Status, &j.CreatedAt); err != nil {
			continue
		}
		json.Unmarshal([]byte(payloadJSON), &j.Payload)
		jobs = append(jobs, j)
	}
	return jobs, rows.Err()
}
