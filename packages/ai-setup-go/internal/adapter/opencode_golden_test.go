package adapter

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// Run with -update to regenerate golden files after an intentional schema change:
//
//	go test ./internal/adapter/... -run TestOpenCodeConfig_Golden -update
var update = flag.Bool("update", false, "overwrite golden files with current output")

// goldenPath returns the path to a golden file relative to the test source.
func goldenPath(name string) string {
	return filepath.Join("testdata", "golden", name)
}

// checkGolden reads back the file at gotPath and compares it to the named
// golden file. When -update is set it writes gotPath's content to the golden
// file instead of comparing.
func checkGolden(t *testing.T, gotPath, golden string) {
	t.Helper()
	got, err := os.ReadFile(gotPath)
	if err != nil {
		t.Fatalf("read generated file %s: %v", gotPath, err)
	}
	gp := goldenPath(golden)
	if *update {
		if err := os.MkdirAll(filepath.Dir(gp), 0o755); err != nil {
			t.Fatalf("mkdir golden dir: %v", err)
		}
		if err := os.WriteFile(gp, got, 0o644); err != nil {
			t.Fatalf("write golden %s: %v", gp, err)
		}
		t.Logf("updated golden: %s", gp)
		return
	}
	want, err := os.ReadFile(gp)
	if err != nil {
		t.Fatalf("read golden %s: %v\n(run with -update to create it)", gp, err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("output does not match golden %s\n\n--- got ---\n%s\n--- want ---\n%s",
			golden, got, want)
	}
}

// TestOpenCodeConfig_Golden verifies the exact on-disk content of the
// generated opencode.jsonc at install time and after MCP compile. These tests
// act as a schema regression guard: if the output shape changes, the test
// fails and requires an explicit golden update (-update flag).
func TestOpenCodeConfig_Golden(t *testing.T) {
	t.Run("install/project", func(t *testing.T) {
		targetDir := t.TempDir()
		ctx := &AdapterContext{
			TargetDir:  targetDir,
			HomeDir:    t.TempDir(),
			SetupScope: types.SetupScopeProject,
			LibraryFS:  createTestFS(),
			Strategy:   types.ConflictStrategyAlign,
			Selections: AdapterSelections{
				Agents: []types.AgentId{"builder"},
				Skills: []types.SkillId{"implement"},
			},
		}
		if _, err := (&OpenCodeAdapter{}).Install(ctx); err != nil {
			t.Fatalf("Install: %v", err)
		}
		checkGolden(t,
			filepath.Join(targetDir, "opencode.jsonc"),
			"opencode_install_project.jsonc",
		)
	})

	t.Run("install/global", func(t *testing.T) {
		homeDir := t.TempDir()
		ctx := &AdapterContext{
			TargetDir:  t.TempDir(),
			HomeDir:    homeDir,
			SetupScope: types.SetupScopeGlobal,
			LibraryFS:  createTestFS(),
			Strategy:   types.ConflictStrategyAlign,
			Selections: AdapterSelections{
				Agents: []types.AgentId{"builder"},
				Skills: []types.SkillId{"implement"},
			},
		}
		if _, err := (&OpenCodeAdapter{}).Install(ctx); err != nil {
			t.Fatalf("Install: %v", err)
		}
		checkGolden(t,
			filepath.Join(homeDir, ".config", "opencode", "opencode.jsonc"),
			"opencode_install_global.jsonc",
		)
	})

	// Deterministic catalog used by both mcp_compile tests.
	const testCatalog = `{
  "servers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem"],
      "enabled": true
    },
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {"GITHUB_PERSONAL_ACCESS_TOKEN": "${GITHUB_TOKEN}"},
      "enabled": true
    }
  }
}`

	// mcp_compile/standalone: compile without a prior install.
	// Verifies that compile alone produces a minimal but valid config.
	t.Run("mcp_compile/standalone", func(t *testing.T) {
		targetDir := t.TempDir()
		aiDir := filepath.Join(targetDir, ".ai")
		if err := files.EnsureDir(aiDir); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(aiDir, "mcp.json"), []byte(testCatalog), 0o644); err != nil {
			t.Fatal(err)
		}

		if _, err := CompileMCPForTool(types.ToolIdOpenCode, CompileContext{
			TargetDir:  targetDir,
			HomeDir:    t.TempDir(),
			SetupScope: types.SetupScopeProject,
		}); err != nil {
			t.Fatalf("CompileMCPForTool: %v", err)
		}
		checkGolden(t,
			filepath.Join(targetDir, ".opencode", "opencode.jsonc"),
			"opencode_mcp_standalone.jsonc",
		)
	})

	// mcp_compile/after_install: install then compile — the normal user flow.
	// This golden captures the full config shape a user will actually see.
	t.Run("mcp_compile/after_install", func(t *testing.T) {
		targetDir := t.TempDir()
		homeDir := t.TempDir()
		aiDir := filepath.Join(targetDir, ".ai")
		if err := files.EnsureDir(aiDir); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(aiDir, "mcp.json"), []byte(testCatalog), 0o644); err != nil {
			t.Fatal(err)
		}

		ctx := &AdapterContext{
			TargetDir:  targetDir,
			HomeDir:    homeDir,
			SetupScope: types.SetupScopeProject,
			LibraryFS:  createTestFS(),
			Strategy:   types.ConflictStrategyAlign,
			Selections: AdapterSelections{
				Agents: []types.AgentId{"builder"},
				Skills: []types.SkillId{"implement"},
			},
		}
		if _, err := (&OpenCodeAdapter{}).Install(ctx); err != nil {
			t.Fatalf("Install: %v", err)
		}
		if _, err := CompileMCPForTool(types.ToolIdOpenCode, CompileContext{
			TargetDir:  targetDir,
			HomeDir:    homeDir,
			SetupScope: types.SetupScopeProject,
		}); err != nil {
			t.Fatalf("CompileMCPForTool: %v", err)
		}
		checkGolden(t,
			filepath.Join(targetDir, ".opencode", "opencode.jsonc"),
			"opencode_mcp_after_install.jsonc",
		)
	})
}
