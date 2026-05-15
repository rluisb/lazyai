package agentmemory

import "testing"

func TestMemoryStoreSaveAndList(t *testing.T) {
	db := testDB(t)
	store := NewMemoryStore(db)
	now := testNow()

	id, err := store.SaveMemory(t.Context(), Memory{Namespace: "ns", Content: "remember sk-abcdefghijklmnopqrstuvwxyz1234567890", SourceTaskID: "task-1", SourceStepID: "step-1", Tags: "handoff,api", Importance: ImportanceHigh, CreatedAt: now})
	if err != nil {
		t.Fatalf("SaveMemory() error = %v", err)
	}
	if id <= 0 {
		t.Fatalf("SaveMemory() id = %d, want > 0", id)
	}
	if _, err := store.SaveMemory(t.Context(), Memory{Namespace: "other", Content: "other", Tags: "api", Importance: ImportanceNormal, CreatedAt: now}); err != nil {
		t.Fatalf("SaveMemory(other) error = %v", err)
	}

	listed, err := store.ListMemories(t.Context(), "ns", "api", 10)
	if err != nil {
		t.Fatalf("ListMemories() error = %v", err)
	}
	if len(listed) != 1 || listed[0].ID != id {
		t.Fatalf("ListMemories() = %+v, want saved memory", listed)
	}
	if listed[0].Content != "remember sk-REDACTED" {
		t.Fatalf("content not redacted: %q", listed[0].Content)
	}
}
