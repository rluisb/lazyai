package scaffold

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

func TestScaffoldMcp_PreparesManagedOrchestratorServerFromLocalBuild(t *testing.T) {
	targetDir := t.TempDir()
	repoRoot := t.TempDir()
	libraryDir := filepath.Join(repoRoot, "library")
	orchestratorDir := filepath.Join(repoRoot, "packages", "orchestrator-go")
	t.Setenv("HOME", t.TempDir())
	mustWriteTestFile(t, filepath.Join(libraryDir, "mcp", "catalog.json"), `{
  "servers": {
    "orchestrator": {
      "description": "orchestrator",
      "command": "ai-setup-orchestrator",
      "args": ["connect"],
      "enabled": false,
      "requiresInstall": false
    }
  }
}`)
	mustWriteTestFile(t, filepath.Join(orchestratorDir, "go.mod"), `module example.com/orchestrator`)

	originalLookPath := orchestratorLookPath
	originalRunner := orchestratorCommandRunner
	t.Cleanup(func() {
		orchestratorLookPath = originalLookPath
		orchestratorCommandRunner = originalRunner
	})
	orchestratorLookPath = func(file string) (string, error) {
		switch file {
		case "go":
			return "/fake/bin/go", nil
		default:
			return "", os.ErrNotExist
		}
	}
	orchestratorCommandRunner = func(ctx context.Context, name string, args ...string) error {
		if name == "/fake/bin/go" {
			for i, arg := range args {
				if arg == "-o" && i+1 < len(args) {
					mustWriteTestFile(t, args[i+1], "binary")
				}
			}
		}
		return nil
	}

	var records []types.TrackedFile
	if err := ScaffoldMcp(targetDir, libraryDir, os.DirFS(libraryDir), nil, []string{"orchestrator"}, &records, types.ConflictStrategyAlign, nil); err != nil {
		t.Fatalf("ScaffoldMcp: %v", err)
	}

	parsed := readScaffoldedMcpCatalog(t, filepath.Join(targetDir, ".ai", "mcp.json"))
	entry := parsed.Servers["orchestrator"]
	if !filepath.IsAbs(entry.Command) || filepath.Base(entry.Command) != "ai-setup-orchestrator" {
		t.Fatalf("command = %q, want prepared absolute ai-setup-orchestrator binary", entry.Command)
	}
	if want := []string{"connect", "--project", targetDir}; !reflect.DeepEqual(entry.Args, want) {
		t.Fatalf("args = %#v, want %#v", entry.Args, want)
	}
	if entry.RequiresInstall {
		t.Fatal("requiresInstall should be false for managed orchestrator")
	}
	if len(records) != 2 || records[0].Owner != types.FileOwnerLibrary {
		t.Fatalf("records = %#v, want two library-owned tracked files (one for .ai/mcp.json, one for .mcp.json)", records)
	}
}

func TestScaffoldMcp_BuildsOrchestratorWhenDistMissing(t *testing.T) {
	targetDir := t.TempDir()
	repoRoot := t.TempDir()
	libraryDir := filepath.Join(repoRoot, "library")
	orchestratorDir := filepath.Join(repoRoot, "packages", "orchestrator-go")
	t.Setenv("HOME", t.TempDir())
	mustWriteTestFile(t, filepath.Join(libraryDir, "mcp", "catalog.json"), `{
  "servers": {
    "orchestrator": {
      "command": "ai-setup-orchestrator",
      "args": ["connect"]
    }
  }
}`)
	mustWriteTestFile(t, filepath.Join(orchestratorDir, "go.mod"), `module example.com/orchestrator`)

	originalLookPath := orchestratorLookPath
	originalRunner := orchestratorCommandRunner
	t.Cleanup(func() {
		orchestratorLookPath = originalLookPath
		orchestratorCommandRunner = originalRunner
	})
	orchestratorLookPath = func(file string) (string, error) {
		switch file {
		case "go":
			return "/fake/bin/go", nil
		default:
			return "", os.ErrNotExist
		}
	}
	var calls []string
	orchestratorCommandRunner = func(ctx context.Context, name string, args ...string) error {
		calls = append(calls, name+" "+strings.Join(args, " "))
		if name == "/fake/bin/go" {
			for i, arg := range args {
				if arg == "-o" && i+1 < len(args) {
					mustWriteTestFile(t, args[i+1], "binary")
				}
			}
		}
		return nil
	}

	if err := ScaffoldMcp(targetDir, libraryDir, os.DirFS(libraryDir), nil, []string{"orchestrator"}, &[]types.TrackedFile{}, types.ConflictStrategyAlign, nil); err != nil {
		t.Fatalf("ScaffoldMcp: %v", err)
	}

	if len(calls) != 1 {
		t.Fatalf("runner calls = %#v, want one go build", calls)
	}
	if !strings.Contains(calls[0], "-C "+orchestratorDir+" build -o ") || !strings.Contains(calls[0], " ./cmd/orchestrator") {
		t.Fatalf("call = %q, want go build for orchestrator cmd", calls[0])
	}
}

func TestScaffoldMcp_OptionalOrchestratorSmokeTest(t *testing.T) {
	targetDir := t.TempDir()
	repoRoot := t.TempDir()
	libraryDir := filepath.Join(repoRoot, "library")
	orchestratorDir := filepath.Join(repoRoot, "packages", "orchestrator-go")
	t.Setenv("HOME", t.TempDir())
	mustWriteTestFile(t, filepath.Join(libraryDir, "mcp", "catalog.json"), `{"servers":{"orchestrator":{"command":"ai-setup-orchestrator","args":["connect"]}}}`)
	mustWriteTestFile(t, filepath.Join(orchestratorDir, "go.mod"), `module example.com/orchestrator`)
	t.Setenv("AI_SETUP_ORCHESTRATOR_SMOKE", "true")

	originalLookPath := orchestratorLookPath
	originalRunner := orchestratorCommandRunner
	t.Cleanup(func() {
		orchestratorLookPath = originalLookPath
		orchestratorCommandRunner = originalRunner
	})
	orchestratorLookPath = func(file string) (string, error) {
		if file == "go" {
			return "/fake/bin/go", nil
		}
		return "", os.ErrNotExist
	}
	var calls []string
	orchestratorCommandRunner = func(ctx context.Context, name string, args ...string) error {
		calls = append(calls, name+" "+strings.Join(args, " "))
		if name == "/fake/bin/go" {
			for i, arg := range args {
				if arg == "-o" && i+1 < len(args) {
					mustWriteTestFile(t, args[i+1], "binary")
				}
			}
		}
		return nil
	}

	if err := ScaffoldMcp(targetDir, libraryDir, os.DirFS(libraryDir), nil, []string{"orchestrator"}, &[]types.TrackedFile{}, types.ConflictStrategyAlign, nil); err != nil {
		t.Fatalf("ScaffoldMcp: %v", err)
	}

	if got := calls[len(calls)-1]; !strings.HasSuffix(got, "ai-setup-orchestrator status") {
		t.Fatalf("last call = %q, want smoke test command", got)
	}
}

func TestScaffoldMcp_ReportsMissingGoForOrchestrator(t *testing.T) {
	targetDir := t.TempDir()
	repoRoot := t.TempDir()
	libraryDir := filepath.Join(repoRoot, "library")
	orchestratorDir := filepath.Join(repoRoot, "packages", "orchestrator-go")
	mustWriteTestFile(t, filepath.Join(libraryDir, "mcp", "catalog.json"), `{"servers":{"orchestrator":{"command":"ai-setup-orchestrator","args":["connect"]}}}`)
	mustWriteTestFile(t, filepath.Join(orchestratorDir, "go.mod"), `module example.com/orchestrator`)

	originalLookPath := orchestratorLookPath
	t.Cleanup(func() { orchestratorLookPath = originalLookPath })
	orchestratorLookPath = func(file string) (string, error) { return "", os.ErrNotExist }

	err := ScaffoldMcp(targetDir, libraryDir, os.DirFS(libraryDir), nil, []string{"orchestrator"}, &[]types.TrackedFile{}, types.ConflictStrategyAlign, nil)
	if err == nil || !strings.Contains(err.Error(), "go executable not found") {
		t.Fatalf("err = %v, want missing go error", err)
	}
}

func TestScaffoldMcp_WritesInternalAndRootSchemas(t *testing.T) {
	targetDir := t.TempDir()
	libraryDir := t.TempDir()
	mustWriteTestFile(t, filepath.Join(libraryDir, "mcp", "catalog.json"), `{
  "servers": {
    "context7": {
      "command": "npx",
      "args": ["-y", "@upstash/context7-mcp"]
    }
  }
}`)

	var records []types.TrackedFile
	if err := ScaffoldMcp(targetDir, libraryDir, os.DirFS(libraryDir), nil, []string{"context7"}, &records, types.ConflictStrategyAlign, nil); err != nil {
		t.Fatalf("ScaffoldMcp: %v", err)
	}

	internalData, err := os.ReadFile(filepath.Join(targetDir, ".ai", "mcp.json"))
	if err != nil {
		t.Fatalf("read internal mcp: %v", err)
	}
	var internal map[string]json.RawMessage
	if err := json.Unmarshal(internalData, &internal); err != nil {
		t.Fatalf("unmarshal internal mcp: %v", err)
	}
	if _, ok := internal["servers"]; !ok {
		t.Fatalf("internal .ai/mcp.json keys = %#v, want top-level servers", internal)
	}
	if _, ok := internal["mcpServers"]; ok {
		t.Fatalf("internal .ai/mcp.json should not contain top-level mcpServers: %#v", internal)
	}

	rootData, err := os.ReadFile(filepath.Join(targetDir, ".mcp.json"))
	if err != nil {
		t.Fatalf("read root mcp: %v", err)
	}
	var root map[string]json.RawMessage
	if err := json.Unmarshal(rootData, &root); err != nil {
		t.Fatalf("unmarshal root mcp: %v", err)
	}
	if _, ok := root["mcpServers"]; !ok {
		t.Fatalf("root .mcp.json keys = %#v, want top-level mcpServers", root)
	}
	if _, ok := root["servers"]; ok {
		t.Fatalf("root .mcp.json should not contain top-level servers: %#v", root)
	}
}

func readScaffoldedMcpCatalog(t *testing.T, path string) mcpCatalog {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %q: %v", path, err)
	}
	var parsed mcpCatalog
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal %q: %v", path, err)
	}
	return parsed
}

func mustWriteTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %q: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %q: %v", path, err)
	}
}
