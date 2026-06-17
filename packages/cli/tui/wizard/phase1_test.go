package wizard

import (
	"errors"
	"reflect"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

func TestRunPhase1NonInteractiveDefaults(t *testing.T) {
	t.Parallel()

	defaults := &Phase1Result{
		Scope:         types.SetupScopeProject,
		Tools:         []types.ToolId{types.ToolIdOpenCode, types.ToolIdClaudeCode},
		Skills:        []types.SkillId{types.SkillIdDiagnose},
		Agents:        []types.AgentId{types.AgentIdImplementer},
		McpPreset:     McpPresetRecommended,
		ProjectName:   "demo-app",
		CliTools:      []string{"gh"},
		EnableServers: []string{"filesystem"},
	}

	result, action, err := RunPhase1(defaults, true)
	if err != nil {
		t.Fatalf("RunPhase1: %v", err)
	}
	if action != PhaseContinue {
		t.Fatalf("action = %v, want %v", action, PhaseContinue)
	}
	if !reflect.DeepEqual(result, defaults) {
		t.Fatalf("result = %#v, want %#v", result, defaults)
	}
}

func TestBuildPhase1Result(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		scope    types.SetupScope
		project  string
		wantName string
	}{
		{name: "project", scope: types.SetupScopeProject, project: "demo-app", wantName: "demo-app"},
		{name: "workspace", scope: types.SetupScopeWorkspace, project: "workspace-root", wantName: "workspace-root"},
		{name: "global", scope: types.SetupScopeGlobal, project: "ignored", wantName: "global"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tools := []types.ToolId{types.ToolIdOpenCode, types.ToolIdCopilot}
			skills := []types.SkillId{types.SkillIdDiagnose, types.SkillIdPrReview}
			agents := []types.AgentId{types.AgentIdImplementer, types.AgentIdReviewer}
			cliTools := []string{"gh"}
			servers := []string{"filesystem"}

			result := buildPhase1Result(tt.scope, tools, skills, agents, McpPresetFull, tt.project, cliTools, servers, "", "")

			if result.Scope != tt.scope {
				t.Fatalf("Scope = %q, want %q", result.Scope, tt.scope)
			}
			if result.ProjectName != tt.wantName {
				t.Fatalf("ProjectName = %q, want %q", result.ProjectName, tt.wantName)
			}
			if !reflect.DeepEqual(result.Tools, tools) {
				t.Fatalf("Tools = %#v, want %#v", result.Tools, tools)
			}
			if !reflect.DeepEqual(result.Skills, skills) {
				t.Fatalf("Skills = %#v, want %#v", result.Skills, skills)
			}
			if !reflect.DeepEqual(result.Agents, agents) {
				t.Fatalf("Agents = %#v, want %#v", result.Agents, agents)
			}
			if result.McpPreset != McpPresetFull {
				t.Fatalf("McpPreset = %q, want %q", result.McpPreset, McpPresetFull)
			}
			if !reflect.DeepEqual(result.CliTools, cliTools) {
				t.Fatalf("CliTools = %#v, want %#v", result.CliTools, cliTools)
			}
			if !reflect.DeepEqual(result.EnableServers, servers) {
				t.Fatalf("EnableServers = %#v, want %#v", result.EnableServers, servers)
			}

			tools[0] = types.ToolIdClaudeCode
			skills[0] = types.SkillIdCodebaseExploration
			agents[0] = types.AgentIdResearcher
			cliTools[0] = "ai-jail"
			servers[0] = "ai-memory"
			if result.Tools[0] != types.ToolIdOpenCode {
				t.Fatalf("result.Tools was not copied")
			}
			if result.Skills[0] != types.SkillIdDiagnose {
				t.Fatalf("result.Skills was not copied")
			}
			if result.Agents[0] != types.AgentIdImplementer {
				t.Fatalf("result.Agents was not copied")
			}
			if result.CliTools[0] != "gh" {
				t.Fatalf("result.CliTools was not copied")
			}
			if result.EnableServers[0] != "filesystem" {
				t.Fatalf("result.EnableServers was not copied")
			}
		})
	}
}

func TestValidateProjectName(t *testing.T) {
	t.Parallel()

	valid := []string{"demo", "demo-app", "project_1", "my project"}
	for _, name := range valid {
		name := name
		t.Run("valid/"+name, func(t *testing.T) {
			t.Parallel()
			if err := validateProjectName(name); err != nil {
				t.Fatalf("validateProjectName(%q) = %v, want nil", name, err)
			}
		})
	}

	invalid := []string{"", "   ", "bad/name", `bad\\name`, "..", "my..project", ".hidden", "demo-app ", "demo\t"}
	for _, name := range invalid {
		name := name
		t.Run("invalid/"+name, func(t *testing.T) {
			t.Parallel()
			if err := validateProjectName(name); err == nil {
				t.Fatalf("validateProjectName(%q) = nil, want error", name)
			}
		})
	}
}

func TestAskProjectNameGlobalImplicitName(t *testing.T) {
	t.Parallel()

	name, action, err := askProjectName("anything", nil, types.SetupScopeGlobal, phase1StepInfo{Current: 3, Total: 4, StepTitle: "Project Name"})
	if err != nil {
		t.Fatalf("askProjectName: %v", err)
	}
	if action != PhaseContinue {
		t.Fatalf("action = %v, want %v", action, PhaseContinue)
	}
	if name != "global" {
		t.Fatalf("name = %q, want global", name)
	}
}

func TestPreviousPhase1Step(t *testing.T) {
	t.Parallel()

	if got := previousPhase1Step(2, types.SetupScopeProject); got != 1 {
		t.Fatalf("previousPhase1Step(project tools) = %d, want 1", got)
	}
	if got := previousPhase1Step(8, types.SetupScopeGlobal); got != 6 {
		t.Fatalf("previousPhase1Step(global cli tools) = %d, want 6", got)
	}
	if got := previousPhase1Step(6, types.SetupScopeGlobal); got != 5 {
		t.Fatalf("previousPhase1Step(global servers) = %d, want 5", got)
	}
}

func TestPhase1StepInfoFor(t *testing.T) {
	t.Parallel()

	defaults := &Phase1Result{
		Scope:       types.SetupScopeProject,
		Tools:       []types.ToolId{types.ToolIdOpenCode, types.ToolIdClaudeCode},
		Skills:      []types.SkillId{types.SkillIdDiagnose, types.SkillIdPrReview},
		Agents:      []types.AgentId{types.AgentIdImplementer},
		McpPreset:   McpPresetRecommended,
		ProjectName: "demo-app",
	}

	info := phase1StepInfoFor(2, types.SetupScopeProject, defaults)
	if got, want := info.Title(), "Setup Context — 2/9: AI Tools (previous: opencode, claude-code)"; got != want {
		t.Fatalf("Title() = %q, want %q", got, want)
	}

	skillInfo := phase1StepInfoFor(3, types.SetupScopeProject, defaults)
	if got, want := skillInfo.Title(), "Setup Context — 3/9: Skills (previous: diagnose, pr-review)"; got != want {
		t.Fatalf("Title() = %q, want %q", got, want)
	}

	agentInfo := phase1StepInfoFor(4, types.SetupScopeProject, defaults)
	if got, want := agentInfo.Title(), "Setup Context — 4/9: Agents (previous: implementer)"; got != want {
		t.Fatalf("Title() = %q, want %q", got, want)
	}

	presetInfo := phase1StepInfoFor(5, types.SetupScopeProject, defaults)
	if got, want := presetInfo.Title(), "Setup Context — 5/9: MCP Preset (previous: recommended)"; got != want {
		t.Fatalf("Title() = %q, want %q", got, want)
	}

	globalInfo := phase1StepInfoFor(8, types.SetupScopeGlobal, defaults)
	if got, want := globalInfo.Title(), "Setup Context — 7/8: CLI Tools"; got != want {
		t.Fatalf("Title() = %q, want %q", got, want)
	}

	// Step 9 is the Project Identity step.
	identity := phase1StepInfoFor(9, types.SetupScopeProject, &Phase1Result{
		Organization: "Acme",
		Team:         "Platform",
	})
	if got, want := identity.Title(), "Setup Context — 9/9: Project Identity (optional) (previous: org=\"Acme\" team=\"Platform\")"; got != want {
		t.Fatalf("Title() = %q, want %q", got, want)
	}
}

func TestDetectInstalledCliTools(t *testing.T) {
	t.Parallel()

	original := cliToolLookPath
	t.Cleanup(func() { cliToolLookPath = original })

	cliToolLookPath = func(file string) (string, error) {
		switch file {
		case "gh":
			return "/usr/bin/gh", nil
		case "ai-jail":
			return "", errors.New("not found")
		default:
			return "", errors.New("unexpected")
		}
	}

	catalog := &McpCatalog{CliTools: map[string]CliTool{"gh": {}, "ai-jail": {}}}
	got := detectInstalledCliTools(catalog)
	if want := []string{"gh"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("detectInstalledCliTools() = %#v, want %#v", got, want)
	}
}

func TestToolOptionsForScope_ReturnsSupportedTools(t *testing.T) {
	t.Parallel()

	globalOpts := toolOptionsForScope(types.SetupScopeGlobal)
	if len(globalOpts) != 3 {
		t.Errorf("global options count = %d, want 3", len(globalOpts))
	}

	projectOpts := toolOptionsForScope(types.SetupScopeProject)
	if len(projectOpts) != 5 {
		t.Errorf("project options count = %d, want 5", len(projectOpts))
	}

	workspaceOpts := toolOptionsForScope(types.SetupScopeWorkspace)
	if len(workspaceOpts) != 5 {
		t.Errorf("workspace options count = %d, want 5", len(workspaceOpts))
	}
}

func TestFilterToolsByScope_DropsIncompatible(t *testing.T) {
	t.Parallel()

	tools := []types.ToolId{
		types.ToolIdClaudeCode,
		types.ToolIdCopilot,
		types.ToolId("gemini"),
	}
	got := filterToolsByScope(tools, types.SetupScopeGlobal)
	if len(got) != 2 {
		t.Errorf("filtered tools count = %d, want 2", len(got))
	}
}

func TestNewCliToolsSelectUsesPreSelectedWhenNoDefaults(t *testing.T) {
	t.Parallel()

	selectField := NewCliToolsSelect(nil, []string{"gh"})
	got, ok := selectField.GetValue().([]string)
	if !ok {
		t.Fatalf("GetValue() type = %T, want []string", selectField.GetValue())
	}
	if want := []string{"gh"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("GetValue() = %#v, want %#v", got, want)
	}

	selectField = NewCliToolsSelect([]string{"ai-jail"}, []string{"gh"})
	got = selectField.GetValue().([]string)
	if want := []string{"ai-jail"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("GetValue() with defaults = %#v, want %#v", got, want)
	}
}
