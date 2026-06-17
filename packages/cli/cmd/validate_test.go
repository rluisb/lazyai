package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateAgentsPassesOnCanonicalShape(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, ".opencode", "agents")
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		t.Fatalf("mkdir agents dir: %v", err)
	}

	agentPath := filepath.Join(agentsDir, "canonical-agent.md")
	agentContent := `---
description: test agent
---
<!-- vibe-lab:managed kind=agent role=architector -->
# System Prompt
## Rules
Use stable, minimal changes.
`
	if err := os.WriteFile(agentPath, []byte(agentContent), 0o644); err != nil {
		t.Fatalf("write agent file: %v", err)
	}

	withWorkingDir(t, dir)

	var err error
	_, _ = captureOutput(t, func() {
		err = runValidateAgents(nil, nil)
	})
	if err != nil {
		t.Fatalf("expected validation to pass, got: %v", err)
	}
}

func TestValidateAgentsFailsOnMissingFrontmatter(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, ".opencode", "agents")
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		t.Fatalf("mkdir agents dir: %v", err)
	}

	agentPath := filepath.Join(agentsDir, "bad-agent.md")
	if err := os.WriteFile(agentPath, []byte(`# System Prompt
## Rules
`), 0o644); err != nil {
		t.Fatalf("write agent file: %v", err)
	}

	withWorkingDir(t, dir)

	var err error
	_, _ = captureOutput(t, func() {
		err = runValidateAgents(nil, nil)
	})
	if err == nil {
		t.Fatalf("expected validation failure for missing frontmatter")
	}
	if !strings.Contains(err.Error(), "validation failed") {
		t.Fatalf("expected validation failed error, got: %v", err)
	}
}

func TestValidateAgentsFailsOnMissingSystemPrompt(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, ".opencode", "agents")
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		t.Fatalf("mkdir agents dir: %v", err)
	}

	agentPath := filepath.Join(agentsDir, "missing-system-prompt.md")
	if err := os.WriteFile(agentPath, []byte(`---
description: test
---
## Rules
`), 0o644); err != nil {
		t.Fatalf("write agent file: %v", err)
	}

	withWorkingDir(t, dir)

	var err error
	_, _ = captureOutput(t, func() {
		err = runValidateAgents(nil, nil)
	})
	if err == nil {
		t.Fatalf("expected validation failure for missing system prompt")
	}
}
