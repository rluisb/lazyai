package scaffold

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

func TestScaffoldMcp_EnablesSelectedServersAndWritesSchemas(t *testing.T) {
	repoRoot := t.TempDir()
	libraryDir := filepath.Join(repoRoot, "library")
	targetDir := t.TempDir()

	mustWriteTestFile(t, filepath.Join(libraryDir, "mcp", "catalog.json"), `{
  "servers": {
    "filesystem": {
      "description": "filesystem",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem"]
    },
    "codegraph": {
      "description": "codegraph",
      "command": "codegraph",
      "args": ["serve", "--mcp"]
    }
  }
}`)

	var records []types.TrackedFile
	if err := ScaffoldMcp(targetDir, libraryDir, os.DirFS(libraryDir), []string{"filesystem"}, []string{"codegraph"}, &records, types.ConflictStrategyAlign, nil); err != nil {
		t.Fatalf("ScaffoldMcp: %v", err)
	}

	catalog := readScaffoldedMcpCatalog(t, filepath.Join(targetDir, ".ai", "mcp.json"))
	if catalog.Version != mcpCatalogVersion {
		t.Fatalf("catalog version = %q, want %q", catalog.Version, mcpCatalogVersion)
	}
	if got := enabledValue(t, catalog.Servers["filesystem"].Enabled); !got {
		t.Fatal("filesystem server should be enabled")
	}
	if got := enabledValue(t, catalog.Servers["codegraph"].Enabled); !got {
		t.Fatal("codegraph server should be enabled")
	}

	rootPath := filepath.Join(targetDir, ".mcp.json")
	data, err := os.ReadFile(rootPath)
	if err != nil {
		t.Fatalf("read .mcp.json: %v", err)
	}
	var root struct {
		MCPServers map[string]mcpServer `json:"mcpServers"`
	}
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatalf("parse .mcp.json: %v", err)
	}
	if len(root.MCPServers) != 2 {
		t.Fatalf("root mcpServers = %d, want 2", len(root.MCPServers))
	}

	if len(records) != 2 {
		t.Fatalf("records = %d, want 2", len(records))
	}
	if records[0].Owner != types.FileOwnerLibrary || records[1].Owner != types.FileOwnerLibrary {
		t.Fatalf("records owners = %#v", records)
	}
}

func TestScaffoldMcp_EnableServersActsAsAllowlist(t *testing.T) {
	repoRoot := t.TempDir()
	libraryDir := filepath.Join(repoRoot, "library")
	targetDir := t.TempDir()

	mustWriteTestFile(t, filepath.Join(libraryDir, "mcp", "catalog.json"), `{
  "servers": {
    "filesystem": { "description": "filesystem", "command": "fs", "enabled": true },
    "codegraph": { "description": "codegraph", "command": "cg", "enabled": true },
    "obsidian": { "description": "obsidian", "command": "ob", "enabled": true }
  }
}`)

	// obsidian is selected as a CLI tool but NOT named in --enable-servers; under
	// the strict allowlist it must stay disabled (CLI-tool implication ignored).
	var records []types.TrackedFile
	if err := ScaffoldMcp(targetDir, libraryDir, os.DirFS(libraryDir), []string{"obsidian"}, []string{"codegraph"}, &records, types.ConflictStrategyAlign, nil); err != nil {
		t.Fatalf("ScaffoldMcp: %v", err)
	}

	catalog := readScaffoldedMcpCatalog(t, filepath.Join(targetDir, ".ai", "mcp.json"))
	if !enabledValue(t, catalog.Servers["codegraph"].Enabled) {
		t.Error("codegraph should stay enabled (named in allowlist)")
	}
	if !enabledValue(t, catalog.Servers["filesystem"].Enabled) {
		t.Error("filesystem should stay enabled (always-on floor)")
	}
	if enabledValue(t, catalog.Servers["obsidian"].Enabled) {
		t.Error("obsidian should be disabled (CLI-tool implication ignored in strict allowlist)")
	}

	// The root .mcp.json is consumed natively, so disabled servers must be absent.
	rootPath := filepath.Join(targetDir, ".mcp.json")
	data, err := os.ReadFile(rootPath)
	if err != nil {
		t.Fatalf("read .mcp.json: %v", err)
	}
	var root struct {
		MCPServers map[string]mcpServer `json:"mcpServers"`
	}
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatalf("parse .mcp.json: %v", err)
	}
	if _, ok := root.MCPServers["obsidian"]; ok {
		t.Error("root .mcp.json should omit disabled obsidian server")
	}
	if len(root.MCPServers) != 2 {
		t.Fatalf("root mcpServers = %d, want 2 (codegraph + filesystem)", len(root.MCPServers))
	}
}

func TestScaffoldMcp_PreservesRemoteServerShape(t *testing.T) {
	repoRoot := t.TempDir()
	libraryDir := filepath.Join(repoRoot, "library")
	targetDir := t.TempDir()

	mustWriteTestFile(t, filepath.Join(libraryDir, "mcp", "catalog.json"), `{
  "servers": {
    "remote-server": {
      "description": "remote-server",
      "url": "https://example.com/mcp",
      "headers": {
        "API_TOKEN": "${API_TOKEN}"
      }
    }
  }
}`)

	var records []types.TrackedFile
	if err := ScaffoldMcp(targetDir, libraryDir, os.DirFS(libraryDir), nil, []string{"remote-server"}, &records, types.ConflictStrategyAlign, nil); err != nil {
		t.Fatalf("ScaffoldMcp: %v", err)
	}

	catalog := readScaffoldedMcpCatalog(t, filepath.Join(targetDir, ".ai", "mcp.json"))
	remote := catalog.Servers["remote-server"]
	if got := enabledValue(t, remote.Enabled); !got {
		t.Fatal("remote-server should be enabled when selected")
	}
	if remote.URL != "https://example.com/mcp" {
		t.Fatalf("remote-server.url = %q, want remote endpoint", remote.URL)
	}
	if got := remote.Headers["API_TOKEN"]; got != "${API_TOKEN}" {
		t.Fatalf("remote-server header = %q, want env placeholder", got)
	}
	if remote.Command != "" || len(remote.Args) > 0 {
		t.Fatalf("remote-server must remain URL-backed, got command=%q args=%v", remote.Command, remote.Args)
	}

	rootPath := filepath.Join(targetDir, ".mcp.json")
	data, err := os.ReadFile(rootPath)
	if err != nil {
		t.Fatalf("read .mcp.json: %v", err)
	}
	var root struct {
		MCPServers map[string]mcpServer `json:"mcpServers"`
	}
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatalf("parse .mcp.json: %v", err)
	}
	rootRemote := root.MCPServers["remote-server"]
	if rootRemote.URL != "https://example.com/mcp" {
		t.Fatalf("root remote-server.url = %q, want remote endpoint", rootRemote.URL)
	}
	if got := rootRemote.Headers["API_TOKEN"]; got != "${API_TOKEN}" {
		t.Fatalf("root remote-server header = %q, want env placeholder", got)
	}
}

func TestScaffoldMcp_IgnoresUnknownServers(t *testing.T) {
	repoRoot := t.TempDir()
	libraryDir := filepath.Join(repoRoot, "library")
	targetDir := t.TempDir()

	mustWriteTestFile(t, filepath.Join(libraryDir, "mcp", "catalog.json"), `{
  "servers": {
    "filesystem": {
      "description": "filesystem",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem"]
    }
  }
}`)

	var records []types.TrackedFile
	if err := ScaffoldMcp(targetDir, libraryDir, os.DirFS(libraryDir), nil, []string{"unknown"}, &records, types.ConflictStrategyAlign, nil); err != nil {
		t.Fatalf("ScaffoldMcp: %v", err)
	}

	catalog := readScaffoldedMcpCatalog(t, filepath.Join(targetDir, ".ai", "mcp.json"))
	if _, ok := catalog.Servers["unknown"]; ok {
		t.Fatal("unknown server should not be added to the catalog")
	}
	// --enable-servers is non-empty, so allowlist mode applies: the unknown name
	// matches nothing, leaving only the always-on filesystem floor enabled.
	if !enabledValue(t, catalog.Servers["filesystem"].Enabled) {
		t.Fatal("filesystem floor should be enabled under allowlist mode")
	}
}

func readScaffoldedMcpCatalog(t *testing.T, path string) mcpCatalog {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var parsed mcpCatalog
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return parsed
}

func enabledValue(t *testing.T, value *bool) bool {
	t.Helper()
	if value == nil {
		t.Fatal("enabled flag is nil")
	}
	return *value
}

func mustWriteTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
