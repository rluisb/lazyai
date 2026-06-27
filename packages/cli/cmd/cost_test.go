package cmd

import (
	"strings"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/db"
)

// TestCostAgentReportsDispatchCounts is a regression test for the cost agent
// command, which previously queried a non-existent agent_dispatches.cost_usd
// column (cost_usd lives on the sessions table). The command must now run
// without error and report per-agent dispatch counts with a completed/failed
// breakdown.
func TestCostAgentReportsDispatchCounts(t *testing.T) {
	dir := t.TempDir()
	withWorkingDir(t, dir)

	database, err := db.Open(db.DefaultDBPath(dir))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer database.Close()
	if err := db.RunMigrations(database); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	// agent_dispatches.session_id is a NOT NULL foreign key into sessions, and
	// foreign keys are enforced, so seed a parent session first.
	if _, err := database.Exec(
		`INSERT INTO sessions (id, goal, status, started_at) VALUES (?, ?, ?, ?)`,
		"s1", "test goal", "active", "2026-06-27T00:00:00Z",
	); err != nil {
		t.Fatalf("seed session: %v", err)
	}

	seed := []struct {
		seq    int
		agent  string
		status string
	}{
		{1, "explore", "completed"},
		{2, "explore", "failed"},
		{3, "task", "completed"},
	}
	for _, d := range seed {
		if _, err := database.Exec(
			`INSERT INTO agent_dispatches (session_id, seq, agent, status) VALUES (?, ?, ?, ?)`,
			"s1", d.seq, d.agent, d.status,
		); err != nil {
			t.Fatalf("seed dispatch %d: %v", d.seq, err)
		}
	}

	var runErr error
	out := captureStdout(t, func() {
		runErr = costAgentCmd.RunE(costAgentCmd, []string{})
	})
	if runErr != nil {
		t.Fatalf("cost agent returned error: %v", runErr)
	}

	wants := []string{
		"explore: 2 dispatches (1 completed, 1 failed)",
		"task: 1 dispatches (1 completed, 0 failed)",
		"cost is tracked per session",
	}
	for _, w := range wants {
		if !strings.Contains(out, w) {
			t.Errorf("output missing %q\n---\n%s", w, out)
		}
	}
}
