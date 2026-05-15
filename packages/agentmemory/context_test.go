package agentmemory

import "testing"

func TestBuildContextAssemblesBundle(t *testing.T) {
	db := testDB(t)
	now := testNow()
	if err := NewTaskStore(db).CreateTask(t.Context(), Task{ID: "task-1", Namespace: "ns", ProjectRoot: "/repo", TaskType: TaskTypeImplement, State: TaskStateRunning, Goal: "deterministic continuity", Tags: "memory", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}
	if err := NewCheckpointStore(db).CreateCheckpoint(t.Context(), Checkpoint{TaskID: "task-1", Namespace: "ns", StepID: "step-1", Summary: "checkpoint", CreatedAt: now}); err != nil {
		t.Fatalf("CreateCheckpoint() error = %v", err)
	}
	if err := NewEventStore(db).AppendEvent(t.Context(), TaskEvent{TaskID: "task-1", Namespace: "ns", EventType: EventTypeStepCompleted, PayloadJSON: `{}`, CreatedAt: now}); err != nil {
		t.Fatalf("AppendEvent() error = %v", err)
	}
	if _, err := NewMemoryStore(db).SaveMemory(t.Context(), Memory{Namespace: "ns", Content: "deterministic continuity lesson", SourceTaskID: "task-1", Tags: "memory", Importance: ImportanceHigh, CreatedAt: now}); err != nil {
		t.Fatalf("SaveMemory() error = %v", err)
	}
	if err := NewArtifactStore(db).RecordArtifact(t.Context(), Artifact{TaskID: "task-1", Namespace: "ns", Path: "plan.md", ContentPreview: "plan", CreatedAt: now}); err != nil {
		t.Fatalf("RecordArtifact() error = %v", err)
	}

	bundle, err := BuildContext(t.Context(), db, "task-1")
	if err != nil {
		t.Fatalf("BuildContext() error = %v", err)
	}
	if bundle.Task.ID != "task-1" || bundle.LatestCheckpoint == nil || len(bundle.RecentEvents) != 1 || len(bundle.RecentMemories) == 0 || len(bundle.Artifacts) != 1 {
		t.Fatalf("BuildContext() = %+v", bundle)
	}
}

func TestBuildContextEmptyTaskError(t *testing.T) {
	db := testDB(t)
	if _, err := BuildContext(t.Context(), db, "missing"); err == nil {
		t.Fatal("BuildContext(missing) error = nil, want error")
	}
}

func TestBuildContextNoCheckpointOrEvents(t *testing.T) {
	db := testDB(t)
	now := testNow()
	if err := NewTaskStore(db).CreateTask(t.Context(), Task{ID: "task-1", Namespace: "ns", ProjectRoot: "/repo", TaskType: TaskTypeResearch, State: TaskStatePending, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}
	bundle, err := BuildContext(t.Context(), db, "task-1")
	if err != nil {
		t.Fatalf("BuildContext() error = %v", err)
	}
	if bundle.LatestCheckpoint != nil {
		t.Fatalf("LatestCheckpoint = %+v, want nil", bundle.LatestCheckpoint)
	}
	if len(bundle.RecentEvents) != 0 {
		t.Fatalf("RecentEvents len = %d, want 0", len(bundle.RecentEvents))
	}
}
