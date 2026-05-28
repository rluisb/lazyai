// Package session provides resource locks for the LazyAI runtime.
package session

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/rluisb/lazyai/packages/cli/internal/runtime"
)

// Lock represents a resource lock.
type Lock struct {
	ID         int
	SessionID  string
	LockName   string
	HeldBy     *string
	AcquiredAt *time.Time
	ReleasedAt *time.Time
	Status     string
}

// AcquireLock attempts to acquire a lock. Returns error if already held.
func (m *Manager) AcquireLock(sessionID string, lockName string, heldBy string) (*Lock, error) {
	acquiredAt := runtime.Now()

	// Try to insert - will fail if lock_name+status='active' unique constraint exists
	// For now, we check first then insert
	var existing Lock
	var heldByStr sql.NullString
	var acquiredAtStr sql.NullString

	err := m.db.QueryRow(
		"SELECT id, session_id, lock_name, held_by, acquired_at, status FROM locks WHERE lock_name = ? AND status = ?",
		lockName, "active",
	).Scan(
		&existing.ID, &existing.SessionID, &existing.LockName, &heldByStr, &acquiredAtStr, &existing.Status,
	)
	if err != sql.ErrNoRows {
		if err != nil {
			return nil, fmt.Errorf("check existing lock: %w", err)
		}
		// Lock is held
		if heldByStr.Valid {
			existing.HeldBy = &heldByStr.String
		}
		if acquiredAtStr.Valid {
			t, _ := time.Parse(time.RFC3339, acquiredAtStr.String)
			existing.AcquiredAt = &t
		}
		return nil, fmt.Errorf("lock %s already held by %s", lockName, *existing.HeldBy)
	}

	// Insert new lock
	result, err := m.db.Exec(
		"INSERT INTO locks (session_id, lock_name, held_by, acquired_at, status) VALUES (?, ?, ?, ?, ?)",
		sessionID, lockName, heldBy, acquiredAt, "active",
	)
	if err != nil {
		return nil, fmt.Errorf("acquire lock: %w", err)
	}

	id, _ := result.LastInsertId()
	return m.GetLock(int(id))
}

// ReleaseLock releases a lock held by an agent.
func (m *Manager) ReleaseLock(lockID int) error {
	releasedAt := runtime.Now()

	_, err := m.db.Exec(
		"UPDATE locks SET status = ?, released_at = ? WHERE id = ?",
		"released", releasedAt, lockID,
	)
	if err != nil {
		return fmt.Errorf("release lock: %w", err)
	}

	return nil
}

// GetLock retrieves a lock by ID.
func (m *Manager) GetLock(lockID int) (*Lock, error) {
	var l Lock
	var heldBy sql.NullString
	var acquiredAt, releasedAt sql.NullString

	err := m.db.QueryRow(
		"SELECT id, session_id, lock_name, held_by, acquired_at, released_at, status FROM locks WHERE id = ?",
		lockID,
	).Scan(
		&l.ID, &l.SessionID, &l.LockName, &heldBy, &acquiredAt, &releasedAt, &l.Status,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("lock not found: %d", lockID)
	}
	if err != nil {
		return nil, fmt.Errorf("get lock: %w", err)
	}

	if heldBy.Valid {
		l.HeldBy = &heldBy.String
	}
	if acquiredAt.Valid {
		t, _ := time.Parse(time.RFC3339, acquiredAt.String)
		l.AcquiredAt = &t
	}
	if releasedAt.Valid {
		t, _ := time.Parse(time.RFC3339, releasedAt.String)
		l.ReleasedAt = &t
	}

	return &l, nil
}

// GetLockStatus retrieves a lock by name.
func (m *Manager) GetLockStatus(lockName string) (*Lock, error) {
	var l Lock
	var heldBy sql.NullString
	var acquiredAt, releasedAt sql.NullString

	err := m.db.QueryRow(
		"SELECT id, session_id, lock_name, held_by, acquired_at, released_at, status FROM locks WHERE lock_name = ? ORDER BY id DESC LIMIT 1",
		lockName,
	).Scan(
		&l.ID, &l.SessionID, &l.LockName, &heldBy, &acquiredAt, &releasedAt, &l.Status,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("lock not found: %s", lockName)
	}
	if err != nil {
		return nil, fmt.Errorf("get lock status: %w", err)
	}

	if heldBy.Valid {
		l.HeldBy = &heldBy.String
	}
	if acquiredAt.Valid {
		t, _ := time.Parse(time.RFC3339, acquiredAt.String)
		l.AcquiredAt = &t
	}
	if releasedAt.Valid {
		t, _ := time.Parse(time.RFC3339, releasedAt.String)
		l.ReleasedAt = &t
	}

	return &l, nil
}

// ListActiveLocks returns all active locks.
func (m *Manager) ListActiveLocks() ([]Lock, error) {
	rows, err := m.db.Query(
		"SELECT id, session_id, lock_name, held_by, acquired_at, released_at, status FROM locks WHERE status = ? ORDER BY acquired_at DESC",
		"active",
	)
	if err != nil {
		return nil, fmt.Errorf("list active locks: %w", err)
	}
	defer rows.Close()

	var locks []Lock
	for rows.Next() {
		var l Lock
		var heldBy sql.NullString
		var acquiredAt, releasedAt sql.NullString

		if err := rows.Scan(
			&l.ID, &l.SessionID, &l.LockName, &heldBy, &acquiredAt, &releasedAt, &l.Status,
		); err != nil {
			continue
		}

		if heldBy.Valid {
			l.HeldBy = &heldBy.String
		}
		if acquiredAt.Valid {
			t, _ := time.Parse(time.RFC3339, acquiredAt.String)
			l.AcquiredAt = &t
		}
		if releasedAt.Valid {
			t, _ := time.Parse(time.RFC3339, releasedAt.String)
			l.ReleasedAt = &t
		}

		locks = append(locks, l)
	}

	return locks, nil
}
