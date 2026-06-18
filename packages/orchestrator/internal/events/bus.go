package events

import (
	"sync"
	"time"

	"github.com/rluisb/lazyai/packages/orchestrator/domain"
	"github.com/rluisb/lazyai/packages/orchestrator/ports"
)

// Event represents a run lifecycle event.
type Event = domain.RunEvent

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
	store       ports.RunEventStore
}

// NewBus creates a new event bus.
func NewBus(store ports.RunEventStore) *Bus {
	return &Bus{
		subscribers: make(map[string][]chan Event),
		store:       store,
	}
}

// Publish emits an event to per-run and global subscribers, and persists it.
func (b *Bus) Publish(runID, eventType string, data map[string]any) {
	now := time.Now().UTC().Format(time.RFC3339)
	if b.store != nil {
		_ = b.store.AppendRunEvent(runID, eventType, data, now)
	}

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

// Subscribe registers a per-run subscriber channel and returns past events from the store.
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
	if b.store == nil {
		return nil
	}
	events, err := b.store.ReplayRunEvents(runID, sinceID)
	if err != nil {
		return nil
	}
	return events
}

// ReplayAll returns events across all runs since a given event ID, ordered by id ASC,
// capped at the provided limit. limit <= 0 returns no events.
func (b *Bus) ReplayAll(sinceID, limit int) []Event {
	if b.store == nil || limit <= 0 {
		return nil
	}
	events, err := b.store.ReplayAllRunEvents(sinceID, limit)
	if err != nil {
		return nil
	}
	return events
}
