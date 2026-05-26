package workflow

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseYAML(t *testing.T) {
	// Create a temporary workflow YAML
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "test.yaml")
	yamlContent := `name: test
trigger: /test
description: Test workflow
modes:
  simple:
    phases: [plan, verify]
    skip_gates: false
default_mode: simple
phases:
  - name: plan
    agent: turbo-crank
    skill: storm-scout
    mode: plan
    feedforward: "Task: {GOAL}. Plan it."
    gate:
      type: human
      description: Human approval
      prompt: Approve plan?
    metrics:
      - record_latency
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("write test yaml: %v", err)
	}

	wf, err := ParseYAML(yamlPath)
	if err != nil {
		t.Fatalf("ParseYAML failed: %v", err)
	}

	if wf.Name != "test" {
		t.Errorf("name = %q, want %q", wf.Name, "test")
	}
	if wf.Trigger != "/test" {
		t.Errorf("trigger = %q, want %q", wf.Trigger, "/test")
	}
	if len(wf.Steps) != 1 {
		t.Fatalf("steps = %d, want 1", len(wf.Steps))
	}
	if wf.Steps[0].Name != "plan" {
		t.Errorf("step[0].name = %q, want %q", wf.Steps[0].Name, "plan")
	}
	if wf.Steps[0].Agent != "turbo-crank" {
		t.Errorf("step[0].agent = %q, want %q", wf.Steps[0].Agent, "turbo-crank")
	}
	if wf.Steps[0].Gate == nil {
		t.Fatal("step[0].gate is nil")
	}
	if wf.Steps[0].Gate.Type != "human" {
		t.Errorf("gate.type = %q, want %q", wf.Steps[0].Gate.Type, "human")
	}
}

func TestParseYAMLInvalid(t *testing.T) {
	dir := t.TempDir()

	// Missing name
	yamlPath := filepath.Join(dir, "invalid.yaml")
	os.WriteFile(yamlPath, []byte("phases:\n  - name: plan\n    agent: test\n"), 0644)
	_, err := ParseYAML(yamlPath)
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestParseRealWorkflows(t *testing.T) {
	// Parse actual workflow files from library
	workflowsDir := "../../../library/fortnite/workflows"
	entries, err := os.ReadDir(workflowsDir)
	if err != nil {
		t.Skipf("cannot read workflows dir: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}

		t.Run(entry.Name(), func(t *testing.T) {
			path := filepath.Join(workflowsDir, entry.Name())
			wf, err := ParseYAML(path)
			if err != nil {
				t.Fatalf("ParseYAML(%s): %v", entry.Name(), err)
			}
			if wf.Name == "" {
				t.Error("workflow name is empty")
			}
			if len(wf.Steps) == 0 {
				t.Error("workflow has no steps")
			}
			if wf.Config.DefaultMode == "" {
				t.Error("workflow has no default_mode")
			}
		})
	}
}
