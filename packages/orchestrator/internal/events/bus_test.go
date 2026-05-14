package events

import (
	"testing"
	"time"

	"github.com/rluisb/lazyai/packages/orchestrator/adapters/sqlite"
	"github.com/rluisb/lazyai/packages/orchestrator/domain"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/db"
)

func newBusTestStore(t *testing.T) *sqlite.RunEventStore {
	t.Helper()
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := database.RunMigrations(); err != nil {
		t.Fatalf("migrate db: %v", err)
	}
	return sqlite.NewRunEventStore(database)
}

func TestBusPublishesThroughRunEventStorePort(t *testing.T) {
	store := &fakeRunEventStore{}
	bus := NewBus(store)

	bus.Publish("chain-1", "chain.started", map[string]any{"step": "research"})

	if len(store.events) != 1 {
		t.Fatalf("expected one persisted event, got %d", len(store.events))
	}
	if store.events[0].RunID != "chain-1" || store.events[0].Type != "chain.started" {
		t.Fatalf("unexpected stored event: %+v", store.events[0])
	}
	if store.events[0].CreatedAt == "" {
		t.Fatalf("expected stored event timestamp")
	}

	replayed := bus.Replay("chain-1", 0)
	if len(replayed) != 1 || replayed[0].Data["step"] != "research" {
		t.Fatalf("unexpected replay through port: %+v", replayed)
	}
}

func TestBusSubscribeAllReceivesEventsFromAllRuns(t *testing.T) {
	bus := NewBus(newBusTestStore(t))

	ch := make(chan Event, 8)
	replay := bus.SubscribeAll(ch)
	if len(replay) != 0 {
		t.Fatalf("expected empty replay on fresh db, got %d", len(replay))
	}
	defer bus.UnsubscribeAll(ch)

	bus.Publish("chain-1", "run_started", map[string]any{"k": "v"})
	bus.Publish("team-9", "task_completed", map[string]any{"k": "v"})

	got := drainBus(t, ch, 2, time.Second)
	if got[0].RunID != "chain-1" || got[0].Type != "run_started" {
		t.Fatalf("first event mismatch: %+v", got[0])
	}
	if got[1].RunID != "team-9" || got[1].Type != "task_completed" {
		t.Fatalf("second event mismatch: %+v", got[1])
	}
}

func TestBusUnsubscribeAllStopsDeliveryWithoutLeakingPublishers(t *testing.T) {
	bus := NewBus(newBusTestStore(t))
	ch := make(chan Event, 4)
	bus.SubscribeAll(ch)
	bus.UnsubscribeAll(ch)

	// After unsubscribe, publishers must not block even if the channel is full.
	for i := 0; i < 16; i++ {
		bus.Publish("chain-1", "tick", map[string]any{})
	}
	if len(ch) != 0 {
		t.Fatalf("expected no events delivered after unsubscribe, got %d", len(ch))
	}
}

func TestBusPublishDropsForSlowGlobalSubscriberWithoutBlocking(t *testing.T) {
	bus := NewBus(newBusTestStore(t))
	slow := make(chan Event) // unbuffered, never read — every send must drop
	bus.SubscribeAll(slow)
	defer bus.UnsubscribeAll(slow)

	done := make(chan struct{})
	go func() {
		for i := 0; i < 256; i++ {
			bus.Publish("chain-1", "tick", map[string]any{})
		}
		close(done)
	}()

	select {
	case <-done:
		// publishers finished; that's the contract.
	case <-time.After(2 * time.Second):
		t.Fatal("Publish blocked behind a slow global subscriber")
	}
}

func TestBusReplayAllOrdersByIDAndCapsLimit(t *testing.T) {
	bus := NewBus(newBusTestStore(t))
	for i := 0; i < 5; i++ {
		bus.Publish("chain-a", "step_completed", map[string]any{"i": i})
		bus.Publish("team-b", "task_completed", map[string]any{"i": i})
	}

	all := bus.ReplayAll(0, 1000)
	if len(all) != 10 {
		t.Fatalf("expected 10 events, got %d", len(all))
	}
	for i := 1; i < len(all); i++ {
		if all[i].ID <= all[i-1].ID {
			t.Fatalf("events not in id ASC order at i=%d: %d after %d", i, all[i].ID, all[i-1].ID)
		}
	}

	limited := bus.ReplayAll(0, 3)
	if len(limited) != 3 {
		t.Fatalf("limit=3 returned %d events", len(limited))
	}

	since := bus.ReplayAll(all[5].ID, 1000)
	if len(since) != len(all)-6 {
		t.Fatalf("since_id replay returned %d, want %d", len(since), len(all)-6)
	}
	if since[0].ID != all[6].ID {
		t.Fatalf("since_id replay first id = %d, want %d", since[0].ID, all[6].ID)
	}
}

func TestBusPerRunSubscribeStillReceivesOnlyMatchingRun(t *testing.T) {
	bus := NewBus(newBusTestStore(t))
	chRun := make(chan Event, 4)
	chAll := make(chan Event, 4)
	bus.Subscribe("chain-x", chRun)
	bus.SubscribeAll(chAll)
	defer bus.Unsubscribe("chain-x", chRun)
	defer bus.UnsubscribeAll(chAll)

	bus.Publish("chain-x", "step_started", map[string]any{})
	bus.Publish("chain-y", "step_started", map[string]any{})

	gotRun := drainBus(t, chRun, 1, time.Second)
	if gotRun[0].RunID != "chain-x" {
		t.Fatalf("per-run subscriber received wrong run: %+v", gotRun[0])
	}
	gotAll := drainBus(t, chAll, 2, time.Second)
	if gotAll[0].RunID != "chain-x" || gotAll[1].RunID != "chain-y" {
		t.Fatalf("global subscriber missed events: %+v", gotAll)
	}
}

func drainBus(t *testing.T, ch chan Event, n int, timeout time.Duration) []Event {
	t.Helper()
	deadline := time.After(timeout)
	out := make([]Event, 0, n)
	for len(out) < n {
		select {
		case e := <-ch:
			out = append(out, e)
		case <-deadline:
			t.Fatalf("timeout waiting for %d events; got %d", n, len(out))
		}
	}
	return out
}

type fakeRunEventStore struct {
	events []domain.RunEvent
}

func (s *fakeRunEventStore) AppendRunEvent(runID, eventType string, data map[string]any, createdAt string) error {
	s.events = append(s.events, domain.RunEvent{
		ID:        len(s.events) + 1,
		RunID:     runID,
		Type:      eventType,
		Data:      data,
		CreatedAt: createdAt,
	})
	return nil
}

func (s *fakeRunEventStore) ReplayRunEvents(runID string, sinceID int) ([]domain.RunEvent, error) {
	var replay []domain.RunEvent
	for _, event := range s.events {
		if event.RunID == runID && event.ID > sinceID {
			replay = append(replay, event)
		}
	}
	return replay, nil
}

func (s *fakeRunEventStore) ReplayAllRunEvents(sinceID, limit int) ([]domain.RunEvent, error) {
	if limit <= 0 {
		return nil, nil
	}
	var replay []domain.RunEvent
	for _, event := range s.events {
		if event.ID <= sinceID {
			continue
		}
		replay = append(replay, event)
		if len(replay) == limit {
			break
		}
	}
	return replay, nil
}
