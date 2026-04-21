package adapter

import (
	"context"
	"testing"
)

// TestLookupClaudeBinary_Found tests that LookupClaudeBinary returns a path
// when the claude binary exists on PATH.
func TestLookupClaudeBinary_Found(t *testing.T) {
	// This test assumes `claude` is on PATH during testing.
	// If the test fails, it means the claude binary is not installed.
	path, found := LookupClaudeBinary()
	if !found {
		t.Skip("claude binary not found on PATH; skipping integration test")
	}
	if path == "" {
		t.Error("path should not be empty when found=true")
	}
}

// TestDefaultClaudeCLIRunner_VersionCommand tests that the runner can execute
// a basic claude command.
func TestDefaultClaudeCLIRunner_VersionCommand(t *testing.T) {
	if _, found := LookupClaudeBinary(); !found {
		t.Skip("claude binary not found on PATH; skipping integration test")
	}

	runner := &DefaultClaudeCLIRunner{}
	ctx := context.Background()
	stdout, stderr, err := runner.Run(ctx, "", "--version")

	// The --version flag should succeed
	if err != nil {
		t.Fatalf("claude --version failed: %v\nstderr: %s", err, string(stderr))
	}

	// Verify stdout contains version info
	if len(stdout) == 0 {
		t.Error("expected stdout from claude --version")
	}
}