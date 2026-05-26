// Package taskqueue provides atomic task claiming for the LazyAI runtime.
package taskqueue

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/rluisb/lazyai/packages/cli/internal/runtime"
)

// Claim atomically claims an available task for the given agent.
// Uses a CTE with RETURNING for atomicity.
// Returns nil if no tasks are available.
func (m *Manager) Claim(sessionID string, topic string, agent string) (*Task, error) {
	claimedAt := runtime.Now()

	// Retry logic: 3 attempts with 1s backoff
	var taskID *int
	var lastErr error

	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(1 * time.Second)
		}

		// Check if this agent already claimed a task in a previous attempt
		// (handles race where we got the lock but another agent succeeded)
		existingTaskID, err := m.findExistingClaim(sessionID, topic, agent)
		if err != nil {
			lastErr = err
			continue
		}
		if existingTaskID != nil {
			taskID = existingTaskID
			break
		}

		// Attempt atomic claim using CTE
		var id sql.NullInt64
		err = m.db.QueryRow(
			`WITH available_task AS (
				SELECT q.id
				FROM task_queue q
				LEFT JOIN task_claims c ON q.id = c.task_id
				WHERE q.session_id = ? AND q.topic = ? AND q.status = 'open'
				  AND q.id NOT IN (SELECT task_id FROM task_claims WHERE agent = ?)
				GROUP BY q.id
				HAVING COUNT(c.agent) < q.max_agents
				ORDER BY q.created_at ASC
				LIMIT 1
			)
			INSERT INTO task_claims (task_id, agent, claimed_at)
			SELECT id, ?, ? FROM available_task
			RETURNING task_id`,
			sessionID, topic, agent, agent, claimedAt,
		).Scan(&id)

		if err == sql.ErrNoRows {
			// No tasks available
			return nil, nil
		}
		if err != nil {
			// Check for database locked (SQLite busy)
			if isBusyError(err) {
				lastErr = err
				continue
			}
			return nil, fmt.Errorf("claim task: %w", err)
		}

		if id.Valid {
			pid := int(id.Int64)
			taskID = &pid
			break
		}
	}

	if taskID == nil {
		if lastErr != nil {
			return nil, fmt.Errorf("claim failed after retries: %w", lastErr)
		}
		return nil, nil // No tasks available
	}

	// Update task status to claimed
	_, err := m.db.Exec(
		"UPDATE task_queue SET status = ? WHERE id = ?",
		string(TaskClaimed), *taskID,
	)
	if err != nil {
		return nil, fmt.Errorf("update task status: %w", err)
	}

	return m.GetTask(*taskID)
}

// findExistingClaim checks if an agent already claimed a task.
func (m *Manager) findExistingClaim(sessionID string, topic string, agent string) (*int, error) {
	var taskID int
	err := m.db.QueryRow(
		`SELECT tc.task_id 
		FROM task_claims tc
		JOIN task_queue q ON tc.task_id = q.id
		WHERE q.session_id = ? AND q.topic = ? AND tc.agent = ?
		LIMIT 1`,
		sessionID, topic, agent,
	).Scan(&taskID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &taskID, nil
}

// isBusyError checks if the error is a SQLite busy/locked error.
func isBusyError(err error) bool {
	if err == nil {
		return false
	}
	// SQLite busy error messages contain "database is locked"
	return strings.Contains(err.Error(), "database is locked") ||
		strings.Contains(err.Error(), "busy")
}
