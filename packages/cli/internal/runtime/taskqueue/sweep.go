// Package taskqueue provides zombie sweeping for the LazyAI runtime.
package taskqueue

import (
	"fmt"
)

// Sweep removes stale claims from agents that have crashed or timed out.
// Only sweeps claims for tasks still in 'open' status.
// Returns the number of removed claims.
func (m *Manager) Sweep(timeoutSeconds int) (int, error) {
	if timeoutSeconds <= 0 {
		timeoutSeconds = 300 // Default 5 minutes
	}

	// Delete stale claims
	result, err := m.db.Exec(
		`DELETE FROM task_claims
		WHERE claimed_at < datetime('now', ?)
		  AND task_id IN (SELECT id FROM task_queue WHERE status = 'open')`,
		fmt.Sprintf("-%d seconds", timeoutSeconds),
	)
	if err != nil {
		return 0, fmt.Errorf("sweep claims: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	return int(rowsAffected), nil
}
