package cmd

import (
	"errors"
	"strings"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/aimanifest"
)

func TestValidateToolFlagAcceptsManifestAliases(t *testing.T) {
	// Aliases and canonical IDs that the manifest accepts must also be valid
	// --tool values.
	for _, token := range []string{"claude", "claude-code", "opencode", "copilot", "pi", "omp", "antigravity", "kiro"} {
		if err := validateToolFlag(token); err != nil {
			t.Errorf("validateToolFlag(%q) = %v, want nil", token, err)
		}
	}
	// Empty string means "no filter" and must be accepted.
	if err := validateToolFlag(""); err != nil {
		t.Errorf(`validateToolFlag("") = %v, want nil`, err)
	}
}

func TestValidateToolFlagRejectsCodexWithHelpfulMessage(t *testing.T) {
	err := validateToolFlag("codex")
	if err == nil {
		t.Fatal("want error for codex, got nil")
	}
	if !errors.Is(err, aimanifest.ErrCodexUnsupported) {
		// The error message should at least mention codex removal.
		if !strings.Contains(err.Error(), "codex") || !strings.Contains(err.Error(), "removed") {
			t.Fatalf("error for codex should mention removal, got %q", err.Error())
		}
	}
}

func TestValidateToolFlagRejectsUnknownWithSupportedList(t *testing.T) {
	err := validateToolFlag("gemini")
	if err == nil {
		t.Fatal("want error for gemini, got nil")
	}
	if !strings.Contains(err.Error(), "gemini") {
		t.Fatalf("error should mention the bad tool, got %q", err.Error())
	}
	// Error should list supported tools.
	for _, want := range []string{"opencode", "claude-code", "copilot", "pi", "omp", "kiro", "antigravity"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error should list %q in supported tools, got %q", want, err.Error())
		}
	}
}
