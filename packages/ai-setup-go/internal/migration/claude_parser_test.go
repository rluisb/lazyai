package migration

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseClaudeSetup_RootOnly(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# My Project\n\nA description.\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	// Need at least one parseable artifact; add a rule.
	if err := os.MkdirAll(filepath.Join(dir, ".claude", "rules"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".claude", "rules", "test.md"), []byte("# Testing\n\nRun tests.\n"), 0o644); err != nil {
		t.Fatalf("rule: %v", err)
	}

	parsed, err := parseClaudeSetup(dir)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if parsed.ProjectName != "My Project" {
		t.Errorf("ProjectName = %q, want \"My Project\"", parsed.ProjectName)
	}
	if parsed.Description == "" {
		t.Errorf("Description should be populated from the first paragraph")
	}
	if parsed.Metadata["adapter"] != "claude-code" {
		t.Errorf("Metadata[adapter] = %q, want claude-code", parsed.Metadata["adapter"])
	}

	var gotRoot bool
	for _, s := range parsed.CustomSections {
		if s.ID == "claude-root" {
			gotRoot = true
		}
	}
	if !gotRoot {
		t.Errorf("expected claude-root custom section")
	}
	if len(parsed.Rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(parsed.Rules))
	}
	if parsed.Rules[0].Category != "claude-rule" {
		t.Errorf("rule category = %q, want claude-rule", parsed.Rules[0].Category)
	}
}

func TestParseClaudeSetup_AgentsAndCommands(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".claude", "commands"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".claude", "reviewer.md"), []byte("# Reviewer Agent\n\nReviews changes.\n"), 0o644); err != nil {
		t.Fatalf("agent: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".claude", "commands", "ship.md"), []byte("# Ship\n\nShip the code.\n"), 0o644); err != nil {
		t.Fatalf("command: %v", err)
	}

	parsed, err := parseClaudeSetup(dir)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(parsed.Agents) != 1 || parsed.Agents[0].ID != "reviewer" {
		t.Errorf("expected reviewer agent, got %+v", parsed.Agents)
	}
	if len(parsed.Commands) != 1 || parsed.Commands[0].ID != "ship" {
		t.Errorf("expected ship command, got %+v", parsed.Commands)
	}
}

func TestParseClaudeSetup_EmptyReturnsError(t *testing.T) {
	dir := t.TempDir()
	_, err := parseClaudeSetup(dir)
	if err == nil {
		t.Error("expected error for empty setup, got nil")
	}
}
