package adapter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// setupCanonicalMcp writes a minimal .ai/mcp.json to targetDir so compile has something to read.
func setupCanonicalMcp(t *testing.T, targetDir string) {
	t.Helper()
	aiDir := filepath.Join(targetDir, ".ai")
	if err := files.EnsureDir(aiDir); err != nil {
		t.Fatalf("EnsureDir: %v", err)
	}
	mcpContent := `{"servers":{"memory":{"command":"npx","args":["-y","@mcp/memory"],"env":{"KEY":"${SECRET}"}}}}`
	if err := os.WriteFile(filepath.Join(aiDir, "mcp.json"), []byte(mcpContent), 0o644); err != nil {
		t.Fatalf("write mcp.json: %v", err)
	}
}

// TestCompileClaudeCodeMCP_LocalSecrets_Project writes to .claude/settings.local.json
// instead of .mcp.json when the flag is set.
func TestCompileClaudeCodeMCP_LocalSecrets_Project(t *testing.T) {
	targetDir := t.TempDir()
	setupCanonicalMcp(t, targetDir)

	records, err := CompileMCPForTool(types.ToolIdClaudeCode, CompileContext{
		TargetDir:    targetDir,
		SetupScope:   types.SetupScopeProject,
		LocalSecrets: true,
	})
	if err != nil {
		t.Fatalf("CompileMCPForTool: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	// Committed .mcp.json must NOT be written when local-secrets is set.
	if _, err := os.Stat(filepath.Join(targetDir, ".mcp.json")); err == nil {
		t.Errorf(".mcp.json must NOT be written when --local-secrets is set")
	}

	// settings.local.json must contain mcpServers.
	settingsLocal := filepath.Join(targetDir, ".claude", "settings.local.json")
	data, err := os.ReadFile(settingsLocal)
	if err != nil {
		t.Fatalf("read settings.local.json: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("parse: %v", err)
	}
	mcpServers, ok := parsed["mcpServers"].(map[string]any)
	if !ok {
		t.Fatalf("mcpServers missing or wrong type: %T", parsed["mcpServers"])
	}
	if _, ok := mcpServers["memory"]; !ok {
		t.Errorf("memory server missing")
	}
}

// TestCompileClaudeCodeMCP_LocalSecrets_Global writes to ~/.claude/settings.local.json.
func TestCompileClaudeCodeMCP_LocalSecrets_Global(t *testing.T) {
	targetDir := t.TempDir()
	home := t.TempDir()
	setupCanonicalMcp(t, targetDir)

	records, err := CompileMCPForTool(types.ToolIdClaudeCode, CompileContext{
		TargetDir:    targetDir,
		HomeDir:      home,
		SetupScope:   types.SetupScopeGlobal,
		LocalSecrets: true,
	})
	if err != nil {
		t.Fatalf("CompileMCPForTool: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	path := filepath.Join(home, ".claude", "settings.local.json")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("~/.claude/settings.local.json not written at global scope: %v", err)
	}
}

// TestCompileClaudeCodeMCP_WithoutLocalSecrets_UnchangedProject retains .mcp.json default path.
func TestCompileClaudeCodeMCP_WithoutLocalSecrets_UnchangedProject(t *testing.T) {
	targetDir := t.TempDir()
	setupCanonicalMcp(t, targetDir)

	records, err := CompileMCPForTool(types.ToolIdClaudeCode, CompileContext{
		TargetDir:  targetDir,
		SetupScope: types.SetupScopeProject,
		// LocalSecrets: false (default)
	})
	if err != nil {
		t.Fatalf("CompileMCPForTool: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	// Default path: .mcp.json written.
	if _, err := os.Stat(filepath.Join(targetDir, ".mcp.json")); err != nil {
		t.Errorf("default path must still write .mcp.json: %v", err)
	}
	// settings.local.json must NOT be touched.
	if _, err := os.Stat(filepath.Join(targetDir, ".claude", "settings.local.json")); err == nil {
		t.Errorf("settings.local.json must NOT be written without --local-secrets")
	}
}

// TestCompileClaudeCodeMCP_LocalSecrets_PreservesUserKeys verifies deep-merge
// preserves user-authored keys in .claude/settings.local.json.
func TestCompileClaudeCodeMCP_LocalSecrets_PreservesUserKeys(t *testing.T) {
	targetDir := t.TempDir()
	setupCanonicalMcp(t, targetDir)

	// Seed a pre-existing settings.local.json with user-authored keys.
	settingsLocal := filepath.Join(targetDir, ".claude", "settings.local.json")
	_ = files.EnsureDir(filepath.Dir(settingsLocal))
	existing := `{
  "mcpServers": {
    "user-server": {"command": "user-mcp"}
  },
  "permissions": {
    "allow": ["custom-tool"]
  }
}`
	if err := os.WriteFile(settingsLocal, []byte(existing), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	if _, err := CompileMCPForTool(types.ToolIdClaudeCode, CompileContext{
		TargetDir:    targetDir,
		SetupScope:   types.SetupScopeProject,
		LocalSecrets: true,
	}); err != nil {
		t.Fatalf("CompileMCPForTool: %v", err)
	}

	data, err := os.ReadFile(settingsLocal)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("parse: %v", err)
	}

	mcpServers := parsed["mcpServers"].(map[string]any)
	if _, ok := mcpServers["user-server"]; !ok {
		t.Errorf("user-server clobbered")
	}
	if _, ok := mcpServers["memory"]; !ok {
		t.Errorf("memory not added")
	}
	perms, ok := parsed["permissions"].(map[string]any)
	if !ok {
		t.Fatalf("permissions missing: %T", parsed["permissions"])
	}
	allow, ok := perms["allow"].([]any)
	if !ok || len(allow) == 0 {
		t.Errorf("permissions.allow clobbered: %v", perms["allow"])
	}

	// .bak from first-touch.
	if _, err := os.Stat(settingsLocal + ".bak"); err != nil {
		t.Errorf("expected .bak after first merge: %v", err)
	}
}
