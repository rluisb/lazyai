// Package taskqueue provides task lifecycle management (complete, fail) for the LazyAI runtime.
package taskqueue

import (
	"fmt"

	"github.com/rluisb/lazyai/packages/cli/internal/runtime"
)

// Complete marks a task as completed and removes the agent's claim.
func (m *Manager) Complete(taskID int, agent string) error {
	return m.db.WithTx(func(tx *sql.Tx) error {
		// Remove claim
		_, err := tx.Exec(
			"DELETE FROM task_claims WHERE task_id = ? AND agent = ?",
			taskID, agent,
		)
		if err != nil {
			return fmt.Errorf("remove claim: %w", err)
		}

		// Update task status
		_, err = tx.Exec(
			"UPDATE task_queue SET status = ? WHERE id = ?",
			string(TaskCompleted), taskID,
		)
		if err != nil {
			return fmt.Errorf("update task status: %w", err)
		}

		return nil
	})
}

// Fail marks a task as failed, removes the agent's claim, and creates a DLQ entry.
func (m *Manager) Fail(taskID int, agent string, errorMessage string, contextDump string) error {
	failedAt := runtime.Now()

	return m.db.WithTx(func(tx *sql.Tx) error {
		// Remove claim
		_, err := tx.Exec(
			"DELETE FROM task_claims WHERE task_id = ? AND agent = ?",
			taskID, agent,
		)
		if err != nil {
			return fmt.Errorf("remove claim: %w", err)
		}

		// Update task status
		_, err = tx.Exec(
			"UPDATE task_queue SET status = ? WHERE id = ?",
			string(TaskFailed), taskID,
		)
		if err != nil {
			return fmt.Errorf("update task status: %w", err)
		}

		// Insert into DLQ
		_, err = tx.Exec(
			"INSERT INTO task_dlq (task_id, failed_agent, error_message, context_dump, failed_at) VALUES (?, ?, ?, ?, ?)",
			taskID, agent, errorMessage, contextDump, failedAt,
		)
		if err != nil {
			return fmt.Errorf("create dlq entry: %w", err)
		}

		return nil
	})
}

// DLQEntry represents a dead letter queue entry.
type DLQEntry struct {
	ID           int
	TaskID       int
	FailedAgent  string
	ErrorMessage string
	ContextDump  string
	FailedAt     time.Time
}

// GetDLQ returns all DLQ entries for a session.
func (m *Manager) GetDLQ(sessionID string) ([]DLQEntry, error) {
	rows, err := m.db.Query(
		`SELECT d.id, d.task_id, d.failed_agent, d.error_message, d.context_dump, d.failed_at
		FROM task_dlq d
		JOIN task_queue q ON d.task_id = q.id
		WHERE q.session_id = ?
		ORDER BY d.failed_at DESC`,
		sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("get dlq: %w", err)
	}
	defer rows.Close()

	var entries []DLQEntry
	for rows.Next() {
		var e DLQEntry
		if err := rows.Scan(
			&e.ID, &e.TaskID, &e.FailedAgent, &e.ErrorMessage, &e.ContextDump, &e.FailedAt,
		); err != nil {
			continue
		}
		entries = append(entries, e)
	}

	return entries, nil
}
