package cmd

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestFormatTaskID(t *testing.T) {
	cases := []struct {
		id       int
		expected string
	}{
		{1, "task_1"},
		{42, "task_42"},
		{0, "task_0"},
		{-1, "task_-1"},
	}
	for _, c := range cases {
		got := formatTaskID(c.id)
		if got != c.expected {
			t.Errorf("formatTaskID(%d) = %q, want %q", c.id, got, c.expected)
		}
	}
}

func TestParseTaskID(t *testing.T) {
	cases := []struct {
		input      string
		wantID     int
		wantErr    bool
		wantErrMsg string
	}{
		{"task_1", 1, false, ""},
		{"task_42", 42, false, ""},
		{"task_0", 0, false, ""},
		{"task_-1", -1, false, ""},
		{"task", 0, true, "invalid task ID format: task"},
		{"task_", 0, true, "invalid task ID format: task_"},
		{"task_abc", 0, true, "invalid task ID format: task_abc"},
		{"task_1_2", 0, true, "invalid task ID format: task_1_2"},
		{"", 0, true, "invalid task ID format: "},
	}
	for _, c := range cases {
		got, err := parseTaskID(c.input)
		if c.wantErr {
			if err == nil {
				t.Errorf("parseTaskID(%q) expected error, got nil", c.input)
				continue
			}
			if err.Error() != c.wantErrMsg {
				t.Errorf("parseTaskID(%q) error = %q, want %q", c.input, err.Error(), c.wantErrMsg)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseTaskID(%q) unexpected error: %v", c.input, err)
			continue
		}
		if got != c.wantID {
			t.Errorf("parseTaskID(%q) = %d, want %d", c.input, got, c.wantID)
		}
	}
}

func TestFormatParseRoundTrip(t *testing.T) {
	for _, id := range []int{1, 2, 99, 12345} {
		formatted := formatTaskID(id)
		parsed, err := parseTaskID(formatted)
		if err != nil {
			t.Fatalf("round-trip for %d: parse error: %v", id, err)
		}
		if parsed != id {
			t.Errorf("round-trip for %d: got %d", id, parsed)
		}
	}
}

// ── Integration tests for task CLI commands ──────────────────────────────────

func TestTaskCreate(t *testing.T) {
	withTempDir(t)

	out := captureStdout(t, func() {
		if err := taskCreateCmd.RunE(taskCreateCmd, []string{"integration-test-task"}); err != nil {
			t.Fatalf("task create failed: %v", err)
		}
	})

	if !strings.Contains(out, "Task created:") {
		t.Errorf("expected output to contain 'Task created:', got:\n%s", out)
	}

	// Verify DB state
	db, err := openRuntimeDB()
	if err != nil {
		t.Fatalf("openRuntimeDB failed: %v", err)
	}
	defer db.Close()

	var status, taskDesc string
	if err := db.QueryRow("SELECT status, task FROM task_queue WHERE task = ?", "integration-test-task").Scan(&status, &taskDesc); err != nil {
		t.Fatalf("failed to query task: %v", err)
	}
	if status != "open" {
		t.Errorf("expected status='open', got %q", status)
	}
	if taskDesc != "integration-test-task" {
		t.Errorf("expected task=%q, got %q", "integration-test-task", taskDesc)
	}
}

func TestTaskList(t *testing.T) {
	withTempDir(t)

	// Seed DB directly
	db, err := openRuntimeDB()
	if err != nil {
		t.Fatalf("openRuntimeDB failed: %v", err)
	}
	if err := ensureSession(db, "cli"); err != nil {
		t.Fatalf("ensureSession failed: %v", err)
	}
	createdAt := time.Now().UTC().Format(time.RFC3339)
	if _, err := db.Exec(
		"INSERT INTO task_queue (session_id, topic, task, status, max_agents, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		"cli", "cli", "list-test-task", "open", 1, createdAt,
	); err != nil {
		t.Fatalf("failed to insert test task: %v", err)
	}
	db.Close()

	out := captureStdout(t, func() {
		if err := taskListCmd.RunE(taskListCmd, []string{}); err != nil {
			t.Fatalf("task list failed: %v", err)
		}
	})

	if !strings.Contains(out, "list-test-task") {
		t.Errorf("expected list output to contain task description, got:\n%s", out)
	}
}

func TestTaskClaimByID(t *testing.T) {
	withTempDir(t)
	t.Setenv("LAZYAI_AGENT", "test-agent")

	// Seed DB directly
	db, err := openRuntimeDB()
	if err != nil {
		t.Fatalf("openRuntimeDB failed: %v", err)
	}
	if err := ensureSession(db, "cli"); err != nil {
		t.Fatalf("ensureSession failed: %v", err)
	}
	createdAt := time.Now().UTC().Format(time.RFC3339)
	result, err := db.Exec(
		"INSERT INTO task_queue (session_id, topic, task, status, max_agents, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		"cli", "cli", "claim-test-task", "open", 1, createdAt,
	)
	if err != nil {
		t.Fatalf("failed to insert test task: %v", err)
	}
	taskID, _ := result.LastInsertId()
	db.Close()

	taskIDStr := fmt.Sprintf("task_%d", taskID)

	out := captureStdout(t, func() {
		if err := taskClaimCmd.RunE(taskClaimCmd, []string{taskIDStr}); err != nil {
			t.Fatalf("task claim failed: %v", err)
		}
	})

	if !strings.Contains(out, "Task claimed:") {
		t.Errorf("expected output to contain 'Task claimed:', got:\n%s", out)
	}

	// Verify DB state
	db, err = openRuntimeDB()
	if err != nil {
		t.Fatalf("openRuntimeDB failed: %v", err)
	}
	defer db.Close()

	var status string
	if err := db.QueryRow("SELECT status FROM task_queue WHERE id = ?", taskID).Scan(&status); err != nil {
		t.Fatalf("failed to query task status: %v", err)
	}
	if status != "claimed" {
		t.Errorf("expected status='claimed', got %q", status)
	}

	var claimCount int
	if err := db.QueryRow("SELECT count(*) FROM task_claims WHERE task_id = ? AND agent = ?", taskID, "test-agent").Scan(&claimCount); err != nil {
		t.Fatalf("failed to query claims: %v", err)
	}
	if claimCount != 1 {
		t.Errorf("expected 1 claim row, got %d", claimCount)
	}
}

func TestTaskClaimNext(t *testing.T) {
	withTempDir(t)
	t.Setenv("LAZYAI_AGENT", "test-agent")

	// Seed DB directly
	db, err := openRuntimeDB()
	if err != nil {
		t.Fatalf("openRuntimeDB failed: %v", err)
	}
	if err := ensureSession(db, "cli"); err != nil {
		t.Fatalf("ensureSession failed: %v", err)
	}
	createdAt := time.Now().UTC().Format(time.RFC3339)
	result, err := db.Exec(
		"INSERT INTO task_queue (session_id, topic, task, status, max_agents, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		"cli", "cli", "claim-next-test-task", "open", 1, createdAt,
	)
	if err != nil {
		t.Fatalf("failed to insert test task: %v", err)
	}
	taskID, _ := result.LastInsertId()
	db.Close()

	out := captureStdout(t, func() {
		if err := taskClaimCmd.RunE(taskClaimCmd, []string{}); err != nil {
			t.Fatalf("task claim failed: %v", err)
		}
	})

	if !strings.Contains(out, "Task claimed:") {
		t.Errorf("expected output to contain 'Task claimed:', got:\n%s", out)
	}

	// Verify DB state
	db, err = openRuntimeDB()
	if err != nil {
		t.Fatalf("openRuntimeDB failed: %v", err)
	}
	defer db.Close()

	var status string
	if err := db.QueryRow("SELECT status FROM task_queue WHERE id = ?", taskID).Scan(&status); err != nil {
		t.Fatalf("failed to query task status: %v", err)
	}
	if status != "claimed" {
		t.Errorf("expected status='claimed', got %q", status)
	}
}

func TestTaskComplete(t *testing.T) {
	withTempDir(t)
	t.Setenv("LAZYAI_AGENT", "test-agent")

	// Seed DB directly with a claimed task
	db, err := openRuntimeDB()
	if err != nil {
		t.Fatalf("openRuntimeDB failed: %v", err)
	}
	if err := ensureSession(db, "cli"); err != nil {
		t.Fatalf("ensureSession failed: %v", err)
	}
	createdAt := time.Now().UTC().Format(time.RFC3339)
	result, err := db.Exec(
		"INSERT INTO task_queue (session_id, topic, task, status, max_agents, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		"cli", "cli", "complete-test-task", "claimed", 1, createdAt,
	)
	if err != nil {
		t.Fatalf("failed to insert test task: %v", err)
	}
	taskID, _ := result.LastInsertId()

	if _, err := db.Exec(
		"INSERT INTO task_claims (task_id, agent, claimed_at) VALUES (?, ?, ?)",
		taskID, "test-agent", createdAt,
	); err != nil {
		t.Fatalf("failed to insert claim: %v", err)
	}
	db.Close()

	taskIDStr := fmt.Sprintf("task_%d", taskID)

	out := captureStdout(t, func() {
		if err := taskCompleteCmd.RunE(taskCompleteCmd, []string{taskIDStr}); err != nil {
			t.Fatalf("task complete failed: %v", err)
		}
	})

	if !strings.Contains(out, "Task completed:") {
		t.Errorf("expected output to contain 'Task completed:', got:\n%s", out)
	}

	// Verify DB state
	db, err = openRuntimeDB()
	if err != nil {
		t.Fatalf("openRuntimeDB failed: %v", err)
	}
	defer db.Close()

	var status string
	if err := db.QueryRow("SELECT status FROM task_queue WHERE id = ?", taskID).Scan(&status); err != nil {
		t.Fatalf("failed to query task status: %v", err)
	}
	if status != "completed" {
		t.Errorf("expected status='completed', got %q", status)
	}

	var claimCount int
	if err := db.QueryRow("SELECT count(*) FROM task_claims WHERE task_id = ?", taskID).Scan(&claimCount); err != nil {
		t.Fatalf("failed to query claims: %v", err)
	}
	if claimCount != 0 {
		t.Errorf("expected 0 claim rows after complete, got %d", claimCount)
	}
}
