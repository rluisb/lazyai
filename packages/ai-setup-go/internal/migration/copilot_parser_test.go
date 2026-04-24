package migration

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseCopilotSetup_InstructionsAndChatModes(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".github", "copilot-chat-modes"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(dir, ".github", "copilot-instructions.md"),
		[]byte("# Copilot for my repo\n\nHelps with code review.\n"),
		0o644,
	); err != nil {
		t.Fatalf("instructions: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(dir, ".github", "copilot-chat-modes", "architect.chatmode.md"),
		[]byte("# Architect\n\nA mode for architecture discussions.\n"),
		0o644,
	); err != nil {
		t.Fatalf("chatmode: %v", err)
	}

	parsed, err := parseCopilotSetup(dir)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if parsed.ProjectName != "Copilot for my repo" {
		t.Errorf("ProjectName = %q, want \"Copilot for my repo\"", parsed.ProjectName)
	}
	if parsed.Metadata["adapter"] != "copilot" {
		t.Errorf("Metadata[adapter] = %q, want copilot", parsed.Metadata["adapter"])
	}
	if len(parsed.Agents) != 1 || parsed.Agents[0].ID != "architect.chatmode" {
		t.Errorf("expected architect.chatmode agent, got %+v", parsed.Agents)
	}
	var gotSection bool
	for _, s := range parsed.CustomSections {
		if s.ID == "copilot-instructions" {
			gotSection = true
		}
	}
	if !gotSection {
		t.Errorf("expected copilot-instructions custom section")
	}
}

func TestParseCopilotSetup_InstructionsFilesAsRules(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".github", "instructions"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".github", "copilot-instructions.md"), []byte("# Repo\n"), 0o644); err != nil {
		t.Fatalf("root: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(dir, ".github", "instructions", "go.instructions.md"),
		[]byte("# Go\n\nUse Go 1.22+.\n"),
		0o644,
	); err != nil {
		t.Fatalf("rule: %v", err)
	}

	parsed, err := parseCopilotSetup(dir)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(parsed.Rules) != 1 || parsed.Rules[0].Category != "copilot-instruction" {
		t.Errorf("expected single copilot-instruction rule, got %+v", parsed.Rules)
	}
}

func TestParseCopilotSetup_EmptyReturnsError(t *testing.T) {
	dir := t.TempDir()
	_, err := parseCopilotSetup(dir)
	if err == nil {
		t.Error("expected error for empty setup, got nil")
	}
}
