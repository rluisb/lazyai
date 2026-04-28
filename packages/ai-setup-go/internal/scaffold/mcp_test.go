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
	orchestratorDir := filepath.Join(repoRoot, "orchestrator")
	mustWriteTestFile(t, filepath.Join(libraryDir, "mcp", "catalog.json"), `{
  "servers": {
    "orchestrator": {
      "description": "orchestrator",
      "command": "npx",
      "args": ["-y", "@ai-setup/orchestrator"],
      "enabled": false,
      "requiresInstall": false
    }
  }
}`)
	mustWriteTestFile(t, filepath.Join(orchestratorDir, "package.json"), `{}`)
	entryPath := filepath.Join(orchestratorDir, "dist", "index.js")
	mustWriteTestFile(t, entryPath, `console.log("ok")`)

	originalLookPath := orchestratorLookPath
	originalRunner := orchestratorCommandRunner
	t.Cleanup(func() {
		orchestratorLookPath = originalLookPath
		orchestratorCommandRunner = originalRunner
	})
	orchestratorLookPath = func(file string) (string, error) {
		switch file {
		case "node":
			return "/fake/bin/node", nil
		case "npm":
			return "/fake/bin/npm", nil
		default:
			return "", os.ErrNotExist
		}
	}
	orchestratorCommandRunner = func(ctx context.Context, name string, args ...string) error { return nil }

	var records []types.TrackedFile
	if err := ScaffoldMcp(targetDir, libraryDir, os.DirFS(libraryDir), nil, []string{"orchestrator"}, &records, types.ConflictStrategyAlign, nil); err != nil {
		t.Fatalf("ScaffoldMcp: %v", err)
	}

	parsed := readScaffoldedMcpCatalog(t, filepath.Join(targetDir, ".ai", "mcp.json"))
	entry := parsed.Servers["orchestrator"]
	if entry.Command != "/fake/bin/node" {
		t.Fatalf("command = %q, want fake node path", entry.Command)
	}
	if want := []string{entryPath}; !reflect.DeepEqual(entry.Args, want) {
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
	orchestratorDir := filepath.Join(repoRoot, "orchestrator")
	mustWriteTestFile(t, filepath.Join(libraryDir, "mcp", "catalog.json"), `{
  "servers": {
    "orchestrator": {
      "command": "npx",
      "args": ["-y", "@ai-setup/orchestrator"]
    }
  }
}`)
	mustWriteTestFile(t, filepath.Join(orchestratorDir, "package.json"), `{}`)

	originalLookPath := orchestratorLookPath
	originalRunner := orchestratorCommandRunner
	t.Cleanup(func() {
		orchestratorLookPath = originalLookPath
		orchestratorCommandRunner = originalRunner
	})
	orchestratorLookPath = func(file string) (string, error) {
		switch file {
		case "node":
			return "/fake/bin/node", nil
		case "npm":
			return "/fake/bin/npm", nil
		default:
			return "", os.ErrNotExist
		}
	}
	var calls []string
	orchestratorCommandRunner = func(ctx context.Context, name string, args ...string) error {
		calls = append(calls, name+" "+strings.Join(args, " "))
		if name == "/fake/bin/npm" && len(args) >= 4 && args[2] == "run" && args[3] == "build" {
			mustWriteTestFile(t, filepath.Join(orchestratorDir, "dist", "index.js"), `console.log("built")`)
		}
		return nil
	}

	if err := ScaffoldMcp(targetDir, libraryDir, os.DirFS(libraryDir), nil, []string{"orchestrator"}, &[]types.TrackedFile{}, types.ConflictStrategyAlign, nil); err != nil {
		t.Fatalf("ScaffoldMcp: %v", err)
	}

	if len(calls) != 2 {
		t.Fatalf("runner calls = %#v, want install and build", calls)
	}
	if !strings.Contains(calls[0], "--prefix "+orchestratorDir+" install") {
		t.Fatalf("first call = %q, want npm install with --prefix", calls[0])
	}
	if !strings.Contains(calls[1], "--prefix "+orchestratorDir+" run build") {
		t.Fatalf("second call = %q, want npm run build with --prefix", calls[1])
	}
}

func TestScaffoldMcp_OptionalOrchestratorSmokeTest(t *testing.T) {
	targetDir := t.TempDir()
	repoRoot := t.TempDir()
	libraryDir := filepath.Join(repoRoot, "library")
	orchestratorDir := filepath.Join(repoRoot, "orchestrator")
	mustWriteTestFile(t, filepath.Join(libraryDir, "mcp", "catalog.json"), `{"servers":{"orchestrator":{"command":"npx","args":["-y","@ai-setup/orchestrator"]}}}`)
	mustWriteTestFile(t, filepath.Join(orchestratorDir, "package.json"), `{}`)
	entryPath := filepath.Join(orchestratorDir, "dist", "index.js")
	mustWriteTestFile(t, entryPath, `console.log("ok")`)
	t.Setenv("AI_SETUP_ORCHESTRATOR_SMOKE", "true")

	originalLookPath := orchestratorLookPath
	originalRunner := orchestratorCommandRunner
	t.Cleanup(func() {
		orchestratorLookPath = originalLookPath
		orchestratorCommandRunner = originalRunner
	})
	orchestratorLookPath = func(file string) (string, error) {
		if file == "node" {
			return "/fake/bin/node", nil
		}
		if file == "npm" {
			return "/fake/bin/npm", nil
		}
		return "", os.ErrNotExist
	}
	var calls []string
	orchestratorCommandRunner = func(ctx context.Context, name string, args ...string) error {
		calls = append(calls, name+" "+strings.Join(args, " "))
		return nil
	}

	if err := ScaffoldMcp(targetDir, libraryDir, os.DirFS(libraryDir), nil, []string{"orchestrator"}, &[]types.TrackedFile{}, types.ConflictStrategyAlign, nil); err != nil {
		t.Fatalf("ScaffoldMcp: %v", err)
	}

	if got := calls[len(calls)-1]; got != "/fake/bin/node "+entryPath+" catalog" {
		t.Fatalf("last call = %q, want smoke test command", got)
	}
}

func TestScaffoldMcp_ReportsMissingNodeForOrchestrator(t *testing.T) {
	targetDir := t.TempDir()
	repoRoot := t.TempDir()
	libraryDir := filepath.Join(repoRoot, "library")
	orchestratorDir := filepath.Join(repoRoot, "orchestrator")
	mustWriteTestFile(t, filepath.Join(libraryDir, "mcp", "catalog.json"), `{"servers":{"orchestrator":{"command":"npx","args":["-y","@ai-setup/orchestrator"]}}}`)
	mustWriteTestFile(t, filepath.Join(orchestratorDir, "package.json"), `{}`)
	mustWriteTestFile(t, filepath.Join(orchestratorDir, "dist", "index.js"), `console.log("ok")`)

	originalLookPath := orchestratorLookPath
	t.Cleanup(func() { orchestratorLookPath = originalLookPath })
	orchestratorLookPath = func(file string) (string, error) { return "", os.ErrNotExist }

	err := ScaffoldMcp(targetDir, libraryDir, os.DirFS(libraryDir), nil, []string{"orchestrator"}, &[]types.TrackedFile{}, types.ConflictStrategyAlign, nil)
	if err == nil || !strings.Contains(err.Error(), "node executable not found") {
		t.Fatalf("err = %v, want missing node error", err)
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
