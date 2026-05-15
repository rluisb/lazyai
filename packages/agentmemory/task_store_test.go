package agentmemory

import "testing"

func TestTaskStoreCreateGetUpdateAndList(t *testing.T) {
	db := testDB(t)
	store := NewTaskStore(db)
	now := testNow()

	first := Task{ID: "task-1", Namespace: "ns", ProjectRoot: "/repo", TaskType: TaskTypeImplement, State: TaskStatePending, StateJSON: `{"api_key":"sk-abcdefghijklmnopqrstuvwxyz1234567890"}`, Goal: "ship memory", Tags: "agentmemory", CreatedAt: now, UpdatedAt: now}
	if err := store.CreateTask(t.Context(), first); err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}
	if err := store.CreateTask(t.Context(), Task{ID: "task-2", Namespace: "other", ProjectRoot: "/repo", TaskType: TaskTypePlan, State: TaskStatePending, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatalf("CreateTask(other) error = %v", err)
	}

	got, err := store.GetTask(t.Context(), "task-1")
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}
	if got.StateJSON != `{"api_key":"sk-REDACTED"}` {
		t.Fatalf("StateJSON was not redacted: %s", got.StateJSON)
	}

	if err := store.UpdateTaskState(t.Context(), "task-1", TaskStateRunning, "write-tests", `{"token":"Bearer abc.def-123"}`); err != nil {
		t.Fatalf("UpdateTaskState() error = %v", err)
	}
	got, err = store.GetTask(t.Context(), "task-1")
	if err != nil {
		t.Fatalf("GetTask(updated) error = %v", err)
	}
	if got.State != TaskStateRunning || got.CurrentStep != "write-tests" || got.StateJSON != `{"token":"Bearer REDACTED"}` {
		t.Fatalf("updated task mismatch: %+v", got)
	}

	listed, err := store.ListTasks(t.Context(), "ns")
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}
	if len(listed) != 1 || listed[0].ID != "task-1" {
		t.Fatalf("ListTasks() = %+v, want only task-1", listed)
	}
}
