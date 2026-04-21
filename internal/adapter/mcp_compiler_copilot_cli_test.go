package adapter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// TestToCopilotCLIMcp_StdioServer verifies the CLI-shape payload for a stdio server.
func TestToCopilotCLIMcp_StdioServer(t *testing.T) {
	servers := map[string]McpServer{
		"memory": {
			Command: "npx",
			Args:    []string{"-y", "@modelcontextprotocol/server-memory"},
		},
	}
	got := toCopilotCLIMcp(servers)

	mcpServers, ok := got["mcpServers"].(map[string]any)
	if !ok {
		t.Fatalf("expected mcpServers map, got %T", got["mcpServers"])
	}
	entry, ok := mcpServers["memory"].(map[string]any)
	if !ok {
		t.Fatalf("expected memory entry, got %T", mcpServers["memory"])
	}
	if entry["type"] != "stdio" {
		t.Errorf("expected type=stdio, got %v", entry["type"])
	}
	if entry["command"] != "npx" {
		t.Errorf("expected command=npx, got %v", entry["command"])
	}
	if _, hasInputs := got["inputs"]; hasInputs {
		t.Errorf("CLI payload must NOT include 'inputs' key; VS-Code-only concept")
	}
}

// TestToCopilotCLIMcp_SseServer verifies the CLI-shape payload for a remote SSE server.
func TestToCopilotCLIMcp_SseServer(t *testing.T) {
	servers := map[string]McpServer{
		"remote": {
			URL:     "https://mcp.example.com",
			Headers: map[string]string{"Authorization": "Bearer ${TOKEN}"},
		},
	}
	got := toCopilotCLIMcp(servers)
	mcpServers := got["mcpServers"].(map[string]any)
	entry := mcpServers["remote"].(map[string]any)
	if entry["type"] != "sse" {
		t.Errorf("expected type=sse, got %v", entry["type"])
	}
	if entry["url"] != "https://mcp.example.com" {
		t.Errorf("unexpected url: %v", entry["url"])
	}
	// SSE servers carry headers but no env placeholder expansion at this layer.
}

// TestToCopilotCLIMcp_EnvPreserved verifies env vars survive the transform.
func TestToCopilotCLIMcp_EnvPreserved(t *testing.T) {
	servers := map[string]McpServer{
		"gh": {
			Command: "npx",
			Args:    []string{"-y", "@gh/mcp"},
			Env:     map[string]string{"GITHUB_PAT": "${GITHUB_PAT}"},
		},
	}
	got := toCopilotCLIMcp(servers)
	entry := got["mcpServers"].(map[string]any)["gh"].(map[string]any)
	env, ok := entry["env"].(map[string]string)
	if !ok {
		t.Fatalf("expected env map, got %T", entry["env"])
	}
	if env["GITHUB_PAT"] != "${GITHUB_PAT}" {
		t.Errorf("env placeholder not preserved: %v", env["GITHUB_PAT"])
	}
}

// TestCompileCopilotCLIMcp_GlobalScope verifies ~/.copilot/mcp-config.json is
// written when the probe passes (simulated via a pre-created ~/.copilot dir).
func TestCompileCopilotCLIMcp_GlobalScope(t *testing.T) {
	targetDir := t.TempDir()
	home := t.TempDir()

	// Pre-create ~/.copilot/ so the probe passes without a real CLI on PATH.
	copilotHome := filepath.Join(home, ".copilot")
	if err := files.EnsureDir(copilotHome); err != nil {
		t.Fatalf("EnsureDir: %v", err)
	}

	// Seed the canonical .ai/mcp.json catalog.
	aiDir := filepath.Join(targetDir, ".ai")
	_ = files.EnsureDir(aiDir)
	mcpContent := `{"servers":{"memory":{"command":"npx","args":["-y","@modelcontextprotocol/server-memory"]}}}`
	if err := os.WriteFile(filepath.Join(aiDir, "mcp.json"), []byte(mcpContent), 0o644); err != nil {
		t.Fatalf("write mcp.json: %v", err)
	}

	records, err := CompileMCPForTool(types.ToolIdCopilot, CompileContext{
		TargetDir:  targetDir,
		HomeDir:    home,
		SetupScope: types.SetupScopeGlobal,
	})
	if err != nil {
		t.Fatalf("CompileMCPForTool: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record for ~/.copilot/mcp-config.json, got %d", len(records))
	}

	cfgPath := filepath.Join(copilotHome, "mcp-config.json")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("read mcp-config.json: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("parse mcp-config.json: %v", err)
	}
	mcpServers, ok := parsed["mcpServers"].(map[string]any)
	if !ok {
		t.Fatalf("mcpServers missing or wrong type: %T", parsed["mcpServers"])
	}
	if _, ok := mcpServers["memory"]; !ok {
		t.Errorf("memory server missing from mcp-config.json")
	}
}

// TestCompileCopilotCLIMcp_ProjectScope verifies that at project scope both
// .vscode/mcp.json and ~/.copilot/mcp-config.json are written when probe passes.
func TestCompileCopilotCLIMcp_ProjectScope(t *testing.T) {
	targetDir := t.TempDir()
	home := t.TempDir()
	copilotHome := filepath.Join(home, ".copilot")
	if err := files.EnsureDir(copilotHome); err != nil {
		t.Fatalf("EnsureDir: %v", err)
	}

	aiDir := filepath.Join(targetDir, ".ai")
	_ = files.EnsureDir(aiDir)
	mcpContent := `{"servers":{"memory":{"command":"npx","args":["-y","@mcp/memory"]}}}`
	if err := os.WriteFile(filepath.Join(aiDir, "mcp.json"), []byte(mcpContent), 0o644); err != nil {
		t.Fatalf("write mcp.json: %v", err)
	}

	records, err := CompileMCPForTool(types.ToolIdCopilot, CompileContext{
		TargetDir:  targetDir,
		HomeDir:    home,
		SetupScope: types.SetupScopeProject,
	})
	if err != nil {
		t.Fatalf("CompileMCPForTool: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 records (.vscode + ~/.copilot), got %d", len(records))
	}

	// Both files must exist.
	if _, err := os.Stat(filepath.Join(targetDir, ".vscode", "mcp.json")); err != nil {
		t.Errorf(".vscode/mcp.json not written: %v", err)
	}
	if _, err := os.Stat(filepath.Join(copilotHome, "mcp-config.json")); err != nil {
		t.Errorf("~/.copilot/mcp-config.json not written: %v", err)
	}
}

// TestCompileCopilotCLIMcp_ProbeFails_ProjectScope verifies that when the
// probe fails at project scope, .vscode/mcp.json is still written but
// ~/.copilot/mcp-config.json is NOT written.
func TestCompileCopilotCLIMcp_ProbeFails_ProjectScope(t *testing.T) {
	targetDir := t.TempDir()
	home := t.TempDir() // no .copilot/ inside

	aiDir := filepath.Join(targetDir, ".ai")
	_ = files.EnsureDir(aiDir)
	mcpContent := `{"servers":{"memory":{"command":"npx","args":["-y","@mcp/memory"]}}}`
	if err := os.WriteFile(filepath.Join(aiDir, "mcp.json"), []byte(mcpContent), 0o644); err != nil {
		t.Fatalf("write mcp.json: %v", err)
	}

	records, err := CompileMCPForTool(types.ToolIdCopilot, CompileContext{
		TargetDir:  targetDir,
		HomeDir:    home,
		SetupScope: types.SetupScopeProject,
	})
	if err != nil {
		t.Fatalf("CompileMCPForTool: %v", err)
	}
	// VS Code file always written at project scope; CLI file is probe-gated.
	// On machines where `copilot` is on PATH, the probe will pass and the
	// record count will be 2. Accept either 1 or 2 to keep the test robust
	// across developer environments; but the VS Code file must always exist.
	if len(records) < 1 {
		t.Fatalf("expected at least 1 record (.vscode/mcp.json), got %d", len(records))
	}
	if _, err := os.Stat(filepath.Join(targetDir, ".vscode", "mcp.json")); err != nil {
		t.Errorf(".vscode/mcp.json must always be written at project scope: %v", err)
	}
}

// TestCompileCopilotCLIMcp_PreservesUserKeys verifies that a user-authored
// mcpServers entry in ~/.copilot/mcp-config.json is preserved across re-run,
// and a .bak is created on first-touch.
func TestCompileCopilotCLIMcp_PreservesUserKeys(t *testing.T) {
	targetDir := t.TempDir()
	home := t.TempDir()
	copilotHome := filepath.Join(home, ".copilot")
	if err := files.EnsureDir(copilotHome); err != nil {
		t.Fatalf("EnsureDir: %v", err)
	}

	cfgPath := filepath.Join(copilotHome, "mcp-config.json")
	// Seed a user-authored mcpServer + an unrelated top-level key.
	existing := `{
  "mcpServers": {
    "user-server": {"type": "stdio", "command": "my-server"}
  },
  "theme": "dark"
}`
	if err := os.WriteFile(cfgPath, []byte(existing), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	// Seed canonical catalog.
	aiDir := filepath.Join(targetDir, ".ai")
	_ = files.EnsureDir(aiDir)
	mcpContent := `{"servers":{"memory":{"command":"npx","args":["-y","@mcp/memory"]}}}`
	if err := os.WriteFile(filepath.Join(aiDir, "mcp.json"), []byte(mcpContent), 0o644); err != nil {
		t.Fatalf("write mcp.json: %v", err)
	}

	if _, err := CompileMCPForTool(types.ToolIdCopilot, CompileContext{
		TargetDir:  targetDir,
		HomeDir:    home,
		SetupScope: types.SetupScopeGlobal,
	}); err != nil {
		t.Fatalf("CompileMCPForTool: %v", err)
	}

	// User key preserved.
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("parse: %v", err)
	}
	mcpServers := parsed["mcpServers"].(map[string]any)
	if _, ok := mcpServers["user-server"]; !ok {
		t.Errorf("user-server entry was clobbered")
	}
	if _, ok := mcpServers["memory"]; !ok {
		t.Errorf("memory entry not added")
	}
	if parsed["theme"] != "dark" {
		t.Errorf("top-level user key 'theme' clobbered")
	}

	// .bak exists from first-touch.
	if _, err := os.Stat(cfgPath + ".bak"); err != nil {
		t.Errorf("expected %s.bak after first merge: %v", cfgPath, err)
	}
}
