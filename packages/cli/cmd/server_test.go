package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// TestServerDoctorL1ConfigCheck verifies that server doctor performs L1 config
// checks (canonical mcp.json existence, entry presence, per-tool config) and
// skips L3 stdio handshake with an accurate message.
func TestServerDoctorL1ConfigCheck(t *testing.T) {
	dir := t.TempDir()

	// Write a minimal store with one tool and one enabled server
	writeManifestStoreData(t, dir, func(data *types.StoreData) {
		data.Config.Tools = []types.ToolId{types.ToolIdOpenCode}
		data.Config.EnableServers = []string{"filesystem"}
	})

	// Write .ai/mcp.json with the filesystem server enabled
	aiDir := filepath.Join(dir, ".ai")
	if err := os.MkdirAll(aiDir, 0o755); err != nil {
		t.Fatalf("mkdir .ai: %v", err)
	}
	mcpContent := `{
  "servers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "."],
      "enabled": true
    }
  }
}`
	if err := os.WriteFile(filepath.Join(aiDir, "mcp.json"), []byte(mcpContent), 0o644); err != nil {
		t.Fatalf("write .ai/mcp.json: %v", err)
	}

	// Write opencode.json (per-tool compiled config) with the server
	ocContent := `{
  "mcp": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "."]
    }
  }
}`
	if err := os.WriteFile(filepath.Join(dir, "opencode.json"), []byte(ocContent), 0o644); err != nil {
		t.Fatalf("write opencode.json: %v", err)
	}

	// Run server doctor
	cmd := newServerDoctorCommand(dir, true)
	stdout, _ := captureOutput(t, func() {
		_ = runServerDoctor(cmd, []string{"filesystem"})
	})

	var payload map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("unmarshal output: %v\noutput=%s", err, stdout)
	}

	reports, ok := payload["reports"].([]any)
	if !ok || len(reports) != 1 {
		t.Fatalf("expected 1 report, got %#v", payload["reports"])
	}

	report := reports[0].(map[string]any)
	checks := report["checks"].([]any)

	// Verify L1 checks pass
	var l1PassCount int
	var l3SkipCount int
	for _, c := range checks {
		check := c.(map[string]any)
		name := check["name"].(string)
		status := check["status"].(string)
		msg := check["message"].(string)

		switch {
		case name == "canonical mcp.json entry" && status == "pass":
			l1PassCount++
		case name == "opencode mcp config" && status == "pass":
			l1PassCount++
		case name == "stdio handshake" && status == "skip":
			l3SkipCount++
			// Must not claim fake support
			if contains(msg, "available") || contains(msg, "performed") {
				t.Errorf("L3 skip message should not claim availability/performed: %q", msg)
			}
		}
	}

	if l1PassCount < 2 {
		t.Errorf("expected at least 2 L1 checks to pass, got %d", l1PassCount)
	}
	if l3SkipCount != 1 {
		t.Errorf("expected exactly 1 L3 skip check, got %d", l3SkipCount)
	}
}

// TestServerDoctorL3SkipMessage verifies the L3 skip message is accurate
// and does not claim fake support.
func TestServerDoctorL3SkipMessage(t *testing.T) {
	dir := t.TempDir()

	// Write a minimal store
	writeManifestStoreData(t, dir, func(data *types.StoreData) {
		data.Config.Tools = []types.ToolId{types.ToolIdOpenCode}
		data.Config.EnableServers = []string{"filesystem"}
	})

	// Write .ai/mcp.json
	aiDir := filepath.Join(dir, ".ai")
	if err := os.MkdirAll(aiDir, 0o755); err != nil {
		t.Fatalf("mkdir .ai: %v", err)
	}
	mcpContent := `{
  "servers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "."],
      "enabled": true
    }
  }
}`
	if err := os.WriteFile(filepath.Join(aiDir, "mcp.json"), []byte(mcpContent), 0o644); err != nil {
		t.Fatalf("write .ai/mcp.json: %v", err)
	}

	// Write opencode.json
	ocContent := `{
  "mcp": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "."]
    }
  }
}`
	if err := os.WriteFile(filepath.Join(dir, "opencode.json"), []byte(ocContent), 0o644); err != nil {
		t.Fatalf("write opencode.json: %v", err)
	}

	// Run server doctor
	cmd := newServerDoctorCommand(dir, true)
	stdout, _ := captureOutput(t, func() {
		_ = runServerDoctor(cmd, []string{"filesystem"})
	})

	var payload map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("unmarshal output: %v\noutput=%s", err, stdout)
	}

	reports := payload["reports"].([]any)
	report := reports[0].(map[string]any)
	checks := report["checks"].([]any)

	for _, c := range checks {
		check := c.(map[string]any)
		if check["name"].(string) == "stdio handshake" {
			msg := check["message"].(string)
			// Must not claim fake support
			if contains(msg, "available") {
				t.Errorf("L3 skip message should not claim availability: %q", msg)
			}
			// Must mention the real reason
			if !contains(msg, "MCP client library") && !contains(msg, "Go binary") {
				t.Errorf("L3 skip message should mention the real limitation: %q", msg)
			}
		}
	}
}

// TestServerDoctorL3SkipRemoteServer verifies that URL-based servers also
// get a skip (not a pass) for the L3 handshake check.
func TestServerDoctorL3SkipRemoteServer(t *testing.T) {
	dir := t.TempDir()

	writeManifestStoreData(t, dir, func(data *types.StoreData) {
		data.Config.Tools = []types.ToolId{types.ToolIdOpenCode}
		data.Config.EnableServers = []string{"ai-memory"}
	})

	aiDir := filepath.Join(dir, ".ai")
	if err := os.MkdirAll(aiDir, 0o755); err != nil {
		t.Fatalf("mkdir .ai: %v", err)
	}
	mcpContent := `{
  "servers": {
    "ai-memory": {
      "url": "http://127.0.0.1:49374/mcp",
      "enabled": true
    }
  }
}`
	if err := os.WriteFile(filepath.Join(aiDir, "mcp.json"), []byte(mcpContent), 0o644); err != nil {
		t.Fatalf("write .ai/mcp.json: %v", err)
	}

	ocContent := `{
  "mcp": {
    "ai-memory": {
      "url": "http://127.0.0.1:49374/mcp"
    }
  }
}`
	if err := os.WriteFile(filepath.Join(dir, "opencode.json"), []byte(ocContent), 0o644); err != nil {
		t.Fatalf("write opencode.json: %v", err)
	}

	cmd := newServerDoctorCommand(dir, true)
	stdout, _ := captureOutput(t, func() {
		_ = runServerDoctor(cmd, []string{"ai-memory"})
	})

	var payload map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("unmarshal output: %v\noutput=%s", err, stdout)
	}

	reports := payload["reports"].([]any)
	report := reports[0].(map[string]any)
	checks := report["checks"].([]any)

	for _, c := range checks {
		check := c.(map[string]any)
		if check["name"].(string) == "stdio handshake" {
			status := check["status"].(string)
			if status != "skip" {
				t.Errorf("expected L3 handshake to be skip for remote server, got %q", status)
			}
		}
	}
}

// TestServerDoctorL1MissingCanonical verifies L1 check fails when .ai/mcp.json is missing.
func TestServerDoctorL1MissingCanonical(t *testing.T) {
	dir := t.TempDir()

	writeManifestStoreData(t, dir, func(data *types.StoreData) {
		data.Config.Tools = []types.ToolId{types.ToolIdOpenCode}
		data.Config.EnableServers = []string{"filesystem"}
	})

	cmd := newServerDoctorCommand(dir, true)
	stdout, _ := captureOutput(t, func() {
		_ = runServerDoctor(cmd, []string{"filesystem"})
	})

	var payload map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("unmarshal output: %v\noutput=%s", err, stdout)
	}

	reports := payload["reports"].([]any)
	report := reports[0].(map[string]any)
	if report["overall"].(string) != "unhealthy" {
		t.Errorf("expected overall=unhealthy when .ai/mcp.json is missing, got %q", report["overall"])
	}
}

// TestServerDoctorL1MissingPerToolConfig verifies L1 check fails when per-tool config is missing.
func TestServerDoctorL1MissingPerToolConfig(t *testing.T) {
	dir := t.TempDir()

	writeManifestStoreData(t, dir, func(data *types.StoreData) {
		data.Config.Tools = []types.ToolId{types.ToolIdOpenCode}
		data.Config.EnableServers = []string{"filesystem"}
	})

	aiDir := filepath.Join(dir, ".ai")
	if err := os.MkdirAll(aiDir, 0o755); err != nil {
		t.Fatalf("mkdir .ai: %v", err)
	}
	mcpContent := `{
  "servers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "."],
      "enabled": true
    }
  }
}`
	if err := os.WriteFile(filepath.Join(aiDir, "mcp.json"), []byte(mcpContent), 0o644); err != nil {
		t.Fatalf("write .ai/mcp.json: %v", err)
	}

	// No opencode.json — per-tool config is missing
	cmd := newServerDoctorCommand(dir, true)
	stdout, _ := captureOutput(t, func() {
		_ = runServerDoctor(cmd, []string{"filesystem"})
	})

	var payload map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("unmarshal output: %v\noutput=%s", err, stdout)
	}

	reports := payload["reports"].([]any)
	report := reports[0].(map[string]any)
	if report["overall"].(string) != "unhealthy" {
		t.Errorf("expected overall=unhealthy when per-tool config is missing, got %q", report["overall"])
	}
}

// TestServerDoctorNoEnabledServers verifies graceful handling when no servers are enabled.
func TestServerDoctorNoEnabledServers(t *testing.T) {
	dir := t.TempDir()

	writeManifestStoreData(t, dir, func(data *types.StoreData) {
		data.Config.Tools = []types.ToolId{types.ToolIdOpenCode}
		data.Config.EnableServers = []string{}
	})

	cmd := newServerDoctorCommand(dir, true)
	stdout, _ := captureOutput(t, func() {
		_ = runServerDoctor(cmd, nil)
	})

	var payload map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("unmarshal output: %v\noutput=%s", err, stdout)
	}

	reports, ok := payload["reports"].([]any)
	if !ok || len(reports) != 0 {
		t.Errorf("expected empty reports when no servers enabled, got %#v", payload["reports"])
	}
}

// TestServerDoctorUnknownServer verifies graceful handling of unknown server names.
func TestServerDoctorUnknownServer(t *testing.T) {
	dir := t.TempDir()

	writeManifestStoreData(t, dir, func(data *types.StoreData) {
		data.Config.Tools = []types.ToolId{types.ToolIdOpenCode}
		data.Config.EnableServers = []string{"nonexistent"}
	})

	cmd := newServerDoctorCommand(dir, true)
	stdout, _ := captureOutput(t, func() {
		_ = runServerDoctor(cmd, []string{"nonexistent"})
	})

	var payload map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("unmarshal output: %v\noutput=%s", err, stdout)
	}

	reports := payload["reports"].([]any)
	report := reports[0].(map[string]any)
	if report["overall"].(string) != "unhealthy" {
		t.Errorf("expected overall=unhealthy for unknown server, got %q", report["overall"])
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newServerDoctorCommand(dir string, jsonOutput bool) *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().String("dir", dir, "")
	cmd.Flags().Bool("json", jsonOutput, "")
	return cmd
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
