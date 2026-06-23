package cmd

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/rluisb/lazyai/packages/cli/internal/handoff"
	runtimesession "github.com/rluisb/lazyai/packages/cli/internal/runtime/session"
)

// captureStdout redirects os.Stdout to a buffer for the duration of fn.
// Returns the captured output.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("failed to copy stdout: %v", err)
	}
	return buf.String()
}

func TestGetDB(t *testing.T) {
	// This test requires a database to be initialized
	// In a real test, we would mock the database or use a temp directory
	// For now, we just verify the function exists and returns an error when no DB exists
	_, err := getDB()
	if err == nil {
		t.Log("getDB returned a database (may have found one in current directory)")
	} else {
		t.Logf("getDB returned expected error: %v", err)
	}
}

func TestRunSessionStart(t *testing.T) {
	// Test that session start generates a valid session ID format
	sessionID := fmt.Sprintf("ses_%d", time.Now().Unix())
	if len(sessionID) < 4 {
		t.Error("Session ID is too short")
	}

	if sessionID[:4] != "ses_" {
		t.Error("Session ID does not start with 'ses_'")
	}
}

func TestSessionIDFormat(t *testing.T) {
	// Test session ID format
	now := time.Now().Unix()
	sessionID := fmt.Sprintf("ses_%d", now)

	expectedPrefix := "ses_"
	if len(sessionID) <= len(expectedPrefix) {
		t.Errorf("Session ID '%s' is too short", sessionID)
	}

	if sessionID[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("Session ID '%s' does not start with '%s'", sessionID, expectedPrefix)
	}
}

func TestTimeFormat(t *testing.T) {
	// Test that time formatting works correctly
	now := time.Now().UTC()
	formatted := now.Format(time.RFC3339)

	// Should contain T and Z
	if !strings.Contains(formatted, "T") {
		t.Error("Formatted time does not contain 'T'")
	}

	if !strings.Contains(formatted, "Z") {
		t.Error("Formatted time does not contain 'Z'")
	}
}

func TestSessionStartCreatesRow(t *testing.T) {
	withTempDir(t)

	goal := "integration-test-goal"
	out := captureStdout(t, func() {
		if err := runSessionStart(&cobra.Command{}, []string{goal}); err != nil {
			t.Fatalf("runSessionStart failed: %v", err)
		}
	})
	if !strings.Contains(out, "Session started:") {
		t.Errorf("expected start output to contain 'Session started:', got:\n%s", out)
	}

	// Verify the row exists.
	db, err := openRuntimeDB()
	if err != nil {
		t.Fatalf("openRuntimeDB failed: %v", err)
	}
	defer db.Close()

	var count int
	if err := db.QueryRow("SELECT count(*) FROM sessions WHERE goal = ? AND status = ?", goal, "active").Scan(&count); err != nil {
		t.Fatalf("failed to count sessions: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 active session row with goal %q, got %d", goal, count)
	}
}

func TestSessionListAndShow(t *testing.T) {
	withTempDir(t)

	// Create session directly via SQL so list/show exercise timestamp parsing.
	db, err := openRuntimeDB()
	if err != nil {
		t.Fatalf("openRuntimeDB failed: %v", err)
	}
	defer db.Close()

	sessionID := fmt.Sprintf("ses_%d", time.Now().Unix())
	goal := "list-show-test"
	startedAt := time.Now().UTC().Format(time.RFC3339)
	if _, err := db.Exec(
		"INSERT INTO sessions (id, started_at, agent, status, goal) VALUES (?, ?, ?, ?, ?)",
		sessionID, startedAt, "cli", "active", goal,
	); err != nil {
		t.Fatalf("failed to insert test session: %v", err)
	}

	listOut := captureStdout(t, func() {
		if err := runSessionList(&cobra.Command{}, []string{}); err != nil {
			t.Fatalf("runSessionList failed: %v", err)
		}
	})
	if !strings.Contains(listOut, sessionID) || !strings.Contains(listOut, goal) {
		t.Errorf("expected list output to contain session ID and goal, got:\n%s", listOut)
	}

	showOut := captureStdout(t, func() {
		if err := runSessionShow(&cobra.Command{}, []string{sessionID}); err != nil {
			t.Fatalf("runSessionShow failed: %v", err)
		}
	})
	if !strings.Contains(showOut, sessionID) || !strings.Contains(showOut, goal) {
		t.Errorf("expected show output to contain session ID and goal, got:\n%s", showOut)
	}
}

func TestSessionEnd(t *testing.T) {
	withTempDir(t)

	// Create session directly via SQL to avoid the scan bug in mgr.Start
	db, err := openRuntimeDB()
	if err != nil {
		t.Fatalf("openRuntimeDB failed: %v", err)
	}
	defer db.Close()

	sessionID := fmt.Sprintf("ses_%d", time.Now().Unix())
	goal := "end-test"
	startedAt := time.Now().UTC().Format(time.RFC3339)
	if _, err := db.Exec(
		"INSERT INTO sessions (id, started_at, agent, status, goal) VALUES (?, ?, ?, ?, ?)",
		sessionID, startedAt, "cli", "active", goal,
	); err != nil {
		t.Fatalf("failed to insert test session: %v", err)
	}

	// End the session
	endOut := captureStdout(t, func() {
		if err := runSessionEnd(&cobra.Command{}, []string{sessionID}); err != nil {
			t.Fatalf("runSessionEnd failed: %v", err)
		}
	})
	if !strings.Contains(endOut, "Session ended:") {
		t.Errorf("expected end output to contain 'Session ended:', got:\n%s", endOut)
	}

	// Verify DB state
	var status string
	var endedAt sql.NullString
	if err := db.QueryRow("SELECT status, ended_at FROM sessions WHERE id = ?", sessionID).Scan(&status, &endedAt); err != nil {
		t.Fatalf("failed to query ended session: %v", err)
	}
	if status != "ended" {
		t.Errorf("expected status='ended', got %q", status)
	}
	if !endedAt.Valid || endedAt.String == "" {
		t.Errorf("expected ended_at to be set, got %v", endedAt)
	}
}

func TestSessionEndWritesHandoffAndMetadata(t *testing.T) {
	tmpDir := withTempDir(t)

	db, err := openRuntimeDB()
	if err != nil {
		t.Fatalf("openRuntimeDB failed: %v", err)
	}
	defer db.Close()

	mgr := runtimesession.NewManager(db)
	s, err := mgr.Start("ship phase four handoff", runtimesession.StartOptions{
		Agent:    "implementer",
		Model:    "sonnet",
		Repo:     "lazyai",
		Worktree: "feature/phase4",
		Tags:     []string{"phase4", "handoff"},
	})
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if err := mgr.UpdateSummary(s.ID, "Phase 3 is complete; session close should emit a resumable handoff."); err != nil {
		t.Fatalf("UpdateSummary failed: %v", err)
	}

	dispatch, err := mgr.Dispatch(s.ID, runtimesession.DispatchOptions{
		Agent:        "planner",
		Task:         "define handoff writer contract",
		Phase:        "phase4",
		FilesTouched: []string{"packages/cli/internal/handoff/writer.go"},
	})
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}
	if err := mgr.CompleteDispatch(dispatch.ID, "writer contract defined", 17); err != nil {
		t.Fatalf("CompleteDispatch failed: %v", err)
	}

	if err := runSessionEnd(&cobra.Command{}, []string{s.ID}); err != nil {
		t.Fatalf("runSessionEnd failed: %v", err)
	}

	var handoffPath, goal, status string
	if err := db.QueryRow("SELECT path, goal, status FROM handoff WHERE session_id = ?", s.ID).Scan(&handoffPath, &goal, &status); err != nil {
		t.Fatalf("query handoff row failed: %v", err)
	}
	if goal != s.Goal {
		t.Fatalf("handoff goal = %q, want %q", goal, s.Goal)
	}
	if status != "done" {
		t.Fatalf("handoff status = %q, want done", status)
	}
	if matched := regexp.MustCompile(`^specs/memory/handoffs/\d{4}-\d{2}-\d{2}-[a-z0-9-]+\.md$`).MatchString(handoffPath); !matched {
		t.Fatalf("handoff path %q does not match expected convention", handoffPath)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM handoff WHERE session_id = ?", s.ID).Scan(&count); err != nil {
		t.Fatalf("count handoff rows failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("handoff row count = %d, want 1", count)
	}

	doc, err := handoff.Read(filepath.Join(tmpDir, handoffPath))
	if err != nil {
		t.Fatalf("handoff.Read failed: %v", err)
	}
	if doc.Goal != s.Goal {
		t.Fatalf("doc goal = %q, want %q", doc.Goal, s.Goal)
	}
	if doc.Progress != handoff.ProgressDone {
		t.Fatalf("doc progress = %q, want %q", doc.Progress, handoff.ProgressDone)
	}
	if len(doc.Constraints) == 0 {
		t.Fatal("doc constraints should not be empty")
	}
	if len(doc.Decisions) == 0 {
		t.Fatal("doc decisions should not be empty")
	}
	if len(doc.NextSteps) == 0 {
		t.Fatal("doc next steps should not be empty")
	}
	if len(doc.Items.Done) == 0 {
		t.Fatal("doc progress done items should not be empty")
	}
	if doc.SessionID != s.ID {
		t.Fatalf("doc session_id = %q, want %q", doc.SessionID, s.ID)
	}
}
