package cmd

import (
	"testing"

	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

func TestAddNonInteractiveMergesIntoExistingSetup(t *testing.T) {
	dir := t.TempDir()
	runSeedInit(t, dir, []types.ToolId{types.ToolIdOpenCode}, types.PresetLevelMinimal)
	withWorkingDir(t, dir)

	if _, _ = captureOutput(t, func() {
		if err := runAddNonInteractive(
			nil,
			[]string{string(types.AgentIdBuilder)},
			[]string{string(types.SkillIdPlan)},
		); err != nil {
			t.Fatalf("runAddNonInteractive: %v", err)
		}
	}); false {
	}

	storeData := readSeededStoreData(t, dir)
	assertToolSet(t, storeData.Config.Tools, types.ToolIdOpenCode)
	assertAgentSet(t, storeData.Selections.Agents, types.AgentIdBuilder)
	assertSkillSet(t, storeData.Selections.Skills, types.SkillIdPlan)
}

func assertToolSet(t *testing.T, got []types.ToolId, want ...types.ToolId) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("tool count = %d, want %d (%v)", len(got), len(want), got)
	}
	for _, item := range want {
		if !containsTool(got, item) {
			t.Fatalf("missing tool %q in %v", item, got)
		}
	}
}

func assertAgentSet(t *testing.T, got []types.AgentId, want ...types.AgentId) {
	t.Helper()
	for _, item := range want {
		if !containsAgent(got, item) {
			t.Fatalf("missing agent %q in %v", item, got)
		}
	}
}

func assertSkillSet(t *testing.T, got []types.SkillId, want ...types.SkillId) {
	t.Helper()
	for _, item := range want {
		if !containsSkill(got, item) {
			t.Fatalf("missing skill %q in %v", item, got)
		}
	}
}

func containsTool(items []types.ToolId, want types.ToolId) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}

func containsAgent(items []types.AgentId, want types.AgentId) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}

func containsSkill(items []types.SkillId, want types.SkillId) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}
