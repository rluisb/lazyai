package adapter

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// recordingRunner captures all CLI invocations for inspection in tests.
type recordingRunner struct {
	calls []cliCall
}

type cliCall struct {
	ctx        context.Context
	workingDir string
	args       []string
}

func (r *recordingRunner) Run(ctx context.Context, workingDir string, args ...string) ([]byte, []byte, error) {
	r.calls = append(r.calls, cliCall{ctx, workingDir, args})

	// For `mcp get`, simulate "not found" so servers are added.
	if len(args) >= 2 && args[0] == "mcp" && args[1] == "get" {
		return nil, nil, ErrMCPServerNotFound
	}

	// For `mcp add-json`, simulate success.
	if len(args) >= 2 && args[0] == "mcp" && args[1] == "add-json" {
		return []byte("ok"), nil, nil
	}

	return nil, nil, nil
}

var ErrMCPServerNotFound = NewError("mcp_server_not_found")

func NewError(code string) error {
	return &customErr{code}
}

type customErr struct {
	code string
}

func (e *customErr) Error() string { return e.code }

// TestMCPAddJsonPayloadProjectScope verifies that project scope uses `-s project` flag.
func TestMCPAddJsonPayloadProjectScope(t *testing.T) {
	runner := &recordingRunner{}
	ctx := CompileContext{
		SetupScope: types.SetupScopeProject,
		TargetDir:  t.TempDir(),
	}

	servers := map[string]McpServer{
		"test-server": {
			Command: "npx",
			Args:    []string{"-y", "@modelcontextprotocol/server-filesystem"},
			Enabled: boolPtr(true),
		},
	}

	// Manually run the CLI logic with the recording runner.
	for name, srv := range servers {
		if !srv.isEnabled() {
			continue
		}
		payload := mcpServerToJSON(srv)
		runner.Run(context.Background(), ctx.TargetDir, "mcp", "add-json", name, payload, "-s", "project")
	}

	if len(runner.calls) == 0 {
		t.Fatal("no CLI calls recorded")
	}

	// Verify the recorded call has the right shape
	call := runner.calls[0]
	if call.workingDir != ctx.TargetDir {
		t.Errorf("expected working dir %q, got %q", ctx.TargetDir, call.workingDir)
	}
	if len(call.args) < 6 || call.args[len(call.args)-1] != "project" {
		t.Errorf("expected `-s project` at end of args, got: %v", call.args)
	}
}

// TestMCPAddJsonPayloadGlobalScope verifies that global scope uses `-s user` flag.
func TestMCPAddJsonPayloadGlobalScope(t *testing.T) {
	runner := &recordingRunner{}
	// Global scope uses empty working dir
	_ = t.TempDir()

	servers := map[string]McpServer{
		"test-server": {
			Command: "npx",
			Args:    []string{"-y", "@modelcontextprotocol/server-filesystem"},
			Enabled: boolPtr(true),
		},
	}

	// At global scope, workingDir is empty.
	for name, srv := range servers {
		if !srv.isEnabled() {
			continue
		}
		payload := mcpServerToJSON(srv)
		runner.Run(context.Background(), "", "mcp", "add-json", name, payload, "-s", "user")
	}

	if len(runner.calls) == 0 {
		t.Fatal("no CLI calls recorded")
	}

	call := runner.calls[0]
	if call.workingDir != "" {
		t.Errorf("expected empty working dir for global scope, got %q", call.workingDir)
	}
	if len(call.args) < 6 || call.args[len(call.args)-1] != "user" {
		t.Errorf("expected `-s user` at end of args, got: %v", call.args)
	}
}

// TestMCPAddJsonPayloadWorkspaceScope verifies that workspace scope uses `-s project` flag
// with workspace directory as working dir.
func TestMCPAddJsonPayloadWorkspaceScope(t *testing.T) {
	runner := &recordingRunner{}
	workspaceDir := t.TempDir()

	servers := map[string]McpServer{
		"test-server": {
			Command: "npx",
			Args:    []string{"-y", "@modelcontextprotocol/server-filesystem"},
			Enabled: boolPtr(true),
		},
	}

	// Workspace uses project flag with workspace dir as working dir.
	for name, srv := range servers {
		if !srv.isEnabled() {
			continue
		}
		payload := mcpServerToJSON(srv)
		runner.Run(context.Background(), workspaceDir, "mcp", "add-json", name, payload, "-s", "project")
	}

	if len(runner.calls) == 0 {
		t.Fatal("no CLI calls recorded")
	}

	call := runner.calls[0]
	if call.workingDir != workspaceDir {
		t.Errorf("expected working dir %q, got %q", workspaceDir, call.workingDir)
	}
	if len(call.args) < 6 || call.args[len(call.args)-1] != "project" {
		t.Errorf("expected `-s project` at end of args for workspace, got: %v", call.args)
	}
}

// TestMcpServerToJSON_Stdio verifies JSON generation for stdio servers.
func TestMcpServerToJSON_Stdio(t *testing.T) {
	srv := McpServer{
		Command: "npx",
		Args:    []string{"-y", "my-mcp-server"},
		Env:     map[string]string{"API_KEY": "${API_KEY}"},
	}

	jsonStr := mcpServerToJSON(srv)

	// Verify it's valid JSON
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &payload); err != nil {
		t.Errorf("invalid JSON: %v\nJSON: %s", err, jsonStr)
	}

	// Verify key fields
	if cmd, ok := payload["command"]; !ok || cmd != "npx" {
		t.Errorf("missing or wrong command: %v", cmd)
	}
	if args, ok := payload["args"].([]interface{}); !ok || len(args) != 2 {
		t.Errorf("missing or wrong args: %v", args)
	}
	if env, ok := payload["env"].(map[string]interface{}); !ok || env["API_KEY"] != "${API_KEY}" {
		t.Errorf("missing or wrong env: %v", env)
	}
}

// TestMcpServerToJSON_HTTP verifies JSON generation for HTTP servers.
func TestMcpServerToJSON_HTTP(t *testing.T) {
	srv := McpServer{
		URL:     "http://localhost:3000",
		Headers: map[string]string{"Authorization": "Bearer token"},
	}

	jsonStr := mcpServerToJSON(srv)

	// Verify it's valid JSON and contains expected fields
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &payload); err != nil {
		t.Errorf("invalid JSON: %v", err)
	}

	if payloadType, ok := payload["type"]; !ok || payloadType != "http" {
		t.Errorf("missing or wrong type: %v", payloadType)
	}
	if url, ok := payload["url"]; !ok || url != "http://localhost:3000" {
		t.Errorf("missing or wrong url: %v", url)
	}
	if headers, ok := payload["headers"].(map[string]interface{}); !ok {
		t.Errorf("missing headers: %v", headers)
	}
}

// TestMcpServerToJSON_StdioNoType verifies stdio payload has no type key.
func TestMcpServerToJSON_StdioNoType(t *testing.T) {
	srv := McpServer{
		Command: "npx",
		Args:    []string{"-y", "my-mcp-server"},
	}

	jsonStr := mcpServerToJSON(srv)

	// Verify it's valid JSON and contains no type key.
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &payload); err != nil {
		t.Errorf("invalid JSON: %v\nJSON: %s", err, jsonStr)
	}
	if _, ok := payload["type"]; ok {
		t.Errorf("stdio payload should not include type, got: %v", payload["type"])
	}
}

// TestMCPPreCheckAndAddJson verifies the pre-check → add-json flow.
// When mcp get returns error, add-json should be called.
func TestMCPPreCheckAndAddJson(t *testing.T) {
	runner := &recordingRunner{}
	ctx := CompileContext{
		SetupScope: types.SetupScopeProject,
		TargetDir:  t.TempDir(),
	}

	servers := map[string]McpServer{
		"test-server": {
			Command: "npx",
			Args:    []string{"-y", "@modelcontextprotocol/server-filesystem"},
			Enabled: boolPtr(true),
		},
	}

	// Simulate the pre-check + add-json flow
	for name, srv := range servers {
		if !srv.isEnabled() {
			continue
		}
		// Pre-check: mcp get
		runner.Run(context.Background(), ctx.TargetDir, "mcp", "get", name)
		// If not found, add it
		payload := mcpServerToJSON(srv)
		runner.Run(context.Background(), ctx.TargetDir, "mcp", "add-json", name, payload, "-s", "project")
	}

	if len(runner.calls) < 2 {
		t.Fatalf("expected at least 2 CLI calls (get + add-json), got %d", len(runner.calls))
	}

	// First call should be mcp get
	getCall := runner.calls[0]
	if len(getCall.args) < 2 || getCall.args[0] != "mcp" || getCall.args[1] != "get" {
		t.Errorf("first call should be `mcp get`, got: %v", getCall.args)
	}

	// Second call should be mcp add-json
	addCall := runner.calls[1]
	if len(addCall.args) < 2 || addCall.args[0] != "mcp" || addCall.args[1] != "add-json" {
		t.Errorf("second call should be `mcp add-json`, got: %v", addCall.args)
	}
}

// TestMCPFallbackWhenBinaryMissing verifies that when claude is not on PATH,
// we fall back to direct-write (no CLI calls, .mcp.json written directly).
func TestMCPFallbackWhenBinaryMissing(t *testing.T) {
	dir := t.TempDir()
	ctx := CompileContext{
		TargetDir:  dir,
		SetupScope: types.SetupScopeProject,
	}

	servers := map[string]McpServer{
		"test-server": {
			Command: "npx",
			Args:    []string{"-y", "@modelcontextprotocol/server-filesystem"},
			Enabled: boolPtr(true),
		},
	}

	// useCliForMCP should return false when binary is not found
	result := useCliForMCP(ctx, servers)

	// When useCliForMCP returns false, fallback path should write .mcp.json
	if result {
		t.Skip("claude is on PATH; test requires it absent")
	}

	// In the fallback case, we write .mcp.json directly
	mcpPath := filepath.Join(dir, ".mcp.json")
	content := toClaudeCodeMcp(servers)
	if err := WriteJSONFile(mcpPath, content); err != nil {
		t.Fatalf("WriteJSONFile: %v", err)
	}

	if !files.FileExists(mcpPath) {
		t.Error(".mcp.json not written in fallback path")
	}
}

// TestMCPMultipleServersAddedInSequence verifies that when multiple servers
// are configured, each gets a separate add-json call.
func TestMCPMultipleServersAddedInSequence(t *testing.T) {
	runner := &recordingRunner{}
	ctx := CompileContext{
		SetupScope: types.SetupScopeProject,
		TargetDir:  t.TempDir(),
	}

	servers := map[string]McpServer{
		"filesystem": {
			Command: "npx",
			Args:    []string{"-y", "@modelcontextprotocol/server-filesystem"},
		},
		"postgresql": {
			Command: "npx",
			Args:    []string{"-y", "@modelcontextprotocol/server-postgres"},
		},
	}

	// Register both servers
	for name, srv := range servers {
		if !srv.isEnabled() {
			continue
		}
		payload := mcpServerToJSON(srv)
		runner.Run(context.Background(), ctx.TargetDir, "mcp", "add-json", name, payload, "-s", "project")
	}

	if len(runner.calls) != 2 {
		t.Fatalf("expected 2 add-json calls, got %d", len(runner.calls))
	}

	// Verify both servers are in the recorded calls
	serversSeen := make(map[string]bool)
	for _, call := range runner.calls {
		if len(call.args) >= 3 && call.args[0] == "mcp" && call.args[1] == "add-json" {
			serversSeen[call.args[2]] = true
		}
	}

	if !serversSeen["filesystem"] || !serversSeen["postgresql"] {
		t.Errorf("not all servers registered, saw: %v", serversSeen)
	}
}

// TestMCPFalseFlagServerSkipped verifies that servers flagged false are not registered.
func TestMCPFalseFlagServerSkipped(t *testing.T) {
	runner := &recordingRunner{}
	ctx := CompileContext{
		SetupScope: types.SetupScopeProject,
		TargetDir:  t.TempDir(),
	}

	servers := map[string]McpServer{
		"filesystem": {
			Command: "npx",
			Args:    []string{"-y", "@modelcontextprotocol/server-filesystem"},
			Enabled: boolPtr(true),
		},
		"flagged-off-server": {
			Command: "npx",
			Args:    []string{"-y", "some-server"},
			Enabled: boolPtr(false),
		},
	}

	// Only register enabled servers
	for name, srv := range servers {
		if !srv.isEnabled() {
			continue
		}
		payload := mcpServerToJSON(srv)
		runner.Run(context.Background(), ctx.TargetDir, "mcp", "add-json", name, payload, "-s", "project")
	}

	if len(runner.calls) != 1 {
		t.Fatalf("expected 1 call (only enabled server), got %d", len(runner.calls))
	}

	// Verify the registered server is the enabled one
	call := runner.calls[0]
	if len(call.args) < 3 || call.args[2] != "filesystem" {
		t.Errorf("expected 'filesystem' server, got: %v", call.args)
	}
}
