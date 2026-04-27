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

// TestCodexAdapter_DriveCLI_FallsBackWhenBinaryAbsent verifies that when
// DriveCLI=true but the codex binary is not on PATH, Install succeeds by
// falling back to direct-write (config.toml still created via configmerge).
func TestCodexAdapter_DriveCLI_FallsBackWhenBinaryAbsent(t *testing.T) {
	ctx, _, _ := newScopeTestContext(t, types.SetupScopeProject)
	ctx.DriveCLI = true

	emptyDir := t.TempDir()
	origPATH := os.Getenv("PATH")
	t.Setenv("PATH", emptyDir+string(os.PathListSeparator)+origPATH)

	adapter := &CodexAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("Install with DriveCLI=true and absent binary must not error: %v", err)
	}

	configPath := filepath.Join(ctx.TargetDir, ".codex", "config.toml")
	if !files.FileExists(configPath) {
		t.Errorf("config.toml not created via direct-write fallback")
	}
}

// TestCodexAdapter_DriveCLI_CallsCodexBinary verifies that when DriveCLI=true
// and a stub codex binary is on PATH, the adapter invokes `codex mcp add`
// with the expected positional/separator structure.
func TestCodexAdapter_DriveCLI_CallsCodexBinary(t *testing.T) {
	ctx, _, _ := newScopeTestContext(t, types.SetupScopeProject)
	ctx.DriveCLI = true

	stubDir := t.TempDir()
	recordFile := filepath.Join(stubDir, "codex-args.txt")
	stubScript := fmt.Sprintf("#!/bin/sh\necho \"$@\" >> %s\n", recordFile)
	stubPath := filepath.Join(stubDir, "codex")
	if err := os.WriteFile(stubPath, []byte(stubScript), 0o755); err != nil {
		t.Fatal(err)
	}

	// Canonical .ai/mcp.json with one server using command + args + env.
	aiDir := filepath.Join(ctx.TargetDir, ".ai")
	if err := files.EnsureDir(aiDir); err != nil {
		t.Fatal(err)
	}
	mcpJSON := `{"servers":{"filesystem":{"command":"npx","args":["-y","@modelcontextprotocol/server-filesystem"],"env":{"FOO":"bar"}}}}`
	if err := files.WriteFile(filepath.Join(aiDir, "mcp.json"), []byte(mcpJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	origPATH := os.Getenv("PATH")
	t.Setenv("PATH", stubDir+string(os.PathListSeparator)+origPATH)

	adapter := &CodexAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("Install: %v", err)
	}

	if !files.FileExists(recordFile) {
		t.Fatal("stub codex binary was not invoked when DriveCLI=true and binary present")
	}
	data, _ := files.ReadFile(recordFile)
	s := string(data)
	if !strings.Contains(s, "mcp add filesystem") {
		t.Errorf("expected 'mcp add filesystem' in stub args, got: %s", s)
	}
	if !strings.Contains(s, "--env FOO=bar") {
		t.Errorf("expected '--env FOO=bar' in stub args, got: %s", s)
	}
	if !strings.Contains(s, "-- npx") {
		t.Errorf("expected '-- npx' separator+command in stub args, got: %s", s)
	}
}
