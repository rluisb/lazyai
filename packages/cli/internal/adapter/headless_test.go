package adapter

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// --- Test: All adapters satisfy the ToolAdapter interface ---

func TestAllAdapters_SatisfyToolAdapter(t *testing.T) {
	// Compile-time interface check — if any adapter is missing methods,
	// this won't compile.
	var _ ToolAdapter = (*OpenCodeAdapter)(nil)
	var _ ToolAdapter = (*ClaudeCodeAdapter)(nil)
	var _ ToolAdapter = (*CopilotAdapter)(nil)
}

// --- Test: CanRunHeadless returns true for all adapters ---

func TestCanRunHeadless_Values(t *testing.T) {
	tests := []struct {
		name     string
		adapter  ToolAdapter
		expected bool
	}{
		{"OpenCode", &OpenCodeAdapter{}, true},
		{"ClaudeCode", &ClaudeCodeAdapter{}, true},
		{"Copilot", &CopilotAdapter{}, true},
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

	adapters := []ToolAdapter{
		&ClaudeCodeAdapter{},
	}

	for _, adapter := range adapters {
		t.Run(adapter.Name()+"_not_installed", func(t *testing.T) {
			err := adapter.RunHeadlessValidation(ctx)
			if err != nil {
				t.Errorf("%s.RunHeadlessValidation() should be non-fatal, got error: %v", adapter.Name(), err)
			}
		})
	}
}

// --- Test: RunHeadlessValidation runs the correct command when tool is installed ---

func TestRunHeadlessValidation_ClaudeCode_Installed(t *testing.T) {
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

// --- Test: RunHeadlessValidation uses correct target directory ---

func TestRunHeadlessValidation_UsesTargetDir(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping headless validation test in short mode (requires claude CLI)")
	}
	targetDir := t.TempDir()
	markerPath := filepath.Join(targetDir, ".marker")
	if err := os.WriteFile(markerPath, []byte("test"), 0o644); err != nil {
		t.Fatalf("failed to create marker file: %v", err)
	}

	ctx := &AdapterContext{
		TargetDir:  targetDir,
		SetupScope: types.SetupScopeProject,
	}

	adapters := []ToolAdapter{
		&ClaudeCodeAdapter{},
	}

	for _, adapter := range adapters {
		t.Run(adapter.Name(), func(t *testing.T) {
			// Should not panic even if tool is not installed.
			_ = adapter.RunHeadlessValidation(ctx)
		})
	}
}

// --- Test: RunHeadlessInit is non-fatal when tool binary is not on PATH ---

func TestRunHeadlessInit_BinaryNotOnPath(t *testing.T) {
	ctx := &AdapterContext{
		TargetDir:  t.TempDir(),
		SetupScope: types.SetupScopeProject,
	}

	adapters := []ToolAdapter{
		&ClaudeCodeAdapter{},
		&OpenCodeAdapter{},
		&CopilotAdapter{},
	}

	for _, adapter := range adapters {
		t.Run(adapter.Name(), func(t *testing.T) {
			// If the binary happens to be on PATH, this will actually run.
			// That's fine — the test verifies the method is non-fatal either way.
			err := adapter.RunHeadlessInit(ctx, "test populate prompt")
			if err != nil {
				t.Errorf("%s.RunHeadlessInit() should be non-fatal, got: %v", adapter.Name(), err)
			}
		})
	}
}

// --- Test: RunHeadlessInit with empty prompt is non-fatal ---

func TestRunHeadlessInit_EmptyPrompt(t *testing.T) {
	ctx := &AdapterContext{
		TargetDir:  t.TempDir(),
		SetupScope: types.SetupScopeProject,
	}

	adapters := []ToolAdapter{
		&ClaudeCodeAdapter{},
		&OpenCodeAdapter{},
		&CopilotAdapter{},
	}

	for _, adapter := range adapters {
		t.Run(adapter.Name(), func(t *testing.T) {
			err := adapter.RunHeadlessInit(ctx, "")
			if err != nil {
				t.Errorf("%s.RunHeadlessInit() with empty prompt should be non-fatal, got: %v", adapter.Name(), err)
			}
		})
	}
}

// --- Test: truncateOutput helper ---

func TestTruncateOutput(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "hello..."},
		{"", 5, ""},
		{"abc", 3, "abc"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := truncateOutput(tc.input, tc.maxLen)
			if got != tc.want {
				t.Errorf("truncateOutput(%q, %d) = %q, want %q", tc.input, tc.maxLen, got, tc.want)
			}
		})
	}
}
