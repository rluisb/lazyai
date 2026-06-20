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
// config is written to <targetDir>/.opencode/lazyai.mcp.jsonc.
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

	configPath := filepath.Join(targetDir, ".opencode", OpenCodeRuntimeMCPFilename)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("expected .opencode/lazyai.mcp.jsonc was not created")
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 tracked file record, got %d", len(records))
	}
}

// TestCompileOpenCodeMCP_GlobalScope verifies that at global scope, the MCP
// config is written to <home>/.config/opencode/lazyai.mcp.jsonc (via
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

	// Global scope: lazyai.mcp.jsonc should live under <home>/.config/opencode.
	globalConfigPath := filepath.Join(homeDir, ".config", "opencode", OpenCodeRuntimeMCPFilename)
	if _, err := os.Stat(globalConfigPath); os.IsNotExist(err) {
		t.Fatalf("expected lazyai.mcp.jsonc at %q, not found", globalConfigPath)
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
	mcpContent := `{"servers":{"filesystem":{"command":"npx","args":["-y","@modelcontextprotocol/server-filesystem","."]}}}`
	if err := os.WriteFile(filepath.Join(aiDir, "mcp.json"), []byte(mcpContent), 0o644); err != nil {
		t.Fatalf("write mcp.json: %v", err)
	}

	// Pre-seed lazyai.mcp.jsonc with a user-authored MCP entry under a name
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

	data, err := os.ReadFile(filepath.Join(ocDir, "lazyai.mcp.jsonc"))
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

	ocDir := filepath.Join(targetDir, ".opencode")
	_ = files.EnsureDir(ocDir)
	preExisting := `{"mcp":{"filesystem":{"type":"local","command":["user-override"]}}}`
	if err := os.WriteFile(filepath.Join(ocDir, "lazyai.mcp.jsonc"), []byte(preExisting), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := CompileMCPForTool(types.ToolIdOpenCode, CompileContext{
		TargetDir:  targetDir,
		SetupScope: types.SetupScopeProject,
	}); err != nil {
		t.Fatalf("CompileMCPForTool: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(ocDir, "lazyai.mcp.jsonc"))
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
      "command": "/tmp/lazyai-orchestrator",
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

	data, err := os.ReadFile(filepath.Join(targetDir, ".opencode", "lazyai.mcp.jsonc"))
	if err != nil {
		t.Fatalf("read lazyai.mcp.jsonc: %v", err)
	}
	contents := string(data)
	if !strings.Contains(contents, `"orchestrator"`) {
		t.Fatalf("compiled config did not preserve user-provided legacy orchestrator entry:\n%s", contents)
	}
	if !strings.Contains(contents, `/tmp/lazyai-orchestrator`) || !strings.Contains(contents, `connect`) || !strings.Contains(contents, `--project`) {
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
      "args": ["connect", "--project", "/tmp/project"]
    },
    "orchestrator-absolute": {
      "command": "/tmp/lazyai-orchestrator",
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

	data, err := os.ReadFile(filepath.Join(targetDir, ".opencode", "lazyai.mcp.jsonc"))
	if err != nil {
		t.Fatalf("read lazyai.mcp.jsonc: %v", err)
	}
	contents := string(data)
	if !strings.Contains(contents, `"lazyai-orchestrator"`) || !strings.Contains(contents, `connect`) || !strings.Contains(contents, `--project`) {
		t.Fatalf("compiled config did not preserve user-provided legacy orchestrator command:\n%s", contents)
	}
	if !strings.Contains(contents, `/tmp/lazyai-orchestrator`) {
		t.Fatalf("compiled config did not preserve user-provided absolute legacy orchestrator command:\n%s", contents)
	}
}

func TestMCPCompilerPreservesUserProvidedLegacyOrchestratorCommandArgsForToolPayloads(t *testing.T) {
	servers := map[string]McpServer{
		"orchestrator": {
			Command: "/tmp/lazyai-orchestrator",
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
		if !strings.Contains(contents, `/tmp/lazyai-orchestrator`) || !strings.Contains(contents, `connect`) || !strings.Contains(contents, `--project`) || !strings.Contains(contents, `/tmp/project`) {
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
	if got := records[0].Path; got != configPath {
		t.Fatalf("record path = %q, want %q", got, configPath)
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
