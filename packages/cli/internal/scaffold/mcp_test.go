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
      "args": ["-y", "@modelcontextprotocol/server-filesystem"],
      "enabled": false
    },
    "memory": {
      "description": "memory",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-memory"],
      "enabled": false
    }
  }
}`)

	var records []types.TrackedFile
	if err := ScaffoldMcp(targetDir, libraryDir, os.DirFS(libraryDir), []string{"filesystem"}, []string{"memory"}, &records, types.ConflictStrategyAlign, nil); err != nil {
		t.Fatalf("ScaffoldMcp: %v", err)
	}

	catalog := readScaffoldedMcpCatalog(t, filepath.Join(targetDir, ".ai", "mcp.json"))
	if got := enabledValue(t, catalog.Servers["filesystem"].Enabled); !got {
		t.Fatal("filesystem server should be enabled")
	}
	if got := enabledValue(t, catalog.Servers["memory"].Enabled); !got {
		t.Fatal("memory server should be enabled")
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

func TestScaffoldMcp_PreservesRemoteServerShape(t *testing.T) {
	repoRoot := t.TempDir()
	libraryDir := filepath.Join(repoRoot, "library")
	targetDir := t.TempDir()

	mustWriteTestFile(t, filepath.Join(libraryDir, "mcp", "catalog.json"), `{
  "servers": {
    "context7": {
      "description": "context7",
      "url": "https://mcp.context7.com/mcp",
      "headers": {
        "CONTEXT7_API_KEY": "${CONTEXT7_API_KEY}"
      },
      "enabled": false
    }
  }
}`)

	var records []types.TrackedFile
	if err := ScaffoldMcp(targetDir, libraryDir, os.DirFS(libraryDir), nil, []string{"context7"}, &records, types.ConflictStrategyAlign, nil); err != nil {
		t.Fatalf("ScaffoldMcp: %v", err)
	}

	catalog := readScaffoldedMcpCatalog(t, filepath.Join(targetDir, ".ai", "mcp.json"))
	context7 := catalog.Servers["context7"]
	if got := enabledValue(t, context7.Enabled); !got {
		t.Fatal("context7 server should be enabled when selected")
	}
	if context7.URL != "https://mcp.context7.com/mcp" {
		t.Fatalf("context7.url = %q, want remote endpoint", context7.URL)
	}
	if got := context7.Headers["CONTEXT7_API_KEY"]; got != "${CONTEXT7_API_KEY}" {
		t.Fatalf("context7 header = %q, want env placeholder", got)
	}
	if context7.Command != "" || len(context7.Args) > 0 {
		t.Fatalf("context7 must remain URL-backed, got command=%q args=%v", context7.Command, context7.Args)
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
	rootContext7 := root.MCPServers["context7"]
	if rootContext7.URL != "https://mcp.context7.com/mcp" {
		t.Fatalf("root context7.url = %q, want remote endpoint", rootContext7.URL)
	}
	if got := rootContext7.Headers["CONTEXT7_API_KEY"]; got != "${CONTEXT7_API_KEY}" {
		t.Fatalf("root context7 header = %q, want env placeholder", got)
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
      "args": ["-y", "@modelcontextprotocol/server-filesystem"],
      "enabled": false
    }
  }
}`)

	var records []types.TrackedFile
	if err := ScaffoldMcp(targetDir, libraryDir, os.DirFS(libraryDir), nil, []string{"unknown"}, &records, types.ConflictStrategyAlign, nil); err != nil {
		t.Fatalf("ScaffoldMcp: %v", err)
	}

	catalog := readScaffoldedMcpCatalog(t, filepath.Join(targetDir, ".ai", "mcp.json"))
	if got := enabledValue(t, catalog.Servers["filesystem"].Enabled); got {
		t.Fatal("filesystem server should remain disabled")
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
