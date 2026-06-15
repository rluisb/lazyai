package sqlite

import (
	"testing"

	"github.com/rluisb/lazyai/packages/orchestrator/internal/db"
)

func TestRunEventStorePersistsAndReplaysEvents(t *testing.T) {
	store := newRunEventStoreTestAdapter(t)

	if err := store.AppendRunEvent("chain-a", "chain.started", map[string]any{"step": "research"}, "2026-05-13T10:00:00Z"); err != nil {
		t.Fatalf("append first event: %v", err)
	}
	if err := store.AppendRunEvent("chain-b", "chain.started", map[string]any{"step": "plan"}, "2026-05-13T10:01:00Z"); err != nil {
		t.Fatalf("append second event: %v", err)
	}
	if err := store.AppendRunEvent("chain-a", "chain.advanced", map[string]any{"step": "implement"}, "2026-05-13T10:02:00Z"); err != nil {
		t.Fatalf("append third event: %v", err)
	}

	chainEvents, err := store.ReplayRunEvents("chain-a", 0)
	if err != nil {
		t.Fatalf("replay chain events: %v", err)
	}
	if len(chainEvents) != 2 {
		t.Fatalf("expected 2 chain-a events, got %d", len(chainEvents))
	}
	if chainEvents[0].Type != "chain.started" || chainEvents[1].Type != "chain.advanced" {
		t.Fatalf("unexpected chain-a event order: %+v", chainEvents)
	}
	if chainEvents[0].Data["step"] != "research" {
		t.Fatalf("unexpected decoded event data: %+v", chainEvents[0].Data)
	}

	allEvents, err := store.ReplayAllRunEvents(chainEvents[0].ID, 2)
	if err != nil {
		t.Fatalf("replay all events: %v", err)
	}
	if len(allEvents) != 2 {
		t.Fatalf("expected capped replay of 2 events, got %d", len(allEvents))
	}
	if allEvents[0].ID <= chainEvents[0].ID || allEvents[1].ID <= allEvents[0].ID {
		t.Fatalf("events were not replayed in ascending id order: %+v", allEvents)
	}
}

func newRunEventStoreTestAdapter(t *testing.T) *RunEventStore {
	t.Helper()
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := database.RunMigrations(); err != nil {
		t.Fatalf("migrate db: %v", err)
	}
	return NewRunEventStore(database)
}
