package adapter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// TestCompileOpenCodeMCP_ProjectScope verifies that in project scope, the MCP
// config is written to .opencode/opencode.jsonc.
func TestCompileOpenCodeMCP_ProjectScope(t *testing.T) {
	targetDir := t.TempDir()

	// Create the canonical .ai/mcp.json.
	aiDir := filepath.Join(targetDir, ".ai")
	_ = files.EnsureDir(aiDir)
	mcpContent := `{"servers":{"memory":{"command":"npx","args":["-y","@modelcontextprotocol/server-memory"]}}}`
	if err := os.WriteFile(filepath.Join(aiDir, "mcp.json"), []byte(mcpContent), 0o644); err != nil {
		t.Fatalf("failed to write mcp.json: %v", err)
	}

	// Call CompileMCPForTool for OpenCode.
	var fileRecords []types.TrackedFile
	records, err := CompileMCPForTool(types.ToolIdOpenCode, targetDir, fileRecords)
	if err != nil {
		t.Fatalf("CompileMCPForTool failed: %v", err)
	}

	// Verify the file was written to .opencode/opencode.jsonc.
	configPath := filepath.Join(targetDir, ".opencode", "opencode.jsonc")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("expected .opencode/opencode.jsonc was not created")
	}

	// Verify the tracked file record has the correct path.
	if len(records) != 1 {
		t.Fatalf("expected 1 tracked file record, got %d", len(records))
	}
	if records[0].Path != ".opencode/opencode.jsonc" {
		t.Errorf("expected path '.opencode/opencode.jsonc', got %q", records[0].Path)
	}
}

// TestCompileOpenCodeMCP_GlobalScope verifies that in global scope, the MCP
// config is written directly to opencode.jsonc (no .opencode subdirectory).
func TestCompileOpenCodeMCP_GlobalScope(t *testing.T) {
	// Create a directory that mimics ~/.config/opencode.
	globalDir := filepath.Join(t.TempDir(), ".config", "opencode")
	_ = files.EnsureDir(globalDir)

	// For global scope, the canonical .ai/mcp.json is still read from a project dir.
	// But compileOpenCodeMCP receives the targetDir where it should write.
	// The function reads from targetDir/.ai/mcp.json — we need to create that too.
	aiDir := filepath.Join(globalDir, ".ai")
	_ = files.EnsureDir(aiDir)
	mcpContent := `{"servers":{"memory":{"command":"npx","args":["-y","@modelcontextprotocol/server-memory"]}}}`
	if err := os.WriteFile(filepath.Join(aiDir, "mcp.json"), []byte(mcpContent), 0o644); err != nil {
		t.Fatalf("failed to write mcp.json: %v", err)
	}

	// Call CompileMCPForTool for OpenCode with the global dir as target.
	var fileRecords []types.TrackedFile
	records, err := CompileMCPForTool(types.ToolIdOpenCode, globalDir, fileRecords)
	if err != nil {
		t.Fatalf("CompileMCPForTool failed: %v", err)
	}

	// Verify the file was written to opencode.jsonc (not .opencode/opencode.jsonc).
	configPath := filepath.Join(globalDir, "opencode.jsonc")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("expected opencode.jsonc was not created in global dir")
	}

	// Verify no .opencode subdirectory was created.
	subDir := filepath.Join(globalDir, ".opencode")
	if _, err := os.Stat(subDir); !os.IsNotExist(err) {
		t.Error("global scope should not create .opencode subdirectory")
	}

	// Verify the tracked file record has the correct path.
	if len(records) != 1 {
		t.Fatalf("expected 1 tracked file record, got %d", len(records))
	}
	if records[0].Path != "opencode.jsonc" {
		t.Errorf("expected path 'opencode.jsonc', got %q", records[0].Path)
	}
}
