package adapter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// TestCompileOpenCodeMCP_ProjectScope verifies that at project scope, the MCP
// config is written to <targetDir>/.opencode/opencode.jsonc.
func TestCompileOpenCodeMCP_ProjectScope(t *testing.T) {
	targetDir := t.TempDir()

	aiDir := filepath.Join(targetDir, ".ai")
	_ = files.EnsureDir(aiDir)
	mcpContent := `{"servers":{"memory":{"command":"npx","args":["-y","@modelcontextprotocol/server-memory"]}}}`
	if err := os.WriteFile(filepath.Join(aiDir, "mcp.json"), []byte(mcpContent), 0o644); err != nil {
		t.Fatalf("failed to write mcp.json: %v", err)
	}

	records, err := CompileMCPForTool(types.ToolIdOpenCode, CompileContext{
		TargetDir:  targetDir,
		SetupScope: types.SetupScopeProject,
	})
	if err != nil {
		t.Fatalf("CompileMCPForTool failed: %v", err)
	}

	configPath := filepath.Join(targetDir, ".opencode", "opencode.jsonc")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("expected .opencode/opencode.jsonc was not created")
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 tracked file record, got %d", len(records))
	}
}

// TestCompileOpenCodeMCP_GlobalScope verifies that at global scope, the MCP
// config is written to <home>/.config/opencode/opencode.jsonc (via
// ResolveToolRoot), not <targetDir>/.opencode/.
func TestCompileOpenCodeMCP_GlobalScope(t *testing.T) {
	targetDir := t.TempDir()
	homeDir := t.TempDir()

	aiDir := filepath.Join(targetDir, ".ai")
	_ = files.EnsureDir(aiDir)
	mcpContent := `{"servers":{"memory":{"command":"npx","args":["-y","@modelcontextprotocol/server-memory"]}}}`
	if err := os.WriteFile(filepath.Join(aiDir, "mcp.json"), []byte(mcpContent), 0o644); err != nil {
		t.Fatalf("failed to write mcp.json: %v", err)
	}

	records, err := CompileMCPForTool(types.ToolIdOpenCode, CompileContext{
		TargetDir:  targetDir,
		HomeDir:    homeDir,
		SetupScope: types.SetupScopeGlobal,
	})
	if err != nil {
		t.Fatalf("CompileMCPForTool failed: %v", err)
	}

	// Global scope: opencode.jsonc should live under <home>/.config/opencode.
	globalConfigPath := filepath.Join(homeDir, ".config", "opencode", "opencode.jsonc")
	if _, err := os.Stat(globalConfigPath); os.IsNotExist(err) {
		t.Fatalf("expected opencode.jsonc at %q, not found", globalConfigPath)
	}

	// Must not write under the project target dir.
	projectSubdir := filepath.Join(targetDir, ".opencode")
	if _, err := os.Stat(projectSubdir); !os.IsNotExist(err) {
		t.Errorf("global scope must not create %q", projectSubdir)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 tracked file record, got %d", len(records))
	}
}

// TestCompileMCPForTool_CopilotGlobalSkips verifies that Copilot × global scope
// is a clean no-op (no error, no records).
func TestCompileMCPForTool_CopilotGlobalSkips(t *testing.T) {
	targetDir := t.TempDir()
	homeDir := t.TempDir()

	aiDir := filepath.Join(targetDir, ".ai")
	_ = files.EnsureDir(aiDir)
	mcpContent := `{"servers":{"memory":{"command":"npx","args":["-y"]}}}`
	if err := os.WriteFile(filepath.Join(aiDir, "mcp.json"), []byte(mcpContent), 0o644); err != nil {
		t.Fatal(err)
	}

	records, err := CompileMCPForTool(types.ToolIdCopilot, CompileContext{
		TargetDir:  targetDir,
		HomeDir:    homeDir,
		SetupScope: types.SetupScopeGlobal,
	})
	if err != nil {
		t.Fatalf("Copilot × global must not error, got: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("Copilot × global must produce 0 records, got %d", len(records))
	}
}

// TestCompileMCPForTool_ClaudeGlobalSkips verifies that Claude Code × global
// skips compile (init's settings.json merge already handles mcpServers).
func TestCompileMCPForTool_ClaudeGlobalSkips(t *testing.T) {
	targetDir := t.TempDir()
	homeDir := t.TempDir()

	aiDir := filepath.Join(targetDir, ".ai")
	_ = files.EnsureDir(aiDir)
	mcpContent := `{"servers":{"memory":{"command":"npx"}}}`
	if err := os.WriteFile(filepath.Join(aiDir, "mcp.json"), []byte(mcpContent), 0o644); err != nil {
		t.Fatal(err)
	}

	records, err := CompileMCPForTool(types.ToolIdClaudeCode, CompileContext{
		TargetDir:  targetDir,
		HomeDir:    homeDir,
		SetupScope: types.SetupScopeGlobal,
	})
	if err != nil {
		t.Fatalf("Claude × global must not error, got: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("Claude × global must produce 0 records; init handles settings.json")
	}
	// No .mcp.json should have been written.
	if files.FileExists(filepath.Join(targetDir, ".mcp.json")) {
		t.Error("Claude × global must not write project-scope .mcp.json")
	}
}
