package adapter

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/library"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// TestCompileOpenCodeMCP_ProjectScope verifies that at project scope, the MCP
// config is written to <targetDir>/opencode.json.
func TestCompileOpenCodeMCP_ProjectScope(t *testing.T) {
	targetDir := t.TempDir()

	aiDir := filepath.Join(targetDir, ".ai")
	_ = files.EnsureDir(aiDir)
	mcpContent := `{"servers":{"filesystem":{"command":"npx","args":["-y","@modelcontextprotocol/server-filesystem","."]}}}`
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

	configPath := filepath.Join(targetDir, OpenCodeConfigFilename)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("expected opencode.json was not created")
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 tracked file record, got %d", len(records))
	}
}

// TestCompileOpenCodeMCP_GlobalScope verifies that at global scope, the MCP
// config is written to <home>/.config/opencode/opencode.json (via
// ResolveToolRoot), not <targetDir>/.opencode/.
func TestCompileOpenCodeMCP_GlobalScope(t *testing.T) {
	targetDir := t.TempDir()
	homeDir := t.TempDir()

	aiDir := filepath.Join(targetDir, ".ai")
	_ = files.EnsureDir(aiDir)
	mcpContent := `{"servers":{"filesystem":{"command":"npx","args":["-y","@modelcontextprotocol/server-filesystem","."]}}}`
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

	// Global scope: opencode.json should live under <home>/.config/opencode.
	globalConfigPath := filepath.Join(homeDir, ".config", "opencode", OpenCodeConfigFilename)
	if _, err := os.Stat(globalConfigPath); os.IsNotExist(err) {
		t.Fatalf("expected opencode.json at %q, not found", globalConfigPath)
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
// user-authored to legacy `.opencode/lazyai.mcp.jsonc` (not present in
// `.ai/mcp.json`) is preserved across an `ai-setup` compile cycle.
func TestCompileOpenCodeMCP_PreservesUserAuthoredServer(t *testing.T) {
	targetDir := t.TempDir()

	aiDir := filepath.Join(targetDir, ".ai")
	_ = files.EnsureDir(aiDir)
	mcpContent := `{"servers":{"filesystem":{"command":"npx","args":["-y","@modelcontextprotocol/server-filesystem","."]}}}`
	if err := os.WriteFile(filepath.Join(aiDir, "mcp.json"), []byte(mcpContent), 0o644); err != nil {
		t.Fatalf("write mcp.json: %v", err)
	}

	// Pre-seed old lazyai.mcp.jsonc with a user-authored MCP entry under a name
	// that ai-setup does not manage.
	ocDir := filepath.Join(targetDir, ".opencode")
	_ = files.EnsureDir(ocDir)
	preExisting := `{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "userAuthored": { "type": "local", "command": ["custom-cli"], "enabled": true }
  }
}`
	if err := os.WriteFile(filepath.Join(ocDir, "lazyai.mcp.jsonc"), []byte(preExisting), 0o644); err != nil {
		t.Fatalf("seed lazyai.mcp.jsonc: %v", err)
	}

	if _, err := CompileMCPForTool(types.ToolIdOpenCode, CompileContext{
		TargetDir:  targetDir,
		SetupScope: types.SetupScopeProject,
	}); err != nil {
		t.Fatalf("CompileMCPForTool: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(targetDir, OpenCodeConfigFilename))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	contents := string(data)
	if !strings.Contains(contents, `"userAuthored"`) {
		t.Errorf("user-authored server was lost on compile:\n%s", contents)
	}
	if !strings.Contains(contents, `"filesystem"`) {
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
	mcpContent := `{"servers":{"filesystem":{"command":"npx","args":["-y","@modelcontextprotocol/server-filesystem","."]}}}`
	if err := os.WriteFile(filepath.Join(aiDir, "mcp.json"), []byte(mcpContent), 0o644); err != nil {
		t.Fatal(err)
	}

	preExisting := `{
  "mcp": {
    "filesystem": { "type": "local", "command": ["user-override"] }
  }
}`
	if err := os.WriteFile(filepath.Join(targetDir, OpenCodeConfigFilename), []byte(preExisting), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := CompileMCPForTool(types.ToolIdOpenCode, CompileContext{
		TargetDir:  targetDir,
		SetupScope: types.SetupScopeProject,
	}); err != nil {
		t.Fatalf("CompileMCPForTool: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(targetDir, OpenCodeConfigFilename))
	if strings.Contains(string(data), "user-override") {
		t.Errorf("user-override survived; managed entry should have won:\n%s", data)
	}
	if !strings.Contains(string(data), "@modelcontextprotocol/server-filesystem") {
		t.Errorf("managed `filesystem` entry not present after collision:\n%s", data)
	}
}

func TestCompileOpenCodeMCP_PreservesUserProvidedLegacyOrchestratorServer(t *testing.T) {
	targetDir := t.TempDir()
	aiDir := filepath.Join(targetDir, ".ai")
	_ = files.EnsureDir(aiDir)
	orchestratorEntry := `{
  "servers": {
    "orchestrator": {
      "command": "/placeholder/lazyai-orchestrator",
      "args": ["connect", "--project", "/placeholder/project"]
    }
  }
}`
	if err := os.WriteFile(filepath.Join(aiDir, "mcp.json"), []byte(orchestratorEntry), 0o644); err != nil {
		t.Fatalf("write mcp.json: %v", err)
	}

	if _, err := CompileMCPForTool(types.ToolIdOpenCode, CompileContext{TargetDir: targetDir, SetupScope: types.SetupScopeProject}); err != nil {
		t.Fatalf("CompileMCPForTool: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(targetDir, OpenCodeConfigFilename))
	if err != nil {
		t.Fatalf("read opencode.json: %v", err)
	}
	contents := string(data)
	if !strings.Contains(contents, `"orchestrator"`) {
		t.Fatalf("compiled config did not preserve user-provided legacy orchestrator entry:\n%s", contents)
	}
	if !strings.Contains(contents, `/placeholder/lazyai-orchestrator`) || !strings.Contains(contents, `connect`) || !strings.Contains(contents, `--project`) {
		t.Fatalf("compiled config did not preserve user-provided legacy orchestrator command/args:\n%s", contents)
	}
	if strings.Contains(contents, `--execution-mode`) || strings.Contains(contents, `a2a`) {
		t.Fatalf("compiled config should not add retired orchestrator execution-mode defaults:\n%s", contents)
	}
}

func TestCompileOpenCodeMCP_PreservesUserProvidedLegacyOrchestratorCommands(t *testing.T) {
	targetDir := t.TempDir()
	aiDir := filepath.Join(targetDir, ".ai")
	_ = files.EnsureDir(aiDir)
	orchestratorEntry := `{
  "servers": {
    "orchestrator": {
      "command": "lazyai-orchestrator",
      "args": ["connect", "--project", "/placeholder/project"]
    },
    "orchestrator-absolute": {
      "command": "/placeholder/lazyai-orchestrator",
      "args": ["connect"]
    }
  }
}`
	if err := os.WriteFile(filepath.Join(aiDir, "mcp.json"), []byte(orchestratorEntry), 0o644); err != nil {
		t.Fatalf("write mcp.json: %v", err)
	}

	if _, err := CompileMCPForTool(types.ToolIdOpenCode, CompileContext{TargetDir: targetDir, SetupScope: types.SetupScopeProject}); err != nil {
		t.Fatalf("CompileMCPForTool: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(targetDir, OpenCodeConfigFilename))
	if err != nil {
		t.Fatalf("read opencode.json: %v", err)
	}
	contents := string(data)
	if !strings.Contains(contents, `"lazyai-orchestrator"`) || !strings.Contains(contents, `connect`) || !strings.Contains(contents, `--project`) {
		t.Fatalf("compiled config did not preserve user-provided legacy orchestrator command:\n%s", contents)
	}
	if !strings.Contains(contents, `/placeholder/lazyai-orchestrator`) {
		t.Fatalf("compiled config did not preserve user-provided absolute legacy orchestrator command:\n%s", contents)
	}
}

func TestMCPCompilerPreservesUserProvidedLegacyOrchestratorCommandArgsForToolPayloads(t *testing.T) {
	servers := map[string]McpServer{
		"orchestrator": {
			Command: "/placeholder/lazyai-orchestrator",
			Args:    []string{"connect", "--project", "/placeholder/project"},
		},
	}

	for name, payload := range map[string]any{
		"opencode": toOpenCodeMcp(servers),
		"claude":   toClaudeCodeMcpInner(servers),
		"copilot": func() any {
			entries, _ := toCopilotServerEntries(servers, "sse")
			return entries
		}(),
	} {
		encoded, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal %s payload: %v", name, err)
		}
		contents := string(encoded)
		if !strings.Contains(contents, `/placeholder/lazyai-orchestrator`) || !strings.Contains(contents, `connect`) || !strings.Contains(contents, `--project`) || !strings.Contains(contents, `/placeholder/project`) {
			t.Fatalf("%s payload did not preserve user-provided legacy orchestrator command/args: %s", name, contents)
		}
		if strings.Contains(contents, `--execution-mode`) || strings.Contains(contents, `a2a`) {
			t.Fatalf("%s payload should not add retired orchestrator execution-mode defaults: %s", name, contents)
		}
	}
}

// TestToClaudeCodeMcpInner_RemoteHasType verifies remote MCP entries carry type: http.
func TestToClaudeCodeMcpInner_RemoteHasType(t *testing.T) {
	servers := map[string]McpServer{
		"ai-memory": {
			URL:     "http://127.0.0.1:49374/mcp",
			Headers: map[string]string{"Authorization": "Bearer token"},
		},
	}

	emitted := toClaudeCodeMcpInner(servers)
	entry, ok := emitted["ai-memory"].(map[string]any)
	if !ok {
		t.Fatalf("emitted ai-memory entry is not a map: %T", emitted["ai-memory"])
	}
	if entry["type"] != "http" {
		t.Fatalf("emitted type = %v, want http", entry["type"])
	}
	if entry["url"] != "http://127.0.0.1:49374/mcp" {
		t.Fatalf("emitted url = %v, want http://127.0.0.1:49374/mcp", entry["url"])
	}
	headers, ok := entry["headers"].(map[string]string)
	if !ok {
		t.Fatalf("emitted headers is not a map: %T", entry["headers"])
	}
	if headers["Authorization"] != "Bearer token" {
		t.Fatalf("emitted Authorization header = %v, want Bearer token", headers["Authorization"])
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
	mcpContent := `{"servers":{"filesystem":{"command":"npx","args":["-y","@modelcontextprotocol/server-filesystem","."]}}}`
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
	mcpContent := `{"servers":{"filesystem":{"command":"npx","args":["-y","@modelcontextprotocol/server-filesystem","."]}}}`
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

func TestCompileKiroMCP_ProjectScope(t *testing.T) {
	targetDir := t.TempDir()

	aiDir := filepath.Join(targetDir, ".ai")
	_ = files.EnsureDir(aiDir)
	mcpContent := `{"servers":{"filesystem":{"command":"npx","args":["-y","@modelcontextprotocol/server-filesystem","."]}}}`
	if err := os.WriteFile(filepath.Join(aiDir, "mcp.json"), []byte(mcpContent), 0o644); err != nil {
		t.Fatal(err)
	}

	records, err := CompileMCPForTool(types.ToolIdKiro, CompileContext{
		TargetDir:  targetDir,
		SetupScope: types.SetupScopeProject,
	})
	if err != nil {
		t.Fatalf("CompileMCPForTool failed: %v", err)
	}

	configPath := filepath.Join(targetDir, ".kiro", "settings", "mcp.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("expected %s: %v", configPath, err)
	}
	if !strings.Contains(string(data), `"mcpServers"`) {
		t.Fatalf("kiro mcp config missing mcpServers: %s", string(data))
	}
	if len(records) != 1 {
		t.Fatalf("records len = %d, want 1", len(records))
	}
	if got := records[0].Path; got != ".kiro/settings/mcp.json" {
		t.Fatalf("record path = %q, want %q", got, ".kiro/settings/mcp.json")
	}
}

func TestCompileKiroMCP_GlobalScope(t *testing.T) {
	targetDir := t.TempDir()
	homeDir := t.TempDir()

	aiDir := filepath.Join(targetDir, ".ai")
	_ = files.EnsureDir(aiDir)
	mcpContent := `{"servers":{"filesystem":{"command":"npx","args":["-y","@modelcontextprotocol/server-filesystem","."]}}}`
	if err := os.WriteFile(filepath.Join(aiDir, "mcp.json"), []byte(mcpContent), 0o644); err != nil {
		t.Fatal(err)
	}

	records, err := CompileMCPForTool(types.ToolIdKiro, CompileContext{
		TargetDir:  targetDir,
		HomeDir:    homeDir,
		SetupScope: types.SetupScopeGlobal,
	})
	if err != nil {
		t.Fatalf("CompileMCPForTool failed: %v", err)
	}

	configPath := filepath.Join(homeDir, ".kiro", "settings", "mcp.json")
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("expected %s: %v", configPath, err)
	}
	if len(records) != 1 {
		t.Fatalf("records len = %d, want 1", len(records))
	}
}

func TestCompileAntigravityMCP_ProjectScopeWritesGlobalConfig(t *testing.T) {
	targetDir := t.TempDir()
	homeDir := t.TempDir()

	aiDir := filepath.Join(targetDir, ".ai")
	_ = files.EnsureDir(aiDir)
	mcpContent := `{
  "servers": {
    "stdio-server": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "."]
    },
    "http-server": {
      "url": "https://example.com/mcp",
      "headers": {"Authorization": "Bearer token"}
    }
  }
}`
	if err := os.WriteFile(filepath.Join(aiDir, "mcp.json"), []byte(mcpContent), 0o644); err != nil {
		t.Fatalf("failed to write mcp.json: %v", err)
	}

	records, err := CompileMCPForTool(types.ToolIdAntigravity, CompileContext{
		TargetDir:  targetDir,
		HomeDir:    homeDir,
		SetupScope: types.SetupScopeProject,
	})
	if err != nil {
		t.Fatalf("CompileMCPForTool failed: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 tracked file records (mcp_config.json + settings.json), got %d", len(records))
	}
	var sawConfig, sawSettings bool
	for _, r := range records {
		switch r.Source {
		case "compiled:mcp:antigravity":
			sawConfig = true
		case "compiled:mcp:antigravity:settings":
			sawSettings = true
		}
	}
	if !sawConfig || !sawSettings {
		t.Fatalf("expected both antigravity mcp_config and settings records; got %#v", records)
	}

	configPath := filepath.Join(homeDir, ".gemini", "config", "mcp_config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("expected %s: %v", configPath, err)
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("unmarshal mcp_config.json: %v", err)
	}

	servers, ok := payload["mcpServers"].(map[string]any)
	if !ok {
		t.Fatalf("payload missing mcpServers map: %#v", payload["mcpServers"])
	}
	httpEntryRaw, ok := servers["http-server"].(map[string]any)
	if !ok {
		t.Fatalf("missing http-server entry: %#v", servers)
	}
	if got, ok := httpEntryRaw["serverUrl"].(string); !ok || got != "https://example.com/mcp" {
		t.Fatalf("http-server missing serverUrl; got=%v", httpEntryRaw["serverUrl"])
	}
}

// TestAIMemoryCatalogEntryUsesRemoteURL verifies the retained ai-memory entry
// stays URL-backed and compiles to the native remote MCP shape.
func TestAIMemoryCatalogEntryUsesRemoteURL(t *testing.T) {
	libFS := library.GetLibraryFS()
	if libFS == nil {
		t.Fatal("library.GetLibraryFS returned nil")
	}
	data, err := fs.ReadFile(libFS, "mcp/catalog.json")
	if err != nil {
		t.Fatalf("read catalog.json: %v", err)
	}
	var catalog struct {
		Servers map[string]McpServer `json:"servers"`
	}
	if err := json.Unmarshal(data, &catalog); err != nil {
		t.Fatalf("unmarshal catalog: %v", err)
	}
	server, ok := catalog.Servers["ai-memory"]
	if !ok {
		t.Fatal("catalog.servers missing 'ai-memory' entry")
	}
	if server.Enabled == nil || !*server.Enabled {
		t.Fatal("ai-memory must be enabled by default")
	}
	if want := "http://127.0.0.1:49374/mcp"; server.URL != want {
		t.Fatalf("ai-memory.url = %q, want %q", server.URL, want)
	}
	if server.Command != "" || len(server.Args) > 0 {
		t.Fatalf("ai-memory entry must not declare stdio command/args: cmd=%q args=%v", server.Command, server.Args)
	}

	emitted := toOpenCodeMcp(map[string]McpServer{"ai-memory": server})
	entry, ok := emitted["ai-memory"].(map[string]any)
	if !ok {
		t.Fatalf("emitted ai-memory entry is not a map: %T", emitted["ai-memory"])
	}
	if entry["type"] != "remote" {
		t.Fatalf("emitted type = %v, want remote", entry["type"])
	}
	if entry["enabled"] != true {
		t.Fatalf("emitted enabled = %v, want true", entry["enabled"])
	}
	if entry["url"] != "http://127.0.0.1:49374/mcp" {
		t.Fatalf("emitted url = %v, want ai-memory endpoint", entry["url"])
	}
}

func TestMcpCatalogOmitsDormantEntries(t *testing.T) {
	libFS := library.GetLibraryFS()
	if libFS == nil {
		t.Fatal("library.GetLibraryFS returned nil")
	}
	data, err := fs.ReadFile(libFS, "mcp/catalog.json")
	if err != nil {
		t.Fatalf("read catalog.json: %v", err)
	}
	var catalog struct {
		Servers map[string]McpServer `json:"servers"`
	}
	if err := json.Unmarshal(data, &catalog); err != nil {
		t.Fatalf("unmarshal catalog: %v", err)
	}
	if got := len(catalog.Servers); got != 5 {
		t.Fatalf("catalog server count = %d, want 5", got)
	}
	for _, excluded := range []string{"figma", "slack", "context7", "github", "playwright", "atlassian", "fetch", "memory", "memoria", "qmd", "graphify"} {
		if _, ok := catalog.Servers[excluded]; ok {
			t.Fatalf("catalog must not ship %q MCP server", excluded)
		}
	}
}

func TestMcpCatalogExcludesRetiredOrchestratorDefault(t *testing.T) {
	libFS := library.GetLibraryFS()
	if libFS == nil {
		t.Fatal("library.GetLibraryFS returned nil")
	}
	data, err := fs.ReadFile(libFS, "mcp/catalog.json")
	if err != nil {
		t.Fatalf("read catalog.json: %v", err)
	}
	var catalog struct {
		Servers map[string]McpServer `json:"servers"`
	}
	if err := json.Unmarshal(data, &catalog); err != nil {
		t.Fatalf("unmarshal catalog: %v", err)
	}
	if _, ok := catalog.Servers["orchestrator"]; ok {
		t.Fatal("catalog must not ship retired orchestrator as a default MCP server")
	}
}

// TestCompileOpenCodeMCP_JsoncFallback verifies that when only .ai/mcp.jsonc
// exists (no .ai/mcp.json), CompileMCPForTool still reads and compiles the
// canonical MCP config. This is the regression test for issue #369.
func TestCompileOpenCodeMCP_JsoncFallback(t *testing.T) {
	targetDir := t.TempDir()

	aiDir := filepath.Join(targetDir, ".ai")
	_ = files.EnsureDir(aiDir)
	// JSONC with comments — only .jsonc exists, no .json
	mcpContent := `{
	// This is a comment in JSONC
	"servers": {
		"filesystem": {
			"command": "npx",
			"args": ["-y", "@modelcontextprotocol/server-filesystem", "."]
		}
	}
}`
	if err := os.WriteFile(filepath.Join(aiDir, "mcp.jsonc"), []byte(mcpContent), 0o644); err != nil {
		t.Fatalf("failed to write mcp.jsonc: %v", err)
	}

	// Verify no .ai/mcp.json exists
	if _, err := os.Stat(filepath.Join(aiDir, "mcp.json")); !os.IsNotExist(err) {
		t.Fatal("mcp.json should not exist for this test")
	}

	records, err := CompileMCPForTool(types.ToolIdOpenCode, CompileContext{
		TargetDir:  targetDir,
		SetupScope: types.SetupScopeProject,
	})
	if err != nil {
		t.Fatalf("CompileMCPForTool failed: %v", err)
	}

	configPath := filepath.Join(targetDir, OpenCodeConfigFilename)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("expected opencode.json was not created from .ai/mcp.jsonc")
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 tracked file record, got %d", len(records))
	}
}

// TestReadCanonicalMcp_JsoncFallback verifies that ReadCanonicalMcp falls back
// to .ai/mcp.jsonc when .ai/mcp.json does not exist.
func TestReadCanonicalMcp_JsoncFallback(t *testing.T) {
	targetDir := t.TempDir()

	aiDir := filepath.Join(targetDir, ".ai")
	_ = files.EnsureDir(aiDir)
	mcpContent := `{
	// comment
	"servers": {
		"test": {
			"command": "echo",
			"args": ["hello"]
		}
	}
}`
	if err := os.WriteFile(filepath.Join(aiDir, "mcp.jsonc"), []byte(mcpContent), 0o644); err != nil {
		t.Fatalf("failed to write mcp.jsonc: %v", err)
	}

	catalog := ReadCanonicalMcp(targetDir)
	if catalog == nil {
		t.Fatal("ReadCanonicalMcp returned nil, expected catalog from .ai/mcp.jsonc")
	}
	if len(catalog.Servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(catalog.Servers))
	}
	svr, ok := catalog.Servers["test"]
	if !ok {
		t.Fatal("expected server 'test' in catalog")
	}
	if svr.Command != "echo" {
		t.Fatalf("expected command 'echo', got %q", svr.Command)
	}
}

// TestReadCanonicalMcp_JsonPrecedence verifies that .ai/mcp.json takes
// precedence over .ai/mcp.jsonc when both exist.
func TestReadCanonicalMcp_JsonPrecedence(t *testing.T) {
	targetDir := t.TempDir()

	aiDir := filepath.Join(targetDir, ".ai")
	_ = files.EnsureDir(aiDir)
	// Write both .json and .jsonc with different content
	jsonContent := `{"servers":{"from-json":{"command":"json","args":[]}}}`
	jsoncContent := `{"servers":{"from-jsonc":{"command":"jsonc","args":[]}}}`
	if err := os.WriteFile(filepath.Join(aiDir, "mcp.json"), []byte(jsonContent), 0o644); err != nil {
		t.Fatalf("failed to write mcp.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(aiDir, "mcp.jsonc"), []byte(jsoncContent), 0o644); err != nil {
		t.Fatalf("failed to write mcp.jsonc: %v", err)
	}

	catalog := ReadCanonicalMcp(targetDir)
	if catalog == nil {
		t.Fatal("ReadCanonicalMcp returned nil")
	}
	if _, ok := catalog.Servers["from-json"]; !ok {
		t.Fatal("expected server 'from-json' (from .json) to take precedence")
	}
	if _, ok := catalog.Servers["from-jsonc"]; ok {
		t.Fatal("server 'from-jsonc' should not appear when .json takes precedence")
	}
}

// TestTrackedRecordPath_Normalization verifies that trackedRecordPath produces
// slash-normalized relative paths for files under workspaceRoot, and
// slash-normalized absolute paths for files outside workspaceRoot.
func TestTrackedRecordPath_Normalization(t *testing.T) {
	t.Run("relative path under root", func(t *testing.T) {
		got := trackedRecordPath("/home/user/project", "/home/user/project/.kiro/settings/mcp.json")
		if got != ".kiro/settings/mcp.json" {
			t.Errorf("got %q, want %q", got, ".kiro/settings/mcp.json")
		}
	})

	t.Run("absolute path outside root", func(t *testing.T) {
		got := trackedRecordPath("/home/user/project", "/home/user/.gemini/config/mcp_config.json")
		want := "/home/user/.gemini/config/mcp_config.json"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("path equals root returns dot", func(t *testing.T) {
		got := trackedRecordPath("/home/user/project", "/home/user/project")
		if got != "." {
			t.Errorf("got %q, want %q", got, ".")
		}
	})

	t.Run("Windows-style backslash under root normalized to forward slash", func(t *testing.T) {
		root := filepath.FromSlash("C:/Users/me/project")
		fp := filepath.FromSlash("C:/Users/me/project/.omp/mcp.json")
		got := trackedRecordPath(root, fp)
		if got != ".omp/mcp.json" {
			t.Errorf("got %q, want %q", got, ".omp/mcp.json")
		}
	})

	t.Run("Windows-style absolute outside root normalized to forward slash", func(t *testing.T) {
		root := filepath.FromSlash("C:/Users/me/project")
		fp := filepath.FromSlash("C:/Users/me/.kiro/settings/mcp.json")
		got := trackedRecordPath(root, fp)
		// On Windows, filepath.Rel succeeds and produces a relative path;
		// on other platforms, the paths are treated as regular directories
		// and Rel produces "../" — the helper falls back to absolute.
		if got != "C:/Users/me/.kiro/settings/mcp.json" && got != "../.kiro/settings/mcp.json" {
			t.Errorf("got %q, want either %q or %q", got, "C:/Users/me/.kiro/settings/mcp.json", "../.kiro/settings/mcp.json")
		}
	})
}

// TestToKiroMcp_RemoteNoType verifies that remote Kiro MCP entries emit {url, headers}
// with NO "type" key, per https://kiro.dev/docs/mcp/configuration.
func TestToKiroMcp_RemoteNoType(t *testing.T) {
	servers := map[string]McpServer{
		"my-remote": {
			URL: "https://example.com/mcp",
			Headers: map[string]string{
				"Authorization": "Bearer token",
			},
		},
	}
	result := toKiroMcp(servers)

	mcpServers, ok := result["mcpServers"].(map[string]any)
	if !ok {
		t.Fatalf("mcpServers not a map[string]any: %T", result["mcpServers"])
	}
	entry, ok := mcpServers["my-remote"].(map[string]any)
	if !ok {
		t.Fatalf("my-remote entry not a map[string]any: %T", mcpServers["my-remote"])
	}

	// Must have url
	if got, ok := entry["url"]; !ok || got != "https://example.com/mcp" {
		t.Errorf("url = %v, want %q", got, "https://example.com/mcp")
	}
	// Must have headers
	if _, ok := entry["headers"]; !ok {
		t.Error("headers missing from remote entry")
	}
	// Must NOT have type
	if _, ok := entry["type"]; ok {
		t.Errorf("type key must not be present in Kiro remote MCP entry, got %v", entry["type"])
	}
}

// TestToKiroMcp_RemoteNoHeadersWhenNil verifies that headers is omitted when nil.
func TestToKiroMcp_RemoteNoHeadersWhenNil(t *testing.T) {
	servers := map[string]McpServer{
		"bare-remote": {URL: "https://example.com/mcp"},
	}
	result := toKiroMcp(servers)

	mcpServers := result["mcpServers"].(map[string]any)
	entry := mcpServers["bare-remote"].(map[string]any)

	if _, ok := entry["type"]; ok {
		t.Error("type key must not be present")
	}
	if _, ok := entry["headers"]; ok {
		t.Error("headers key must not be present when nil")
	}
	if got := entry["url"]; got != "https://example.com/mcp" {
		t.Errorf("url = %v, want https://example.com/mcp", got)
	}
}

// TestToKiroMcp_LocalShape verifies that local/stdio Kiro MCP entries emit
// {command, args, env} matching the documented local server shape.
func TestToKiroMcp_LocalShape(t *testing.T) {
	servers := map[string]McpServer{
		"fs": {
			Command: "npx",
			Args:    []string{"-y", "@modelcontextprotocol/server-filesystem", "."},
			Env:     map[string]string{"DEBUG": "1"},
		},
	}
	result := toKiroMcp(servers)

	mcpServers := result["mcpServers"].(map[string]any)
	entry := mcpServers["fs"].(map[string]any)

	if got := entry["command"]; got != "npx" {
		t.Errorf("command = %v, want npx", got)
	}
	args, ok := entry["args"].([]string)
	if !ok || len(args) != 3 {
		t.Errorf("args = %v, unexpected", entry["args"])
	}
	if _, ok := entry["env"]; !ok {
		t.Error("env missing from local entry")
	}
	if _, ok := entry["type"]; ok {
		t.Error("type key must not be present in local entry")
	}
	if _, ok := entry["url"]; ok {
		t.Error("url key must not be present in local entry")
	}
}
