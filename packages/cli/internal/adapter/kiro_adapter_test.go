package adapter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

func TestKiroAdapter_Install_EmitsAgentProfilesAndSkills(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)
	ctx.Selections = AdapterSelections{
		Agents: []types.AgentId{types.AgentIdReviewer},
		Skills: []types.SkillId{types.SkillIdDiagnose},
	}

	adapter := &KiroAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("Kiro Install failed: %v", err)
	}

	defaultAgentPath := filepath.Join(targetDir, ".kiro", "agents", "guide.md")
	selectedAgentPath := filepath.Join(targetDir, ".kiro", "agents", "reviewer.md")
	selectedSkillPath := filepath.Join(targetDir, ".kiro", "skills", "diagnose", "SKILL.md")

	for _, path := range []string{defaultAgentPath, selectedAgentPath, selectedSkillPath} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected %s: %v", path, err)
		}
	}

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
