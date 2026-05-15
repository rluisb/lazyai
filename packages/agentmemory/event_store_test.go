package agentmemory

import "testing"

func TestEventStoreAppendListAndRecent(t *testing.T) {
	db := testDB(t)
	store := NewEventStore(db)
	now := testNow()

	events := []TaskEvent{
		{TaskID: "task-1", Namespace: "ns", RunID: "run-1", EventType: EventTypeTaskStarted, PayloadJSON: `{"auth":"Bearer abc.def-123"}`, CreatedAt: now},
		{TaskID: "task-1", Namespace: "ns", RunID: "run-1", EventType: EventTypeStepCompleted, PayloadJSON: `{"ok":true}`, CreatedAt: now},
		{TaskID: "task-2", Namespace: "other", EventType: EventTypeTaskStarted, PayloadJSON: `{}`, CreatedAt: now},
	}
	for _, event := range events {
		if err := store.AppendEvent(t.Context(), event); err != nil {
			t.Fatalf("AppendEvent() error = %v", err)
		}
	}

	listed, err := store.ListEvents(t.Context(), "task-1", 0, 10)
	if err != nil {
		t.Fatalf("ListEvents() error = %v", err)
	}
	if len(listed) != 2 {
		t.Fatalf("ListEvents() len = %d, want 2", len(listed))
	}
	if listed[0].PayloadJSON != `{"auth":"Bearer REDACTED"}` {
		t.Fatalf("payload not redacted: %s", listed[0].PayloadJSON)
	}

	recent, err := store.RecentEvents(t.Context(), "ns", 1)
	if err != nil {
		t.Fatalf("RecentEvents() error = %v", err)
	}
	if len(recent) != 1 || recent[0].TaskID != "task-1" {
		t.Fatalf("RecentEvents() = %+v", recent)
	}
}
