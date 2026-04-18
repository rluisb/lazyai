package wizard

import (
	"fmt"
	"strings"
	"unicode"

	"charm.land/huh/v2"

	"github.com/ricardoborges-teachable/ai-setup/internal/adapter"
	"github.com/ricardoborges-teachable/ai-setup/internal/detect"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

const phase1BackValue = "__phase1_back__"

type phase1StepInfo struct {
	Current  int
	Total    int
	StepTitle string
	Previous string
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
	ProjectName   string
	CliTools      []string
	EnableServers []string
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
		return nil, PhaseCancel, fmt.Errorf("--tools is required in non-interactive mode (opencode, claude-code, gemini, copilot, codex)")
	}
	if defaults.ProjectName == "" {
		return nil, PhaseCancel, fmt.Errorf("project name is required in non-interactive mode")
	}
	return defaults, PhaseContinue, nil
}

func runPhase1Interactive(defaults *Phase1Result) (*Phase1Result, PhaseAction, error) {
	state := phase1InteractiveState{Scope: defaultPhase1Scope()}
	if defaults != nil {
		state.Scope = defaults.Scope
		state.Tools = cloneToolIDs(defaults.Tools)
		state.ProjectName = defaults.ProjectName
		state.CliTools = cloneStrings(defaults.CliTools)
		state.Servers = cloneStrings(defaults.EnableServers)
	}
	if state.Scope == "" {
		state.Scope = defaultPhase1Scope()
	}
	if state.ProjectName == "" {
		state.ProjectName = defaultPhase1ProjectName()
	}

	currentStep := 1
	for currentStep >= 1 && currentStep <= 5 {
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
			name, action, err := askProjectName(state.ProjectName, defaults, state.Scope, phase1StepInfoFor(3, scopeForStep, defaults))
			if err != nil {
				return nil, action, err
			}
			if action == PhaseBack {
				currentStep = previousPhase1Step(currentStep, state.Scope)
				continue
			}
			state.ProjectName = name
			currentStep++
		case 4:
			cliTools, action, err := askCliTools(state.CliTools, phase1StepInfoFor(4, scopeForStep, defaults))
			if err != nil {
				return nil, action, err
			}
			if action == PhaseBack {
				currentStep = previousPhase1Step(currentStep, state.Scope)
				continue
			}
			state.CliTools = cliTools
			currentStep++
		case 5:
			servers, action, err := askMcpServers(state.Servers, phase1StepInfoFor(5, scopeForStep, defaults))
			if err != nil {
				return nil, action, err
			}
			if action == PhaseBack {
				currentStep = previousPhase1Step(currentStep, state.Scope)
				continue
			}
			state.Servers = servers
			currentStep++
		}
	}

	return buildPhase1Result(state.Scope, state.Tools, state.ProjectName, state.CliTools, state.Servers), PhaseContinue, nil
}

type phase1InteractiveState struct {
	Scope       types.SetupScope
	Tools       []types.ToolId
	ProjectName string
	CliTools    []string
	Servers     []string
}

func buildPhase1Result(scope types.SetupScope, tools []types.ToolId, name string, cliTools, servers []string) *Phase1Result {
	projectName := name
	if scope == types.SetupScopeGlobal {
		projectName = "global"
	}

	return &Phase1Result{
		Scope:         scope,
		Tools:         cloneToolIDs(tools),
		ProjectName:   projectName,
		CliTools:      cloneStrings(cliTools),
		EnableServers: cloneStrings(servers),
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

	if err := huh.NewForm(huh.NewGroup(field)).Run(); err != nil {
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

	if err := huh.NewForm(huh.NewGroup(field)).Run(); err != nil {
		return nil, PhaseCancel, fmt.Errorf("phase 1 cancelled: %w", err)
	}
	if containsString(selectedTools, phase1BackValue) {
		return nil, PhaseBack, nil
	}

	tools := stringsToToolIDs(selectedTools)
	filteredTools, err := filterUninstalledCodex(tools)
	if err != nil {
		return nil, PhaseCancel, err
	}
	return filteredTools, PhaseContinue, nil
}

// toolOptionsForScope returns the AI tool multi-select options filtered to
// only tools that support the given scope (see adapter.IsScopeSupported).
func toolOptionsForScope(scope types.SetupScope) []huh.Option[string] {
	all := []huh.Option[string]{
		huh.NewOption("OpenCode", "opencode"),
		huh.NewOption("Claude Code", "claude-code"),
		huh.NewOption("Gemini CLI", "gemini"),
		huh.NewOption("GitHub Copilot", "copilot"),
		huh.NewOption("Codex (OpenAI)", "codex"),
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

	if err := huh.NewForm(huh.NewGroup(field)).Run(); err != nil {
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

	if err := huh.NewForm(huh.NewGroup(field)).Run(); err != nil {
		return nil, PhaseCancel, fmt.Errorf("phase 1 cancelled: %w", err)
	}
	if containsString(selected, phase1BackValue) {
		return nil, PhaseBack, nil
	}

	return selected, PhaseContinue, nil
}

func askMcpServers(current []string, info phase1StepInfo) ([]string, PhaseAction, error) {
	field := NewMcpServersSelect(current)
	field.Title(info.Title())

	selected := cloneStrings(current)
	field.Value(&selected)
	field.Options(appendPhase1BackOption(mcpServerOptionsForSelect())...)

	if err := huh.NewForm(huh.NewGroup(field)).Run(); err != nil {
		return nil, PhaseCancel, fmt.Errorf("phase 1 cancelled: %w", err)
	}
	if containsString(selected, phase1BackValue) {
		return nil, PhaseBack, nil
	}

	return selected, PhaseContinue, nil
}

func filterUninstalledCodex(tools []types.ToolId) ([]types.ToolId, error) {
	if !containsTool(tools, string(types.ToolIdCodex)) || detect.IsCodexInstalled() {
		return tools, nil
	}

	confirm := false
	if err := huh.NewConfirm().
		Title(fmt.Sprintf("Codex CLI is not installed.\n\n%s\n\nContinue anyway?", detect.CodexInstallHint())).
		Affirmative("Continue").
		Negative("Remove codex from selection").
		Value(&confirm).
		Run(); err != nil {
		return nil, fmt.Errorf("phase 1 cancelled: %w", err)
	}
	if confirm {
		return tools, nil
	}

	filtered := make([]types.ToolId, 0, len(tools))
	for _, tool := range tools {
		if tool != types.ToolIdCodex {
			filtered = append(filtered, tool)
		}
	}
	return filtered, nil
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
	if scope == types.SetupScopeGlobal && step >= 4 {
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
		info.StepTitle = "Project Name"
		if defaults != nil && defaults.ProjectName != "" {
			info.Previous = defaults.ProjectName
		}
	case 4:
		info.StepTitle = "CLI Tools"
	case 5:
		info.StepTitle = "MCP Servers"
	}
	return info
}

func phase1Total(scope types.SetupScope) int {
	if scope == types.SetupScopeGlobal {
		return 4
	}
	return 5
}

func previousPhase1Step(current int, scope types.SetupScope) int {
	previous := current - 1
	if scope == types.SetupScopeGlobal && previous == 3 {
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

func cloneToolIDs(values []types.ToolId) []types.ToolId {
	if len(values) == 0 {
		return nil
	}
	return append([]types.ToolId(nil), values...)
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
