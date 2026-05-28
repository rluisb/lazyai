package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTestWorkflow creates a minimal valid workflow YAML under
// <dir>/.opencode/workflows/<name>.yaml suitable for both the legacy CLI
// parser (workflow.go) and the runtime parser (internal/runtime/workflow).
func writeTestWorkflow(t *testing.T, dir, name string) {
	t.Helper()
	workflowsDir := filepath.Join(dir, ".opencode", "workflows")
	if err := os.MkdirAll(workflowsDir, 0o755); err != nil {
		t.Fatalf("failed to create workflows dir: %v", err)
	}
	content := fmt.Sprintf(`name: %s
version: "1.0"
description: A test workflow
phases:
  - name: Phase 1
    description: First phase
    agent: test-agent
    mode: standard
  - name: Phase 2
    description: Second phase
    agent: test-agent
    mode: standard
`, name)
	path := filepath.Join(workflowsDir, name+".yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write workflow file: %v", err)
	}
}

func TestWorkflowList(t *testing.T) {
	withTempDir(t)
	writeTestWorkflow(t, ".", "test-workflow")

	out := captureStdout(t, func() {
		if err := workflowListCmd.RunE(workflowListCmd, []string{}); err != nil {
			t.Fatalf("workflow list failed: %v", err)
		}
	})

	if !strings.Contains(out, "test-workflow") {
		t.Errorf("expected output to contain workflow name, got:\n%s", out)
	}
	if !strings.Contains(out, "2 phases") {
		t.Errorf("expected output to contain phase count, got:\n%s", out)
	}
}

func TestWorkflowShow(t *testing.T) {
	withTempDir(t)
	writeTestWorkflow(t, ".", "test-workflow")

	out := captureStdout(t, func() {
		if err := workflowShowCmd.RunE(workflowShowCmd, []string{"test-workflow"}); err != nil {
			t.Fatalf("workflow show failed: %v", err)
		}
	})

	if !strings.Contains(out, "Workflow: test-workflow") {
		t.Errorf("expected output to contain 'Workflow: test-workflow', got:\n%s", out)
	}
	if !strings.Contains(out, "Phase 1") {
		t.Errorf("expected output to contain Phase 1, got:\n%s", out)
	}
	if !strings.Contains(out, "Phase 2") {
		t.Errorf("expected output to contain Phase 2, got:\n%s", out)
	}
	if !strings.Contains(out, "Agent: test-agent") {
		t.Errorf("expected output to contain agent info, got:\n%s", out)
	}
}

func TestWorkflowRunDryRun(t *testing.T) {
	withTempDir(t)
	writeTestWorkflow(t, ".", "test-workflow")

	// Preserve and restore the --dry-run flag so other tests are not affected.
	oldDryRun, _ := workflowRunCmd.Flags().GetBool("dry-run")
	_ = workflowRunCmd.Flags().Set("dry-run", "true")
	t.Cleanup(func() {
		_ = workflowRunCmd.Flags().Set("dry-run", fmt.Sprintf("%t", oldDryRun))
	})

	out := captureStdout(t, func() {
		if err := workflowRunCmd.RunE(workflowRunCmd, []string{"test-workflow"}); err != nil {
			t.Fatalf("workflow run failed: %v", err)
		}
	})

	if !strings.Contains(out, "DRY RUN MODE") {
		t.Errorf("expected output to contain 'DRY RUN MODE', got:\n%s", out)
	}
	if !strings.Contains(out, "Dry run complete") {
		t.Errorf("expected output to contain 'Dry run complete', got:\n%s", out)
	}

	// Verify runtime DB tracking records.
	db, err := openRuntimeDB()
	if err != nil {
		t.Fatalf("openRuntimeDB failed: %v", err)
	}
	defer db.Close()

	var status string
	var currentStep, totalSteps int
	if err := db.QueryRow(
		"SELECT status, current_step, total_steps FROM workflow_instances WHERE workflow_name = ?",
		"test-workflow",
	).Scan(&status, &currentStep, &totalSteps); err != nil {
		t.Fatalf("failed to query workflow instance: %v", err)
	}
	if status != "dry_run" {
		t.Errorf("expected status='dry_run', got %q", status)
	}
	if currentStep != 2 {
		t.Errorf("expected current_step=2, got %d", currentStep)
	}
	if totalSteps != 2 {
		t.Errorf("expected total_steps=2, got %d", totalSteps)
	}
}

func TestWorkflowSync(t *testing.T) {
	withTempDir(t)
	writeTestWorkflow(t, ".", "test-workflow")

	out := captureStdout(t, func() {
		if err := workflowSyncCmd.RunE(workflowSyncCmd, []string{}); err != nil {
			t.Fatalf("workflow sync failed: %v", err)
		}
	})

	if !strings.Contains(out, "Workflows synced") {
		t.Errorf("expected output to contain 'Workflows synced', got:\n%s", out)
	}

	// upsertWorkflow in internal/runtime/workflow/sync.go is currently a TODO
	// and does not write to the DB. We verify no panic occurred and the command
	// reported success; DB evidence will appear once the TODO is implemented.
}
