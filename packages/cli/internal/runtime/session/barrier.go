// Package session provides synchronization barriers for the LazyAI runtime.
package session

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/rluisb/lazyai/packages/cli/internal/runtime"
)

// Barrier represents a synchronization barrier.
type Barrier struct {
	ID            int
	SessionID     string
	BarrierID     string
	ExpectedCount int
	ArrivedCount  int
	Status        string
	CreatedAt     time.Time
	ResolvedAt    *time.Time
}

// CreateBarrier creates a new barrier for a session.
func (m *Manager) CreateBarrier(sessionID string, barrierID string, expectedCount int) (*Barrier, error) {
	createdAt := runtime.Now()

	result, err := m.db.Exec(
		"INSERT INTO barriers (session_id, barrier_id, expected_count, arrived_count, status, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		sessionID, barrierID, expectedCount, 0, "waiting", createdAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create barrier: %w", err)
	}

	id, _ := result.LastInsertId()
	return m.GetBarrierByDBID(int(id))
}

// ArriveAtBarrier increments the arrived count and checks if resolved.
func (m *Manager) ArriveAtBarrier(sessionID string, barrierID string) (*Barrier, error) {
	// Use a transaction to ensure atomicity
	var barrier *Barrier
	err := m.db.WithTx(func(tx *sql.Tx) error {
		// Get current state
		var b Barrier
		var resolvedAt sql.NullString
		err := tx.QueryRow(
			"SELECT id, session_id, barrier_id, expected_count, arrived_count, status, created_at, resolved_at FROM barriers WHERE session_id = ? AND barrier_id = ?",
			sessionID, barrierID,
		).Scan(
			&b.ID, &b.SessionID, &b.BarrierID, &b.ExpectedCount, &b.ArrivedCount,
			&b.Status, &b.CreatedAt, &resolvedAt,
		)
		if err == sql.ErrNoRows {
			return fmt.Errorf("barrier not found: %s", barrierID)
		}
		if err != nil {
			return err
		}

		if resolvedAt.Valid {
			t, _ := time.Parse(time.RFC3339, resolvedAt.String)
			b.ResolvedAt = &t
		}

		// Increment arrived count
		b.ArrivedCount++

		// Check if resolved
		if b.ArrivedCount >= b.ExpectedCount {
			b.Status = "resolved"
			resolvedAt := runtime.Now()
			_, err = tx.Exec(
				"UPDATE barriers SET arrived_count = ?, status = ?, resolved_at = ? WHERE id = ?",
				b.ArrivedCount, b.Status, resolvedAt, b.ID,
			)
		} else {
			_, err = tx.Exec(
				"UPDATE barriers SET arrived_count = ? WHERE id = ?",
				b.ArrivedCount, b.ID,
			)
		}
		if err != nil {
			return err
		}

		barrier = &b
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("arrive at barrier: %w", err)
	}

	return barrier, nil
}

// GetBarrierByDBID retrieves a barrier by its database ID.
func (m *Manager) GetBarrierByDBID(id int) (*Barrier, error) {
	var b Barrier
	var resolvedAt sql.NullString

	err := m.db.QueryRow(
		"SELECT id, session_id, barrier_id, expected_count, arrived_count, status, created_at, resolved_at FROM barriers WHERE id = ?",
		id,
	).Scan(
		&b.ID, &b.SessionID, &b.BarrierID, &b.ExpectedCount, &b.ArrivedCount,
		&b.Status, &b.CreatedAt, &resolvedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("barrier not found: %d", id)
	}
	if err != nil {
		return nil, fmt.Errorf("get barrier: %w", err)
	}

	if resolvedAt.Valid {
		t, _ := time.Parse(time.RFC3339, resolvedAt.String)
		b.ResolvedAt = &t
	}

	return &b, nil
}

// GetBarrierStatus retrieves a barrier by session and barrier ID.
func (m *Manager) GetBarrierStatus(sessionID string, barrierID string) (*Barrier, error) {
	var b Barrier
	var resolvedAt sql.NullString

	err := m.db.QueryRow(
		"SELECT id, session_id, barrier_id, expected_count, arrived_count, status, created_at, resolved_at FROM barriers WHERE session_id = ? AND barrier_id = ?",
		sessionID, barrierID,
	).Scan(
		&b.ID, &b.SessionID, &b.BarrierID, &b.ExpectedCount, &b.ArrivedCount,
		&b.Status, &b.CreatedAt, &resolvedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("barrier not found: %s", barrierID)
	}
	if err != nil {
		return nil, fmt.Errorf("get barrier status: %w", err)
	}

	if resolvedAt.Valid {
		t, _ := time.Parse(time.RFC3339, resolvedAt.String)
		b.ResolvedAt = &t
	}

	return &b, nil
}

// ResolveBarrier manually resolves a barrier.
func (m *Manager) ResolveBarrier(sessionID string, barrierID string) error {
	resolvedAt := runtime.Now()

	_, err := m.db.Exec(
		"UPDATE barriers SET status = ?, resolved_at = ? WHERE session_id = ? AND barrier_id = ?",
		"resolved", resolvedAt, sessionID, barrierID,
	)
	if err != nil {
		return fmt.Errorf("resolve barrier: %w", err)
	}

	return nil
}
