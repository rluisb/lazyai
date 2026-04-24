package migration

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseGeminiSetup_RootAndCommands(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "GEMINI.md"), []byte("# Gem Project\n\nA description.\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".gemini", "commands"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".gemini", "commands", "review.toml"), []byte("prompt = \"Review this code\"\n"), 0o644); err != nil {
		t.Fatalf("command: %v", err)
	}

	parsed, err := parseGeminiSetup(dir)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if parsed.ProjectName != "Gem Project" {
		t.Errorf("ProjectName = %q, want \"Gem Project\"", parsed.ProjectName)
	}
	if parsed.Metadata["adapter"] != "gemini" {
		t.Errorf("Metadata[adapter] = %q, want gemini", parsed.Metadata["adapter"])
	}
	if len(parsed.Commands) != 1 || parsed.Commands[0].ID != "review" {
		t.Errorf("expected review command, got %+v", parsed.Commands)
	}
}

func TestParseGeminiSetup_EmptyReturnsError(t *testing.T) {
	dir := t.TempDir()
	_, err := parseGeminiSetup(dir)
	if err == nil {
		t.Error("expected error for empty setup, got nil")
	}
}
