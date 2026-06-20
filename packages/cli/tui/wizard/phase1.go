package wizard

import (
	"fmt"
	"strings"
	"unicode"

	"charm.land/huh/v2"
	"github.com/rluisb/lazyai/packages/cli/internal/theme"

	"github.com/rluisb/lazyai/packages/cli/internal/adapter"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

const phase1BackValue = "__phase1_back__"

type phase1StepInfo struct {
	Current   int
	Total     int
	StepTitle string
	Previous  string
}

func (s phase1StepInfo) Title() string {
	title := fmt.Sprintf("Setup Context — %d/%d: %s", s.Current, s.Total, s.StepTitle)
	if s.Previous == "" {
		return title
	}
	return fmt.Sprintf("%s (previous: %s)", title, s.Previous)
}

// Phase1Result holds the collected context from the first wizard phase.
type Phase1Result struct {
	Scope         types.SetupScope
	Tools         []types.ToolId
	Skills        []types.SkillId
	Agents        []types.AgentId
	McpPreset     McpPreset
	ProjectName   string
	CliTools      []string
	EnableServers []string
	// Organization and Team are optional identity fields that populate the
	// [YOUR_ORG] / [YOUR_TEAM] placeholders in the generated AGENTS.md.
	// Empty → left as <!-- fill-in --> markers.
	Organization string
	Team         string
}

// RunPhase1 runs the context collection phase.
//
// If nonInteractive is true, required fields must be pre-populated in
// defaults (from CLI flags); otherwise interactive prompts are presented.
//
// The action indicates what the user wants to do next:
//   - PhaseContinue: proceed to Phase 2
//   - PhaseBack: go back (not applicable for Phase 1, but kept for consistency)
//   - PhaseCancel: user cancelled
func RunPhase1(defaults *Phase1Result, nonInteractive bool) (*Phase1Result, PhaseAction, error) {
	if nonInteractive {
		return runPhase1NonInteractive(defaults)
	}
	return runPhase1Interactive(defaults)
}

func runPhase1NonInteractive(defaults *Phase1Result) (*Phase1Result, PhaseAction, error) {
	if defaults == nil {
		return nil, PhaseCancel, fmt.Errorf("non-interactive mode requires defaults with scope, tools, and project name")
	}
	if defaults.Scope == "" {
		return nil, PhaseCancel, fmt.Errorf("--scope is required in non-interactive mode (global | workspace | project)")
	}
	if len(defaults.Tools) == 0 {
		return nil, PhaseCancel, fmt.Errorf("--tools is required in non-interactive mode (opencode, claude-code, copilot, pi, antigravity)")
	}
	if defaults.ProjectName == "" {
		return nil, PhaseCancel, fmt.Errorf("project name is required in non-interactive mode")
	}

	if defaults.McpPreset == "" {
		defaults.McpPreset = McpPresetRecommended
	}
	if len(defaults.EnableServers) == 0 {
		defaults.EnableServers = defaultMcpServersForPreset(defaults.McpPreset)
	}
	if len(defaults.CliTools) == 0 {
		defaults.CliTools = detectInstalledCliToolsFromCatalog()
	}
	if len(defaults.Skills) == 0 {
		defaults.Skills = types.ALL_SKILLS
	}
	if len(defaults.Agents) == 0 {
		defaults.Agents = types.ALL_AGENTS
	}

	return defaults, PhaseContinue, nil
}

func runPhase1Interactive(defaults *Phase1Result) (*Phase1Result, PhaseAction, error) {
	state := phase1InteractiveState{Scope: defaultPhase1Scope()}
	if defaults != nil {
		state.Scope = defaults.Scope
		state.Tools = cloneToolIDs(defaults.Tools)
		state.Skills = cloneSkillIDs(defaults.Skills)
		state.Agents = cloneAgentIDs(defaults.Agents)
		state.McpPreset = defaults.McpPreset
		state.ProjectName = defaults.ProjectName
		state.CliTools = cloneStrings(defaults.CliTools)
		state.Servers = cloneStrings(defaults.EnableServers)
		state.Organization = defaults.Organization
		state.Team = defaults.Team
	}
	if state.Scope == "" {
		state.Scope = defaultPhase1Scope()
	}
	if state.ProjectName == "" {
		state.ProjectName = defaultPhase1ProjectName()
	}
	if state.McpPreset == "" {
		state.McpPreset = McpPresetRecommended
	}

	currentStep := 1
	for currentStep >= 1 && currentStep <= 9 {
		scopeForStep := state.Scope
		if scopeForStep == "" {
			scopeForStep = defaultPhase1Scope()
		}

		switch currentStep {
		case 1:
			scope, action, err := askScope(state.Scope, phase1StepInfoFor(1, scopeForStep, defaults))
			if err != nil {
				return nil, action, err
			}
			state.Scope = scope
			if action == PhaseBack {
				return nil, PhaseBack, nil
			}
			currentStep++
		case 2:
			tools, action, err := askTools(state.Tools, scopeForStep, phase1StepInfoFor(2, scopeForStep, defaults))
			if err != nil {
				return nil, action, err
			}
			if action == PhaseBack {
				currentStep = previousPhase1Step(currentStep, state.Scope)
				continue
			}
			state.Tools = tools
			currentStep++
		case 3:
			skills, action, err := askSkills(state.Skills, phase1StepInfoFor(3, scopeForStep, defaults))
			if err != nil {
				return nil, action, err
			}
			if action == PhaseBack {
				currentStep = previousPhase1Step(currentStep, state.Scope)
				continue
			}
			state.Skills = skills
			currentStep++
		case 4:
			agents, action, err := askAgents(state.Agents, phase1StepInfoFor(4, scopeForStep, defaults))
			if err != nil {
				return nil, action, err
			}
			if action == PhaseBack {
				currentStep = previousPhase1Step(currentStep, state.Scope)
				continue
			}
			state.Agents = agents
			currentStep++
		case 5:
			mcpPreset, action, err := askMcpPreset(state.McpPreset, phase1StepInfoFor(5, scopeForStep, defaults))
			if err != nil {
				return nil, action, err
			}
			if action == PhaseBack {
				currentStep = previousPhase1Step(currentStep, state.Scope)
				continue
			}
			state.McpPreset = mcpPreset
			if !state.mcpSelectionCustomized {
				state.Servers = defaultMcpServersForPreset(mcpPreset)
			}
			currentStep++
		case 6:
			servers, action, err := askMcpServers(defaultMcpSelection(state.Servers, state.McpPreset), phase1StepInfoFor(6, scopeForStep, defaults))
			if err != nil {
				return nil, action, err
			}
			if action == PhaseBack {
				currentStep = previousPhase1Step(currentStep, state.Scope)
				continue
			}
			state.Servers = servers
			state.mcpSelectionCustomized = true
			currentStep++
		case 7:
			name, action, err := askProjectName(state.ProjectName, defaults, state.Scope, phase1StepInfoFor(7, scopeForStep, defaults))
			if err != nil {
				return nil, action, err
			}
			if action == PhaseBack {
				currentStep = previousPhase1Step(currentStep, state.Scope)
				continue
			}
			state.ProjectName = name
			currentStep++
		case 8:
			cliTools, action, err := askCliTools(state.CliTools, phase1StepInfoFor(8, scopeForStep, defaults))
			if err != nil {
				return nil, action, err
			}
			if action == PhaseBack {
				currentStep = previousPhase1Step(currentStep, state.Scope)
				continue
			}
			state.CliTools = cliTools
			currentStep++
		case 9:
			org, team, action, err := askProjectIdentity(state.Organization, state.Team, phase1StepInfoFor(9, scopeForStep, defaults))
			if err != nil {
				return nil, action, err
			}
			if action == PhaseBack {
				currentStep = previousPhase1Step(currentStep, state.Scope)
				continue
			}
			state.Organization = org
			state.Team = team
			currentStep++
		}
	}

	return buildPhase1Result(state.Scope, state.Tools, state.Skills, state.Agents, state.McpPreset, state.ProjectName, state.CliTools, state.Servers, state.Organization, state.Team), PhaseContinue, nil
}

type phase1InteractiveState struct {
	Scope                  types.SetupScope
	Tools                  []types.ToolId
	Skills                 []types.SkillId
	Agents                 []types.AgentId
	McpPreset              McpPreset
	ProjectName            string
	CliTools               []string
	Servers                []string
	Organization           string
	Team                   string
	mcpSelectionCustomized bool
}

func buildPhase1Result(scope types.SetupScope, tools []types.ToolId, skills []types.SkillId, agents []types.AgentId, mcpPreset McpPreset, name string, cliTools, servers []string, org, team string) *Phase1Result {
	projectName := name
	if scope == types.SetupScopeGlobal {
		projectName = "global"
	}

	return &Phase1Result{
		Scope:         scope,
		Tools:         cloneToolIDs(tools),
		Skills:        cloneSkillIDs(skills),
		Agents:        cloneAgentIDs(agents),
		McpPreset:     normalizeMcpPreset(mcpPreset),
		ProjectName:   projectName,
		CliTools:      cloneStrings(cliTools),
		EnableServers: cloneStrings(servers),
		Organization:  org,
		Team:          team,
	}
}

func askScope(current types.SetupScope, info phase1StepInfo) (types.SetupScope, PhaseAction, error) {
	scopeValue := string(current)
	if scopeValue == "" {
		scopeValue = string(defaultPhase1Scope())
	}

	field := huh.NewSelect[string]().
		Title(info.Title()).
		Options(
			huh.NewOption("Global  — Install to ~/.config/opencode/ + native tool global paths", "global"),
			huh.NewOption("Workspace  — Planning repo with multi-project management", "workspace"),
			huh.NewOption("Project (recommended)  — Self-contained single repository", "project"),
		).
		Value(&scopeValue)

	if err := theme.NewForm(huh.NewGroup(field)).Run(); err != nil {
		return "", PhaseCancel, fmt.Errorf("phase 1 cancelled: %w", err)
	}

	return types.SetupScope(scopeValue), PhaseContinue, nil
}

func askTools(current []types.ToolId, scope types.SetupScope, info phase1StepInfo) ([]types.ToolId, PhaseAction, error) {
	selectedTools := toolIDsToStrings(filterToolsByScope(current, scope))
	field := huh.NewMultiSelect[string]().
		Title(info.Title()).
		Options(appendPhase1BackOption(toolOptionsForScope(scope))...).
		Value(&selectedTools)

	if err := theme.NewForm(huh.NewGroup(field)).Run(); err != nil {
		return nil, PhaseCancel, fmt.Errorf("phase 1 cancelled: %w", err)
	}
	if containsString(selectedTools, phase1BackValue) {
		return nil, PhaseBack, nil
	}

	return stringsToToolIDs(selectedTools), PhaseContinue, nil
}

// toolOptionsForScope returns the AI tool multi-select options filtered to
// only tools that support the given scope (see adapter.IsScopeSupported).
func toolOptionsForScope(scope types.SetupScope) []huh.Option[string] {
	all := []huh.Option[string]{
		huh.NewOption("OpenCode", string(types.ToolIdOpenCode)),
		huh.NewOption("Claude Code", string(types.ToolIdClaudeCode)),
		huh.NewOption("GitHub Copilot", string(types.ToolIdCopilot)),
		huh.NewOption("OMP/Pi", string(types.ToolIdPi)),
		huh.NewOption("OMP", string(types.ToolIdOmp)),
		huh.NewOption("Kiro", string(types.ToolIdKiro)),
		huh.NewOption("Antigravity", string(types.ToolIdAntigravity)),
	}
	if scope == "" {
		return all
	}
	out := make([]huh.Option[string], 0, len(all))
	for _, opt := range all {
		if adapter.IsScopeSupported(types.ToolId(opt.Value), scope) {
			out = append(out, opt)
		}
	}
	return out
}

// filterToolsByScope drops previously-selected tool IDs that are not supported
// at the current scope. Prevents stale selections from bypassing the scope filter
// when the user navigates back and changes scope.
func filterToolsByScope(tools []types.ToolId, scope types.SetupScope) []types.ToolId {
	if scope == "" {
		return tools
	}
	out := make([]types.ToolId, 0, len(tools))
	for _, t := range tools {
		if adapter.IsScopeSupported(t, scope) {
			out = append(out, t)
		}
	}
	return out
}

func askProjectName(current string, defaults *Phase1Result, scope types.SetupScope, info phase1StepInfo) (string, PhaseAction, error) {
	if scope == types.SetupScopeGlobal {
		return "global", PhaseContinue, nil
	}

	nameValue := current
	placeholder := defaultPhase1ProjectName()
	if defaults != nil && defaults.ProjectName != "" {
		placeholder = defaults.ProjectName
	}
	if nameValue == "" {
		nameValue = placeholder
	}

	field := huh.NewInput().
		Title(info.Title()).
		Placeholder(placeholder).
		Value(&nameValue).
		Validate(validateProjectName)

	if err := theme.NewForm(huh.NewGroup(field)).Run(); err != nil {
		return "", PhaseCancel, fmt.Errorf("phase 1 cancelled: %w", err)
	}

	return nameValue, PhaseContinue, nil
}

func askCliTools(current []string, info phase1StepInfo) ([]string, PhaseAction, error) {
	preSelected := detectInstalledCliToolsFromCatalog()
	field := NewCliToolsSelect(current, preSelected)
	field.Title(info.Title())

	selected := cloneStrings(current)
	if len(selected) == 0 {
		selected = cloneStrings(preSelected)
	}
	field.Value(&selected)
	field.Options(appendPhase1BackOption(cliToolOptionsFromCatalogForSelect())...)

	if err := theme.NewForm(huh.NewGroup(field)).Run(); err != nil {
		return nil, PhaseCancel, fmt.Errorf("phase 1 cancelled: %w", err)
	}
	if containsString(selected, phase1BackValue) {
		return nil, PhaseBack, nil
	}

	return selected, PhaseContinue, nil
}

func askSkills(current []types.SkillId, info phase1StepInfo) ([]types.SkillId, PhaseAction, error) {
	selected := skillIDsToStrings(current)
	if len(selected) == 0 {
		selected = skillIDsToStrings(types.ALL_SKILLS)
	}

	field := huh.NewMultiSelect[string]().
		Title(info.Title()).
		Options(appendPhase1BackOption(skillOptions())...).
		Value(&selected)

	if err := theme.NewForm(huh.NewGroup(field)).Run(); err != nil {
		return nil, PhaseCancel, fmt.Errorf("phase 1 cancelled: %w", err)
	}
	if containsString(selected, phase1BackValue) {
		return nil, PhaseBack, nil
	}

	return stringsToSkillIDs(selected), PhaseContinue, nil
}

func askAgents(current []types.AgentId, info phase1StepInfo) ([]types.AgentId, PhaseAction, error) {
	selected := agentIDsToStrings(current)
	if len(selected) == 0 {
		selected = agentIDsToStrings(types.ALL_AGENTS)
	}

	field := huh.NewMultiSelect[string]().
		Title(info.Title()).
		Options(appendPhase1BackOption(agentOptions())...).
		Value(&selected)

	if err := theme.NewForm(huh.NewGroup(field)).Run(); err != nil {
		return nil, PhaseCancel, fmt.Errorf("phase 1 cancelled: %w", err)
	}
	if containsString(selected, phase1BackValue) {
		return nil, PhaseBack, nil
	}

	return stringsToAgentIDs(selected), PhaseContinue, nil
}

func askMcpPreset(current McpPreset, info phase1StepInfo) (McpPreset, PhaseAction, error) {
	presetValue := string(normalizeMcpPreset(current))
	field := huh.NewSelect[string]().
		Title(info.Title()).
		Description("Choose a starting MCP set, then refine individual servers next.").
		Options(
			huh.NewOption("Minimal — core local setup tools", string(McpPresetMinimal)),
			huh.NewOption("Recommended — balanced default", string(McpPresetRecommended)),
			huh.NewOption("Full — all catalog servers", string(McpPresetFull)),
			huh.NewOption("↩ Back", phase1BackValue),
		).
		Value(&presetValue)

	if err := theme.NewForm(huh.NewGroup(field)).Run(); err != nil {
		return "", PhaseCancel, fmt.Errorf("phase 1 cancelled: %w", err)
	}
	if presetValue == phase1BackValue {
		return "", PhaseBack, nil
	}

	return normalizeMcpPreset(McpPreset(presetValue)), PhaseContinue, nil
}

func askMcpServers(current []string, info phase1StepInfo) ([]string, PhaseAction, error) {
	field := NewMcpServersSelect(current)
	field.Title(info.Title())

	selected := cloneStrings(current)
	field.Value(&selected)
	field.Options(appendPhase1BackOption(mcpServerOptionsForSelect())...)

	if err := theme.NewForm(huh.NewGroup(field)).Run(); err != nil {
		return nil, PhaseCancel, fmt.Errorf("phase 1 cancelled: %w", err)
	}
	if containsString(selected, phase1BackValue) {
		return nil, PhaseBack, nil
	}

	return selected, PhaseContinue, nil
}

// askProjectIdentity collects optional Organization and Team values used to
// fill [YOUR_ORG] / [YOUR_TEAM] in the generated AGENTS.md. Both fields
// accept empty input (enter to skip — stays as <!-- fill-in --> marker).
func askProjectIdentity(currentOrg, currentTeam string, info phase1StepInfo) (string, string, PhaseAction, error) {
	org := currentOrg
	team := currentTeam
	decision := "continue"

	orgField := huh.NewInput().
		Title(info.Title()).
		Description("Organization name (leave blank to skip — stays as <!-- fill-in --> in AGENTS.md).").
		Value(&org)
	teamField := huh.NewInput().
		Title("Team").
		Description("Team name (optional).").
		Value(&team)
	decisionField := huh.NewSelect[string]().
		Title("Save identity or go back?").
		Options(
			huh.NewOption("Save and continue", "continue"),
			huh.NewOption("↩ Back", phase1BackValue),
		).
		Value(&decision)

	if err := theme.NewForm(huh.NewGroup(orgField, teamField, decisionField)).Run(); err != nil {
		return "", "", PhaseCancel, fmt.Errorf("phase 1 cancelled: %w", err)
	}
	if decision == phase1BackValue {
		return "", "", PhaseBack, nil
	}
	return strings.TrimSpace(org), strings.TrimSpace(team), PhaseContinue, nil
}

func validateProjectName(name string) error {
	if name == "" || strings.TrimSpace(name) == "" {
		return fmt.Errorf("project name is required")
	}
	if strings.TrimRightFunc(name, unicode.IsSpace) != name {
		return fmt.Errorf("project name cannot have trailing whitespace")
	}
	if strings.Contains(name, "/") || strings.Contains(name, `\\`) {
		return fmt.Errorf("project name cannot contain path separators")
	}
	if strings.Contains(name, "..") {
		return fmt.Errorf("project name cannot contain '..'")
	}
	if strings.HasPrefix(name, ".") {
		return fmt.Errorf("project name cannot start with '.'")
	}
	return nil
}

func defaultPhase1Scope() types.SetupScope {
	return types.SetupScopeProject
}

func defaultPhase1ProjectName() string {
	return "my-project"
}

func phase1StepInfoFor(step int, scope types.SetupScope, defaults *Phase1Result) phase1StepInfo {
	total := phase1Total(scope)
	current := step
	if scope == types.SetupScopeGlobal && step >= 8 {
		current--
	}

	info := phase1StepInfo{Current: current, Total: total}
	switch step {
	case 1:
		info.StepTitle = "Scope"
		if defaults != nil && defaults.Scope != "" {
			info.Previous = string(defaults.Scope)
		}
	case 2:
		info.StepTitle = "AI Tools"
		if defaults != nil && len(defaults.Tools) > 0 {
			info.Previous = strings.Join(toolIDsToStrings(defaults.Tools), ", ")
		}
	case 3:
		info.StepTitle = "Skills"
		if defaults != nil && len(defaults.Skills) > 0 {
			info.Previous = strings.Join(skillIDsToStrings(defaults.Skills), ", ")
		}
	case 4:
		info.StepTitle = "Agents"
		if defaults != nil && len(defaults.Agents) > 0 {
			info.Previous = strings.Join(agentIDsToStrings(defaults.Agents), ", ")
		}
	case 5:
		info.StepTitle = "MCP Preset"
		if defaults != nil && defaults.McpPreset != "" {
			info.Previous = string(normalizeMcpPreset(defaults.McpPreset))
		}
	case 6:
		info.StepTitle = "MCP Servers"
		if defaults != nil && len(defaults.EnableServers) > 0 {
			info.Previous = strings.Join(defaults.EnableServers, ", ")
		}
	case 7:
		info.StepTitle = "Project Name"
		if defaults != nil && defaults.ProjectName != "" {
			info.Previous = defaults.ProjectName
		}
	case 8:
		info.StepTitle = "CLI Tools"
	case 9:
		info.StepTitle = "Project Identity (optional)"
		if defaults != nil && (defaults.Organization != "" || defaults.Team != "") {
			info.Previous = fmt.Sprintf("org=%q team=%q", defaults.Organization, defaults.Team)
		}
	}
	return info
}

func phase1Total(scope types.SetupScope) int {
	if scope == types.SetupScopeGlobal {
		return 8
	}
	return 9
}

func previousPhase1Step(current int, scope types.SetupScope) int {
	previous := current - 1
	if scope == types.SetupScopeGlobal && previous == 7 {
		previous--
	}
	if previous < 1 {
		return 1
	}
	return previous
}

func appendPhase1BackOption(options []huh.Option[string], extra ...huh.Option[string]) []huh.Option[string] {
	result := append([]huh.Option[string]{}, options...)
	if len(extra) > 0 {
		result = append(result, extra...)
	} else {
		result = append(result, huh.NewOption("↩ Back", phase1BackValue))
	}
	return result
}

func toolIDsToStrings(tools []types.ToolId) []string {
	values := make([]string, len(tools))
	for i, tool := range tools {
		values[i] = string(tool)
	}
	return values
}

func stringsToToolIDs(values []string) []types.ToolId {
	tools := make([]types.ToolId, len(values))
	for i, value := range values {
		tools[i] = types.ToolId(value)
	}
	return tools
}

func skillIDsToStrings(skills []types.SkillId) []string {
	values := make([]string, len(skills))
	for i, skill := range skills {
		values[i] = string(skill)
	}
	return values
}

func stringsToSkillIDs(values []string) []types.SkillId {
	skills := make([]types.SkillId, len(values))
	for i, value := range values {
		skills[i] = types.SkillId(value)
	}
	return skills
}

func agentIDsToStrings(agents []types.AgentId) []string {
	values := make([]string, len(agents))
	for i, agent := range agents {
		values[i] = string(agent)
	}
	return values
}

func stringsToAgentIDs(values []string) []types.AgentId {
	agents := make([]types.AgentId, len(values))
	for i, value := range values {
		agents[i] = types.AgentId(value)
	}
	return agents
}

func cloneToolIDs(values []types.ToolId) []types.ToolId {
	if len(values) == 0 {
		return nil
	}
	return append([]types.ToolId(nil), values...)
}

func cloneSkillIDs(values []types.SkillId) []types.SkillId {
	if len(values) == 0 {
		return nil
	}
	return append([]types.SkillId(nil), values...)
}

func cloneAgentIDs(values []types.AgentId) []types.AgentId {
	if len(values) == 0 {
		return nil
	}
	return append([]types.AgentId(nil), values...)
}

func cloneStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	return append([]string(nil), values...)
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func cliToolOptionsFromCatalogForSelect() []huh.Option[string] {
	catalog, err := loadMcpCatalog()
	if err != nil {
		return nil
	}
	return cliToolOptionsFromCatalog(catalog)
}

func mcpServerOptionsForSelect() []huh.Option[string] {
	catalog, err := loadMcpCatalog()
	if err != nil {
		return nil
	}
	return mcpServerOptionsFromCatalog(catalog)
}

// containsTool reports whether the given tool ID is present in the list.
func containsTool(tools []types.ToolId, id string) bool {
	for _, t := range tools {
		if string(t) == id {
			return true
		}
	}
	return false
}
