// Package session provides parallel task tracking for the LazyAI runtime.
package session

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/rluisb/lazyai/packages/cli/internal/runtime"
)

// ParallelTask represents a task executed in parallel.
type ParallelTask struct {
	ID               int
	SessionID        string
	ParentDispatchID *int
	WaveID           string
	Agent            string
	Task             string
	Status           string
	StartedAt        *time.Time
	CompletedAt      *time.Time
	Result           string
	OutputPath       string
	ErrorMessage     string
	CreatedAt        time.Time
}

// CreateParallelTask creates a new parallel task for a session.
func (m *Manager) CreateParallelTask(sessionID string, waveID string, agent string, task string, parentDispatchID *int) (*ParallelTask, error) {
	createdAt := runtime.Now()

	result, err := m.db.Exec(
		"INSERT INTO parallel_tasks (session_id, parent_dispatch_id, wave_id, agent, task, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		sessionID, parentDispatchID, waveID, agent, task, "pending", createdAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create parallel task: %w", err)
	}

	id, _ := result.LastInsertId()
	return m.GetParallelTask(int(id))
}

// UpdateParallelTask updates the status and result of a parallel task.
func (m *Manager) UpdateParallelTask(taskID int, status string, result string, outputPath string, errorMessage string) error {
	completedAt := runtime.Now()

	_, err := m.db.Exec(
		"UPDATE parallel_tasks SET status = ?, result = ?, output_path = ?, error_message = ?, completed_at = ? WHERE id = ?",
		status, result, outputPath, errorMessage, completedAt, taskID,
	)
	if err != nil {
		return fmt.Errorf("update parallel task: %w", err)
	}

	return nil
}

// GetParallelTask retrieves a parallel task by ID.
func (m *Manager) GetParallelTask(taskID int) (*ParallelTask, error) {
	var pt ParallelTask
	var startedAt, completedAt sql.NullString
	var parentID sql.NullInt64

	err := m.db.QueryRow(
		"SELECT id, session_id, parent_dispatch_id, wave_id, agent, task, status, started_at, completed_at, result, output_path, error_message, created_at FROM parallel_tasks WHERE id = ?",
		taskID,
	).Scan(
		&pt.ID, &pt.SessionID, &parentID, &pt.WaveID, &pt.Agent, &pt.Task,
		&pt.Status, &startedAt, &completedAt, &pt.Result, &pt.OutputPath,
		&pt.ErrorMessage, &pt.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("parallel task not found: %d", taskID)
	}
	if err != nil {
		return nil, fmt.Errorf("get parallel task: %w", err)
	}

	if parentID.Valid {
		pid := int(parentID.Int64)
		pt.ParentDispatchID = &pid
	}

	if startedAt.Valid {
		t, _ := time.Parse(time.RFC3339, startedAt.String)
		pt.StartedAt = &t
	}
	if completedAt.Valid {
		t, _ := time.Parse(time.RFC3339, completedAt.String)
		pt.CompletedAt = &t
	}

	return &pt, nil
}

// ListParallelTasks returns all parallel tasks for a session.
func (m *Manager) ListParallelTasks(sessionID string) ([]ParallelTask, error) {
	rows, err := m.db.Query(
		"SELECT id, session_id, parent_dispatch_id, wave_id, agent, task, status, started_at, completed_at, result, output_path, error_message, created_at FROM parallel_tasks WHERE session_id = ? ORDER BY created_at DESC",
		sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("list parallel tasks: %w", err)
	}
	defer rows.Close()

	var tasks []ParallelTask
	for rows.Next() {
		var pt ParallelTask
		var startedAt, completedAt sql.NullString
		var parentID sql.NullInt64

		if err := rows.Scan(
			&pt.ID, &pt.SessionID, &parentID, &pt.WaveID, &pt.Agent, &pt.Task,
			&pt.Status, &startedAt, &completedAt, &pt.Result, &pt.OutputPath,
			&pt.ErrorMessage, &pt.CreatedAt,
		); err != nil {
			continue
		}

		if parentID.Valid {
			pid := int(parentID.Int64)
			pt.ParentDispatchID = &pid
		}

		if startedAt.Valid {
			t, _ := time.Parse(time.RFC3339, startedAt.String)
			pt.StartedAt = &t
		}
		if completedAt.Valid {
			t, _ := time.Parse(time.RFC3339, completedAt.String)
			pt.CompletedAt = &t
		}

		tasks = append(tasks, pt)
	}

	return tasks, nil
}

// GetWaveSummary aggregates results for a wave.
func (m *Manager) GetWaveSummary(sessionID string, waveID string) (map[string]int, error) {
	rows, err := m.db.Query(
		"SELECT status, COUNT(*) FROM parallel_tasks WHERE session_id = ? AND wave_id = ? GROUP BY status",
		sessionID, waveID,
	)
	if err != nil {
		return nil, fmt.Errorf("get wave summary: %w", err)
	}
	defer rows.Close()

	summary := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			continue
		}
		summary[status] = count
	}

	return summary, nil
}
