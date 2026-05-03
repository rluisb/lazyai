package adapter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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

// TestCompileOpenCodeMCP_PreservesUserAuthoredServer verifies that a server
// the user hand-added to opencode.jsonc (not present in .ai/mcp.json) is
// preserved across an ai-setup compile cycle. This is the core per-server
// deep-merge behavior introduced in spec 011 phase 2.
func TestCompileOpenCodeMCP_PreservesUserAuthoredServer(t *testing.T) {
	targetDir := t.TempDir()

	aiDir := filepath.Join(targetDir, ".ai")
	_ = files.EnsureDir(aiDir)
	mcpContent := `{"servers":{"memory":{"command":"npx","args":["-y","@modelcontextprotocol/server-memory"]}}}`
	if err := os.WriteFile(filepath.Join(aiDir, "mcp.json"), []byte(mcpContent), 0o644); err != nil {
		t.Fatalf("write mcp.json: %v", err)
	}

	// Pre-seed opencode.jsonc with a user-authored MCP entry under a name
	// that ai-setup does not manage.
	ocDir := filepath.Join(targetDir, ".opencode")
	_ = files.EnsureDir(ocDir)
	preExisting := `{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "userAuthored": { "type": "local", "command": ["custom-cli"], "enabled": true }
  }
}`
	if err := os.WriteFile(filepath.Join(ocDir, "opencode.jsonc"), []byte(preExisting), 0o644); err != nil {
		t.Fatalf("seed opencode.jsonc: %v", err)
	}

	if _, err := CompileMCPForTool(types.ToolIdOpenCode, CompileContext{
		TargetDir:  targetDir,
		SetupScope: types.SetupScopeProject,
	}); err != nil {
		t.Fatalf("CompileMCPForTool: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(ocDir, "opencode.jsonc"))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	contents := string(data)
	if !strings.Contains(contents, `"userAuthored"`) {
		t.Errorf("user-authored server was lost on compile:\n%s", contents)
	}
	if !strings.Contains(contents, `"memory"`) {
		t.Errorf("managed server missing after merge:\n%s", contents)
	}
	if !strings.Contains(contents, `"custom-cli"`) {
		t.Errorf("user-authored server's command was rewritten:\n%s", contents)
	}
}

// TestCompileOpenCodeMCP_ManagedWinsOnNameCollision documents the
// collision rule: if the user hand-authored a server under a name that is
// also managed by ai-setup, the managed entry wins.
func TestCompileOpenCodeMCP_ManagedWinsOnNameCollision(t *testing.T) {
	targetDir := t.TempDir()

	aiDir := filepath.Join(targetDir, ".ai")
	_ = files.EnsureDir(aiDir)
	mcpContent := `{"servers":{"memory":{"command":"npx","args":["-y","@modelcontextprotocol/server-memory"]}}}`
	if err := os.WriteFile(filepath.Join(aiDir, "mcp.json"), []byte(mcpContent), 0o644); err != nil {
		t.Fatal(err)
	}

	ocDir := filepath.Join(targetDir, ".opencode")
	_ = files.EnsureDir(ocDir)
	preExisting := `{"mcp":{"memory":{"type":"local","command":["user-override"]}}}`
	if err := os.WriteFile(filepath.Join(ocDir, "opencode.jsonc"), []byte(preExisting), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := CompileMCPForTool(types.ToolIdOpenCode, CompileContext{
		TargetDir:  targetDir,
		SetupScope: types.SetupScopeProject,
	}); err != nil {
		t.Fatalf("CompileMCPForTool: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(ocDir, "opencode.jsonc"))
	if strings.Contains(string(data), "user-override") {
		t.Errorf("user-override survived; managed entry should have won:\n%s", data)
	}
	if !strings.Contains(string(data), "@modelcontextprotocol/server-memory") {
		t.Errorf("managed `memory` entry not present after collision:\n%s", data)
	}
}

func TestCompileOpenCodeMCP_WiresManagedOrchestratorServer(t *testing.T) {
	targetDir := t.TempDir()
	aiDir := filepath.Join(targetDir, ".ai")
	_ = files.EnsureDir(aiDir)
	orchestratorEntry := `{
  "servers": {
    "orchestrator": {
      "command": "/tmp/ai-setup-orchestrator",
      "args": ["connect", "--project", "/tmp/project"]
    }
  }
}`
	if err := os.WriteFile(filepath.Join(aiDir, "mcp.json"), []byte(orchestratorEntry), 0o644); err != nil {
		t.Fatalf("write mcp.json: %v", err)
	}

	if _, err := CompileMCPForTool(types.ToolIdOpenCode, CompileContext{TargetDir: targetDir, SetupScope: types.SetupScopeProject}); err != nil {
		t.Fatalf("CompileMCPForTool: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(targetDir, ".opencode", "opencode.jsonc"))
	if err != nil {
		t.Fatalf("read opencode.jsonc: %v", err)
	}
	contents := string(data)
	if !strings.Contains(contents, `"orchestrator"`) {
		t.Fatalf("compiled config missing orchestrator entry:\n%s", contents)
	}
	if !strings.Contains(contents, `/tmp/ai-setup-orchestrator`) || !strings.Contains(contents, `connect`) || !strings.Contains(contents, `--project`) {
		t.Fatalf("compiled config missing orchestrator command/args:\n%s", contents)
	}
}

func TestCompileOpenCodeMCP_PreservesLegacyOrchestratorCommand(t *testing.T) {
	targetDir := t.TempDir()
	aiDir := filepath.Join(targetDir, ".ai")
	_ = files.EnsureDir(aiDir)
	orchestratorEntry := `{
  "servers": {
    "orchestrator": {
      "command": "npx",
      "args": ["-y", "@ai-setup/orchestrator"]
    },
    "orchestrator-node": {
      "command": "/tmp/fake-node",
      "args": ["/tmp/orchestrator/dist/index.js"]
    }
  }
}`
	if err := os.WriteFile(filepath.Join(aiDir, "mcp.json"), []byte(orchestratorEntry), 0o644); err != nil {
		t.Fatalf("write mcp.json: %v", err)
	}

	if _, err := CompileMCPForTool(types.ToolIdOpenCode, CompileContext{TargetDir: targetDir, SetupScope: types.SetupScopeProject}); err != nil {
		t.Fatalf("CompileMCPForTool: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(targetDir, ".opencode", "opencode.jsonc"))
	if err != nil {
		t.Fatalf("read opencode.jsonc: %v", err)
	}
	contents := string(data)
	if !strings.Contains(contents, `"npx"`) || !strings.Contains(contents, `@ai-setup/orchestrator`) {
		t.Fatalf("compiled config did not preserve legacy npx orchestrator command:\n%s", contents)
	}
	if !strings.Contains(contents, `/tmp/fake-node`) || !strings.Contains(contents, `/tmp/orchestrator/dist/index.js`) {
		t.Fatalf("compiled config did not preserve legacy node orchestrator command:\n%s", contents)
	}
}

func TestMCPCompilerPreservesGoOrchestratorCommandArgsForToolPayloads(t *testing.T) {
	servers := map[string]McpServer{
		"orchestrator": {
			Command: "/tmp/ai-setup-orchestrator",
			Args:    []string{"connect", "--project", "/tmp/project"},
		},
	}

	for name, payload := range map[string]any{
		"opencode": toOpenCodeMcp(servers),
		"claude":   toClaudeCodeMcpInner(servers),
		"copilot": func() any {
			entries, _ := toCopilotServerEntries(servers)
			return entries
		}(),
	} {
		encoded, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal %s payload: %v", name, err)
		}
		contents := string(encoded)
		if !strings.Contains(contents, `/tmp/ai-setup-orchestrator`) || !strings.Contains(contents, `connect`) || !strings.Contains(contents, `--project`) || !strings.Contains(contents, `/tmp/project`) {
			t.Fatalf("%s payload did not preserve Go orchestrator command/args: %s", name, contents)
		}
	}
}

// TestCompileMCPForTool_CopilotGlobalSkips verifies that Copilot × global scope
// is a clean no-op (no error, no records).
func TestCompileMCPForTool_CopilotGlobalSkips(t *testing.T) {
	targetDir := t.TempDir()
	homeDir := t.TempDir()
	t.Setenv("PATH", t.TempDir())

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
