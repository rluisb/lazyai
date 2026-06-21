package adapter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

func TestKiroAdapter_Install_EmitsAgentProfilesSkillsAndPrompts(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)
	libFS, ok := ctx.LibraryFS.(fstest.MapFS)
	if !ok {
		t.Fatalf("expected test library fs")
	}
	libFS["prompts/plan.md"] = &fstest.MapFile{
		Data: []byte("---\nname: plan\n---\n\n# plan\n\nPlanning prompt body.\n"),
	}
	libFS["prompts/implement.md"] = &fstest.MapFile{
		Data: []byte("---\nname: implement\n---\n\n# implement\n\nImplement prompt body.\n"),
	}
	ctx.Selections = AdapterSelections{
		Agents:  []types.AgentId{types.AgentIdReviewer},
		Skills:  []types.SkillId{types.SkillIdDiagnose},
		Prompts: []types.PromptId{"plan", "implement"},
	}

	adapter := &KiroAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("Kiro Install failed: %v", err)
	}

	expectedPaths := []string{
		filepath.Join(targetDir, ".kiro", "agents", "guide.md"),
		filepath.Join(targetDir, ".kiro", "agents", "reviewer.md"),
		filepath.Join(targetDir, ".kiro", "skills", "diagnose", "SKILL.md"),
		filepath.Join(targetDir, ".kiro", "prompts", "plan.md"),
		filepath.Join(targetDir, ".kiro", "prompts", "implement.md"),
	}
	for _, path := range expectedPaths {
		assertExists(t, path)
	}

	selectedAgentPath := filepath.Join(targetDir, ".kiro", "agents", "reviewer.md")
	data, err := os.ReadFile(selectedAgentPath)
	if err != nil {
		t.Fatalf("read selected agent: %v", err)
	}
	got := string(data)
	if !strings.Contains(got, "description:") {
		t.Fatalf("Kiro agent profile missing description frontmatter: %q", got)
	}
	if !strings.Contains(got, "\n---\n\n# ") {
		t.Fatalf("Kiro agent profile missing markdown body after frontmatter: %q", got)
	}
}
