package adapter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// TestClaudeAdapter_DriveCLI_FallsBackWhenBinaryAbsent verifies that when
// DriveCLI=true but the claude binary is not on PATH, Install succeeds by
// falling back to direct-write (settings.json still created via configmerge).
func TestClaudeAdapter_DriveCLI_FallsBackWhenBinaryAbsent(t *testing.T) {
	ctx, _, _ := newScopeTestContext(t, types.SetupScopeProject)
	ctx.DriveCLI = true

	// Prepend an empty tmpdir to PATH so `claude` isn't found.
	emptyDir := t.TempDir()
	origPATH := os.Getenv("PATH")
	t.Setenv("PATH", emptyDir+string(os.PathListSeparator)+origPATH)

	adapter := &ClaudeCodeAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("Install with DriveCLI=true and absent binary must not error: %v", err)
	}

	settingsPath := filepath.Join(ctx.TargetDir, ".claude", "settings.json")
	if !files.FileExists(settingsPath) {
		t.Errorf("settings.json not created via direct-write fallback")
	}
}

// TestClaudeAdapter_DriveCLI_CallsClaudeBinary verifies that when DriveCLI=true
// and a stub claude binary is on PATH, the adapter invokes `claude mcp` commands.
// (spec 012: uses `mcp get` to pre-check, then `mcp add-json` to register)
func TestClaudeAdapter_DriveCLI_CallsClaudeBinary(t *testing.T) {
	ctx, _, _ := newScopeTestContext(t, types.SetupScopeProject)
	ctx.DriveCLI = true

	// Stub claude binary that records its args.
	stubDir := t.TempDir()
	recordFile := filepath.Join(stubDir, "claude-args.txt")
	stubScript := fmt.Sprintf("#!/bin/sh\necho \"$@\" >> %s\nexit 1\n", recordFile)
	stubPath := filepath.Join(stubDir, "claude")
	if err := os.WriteFile(stubPath, []byte(stubScript), 0o755); err != nil {
		t.Fatal(err)
	}

	// Canonical .ai/mcp.json so the adapter has a server to register.
	aiDir := filepath.Join(ctx.TargetDir, ".ai")
	if err := files.EnsureDir(aiDir); err != nil {
		t.Fatal(err)
	}
	mcpJSON := `{"servers":{"filesystem":{"command":"npx","args":["-y","@modelcontextprotocol/server-filesystem"]}}}`
	if err := files.WriteFile(filepath.Join(aiDir, "mcp.json"), []byte(mcpJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	origPATH := os.Getenv("PATH")
	t.Setenv("PATH", stubDir+string(os.PathListSeparator)+origPATH)

	adapter := &ClaudeCodeAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("Install: %v", err)
	}

	if !files.FileExists(recordFile) {
		t.Fatal("stub claude binary was not called when DriveCLI=true and binary present")
	}
	data, _ := files.ReadFile(recordFile)
	// Expect either 'mcp get' (pre-check) or 'mcp add-json' (registration).
	contents := string(data)
	if !strings.Contains(contents, "mcp get") && !strings.Contains(contents, "mcp add-json") {
		t.Errorf("expected 'mcp get' or 'mcp add-json' in stub args, got: %s", contents)
	}
}
