package sqlite

import (
	"encoding/json"

	"github.com/rluisb/lazyai/packages/orchestrator/domain"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/db"
	"github.com/rluisb/lazyai/packages/orchestrator/ports"
)

var _ ports.RunEventStore = (*RunEventStore)(nil)

// RunEventStore persists run events in the orchestrator SQLite database.
type RunEventStore struct {
	database *db.DB
}

// NewRunEventStore creates a SQLite-backed run event store adapter.
func NewRunEventStore(database *db.DB) *RunEventStore {
	return &RunEventStore{database: database}
}

// AppendRunEvent records a run event.
func (s *RunEventStore) AppendRunEvent(runID, eventType string, data map[string]any, createdAt string) error {
	dataJSON, _ := json.Marshal(data)
	_, err := s.database.Exec(
		`INSERT INTO run_events (run_id, event_type, event_json, created_at) VALUES (?, ?, ?, ?)`,
		runID, eventType, string(dataJSON), createdAt)
	return err
}

// ReplayRunEvents returns events for one run since a given event ID.
func (s *RunEventStore) ReplayRunEvents(runID string, sinceID int) ([]domain.RunEvent, error) {
	var query string
	var args []any
	if sinceID > 0 {
		query = `SELECT id, run_id, event_type, event_json, created_at FROM run_events WHERE run_id = ? AND id > ? ORDER BY id ASC`
		args = []any{runID, sinceID}
	} else {
		query = `SELECT id, run_id, event_type, event_json, created_at FROM run_events WHERE run_id = ? ORDER BY id ASC`
		args = []any{runID}
	}

	rows, err := s.database.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanRunEvents(rows), rows.Err()
}

// ReplayAllRunEvents returns events across all runs since a given event ID.
func (s *RunEventStore) ReplayAllRunEvents(sinceID, limit int) ([]domain.RunEvent, error) {
	if limit <= 0 {
		return nil, nil
	}
	var query string
	var args []any
	if sinceID > 0 {
		query = `SELECT id, run_id, event_type, event_json, created_at FROM run_events WHERE id > ? ORDER BY id ASC LIMIT ?`
		args = []any{sinceID, limit}
	} else {
		query = `SELECT id, run_id, event_type, event_json, created_at FROM run_events ORDER BY id ASC LIMIT ?`
		args = []any{limit}
	}

	rows, err := s.database.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanRunEvents(rows), rows.Err()
}

// scanRunEvents reads rows from a `SELECT id, run_id, event_type, event_json, created_at` query.
func scanRunEvents(rows interface {
	Next() bool
	Scan(dest ...any) error
}) []domain.RunEvent {
	var events []domain.RunEvent
	for rows.Next() {
		var e domain.RunEvent
		var dataJSON string
		if err := rows.Scan(&e.ID, &e.RunID, &e.Type, &dataJSON, &e.CreatedAt); err != nil {
			continue
		}
		_ = json.Unmarshal([]byte(dataJSON), &e.Data)
		events = append(events, e)
	}
	return events
}
