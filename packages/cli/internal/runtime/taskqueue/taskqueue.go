// Package taskqueue provides atomic task claiming, dead-letter queue, and zombie sweeping.
package taskqueue

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/rluisb/lazyai/packages/cli/internal/runtime"
)

// Manager handles task queue operations.
type Manager struct {
	db *runtime.DB
}

// NewManager creates a task queue manager backed by the given database.
func NewManager(db *runtime.DB) *Manager {
	return &Manager{db: db}
}

// TaskStatus represents the lifecycle state of a task.
type TaskStatus string

const (
	TaskOpen      TaskStatus = "open"
	TaskClaimed   TaskStatus = "claimed"
	TaskCompleted TaskStatus = "completed"
	TaskFailed    TaskStatus = "failed"
)

// Task represents a task in the queue.
type Task struct {
	ID        int
	SessionID string
	Topic     string
	Task      string
	Status    TaskStatus
	MaxAgents int
	DedupeKey *string
	CreatedAt time.Time
	Claims    []TaskClaim
}

// TaskClaim represents an agent's claim on a task.
type TaskClaim struct {
	TaskID    int
	Agent     string
	ClaimedAt time.Time
}

// EnqueueOptions configures a new task.
type EnqueueOptions struct {
	MaxAgents int
	DedupeKey *string
}

// Enqueue adds a task to the queue.
func (m *Manager) Enqueue(sessionID string, topic string, task string, opts EnqueueOptions) (*Task, error) {
	createdAt := runtime.Now()
	maxAgents := opts.MaxAgents
	if maxAgents <= 0 {
		maxAgents = 1
	}

	result, err := m.db.Exec(
		"INSERT INTO task_queue (session_id, topic, task, status, max_agents, dedupe_key, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		sessionID, topic, task, string(TaskOpen), maxAgents, opts.DedupeKey, createdAt,
	)
	if err != nil {
		return nil, fmt.Errorf("enqueue task: %w", err)
	}

	id, _ := result.LastInsertId()
	return m.GetTask(int(id))
}

// GetTask retrieves a task by ID with its claims.
func (m *Manager) GetTask(taskID int) (*Task, error) {
	var t Task
	var dedupeKey sql.NullString

	err := m.db.QueryRow(
		"SELECT id, session_id, topic, task, status, max_agents, dedupe_key, created_at FROM task_queue WHERE id = ?",
		taskID,
	).Scan(
		&t.ID, &t.SessionID, &t.Topic, &t.Task, &t.Status, &t.MaxAgents, &dedupeKey, &t.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task not found: %d", taskID)
	}
	if err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}

	if dedupeKey.Valid {
		t.DedupeKey = &dedupeKey.String
	}

	// Load claims
	claims, err := m.GetClaims(taskID)
	if err != nil {
		return nil, err
	}
	t.Claims = claims

	return &t, nil
}

// GetClaims returns all claims for a task.
func (m *Manager) GetClaims(taskID int) ([]TaskClaim, error) {
	rows, err := m.db.Query(
		"SELECT task_id, agent, claimed_at FROM task_claims WHERE task_id = ?",
		taskID,
	)
	if err != nil {
		return nil, fmt.Errorf("get claims: %w", err)
	}
	defer rows.Close()

	var claims []TaskClaim
	for rows.Next() {
		var c TaskClaim
		if err := rows.Scan(&c.TaskID, &c.Agent, &c.ClaimedAt); err != nil {
			continue
		}
		claims = append(claims, c)
	}

	return claims, nil
}

// List returns tasks in the queue, optionally filtered by session and status.
func (m *Manager) List(sessionID string, status TaskStatus) ([]Task, error) {
	var rows *sql.Rows
	var err error

	if status != "" {
		rows, err = m.db.Query(
			"SELECT id, session_id, topic, task, status, max_agents, dedupe_key, created_at FROM task_queue WHERE session_id = ? AND status = ? ORDER BY created_at ASC",
			sessionID, string(status),
		)
	} else {
		rows, err = m.db.Query(
			"SELECT id, session_id, topic, task, status, max_agents, dedupe_key, created_at FROM task_queue WHERE session_id = ? ORDER BY created_at ASC",
			sessionID,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		var dedupeKey sql.NullString

		if err := rows.Scan(
			&t.ID, &t.SessionID, &t.Topic, &t.Task, &t.Status, &t.MaxAgents, &dedupeKey, &t.CreatedAt,
		); err != nil {
			continue
		}

		if dedupeKey.Valid {
			t.DedupeKey = &dedupeKey.String
		}

		tasks = append(tasks, t)
	}

	return tasks, nil
}
