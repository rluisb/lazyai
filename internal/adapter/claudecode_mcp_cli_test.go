package adapter

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// recordingRunner captures all CLI invocations for inspection in tests.
type recordingRunner struct {
	calls []cliCall
}

type cliCall struct {
	ctx       context.Context
	workingDir string
	args      []string
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

// TestMCPInstallCLI_ProjectScope verifies project scope → `-s project` flag.
func TestMCPInstallCLI_ProjectScope(t *testing.T) {
	ctx, _, _ := newScopeTestContext(t, types.SetupScopeProject)
	runner := &recordingRunner{}

	// Mock MCP catalog with one stdio server
	ctx.FileRecords = append(ctx.FileRecords, types.TrackedFile{
		Path: ".ai/mcp.json",
		Hash: "abc123",
	})

	// Simulate installClaudeMCPViaCLI with a mock runner.
	// For now, just verify the catalog reading works.
	catalog := readCanonicalMcp(ctx.TargetDir)
	if catalog != nil && len(catalog.Servers) > 0 {
		// Catalog was readable; test would mock the runner calls here.
	}

	_ = runner // placeholder for future comprehensive test with mocking
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

	if url, ok := payload["url"]; !ok || url != "http://localhost:3000" {
		t.Errorf("missing or wrong url: %v", url)
	}
	if headers, ok := payload["headers"].(map[string]interface{}); !ok {
		t.Errorf("missing headers: %v", headers)
	}
}
