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
//
// It supports two subscription modes:
//   - Per-run subscribers (Subscribe/Unsubscribe), which only see events for one run.
//   - Global subscribers (SubscribeAll/UnsubscribeAll), which see every event published.
//
// Both modes use non-blocking sends; slow consumers drop events rather than block publishers.
type Bus struct {
	mu          sync.RWMutex
	subscribers map[string][]chan Event
	globalSubs  []chan Event
	database    *db.DB
}

// NewBus creates a new event bus.
func NewBus(database *db.DB) *Bus {
	return &Bus{
		subscribers: make(map[string][]chan Event),
		database:    database,
	}
}

// Publish emits an event to per-run and global subscribers, and persists it.
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
	for _, ch := range b.globalSubs {
		select {
		case ch <- event:
		default:
		}
	}
}

// Subscribe registers a per-run subscriber channel and returns past events from the DB.
func (b *Bus) Subscribe(runID string, ch chan Event) []Event {
	b.mu.Lock()
	b.subscribers[runID] = append(b.subscribers[runID], ch)
	b.mu.Unlock()

	return b.Replay(runID, 0)
}

// Unsubscribe removes a per-run subscriber channel.
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

// RemoveAll removes all per-run subscribers for a run.
func (b *Bus) RemoveAll(runID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.subscribers, runID)
}

// SubscribeAll registers a global subscriber that receives every event published
// across all runs. Returns no replay; callers that need history should call ReplayAll.
func (b *Bus) SubscribeAll(ch chan Event) []Event {
	b.mu.Lock()
	b.globalSubs = append(b.globalSubs, ch)
	b.mu.Unlock()
	return nil
}

// UnsubscribeAll removes a global subscriber channel.
func (b *Bus) UnsubscribeAll(ch chan Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for i, s := range b.globalSubs {
		if s == ch {
			b.globalSubs = append(b.globalSubs[:i], b.globalSubs[i+1:]...)
			return
		}
	}
}

// Replay returns events for a single run since a given event ID, ordered by id ASC.
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

	return scanEvents(rows)
}

// ReplayAll returns events across all runs since a given event ID, ordered by id ASC,
// capped at the provided limit. limit <= 0 returns no events.
func (b *Bus) ReplayAll(sinceID, limit int) []Event {
	if limit <= 0 {
		return nil
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

	rows, err := b.database.Query(query, args...)
	if err != nil {
		return nil
	}
	defer rows.Close()

	return scanEvents(rows)
}

// scanEvents reads rows from a `SELECT id, run_id, event_type, event_json, created_at` query.
func scanEvents(rows interface {
	Next() bool
	Scan(dest ...any) error
}) []Event {
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
