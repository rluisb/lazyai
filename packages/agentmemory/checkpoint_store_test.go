package agentmemory

import "testing"

func TestCheckpointStoreCreateLatestAndList(t *testing.T) {
	db := testDB(t)
	store := NewCheckpointStore(db)
	now := testNow()

	checkpoints := []Checkpoint{
		{TaskID: "task-1", Namespace: "ns", StepID: "research", Summary: "first", StateJSON: `{"secret":"sk-abcdefghijklmnopqrstuvwxyz1234567890"}`, CreatedAt: now},
		{TaskID: "task-1", Namespace: "ns", StepID: "plan", Summary: "second", StateJSON: `{}`, CreatedAt: now},
	}
	for _, cp := range checkpoints {
		if err := store.CreateCheckpoint(t.Context(), cp); err != nil {
			t.Fatalf("CreateCheckpoint() error = %v", err)
		}
	}

	latest, err := store.LatestCheckpoint(t.Context(), "task-1")
	if err != nil {
		t.Fatalf("LatestCheckpoint() error = %v", err)
	}
	if latest.StepID != "plan" {
		t.Fatalf("LatestCheckpoint().StepID = %q, want plan", latest.StepID)
	}

	listed, err := store.ListCheckpoints(t.Context(), "task-1")
	if err != nil {
		t.Fatalf("ListCheckpoints() error = %v", err)
	}
	if len(listed) != 2 || listed[0].StateJSON != `{"secret":"sk-REDACTED"}` {
		t.Fatalf("ListCheckpoints() = %+v", listed)
	}
}
