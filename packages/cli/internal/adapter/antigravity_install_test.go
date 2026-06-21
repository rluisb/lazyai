package adapter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

const antigravityHooksJSON = `{
  "lazyai-block-destructive-shell": {
    "PreToolUse": [
      {
        "matcher": "run_command",
        "hooks": [
          {
            "type": "command",
            "command": ".gemini/hooks/lazyai/block-destructive-shell.sh",
            "timeout": 10
          }
        ]
      }
    ]
  },
  "lazyai-objective-workflow-gate": {
    "Stop": [
      {
        "type": "command",
        "command": ".gemini/hooks/lazyai/objective-workflow-gate.sh"
      }
    ]
  }
}
`

func TestAntigravityAdapter_Install_ProducesAgentSkillsSurfaceAtAgentsDir(t *testing.T) {
	targetDir := t.TempDir()
	libDir := t.TempDir()

	writeFile(t, filepath.Join(libDir, "antigravity", "settings.json"), "{}\n")
	writeFile(t, filepath.Join(libDir, "antigravity", "hooks", "lazyai", "block-destructive-shell.sh"), "#!/usr/bin/env bash\n")
	writeFile(t, filepath.Join(libDir, "antigravity", "hooks", "lazyai", "objective-workflow-gate.sh"), "#!/usr/bin/env bash\n")
	writeFile(t, filepath.Join(libDir, "antigravity", "hooks.json"), antigravityHooksJSON)
	writeFile(t, filepath.Join(libDir, "skills", "diagnose.md"), "# diagnose\n")
	writeFile(t, filepath.Join(libDir, "skills", "issue-triage.md"), "# issue triage\n")

	ctx := &AdapterContext{
		TargetDir:  targetDir,
		SetupScope: types.SetupScopeProject,
		LibraryDir: libDir,
		LibraryFS:  nil,
		Strategy:   types.ConflictStrategyAlign,
		Selections: AdapterSelections{
			Skills: []types.SkillId{types.SkillIdDiagnose, types.SkillIdIssueTriage},
		},
	}

	adapter := &AntigravityAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("Antigravity Install failed: %v", err)
	}

	assertFileExists(t, filepath.Join(targetDir, ".gemini", "settings.json"))
	assertFileExists(t, filepath.Join(targetDir, ".gemini", "hooks", "lazyai", "block-destructive-shell.sh"))
	assertFileExists(t, filepath.Join(targetDir, ".gemini", "hooks", "lazyai", "objective-workflow-gate.sh"))
	assertFileExists(t, filepath.Join(targetDir, ".agents", "hooks.json"))
	assertFileExists(t, filepath.Join(targetDir, ".agents", "skills", "diagnose", "SKILL.md"))
	assertFileExists(t, filepath.Join(targetDir, ".agents", "skills", "issue-triage", "SKILL.md"))

	content, err := os.ReadFile(filepath.Join(targetDir, ".agents", "hooks.json"))
	if err != nil {
		t.Fatalf("read hooks.json failed: %v", err)
	}
	if !strings.Contains(string(content), ".gemini/hooks/lazyai/block-destructive-shell.sh") {
		t.Fatalf("hooks.json missing block-destructive-shell hook command")
	}
	if _, err := os.Stat(filepath.Join(targetDir, ".agents", "agents")); err == nil {
		t.Fatalf("unexpected .agents/agents directory")
	}
}

func TestAntigravityAdapter_Install_UsesWorkspaceRootForAgentsAndGemini(t *testing.T) {
	workspaceRoot := t.TempDir()
	targetRepo := filepath.Join(workspaceRoot, "repo")
	if err := os.MkdirAll(targetRepo, 0o755); err != nil {
		t.Fatalf("mkdir repo failed: %v", err)
	}
	libDir := t.TempDir()

	writeFile(t, filepath.Join(libDir, "antigravity", "settings.json"), "{}\n")
	writeFile(t, filepath.Join(libDir, "antigravity", "hooks.json"), antigravityHooksJSON)
	writeFile(t, filepath.Join(libDir, "skills", "review.md"), "# review\n")

	ctx := &AdapterContext{
		TargetDir:     targetRepo,
		WorkspaceRoot: workspaceRoot,
		SetupScope:    types.SetupScopeWorkspace,
		LibraryDir:    libDir,
		LibraryFS:     nil,
		Strategy:      types.ConflictStrategyAlign,
		Selections: AdapterSelections{
			Skills: []types.SkillId{types.SkillIdReview},
		},
	}

	adapter := &AntigravityAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("Antigravity Install failed: %v", err)
	}

	assertFileExists(t, filepath.Join(workspaceRoot, ".gemini", "settings.json"))
	assertFileExists(t, filepath.Join(workspaceRoot, ".agents", "skills", "review", "SKILL.md"))
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected %s: %v", path, err)
	}
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s failed: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write %s failed: %v", path, err)
	}
}
