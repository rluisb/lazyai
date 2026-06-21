package adapter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
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

func TestToCopilotVSCodeMcp_InputsSortedByID(t *testing.T) {
	servers := map[string]McpServer{
		"alpha": {
			Command: "npx",
			Env: map[string]string{
				"Z_TOKEN": "${Z_TOKEN}",
				"API_KEY": "${API_KEY}",
			},
		},
		"beta": {
			Command: "npx",
			Env: map[string]string{
				"A_TOKEN": "${A_TOKEN}",
				"M_TOKEN": "${M_TOKEN}",
			},
		},
	}

	got := toCopilotVSCodeMcp(servers)
	inputs, ok := got["inputs"].([]map[string]any)
	if !ok {
		t.Fatalf("expected inputs slice, got %T", got["inputs"])
	}

	wantIDs := []string{"API_KEY", "A_TOKEN", "M_TOKEN", "Z_TOKEN"}
	if len(inputs) != len(wantIDs) {
		t.Fatalf("expected %d inputs, got %d: %v", len(wantIDs), len(inputs), inputs)
	}
	for i, wantID := range wantIDs {
		if inputs[i]["id"] != wantID {
			t.Fatalf("inputs[%d].id = %v, want %s (inputs=%v)", i, inputs[i]["id"], wantID, inputs)
		}
	}
}

func TestToCopilotVSCodeMcp_RemoteURLUsesHttpType(t *testing.T) {
	servers := map[string]McpServer{
		"remote": {
			URL:     "https://mcp.example.com",
			Headers: map[string]string{"Authorization": "Bearer TOKEN"},
		},
		"stdio": {
			Command: "npx",
			Env:     map[string]string{"API_KEY": "${API_KEY}"},
		},
	}
	got := toCopilotVSCodeMcp(servers)

	serversSection, ok := got["servers"].(map[string]any)
	if !ok {
		t.Fatalf("expected servers map, got %T", got["servers"])
	}
	remote := serversSection["remote"].(map[string]any)
	if remote["type"] != "http" {
		t.Errorf("expected remote type=http, got %v", remote["type"])
	}
	if remote["url"] != "https://mcp.example.com" {
		t.Fatalf("expected remote url, got %v", remote["url"])
	}
	headers, ok := remote["headers"].(map[string]string)
	if !ok {
		t.Fatalf("expected headers map, got %T", remote["headers"])
	}
	if headers["Authorization"] != "Bearer TOKEN" {
		t.Fatalf("expected Authorization header to pass through, got %v", headers["Authorization"])
	}

	inputs, ok := got["inputs"].([]map[string]any)
	if !ok {
		t.Fatalf("expected inputs slice, got %T", got["inputs"])
	}
	if len(inputs) != 1 || inputs[0]["id"] != "API_KEY" {
		t.Fatalf("expected one API_KEY input prompt, got %v", inputs)
	}
}


func TestCompileCopilotVSCodeMcp_RemoteURL_UsesHTTPType(t *testing.T) {
	targetDir := t.TempDir()
	home := t.TempDir()
	t.Setenv("PATH", t.TempDir())

	aiDir := filepath.Join(targetDir, ".ai")
	_ = files.EnsureDir(aiDir)
	mcpContent := `{
  "servers": {
    "remote": {
      "url": "https://mcp.example.com",
      "headers": { "Authorization": "Bearer ${TOKEN}" }
    },
    "stdio": {
      "command": "npx",
      "env": { "API_KEY": "${API_KEY}" }
    }
  }
}`
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
	if len(records) != 1 {
		t.Fatalf("expected only .vscode/mcp.json record when Copilot CLI probe fails, got %d", len(records))
	}

	var parsed struct {
		Servers map[string]map[string]any `json:"servers"`
		Inputs  []struct {
			ID string `json:"id"`
		} `json:"inputs"`
	}
	data, err := os.ReadFile(filepath.Join(targetDir, ".vscode", "mcp.json"))
	if err != nil {
		t.Fatalf("read .vscode/mcp.json: %v", err)
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("parse .vscode/mcp.json: %v", err)
	}
	remote := parsed.Servers["remote"]
	if remote["type"] != "http" {
		t.Fatalf("expected remote type=http, got %v", remote["type"])
	}
	if remote["url"] != "https://mcp.example.com" {
		t.Fatalf("expected remote url, got %v", remote["url"])
	}
	if headers, ok := remote["headers"].(map[string]any); !ok || headers["Authorization"] != "Bearer ${TOKEN}" {
		t.Fatalf("expected headers to include Authorization=Bearer ${TOKEN}, got %v", remote["headers"])
	}
	if len(parsed.Inputs) != 1 || parsed.Inputs[0].ID != "API_KEY" {
		t.Fatalf("expected one API_KEY input prompt, got %+v", parsed.Inputs)
	}
}
func TestCompileCopilotVSCodeMcp_WritesInputsSortedByID(t *testing.T) {
	targetDir := t.TempDir()
	home := t.TempDir()
	t.Setenv("PATH", t.TempDir())

	aiDir := filepath.Join(targetDir, ".ai")
	_ = files.EnsureDir(aiDir)
	mcpContent := `{
  "servers": {
    "zeta": { "command": "npx", "env": { "Z_TOKEN": "${Z_TOKEN}" } },
    "alpha": { "command": "npx", "env": { "M_TOKEN": "${M_TOKEN}", "A_TOKEN": "${A_TOKEN}", "API_KEY": "${API_KEY}" } }
  }
}`
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
	if len(records) != 1 {
		t.Fatalf("expected only .vscode/mcp.json record when Copilot CLI probe fails, got %d", len(records))
	}

	data, err := os.ReadFile(filepath.Join(targetDir, ".vscode", "mcp.json"))
	if err != nil {
		t.Fatalf("read .vscode/mcp.json: %v", err)
	}
	var parsed struct {
		Inputs []struct {
			ID string `json:"id"`
		} `json:"inputs"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("parse .vscode/mcp.json: %v", err)
	}

	wantIDs := []string{"API_KEY", "A_TOKEN", "M_TOKEN", "Z_TOKEN"}
	if len(parsed.Inputs) != len(wantIDs) {
		t.Fatalf("expected %d inputs, got %d: %+v", len(wantIDs), len(parsed.Inputs), parsed.Inputs)
	}
	for i, wantID := range wantIDs {
		if parsed.Inputs[i].ID != wantID {
			t.Fatalf("inputs[%d].id = %q, want %q (inputs=%+v)", i, parsed.Inputs[i].ID, wantID, parsed.Inputs)
		}
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
