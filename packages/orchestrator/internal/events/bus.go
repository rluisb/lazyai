package events

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/rluisb/lazyai/packages/orchestrator/internal/db"
)

// Event represents a run lifecycle event.
type Event struct {
	ID        int            `json:"id"`
	RunID     string         `json:"runId"`
	Type      string         `json:"eventType"`
	Data      map[string]any `json:"data"`
	CreatedAt string         `json:"createdAt"`
}

// Bus provides channel-based pub/sub for run events.
type Bus struct {
	mu          sync.RWMutex
	subscribers map[string][]chan Event
	database    *db.DB
}

// NewBus creates a new event bus.
func NewBus(database *db.DB) *Bus {
	return &Bus{
		subscribers: make(map[string][]chan Event),
		database:    database,
	}
}

// Publish emits an event to all subscribers of a run and persists it.
func (b *Bus) Publish(runID, eventType string, data map[string]any) {
	now := time.Now().UTC().Format(time.RFC3339)
	dataJSON, _ := json.Marshal(data)

	b.database.Exec(
		`INSERT INTO run_events (run_id, event_type, event_json, created_at) VALUES (?, ?, ?, ?)`,
		runID, eventType, string(dataJSON), now)

	event := Event{
		RunID:     runID,
		Type:      eventType,
		Data:      data,
		CreatedAt: now,
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, ch := range b.subscribers[runID] {
		select {
		case ch <- event:
		default:
		}
	}
}

// Subscribe registers a subscriber channel for a run.
// Returns past events from DB.
func (b *Bus) Subscribe(runID string, ch chan Event) []Event {
	b.mu.Lock()
	b.subscribers[runID] = append(b.subscribers[runID], ch)
	b.mu.Unlock()

	// Replay past events
	return b.Replay(runID, 0)
}

// Unsubscribe removes a subscriber channel.
func (b *Bus) Unsubscribe(runID string, ch chan Event) {
	b.mu.Lock()
	defer b.mu.Unlock()

	subs := b.subscribers[runID]
	for i, s := range subs {
		if s == ch {
			b.subscribers[runID] = append(subs[:i], subs[i+1:]...)
			return
		}
	}
}

// RemoveAll removes all subscribers for a run.
func (b *Bus) RemoveAll(runID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.subscribers, runID)
}

// Replay returns events from the database since a given event ID.
func (b *Bus) Replay(runID string, sinceID int) []Event {
	var query string
	var args []any
	if sinceID > 0 {
		query = `SELECT id, run_id, event_type, event_json, created_at FROM run_events WHERE run_id = ? AND id > ? ORDER BY id ASC`
		args = []any{runID, sinceID}
	} else {
		query = `SELECT id, run_id, event_type, event_json, created_at FROM run_events WHERE run_id = ? ORDER BY id ASC`
		args = []any{runID}
	}

	rows, err := b.database.Query(query, args...)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		var dataJSON string
		if err := rows.Scan(&e.ID, &e.RunID, &e.Type, &dataJSON, &e.CreatedAt); err != nil {
			continue
		}
		json.Unmarshal([]byte(dataJSON), &e.Data)
		events = append(events, e)
	}
	return events
}
