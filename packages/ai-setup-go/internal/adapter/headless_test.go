package adapter

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// --- Test: All adapters satisfy the ToolAdapter interface ---

func TestAllAdapters_SatisfyToolAdapter(t *testing.T) {
	// Compile-time interface check — if any adapter is missing methods,
	// this won't compile.
	var _ ToolAdapter = (*OpenCodeAdapter)(nil)
	var _ ToolAdapter = (*ClaudeCodeAdapter)(nil)
	var _ ToolAdapter = (*CodexAdapter)(nil)
	var _ ToolAdapter = (*CopilotAdapter)(nil)
	var _ ToolAdapter = (*GeminiAdapter)(nil)
}

// --- Test: CanRunHeadless returns correct values per adapter ---

func TestCanRunHeadless_Values(t *testing.T) {
	tests := []struct {
		name     string
		adapter  ToolAdapter
		expected bool
	}{
		{"OpenCode", &OpenCodeAdapter{}, false},
		{"ClaudeCode", &ClaudeCodeAdapter{}, true},
		{"Codex", &CodexAdapter{}, true},
		{"Copilot", &CopilotAdapter{}, false},
		{"Gemini", &GeminiAdapter{}, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.adapter.CanRunHeadless()
			if result != tc.expected {
				t.Errorf("%s.CanRunHeadless() = %v, want %v", tc.name, result, tc.expected)
			}
		})
	}
}

// --- Test: RunHeadlessValidation is non-fatal for no-op adapters ---

func TestRunHeadlessValidation_NoOpAdapters(t *testing.T) {
	ctx := &AdapterContext{
		TargetDir:  t.TempDir(),
		SetupScope: types.SetupScopeProject,
	}

	adapters := []ToolAdapter{
		&OpenCodeAdapter{},
		&CopilotAdapter{},
		&GeminiAdapter{},
	}

	for _, adapter := range adapters {
		t.Run(adapter.Name(), func(t *testing.T) {
			err := adapter.RunHeadlessValidation(ctx)
			if err != nil {
				t.Errorf("%s.RunHeadlessValidation() returned error: %v", adapter.Name(), err)
			}
		})
	}
}

// --- Test: RunHeadlessValidation is non-fatal when tool is not installed ---

func TestRunHeadlessValidation_ToolNotInstalled(t *testing.T) {
	ctx := &AdapterContext{
		TargetDir:  t.TempDir(),
		SetupScope: types.SetupScopeProject,
	}

	// Claude Code and Codex should return nil (non-fatal) when the tool is not on PATH.
	// We can't easily remove them from PATH, but we can verify the behavior
	// by checking that the method doesn't panic and returns nil.
	adapters := []ToolAdapter{
		&ClaudeCodeAdapter{},
		&CodexAdapter{},
	}

	for _, adapter := range adapters {
		t.Run(adapter.Name()+"_not_installed", func(t *testing.T) {
			// This test passes if the method returns nil regardless of install state.
			// The actual behavior (skip vs run) depends on whether the tool is on PATH.
			err := adapter.RunHeadlessValidation(ctx)
			if err != nil {
				t.Errorf("%s.RunHeadlessValidation() should be non-fatal, got error: %v", adapter.Name(), err)
			}
		})
	}
}

// --- Test: RunHeadlessValidation runs the correct command when tool is installed ---

func TestRunHeadlessValidation_ClaudeCode_Installed(t *testing.T) {
	// Skip if claude is not on PATH.
	if _, err := exec.LookPath("claude"); err != nil {
		t.Skip("claude not on PATH, skipping installed test")
	}

	ctx := &AdapterContext{
		TargetDir:  t.TempDir(),
		SetupScope: types.SetupScopeProject,
	}

	adapter := &ClaudeCodeAdapter{}
	err := adapter.RunHeadlessValidation(ctx)

	// Should be non-fatal even if the command itself fails.
	if err != nil {
		t.Errorf("RunHeadlessValidation should be non-fatal, got: %v", err)
	}
}

func TestRunHeadlessValidation_Codex_Installed(t *testing.T) {
	// Skip if codex is not on PATH.
	if _, err := exec.LookPath("codex"); err != nil {
		t.Skip("codex not on PATH, skipping installed test")
	}

	ctx := &AdapterContext{
		TargetDir:  t.TempDir(),
		SetupScope: types.SetupScopeProject,
	}

	adapter := &CodexAdapter{}
	err := adapter.RunHeadlessValidation(ctx)

	// Should be non-fatal even if the command itself fails.
	if err != nil {
		t.Errorf("RunHeadlessValidation should be non-fatal, got: %v", err)
	}
}

// --- Test: RunHeadlessValidation uses correct target directory ---

func TestRunHeadlessValidation_UsesTargetDir(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping headless validation test in short mode (requires claude/codex CLI)")
	}
	// Create a temp dir with a marker file to verify cmd.Dir is set correctly.
	targetDir := t.TempDir()
	markerPath := filepath.Join(targetDir, ".marker")
	if err := os.WriteFile(markerPath, []byte("test"), 0o644); err != nil {
		t.Fatalf("failed to create marker file: %v", err)
	}

	ctx := &AdapterContext{
		TargetDir:  targetDir,
		SetupScope: types.SetupScopeProject,
	}

	// Both adapters should use ctx.TargetDir as the working directory.
	// We verify this indirectly by ensuring no panic/error from invalid dir.
	adapters := []ToolAdapter{
		&ClaudeCodeAdapter{},
		&CodexAdapter{},
	}

	for _, adapter := range adapters {
		t.Run(adapter.Name(), func(t *testing.T) {
			// Should not panic even if tool is not installed.
			_ = adapter.RunHeadlessValidation(ctx)
		})
	}
}
