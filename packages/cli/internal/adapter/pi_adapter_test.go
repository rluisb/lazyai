package adapter

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

func TestPiAdapter_Install_EmitsAgentsSkillsAndPrompts(t *testing.T) {
	targets := []string{".pi/extensions", ".pi/hooks"}

	ctx, targetDir := createTestAdapterContext(t)
	ctx.Selections = AdapterSelections{
		Agents: []types.AgentId{types.AgentIdResearcher, types.AgentIdImplementer},
		Skills: []types.SkillId{types.SkillIdIssueTriage},
	}
	if testFS, ok := ctx.LibraryFS.(fstest.MapFS); ok {
		testFS["prompts/plan.md"] = &fstest.MapFile{Data: []byte("# plan")}
		testFS["prompts/research.md"] = &fstest.MapFile{Data: []byte("# research")}
	}

	adapter := &PiAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("Pi Install failed: %v", err)
	}

	for _, rel := range []string{
		".pi/agents/researcher.md",
		".pi/agents/implementer.md",
		".pi/skills/issue-triage/SKILL.md",
		".pi/prompts/plan.md",
		".pi/prompts/research.md",
	} {
		assertExists(t, filepath.Join(targetDir, rel))
	}
	for _, rel := range targets {
		assertMissing(t, filepath.Join(targetDir, rel))
	}

	skillsDir := filepath.Join(targetDir, ".pi", "skills")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		t.Fatalf("read skills dir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected exactly one skill emitted, found %d", len(entries))
	}
	if got := entries[0].Name(); got != string(types.SkillIdIssueTriage) {
		t.Fatalf("expected only issue-triage skill, found %q", got)
	}
}
