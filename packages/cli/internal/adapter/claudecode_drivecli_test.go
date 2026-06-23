package adapter

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
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
	var stubPath string
	if runtime.GOOS == "windows" {
		stubScript := fmt.Sprintf("@echo off\r\necho %%* >> %s\r\nexit /b 1\r\n", recordFile)
		stubPath = filepath.Join(stubDir, "claude.bat")
		if err := os.WriteFile(stubPath, []byte(stubScript), 0o644); err != nil {
			t.Fatal(err)
		}
	} else {
		stubScript := fmt.Sprintf("#!/bin/sh\necho \"$@\" >> %s\nexit 1\n", recordFile)
		stubPath = filepath.Join(stubDir, "claude")
		if err := os.WriteFile(stubPath, []byte(stubScript), 0o755); err != nil {
			t.Fatal(err)
		}
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

func TestClaudeAdapter_DriveCLI_WorkspaceUsesWorkspaceRootForMCPRegistration(t *testing.T) {
	ctx, _, _ := newScopeTestContext(t, types.SetupScopeWorkspace)
	planningRepo := t.TempDir()
	workspaceRoot := t.TempDir()
	ctx.TargetDir = planningRepo
	ctx.WorkspaceRoot = workspaceRoot
	ctx.DriveCLI = true

	stubDir := t.TempDir()
	recordFile := filepath.Join(stubDir, "claude-cwd.txt")
	var stubPath string
	if runtime.GOOS == "windows" {
		stubScript := fmt.Sprintf("@echo off\r\necho %%cd%% >> %s\r\nexit /b 1\r\n", recordFile)
		stubPath = filepath.Join(stubDir, "claude.bat")
		if err := os.WriteFile(stubPath, []byte(stubScript), 0o644); err != nil {
			t.Fatal(err)
		}
	} else {
		stubScript := fmt.Sprintf("#!/bin/sh\npwd >> %s\nexit 1\n", recordFile)
		stubPath = filepath.Join(stubDir, "claude")
		if err := os.WriteFile(stubPath, []byte(stubScript), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	aiDir := filepath.Join(workspaceRoot, ".ai")
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

	data, err := files.ReadFile(recordFile)
	if err != nil {
		t.Fatalf("expected claude CLI to be called using workspace canonical MCP: %v", err)
	}
	for _, got := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		got = strings.TrimSpace(got)
		if got != workspaceRoot {
			t.Fatalf("claude CLI working dir = %q, want workspace root %q", got, workspaceRoot)
		}
	}
}
