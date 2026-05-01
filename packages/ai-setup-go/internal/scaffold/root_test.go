package scaffold

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// TestScaffoldCompiledRoot_GlobalRequiresHomeDir verifies that passing an empty
// HomeDir at global scope returns an error rather than falling through to the
// real os.UserHomeDir() (R-3 mitigation from spec 008 risks).
func TestScaffoldCompiledRoot_GlobalRequiresHomeDir(t *testing.T) {
	err := ScaffoldCompiledRoot(ScaffoldCompiledRootOptions{
		TargetDir:  t.TempDir(),
		HomeDir:    "", // intentionally empty
		SetupScope: types.SetupScopeGlobal,
		Tools:      []types.ToolId{types.ToolIdClaudeCode},
	})
	if err == nil {
		t.Fatal("expected error for empty HomeDir at global scope, got nil")
	}
	if !strings.Contains(err.Error(), "HomeDir must be set") {
		t.Errorf("expected 'HomeDir must be set' message, got: %v", err)
	}
}

func TestMemoryDocDestPath(t *testing.T) {
	target := "/tmp/target"
	home := "/tmp/home"

	cases := []struct {
		name  string
		tool  types.ToolId
		scope types.SetupScope
		want  string
		unsup bool
	}{
		// claude-code
		{"claude_project", types.ToolIdClaudeCode, types.SetupScopeProject, filepath.Join(target, "AGENTS.md"), false},
		{"claude_workspace", types.ToolIdClaudeCode, types.SetupScopeWorkspace, filepath.Join(target, "AGENTS.md"), false},
		{"claude_global", types.ToolIdClaudeCode, types.SetupScopeGlobal, filepath.Join(home, ".claude", "AGENTS.md"), false},
		// opencode
		{"opencode_project", types.ToolIdOpenCode, types.SetupScopeProject, filepath.Join(target, "AGENTS.md"), false},
		{"opencode_global", types.ToolIdOpenCode, types.SetupScopeGlobal, filepath.Join(home, ".config", "opencode", "AGENTS.md"), false},
		// copilot
		{"copilot_project", types.ToolIdCopilot, types.SetupScopeProject, filepath.Join(target, ".github", "copilot-instructions.md"), false},
		{"copilot_workspace", types.ToolIdCopilot, types.SetupScopeWorkspace, filepath.Join(target, ".github", "copilot-instructions.md"), false},
		{"copilot_global", types.ToolIdCopilot, types.SetupScopeGlobal, "", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			outputFile := RootFileByTool[c.tool]
			got, err := memoryDocDestPath(c.tool, c.scope, target, home, outputFile)
			if c.unsup {
				if !errors.Is(err, errMemoryDocScopeUnsupported) {
					t.Fatalf("want errMemoryDocScopeUnsupported, got err=%v path=%q", err, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestScaffoldCompiledRootClaudeUsesAgentsWithoutCreatingClaude(t *testing.T) {
	targetDir := t.TempDir()
	records := []types.TrackedFile{}

	if err := ScaffoldCompiledRoot(ScaffoldCompiledRootOptions{
		TargetDir:        targetDir,
		LibraryFS:        rootTestFS(),
		Tools:            []types.ToolId{types.ToolIdClaudeCode},
		ProjectName:      "test-project",
		FileRecords:      &records,
		Strategy:         types.ConflictStrategySkip,
		PerFileOverrides: map[string]types.ConflictStrategy{},
		SetupScope:       types.SetupScopeProject,
	}); err != nil {
		t.Fatalf("ScaffoldCompiledRoot: %v", err)
	}

	if _, err := os.Stat(filepath.Join(targetDir, "AGENTS.md")); err != nil {
		t.Fatalf("expected AGENTS.md: %v", err)
	}
	if _, err := os.Stat(filepath.Join(targetDir, "CLAUDE.md")); !os.IsNotExist(err) {
		t.Fatalf("CLAUDE.md should not be created, stat err=%v", err)
	}
}

func TestScaffoldCompiledRootAppendsClaudeAgentsReferenceOnce(t *testing.T) {
	targetDir := t.TempDir()
	claudePath := filepath.Join(targetDir, "CLAUDE.md")
	if err := os.WriteFile(claudePath, []byte("# Existing Claude\n"), 0o644); err != nil {
		t.Fatalf("seed CLAUDE.md: %v", err)
	}

	for i := 0; i < 2; i++ {
		records := []types.TrackedFile{}
		if err := ScaffoldCompiledRoot(ScaffoldCompiledRootOptions{
			TargetDir:        targetDir,
			LibraryFS:        rootTestFS(),
			Tools:            []types.ToolId{types.ToolIdClaudeCode},
			ProjectName:      "test-project",
			FileRecords:      &records,
			Strategy:         types.ConflictStrategySkip,
			PerFileOverrides: map[string]types.ConflictStrategy{},
			SetupScope:       types.SetupScopeProject,
		}); err != nil {
			t.Fatalf("ScaffoldCompiledRoot pass %d: %v", i, err)
		}
	}

	got, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	content := string(got)
	if !strings.Contains(content, "# Existing Claude") {
		t.Fatalf("existing content was not preserved: %q", content)
	}
	if count := strings.Count(content, claudeAgentsReference); count != 1 {
		t.Fatalf("reference count = %d, want 1\n%s", count, content)
	}
}

func TestScaffoldCompiledRootCopilotUsesRootTemplate(t *testing.T) {
	targetDir := t.TempDir()
	records := []types.TrackedFile{}

	if err := ScaffoldCompiledRoot(ScaffoldCompiledRootOptions{
		TargetDir:        targetDir,
		LibraryFS:        rootTestFS(),
		Tools:            []types.ToolId{types.ToolIdCopilot},
		ProjectName:      "test-project",
		FileRecords:      &records,
		Strategy:         types.ConflictStrategySkip,
		PerFileOverrides: map[string]types.ConflictStrategy{},
		SetupScope:       types.SetupScopeProject,
	}); err != nil {
		t.Fatalf("ScaffoldCompiledRoot: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(targetDir, ".github", "copilot-instructions.md"))
	if err != nil {
		t.Fatalf("read copilot instructions: %v", err)
	}
	if !strings.Contains(string(got), "Copilot root template") {
		t.Fatalf("copilot instructions did not use root template: %q", string(got))
	}
}

func rootTestFS() fstest.MapFS {
	return fstest.MapFS{
		"root/AGENTS.template.md":               &fstest.MapFile{Data: []byte("# AGENTS {{projectName}}")},
		"root/copilot-instructions.template.md": &fstest.MapFile{Data: []byte("# Copilot root template")},
	}
}
