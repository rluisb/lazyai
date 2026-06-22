package wizard

import (
	"strings"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

func TestFormatDryRunSummaryIncludesChoicesAndPlannedFiles(t *testing.T) {
	useReversa := false
	result := &WizardResult{
		Phase1: &Phase1Result{
			Scope:         types.SetupScopeProject,
			Tools:         []types.ToolId{types.ToolIdOpenCode, types.ToolIdPi},
			Skills:        []types.SkillId{types.SkillIdArchitectureReview},
			Agents:        []types.AgentId{types.AgentIdReviewer, types.AgentIdPlanner},
			McpPreset:     McpPresetRecommended,
			ProjectName:   "demo",
			CliTools:      []string{"ai-memory"},
			EnableServers: []string{"filesystem", "ai-memory"},
		},
		Phase2: &Phase2Result{
			Preset: types.PresetLevelStandard,
			GitConv: &types.GitConventions{
				BranchPattern: "{type}/{ticket}-{description}",
				CommitPattern: "{type}: {description}",
				RequireTicket: true,
			},
			UseReversa: &useReversa,
		},
		Phase5: &Phase5Result{
			MemoryPath:        ".specify/memory",
			EnableCodegraph:   true,
			CodegraphDataPath: ".codegraph/",
		},
	}
	plan := &InstallPlan{FilesToInstall: []PlannedFile{
		{Type: "agent"},
		{Type: "mcp", Existing: true},
	}}

	summary := formatDryRunSummary(plan, result)
	for _, want := range []string{
		"Scope: project",
		"Project: demo",
		"AI tools: opencode, pi",
		"Skills: 1 selected",
		"Agents: 2 selected",
		"MCP servers: ai-memory, filesystem",
		"Require ticket: true",
		"Project analysis: disabled",
		"Codegraph: true (.codegraph/)",
		"Agent definitions: 1 new",
		"MCP configuration: 1 updates",
		"Files: 1 new, 1 updates",
	} {
		if !strings.Contains(summary, want) {
			t.Fatalf("formatDryRunSummary() missing %q in:\n%s", want, summary)
		}
	}
}

func TestOptionDescriptionFallsBackForUnknownHoveredOption(t *testing.T) {
	got := optionDescription("custom-tool", nil, defaultHoverHint)
	want := "custom-tool: no extra setup beyond selecting this item."
	if got != want {
		t.Fatalf("optionDescription() = %q, want %q", got, want)
	}
}
