package adapter

import (
	"encoding/json"
	"os"
	"path/filepath"
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
		filepath.Join(targetDir, ".kiro", "agents", "guide.json"),
		filepath.Join(targetDir, ".kiro", "agents", "reviewer.json"),
		filepath.Join(targetDir, ".kiro", "skills", "diagnose", "SKILL.md"),
		filepath.Join(targetDir, ".kiro", "prompts", "plan.md"),
		filepath.Join(targetDir, ".kiro", "prompts", "implement.md"),
		filepath.Join(targetDir, ".kiro", "hooks", "block-destructive-shell.json"),
	}
	for _, path := range expectedPaths {
		assertExists(t, path)
	}
	assertMissing(t, filepath.Join(targetDir, ".kiro", "workflows"))

	// Verify reviewer.json is valid JSON with required fields.
	selectedAgentPath := filepath.Join(targetDir, ".kiro", "agents", "reviewer.json")
	data, err := os.ReadFile(selectedAgentPath)
	if err != nil {
		t.Fatalf("read selected agent: %v", err)
	}
	var agent struct {
		Name         string   `json:"name"`
		Description  string   `json:"description"`
		Tools        []string `json:"tools"`
		AllowedTools []string `json:"allowedTools"`
		Prompt       string   `json:"prompt"`
	}
	if err := json.Unmarshal(data, &agent); err != nil {
		t.Fatalf("reviewer.json is not valid JSON: %v\ncontent: %s", err, data)
	}
	if agent.Name == "" {
		t.Fatalf("reviewer.json missing name field")
	}
	if agent.Description == "" {
		t.Fatalf("reviewer.json missing description field")
	}
	if agent.Tools == nil {
		t.Fatalf("reviewer.json tools must be an array (got nil)")
	}
	if agent.AllowedTools == nil {
		t.Fatalf("reviewer.json allowedTools must be an array (got nil)")
	}
}

// TestKiroAdapter_RewriteAgentForKiro_AllowedToolsNeverSupersetOfTools verifies
// that allowedTools is always a subset of tools (B2-3 from #574 plan).
func TestKiroAdapter_RewriteAgentForKiro_AllowedToolsNeverSupersetOfTools(t *testing.T) {
	source := []byte("---\nname: analyst\ndescription: Read-only analyst.\ntools:\n  - read\n  - search\n---\n\n# Analyst\n\nYou analyse code.\n")

	out, err := RewriteAgentForKiro(source, nil)
	if err != nil {
		t.Fatalf("RewriteAgentForKiro: %v", err)
	}

	var agent struct {
		Name         string   `json:"name"`
		Description  string   `json:"description"`
		Tools        []string `json:"tools"`
		AllowedTools []string `json:"allowedTools"`
		Prompt       string   `json:"prompt"`
	}
	if err := json.Unmarshal(out, &agent); err != nil {
		t.Fatalf("output is not valid JSON: %v\ncontent: %s", err, out)
	}

	// Verify tool values are populated from the source tools: field.
	if len(agent.Tools) != 2 {
		t.Fatalf("expected 2 tools, got %d: %v", len(agent.Tools), agent.Tools)
	}
	if agent.Tools[0] != "read" || agent.Tools[1] != "search" {
		t.Fatalf("unexpected tools: %v", agent.Tools)
	}

	// allowedTools must never be a superset of tools.
	toolSet := make(map[string]bool, len(agent.Tools))
	for _, tool := range agent.Tools {
		toolSet[tool] = true
	}
	for _, allowed := range agent.AllowedTools {
		if !toolSet[allowed] {
			t.Errorf("allowedTools contains %q which is not in tools %v", allowed, agent.Tools)
		}
	}
}

// TestKiroAdapter_RewriteAgentForKiro_EmptyToolsWhenNotDeclared verifies that
// an agent source without a tools: field gets empty arrays for both fields.
func TestKiroAdapter_RewriteAgentForKiro_EmptyToolsWhenNotDeclared(t *testing.T) {
	source := []byte("---\nname: helper\ndescription: A helper agent.\n---\n\n# Helper\n\nHelp the user.\n")

	out, err := RewriteAgentForKiro(source, nil)
	if err != nil {
		t.Fatalf("RewriteAgentForKiro: %v", err)
	}

	var agent struct {
		Tools        []string `json:"tools"`
		AllowedTools []string `json:"allowedTools"`
	}
	if err := json.Unmarshal(out, &agent); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if len(agent.Tools) != 0 {
		t.Fatalf("expected empty tools array, got %v", agent.Tools)
	}
	if len(agent.AllowedTools) != 0 {
		t.Fatalf("expected empty allowedTools array, got %v", agent.AllowedTools)
	}
}
