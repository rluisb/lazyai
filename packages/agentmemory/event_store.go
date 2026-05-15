package agentmemory

import (
	"context"
	"database/sql"
)

// EventStore persists append-only task events.
type EventStore struct{ db *sql.DB }

func NewEventStore(db *sql.DB) *EventStore { return &EventStore{db: db} }

func (s *EventStore) AppendEvent(ctx context.Context, event TaskEvent) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO task_events (task_id, namespace, run_id, event_type, payload_json, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		event.TaskID, event.Namespace, event.RunID, event.EventType, Redact(event.PayloadJSON), event.CreatedAt)
	return err
}

func (s *EventStore) ListEvents(ctx context.Context, taskID string, sinceID int, limit int) ([]TaskEvent, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, task_id, namespace, COALESCE(run_id, ''), event_type, COALESCE(payload_json, ''), created_at
		FROM task_events WHERE task_id = ? AND id > ? ORDER BY id ASC LIMIT ?`, taskID, sinceID, normalizeLimit(limit))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEvents(rows)
}

func (s *EventStore) RecentEvents(ctx context.Context, namespace string, limit int) ([]TaskEvent, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, task_id, namespace, COALESCE(run_id, ''), event_type, COALESCE(payload_json, ''), created_at
		FROM task_events WHERE namespace = ? ORDER BY id DESC LIMIT ?`, namespace, normalizeLimit(limit))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEvents(rows)
}

func scanEvents(rows *sql.Rows) ([]TaskEvent, error) {
	events := []TaskEvent{}
	for rows.Next() {
		var event TaskEvent
		if err := rows.Scan(&event.ID, &event.TaskID, &event.Namespace, &event.RunID, &event.EventType, &event.PayloadJSON, &event.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}
