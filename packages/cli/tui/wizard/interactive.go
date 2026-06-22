package wizard

import (
	"charm.land/huh/v2"
	"github.com/rluisb/lazyai/packages/cli/internal/theme"

	"github.com/rluisb/lazyai/packages/cli/internal/preset"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

type WizardState struct {
	// Phase 1
	Scope        string
	Tools        []string
	Skills       []string
	Agents       []string
	McpPreset    string
	McpServers   []string
	ProjectName  string
	CliTools     []string
	Organization string
	Team         string

	// Phase 2
	Preset              string
	Features            []string
	BranchPattern       string
	CustomBranch        string
	CommitPattern       string
	CustomCommit        string
	RequireTicket       bool
	ChatModes           []string
	OpenCodeCommands    []string
	OpenCodeModes       []string
	AnalyzeExistingCode bool

	MemoryPath        string
	EnableObsidian    bool
	EnableCodegraph   bool
	CodegraphDataPath string
	OpenCodePlugins   []string
	OpenCodeProviders []string
}

func initWizardState(defaults *WizardResult) *WizardState {
	s := &WizardState{}

	// Set Phase 1 Defaults
	s.Scope = string(defaultPhase1Scope())
	s.ProjectName = defaultPhase1ProjectName()
	s.McpPreset = string(McpPresetRecommended)
	s.Skills = skillIDsToStrings(types.ALL_SKILLS)
	s.Agents = agentIDsToStrings(types.ALL_AGENTS)
	s.CliTools = detectInstalledCliToolsFromCatalog() // pre-select installed ones
	s.McpServers = defaultMcpServersForPreset(McpPresetRecommended)

	if defaults != nil && defaults.Phase1 != nil {
		if defaults.Phase1.Scope != "" {
			s.Scope = string(defaults.Phase1.Scope)
		}
		if len(defaults.Phase1.Tools) > 0 {
			s.Tools = toolIDsToStrings(defaults.Phase1.Tools)
		}
		if len(defaults.Phase1.Skills) > 0 {
			s.Skills = skillIDsToStrings(defaults.Phase1.Skills)
		}
		if len(defaults.Phase1.Agents) > 0 {
			s.Agents = agentIDsToStrings(defaults.Phase1.Agents)
		}
		if defaults.Phase1.McpPreset != "" {
			s.McpPreset = string(defaults.Phase1.McpPreset)
		}
		if defaults.Phase1.ProjectName != "" {
			s.ProjectName = defaults.Phase1.ProjectName
		}
		if len(defaults.Phase1.CliTools) > 0 {
			s.CliTools = defaults.Phase1.CliTools
		}
		if len(defaults.Phase1.EnableServers) > 0 {
			s.McpServers = defaults.Phase1.EnableServers
		}
		s.Organization = defaults.Phase1.Organization
		s.Team = defaults.Phase1.Team
	}

	// Set Phase 2 Defaults
	gitDefs := types.DefaultGitConventions()
	s.Preset = string(preset.DefaultPresetForScope(types.SetupScope(s.Scope)))
	s.Features = defaultPhase2FeatureSelection()
	s.BranchPattern = gitDefs.BranchPattern
	s.CustomBranch = gitDefs.BranchPattern
	s.CommitPattern = gitDefs.CommitPattern
	s.CustomCommit = gitDefs.CommitPattern
	s.RequireTicket = gitDefs.RequireTicket
	s.ChatModes = chatModeIdsToStrings(types.ALL_CHATMODES[:])
	s.OpenCodeCommands = opencodeCommandIdsToStrings(types.ALL_OPENCODE_COMMANDS[:])
	s.OpenCodeModes = opencodeModeIdsToStrings(types.ALL_OPENCODE_MODES[:])
	s.AnalyzeExistingCode = true

	if defaults != nil && defaults.Phase2 != nil {
		if defaults.Phase2.Preset != "" {
			s.Preset = string(defaults.Phase2.Preset)
		}
		if defaults.Phase2.Features != nil {
			s.Features = featureSelectionFromFlags(defaults.Phase2.Features)
		}
		if defaults.Phase2.GitConv != nil {
			if defaults.Phase2.GitConv.BranchPattern != "" {
				s.BranchPattern = defaults.Phase2.GitConv.BranchPattern
			}
			if defaults.Phase2.GitConv.CommitPattern != "" {
				s.CommitPattern = defaults.Phase2.GitConv.CommitPattern
			}
			s.RequireTicket = defaults.Phase2.GitConv.RequireTicket
		}
		if len(defaults.Phase2.ChatModes) > 0 {
			s.ChatModes = chatModeIdsToStrings(defaults.Phase2.ChatModes)
		}
		if len(defaults.Phase2.OpenCodeCommands) > 0 {
			s.OpenCodeCommands = opencodeCommandIdsToStrings(defaults.Phase2.OpenCodeCommands)
		}
		if len(defaults.Phase2.OpenCodeModes) > 0 {
			s.OpenCodeModes = opencodeModeIdsToStrings(defaults.Phase2.OpenCodeModes)
		}
		if defaults.Phase2.UseReversa != nil {
			s.AnalyzeExistingCode = *defaults.Phase2.UseReversa
		}
	}

	// Set Phase 5 Defaults
	s.MemoryPath = ".specify/memory"
	s.EnableObsidian = true
	s.EnableCodegraph = true
	s.CodegraphDataPath = ".codegraph/"
	if defaults != nil && defaults.Phase5 != nil {
		if defaults.Phase5.MemoryPath != "" {
			s.MemoryPath = defaults.Phase5.MemoryPath
		}
		if defaults.Phase5.EnableObsidian {
			s.EnableObsidian = defaults.Phase5.EnableObsidian
		}
		if defaults.Phase5.EnableCodegraph {
			s.EnableCodegraph = defaults.Phase5.EnableCodegraph
		}
		if defaults.Phase5.CodegraphDataPath != "" {
			s.CodegraphDataPath = defaults.Phase5.CodegraphDataPath
		}
		if len(defaults.Phase5.OpenCodePlugins) > 0 {
			s.OpenCodePlugins = defaults.Phase5.OpenCodePlugins
		}
		if len(defaults.Phase5.OpenCodeProviders) > 0 {
			s.OpenCodeProviders = defaults.Phase5.OpenCodeProviders
		}
	}

	return s
}

func initExpressWizardState(defaults *WizardResult) *WizardState {
	state := initWizardState(defaults)
	state.McpPreset = string(McpPresetRecommended)
	state.McpServers = []string{"filesystem", "ripgrep", "ai-memory", "codegraph"}
	state.EnableObsidian = false
	state.EnableCodegraph = true
	state.CodegraphDataPath = ".codegraph/"
	return state
}

func buildExpressInteractiveForm(state *WizardState) *huh.Form {
	scopeField := huh.NewSelect[string]().
		Title("Scope").
		Options(
			huh.NewOption("Global  — Install to ~/.config/opencode/ + native tool global paths", "global"),
			huh.NewOption("Workspace  — Planning repo with multi-project management", "workspace"),
			huh.NewOption("Project (recommended)  — Self-contained single repository", "project"),
		).
		Value(&state.Scope)

	toolsField := huh.NewMultiSelect[string]().
		Title("AI Tools").
		OptionsFunc(func() []huh.Option[string] {
			return toolOptionsForScope(types.SetupScope(state.Scope))
		}, &state.Scope).
		Value(&state.Tools)

	groups := []*huh.Group{
		huh.NewGroup(selectFooterDescription(scopeField, func() string {
			return selectHoverDescription(scopeField, scopeDescriptions, defaultHoverHint)
		})),
		huh.NewGroup(multiSelectFooterDescription(toolsField, func() string {
			return multiSelectHoverDescription(toolsField, toolDescriptions, defaultHoverHint)
		})),
		huh.NewGroup(
			huh.NewInput().
				Title("Project Name").
				Description("Used in generated project identity and local config names.").
				Placeholder(defaultPhase1ProjectName()).
				Value(&state.ProjectName).
				Validate(validateProjectName),
		).WithHideFunc(func() bool {
			return state.Scope == "global"
		}),
	}

	return theme.NewForm(groups...)
}

func buildInteractiveForm(state *WizardState) *huh.Form {
	scopeField := huh.NewSelect[string]().
		Title("Scope").
		Options(
			huh.NewOption("Global  — Install to ~/.config/opencode/ + native tool global paths", "global"),
			huh.NewOption("Workspace  — Planning repo with multi-project management", "workspace"),
			huh.NewOption("Project (recommended)  — Self-contained single repository", "project"),
		).
		Value(&state.Scope)

	toolsField := huh.NewMultiSelect[string]().
		Title("AI Tools").
		OptionsFunc(func() []huh.Option[string] {
			return toolOptionsForScope(types.SetupScope(state.Scope))
		}, &state.Scope).
		Value(&state.Tools)

	skillsField := huh.NewMultiSelect[string]().
		Title("Skills").
		Options(skillOptions()...).
		Value(&state.Skills)

	agentsField := huh.NewMultiSelect[string]().
		Title("Agents").
		Options(agentOptions()...).
		Value(&state.Agents)

	mcpPresetField := huh.NewSelect[string]().
		Title("MCP Preset").
		Options(
			huh.NewOption("Minimal — core local setup tools", string(McpPresetMinimal)),
			huh.NewOption("Recommended — balanced default", string(McpPresetRecommended)),
			huh.NewOption("Full — all catalog servers", string(McpPresetFull)),
		).
		Value(&state.McpPreset)

	mcpServersField := NewMcpServersSelect(state.McpServers).
		Title("MCP Servers").
		Value(&state.McpServers)

	cliToolsField := NewCliToolsSelect(state.CliTools, detectInstalledCliToolsFromCatalog()).
		Title("CLI Tools").
		Value(&state.CliTools)

	presetField := huh.NewSelect[string]().
		Title("Preset").
		Options(
			huh.NewOption("Minimal — Quality gates + git only", "minimal"),
			huh.NewOption("Standard (recommended) — +RPI, reasoning, bug resolution", "standard"),
			huh.NewOption("Full — All features enabled", "full"),
			huh.NewOption("Custom — Pick features individually", "custom"),
		).
		Value(&state.Preset)

	featuresField := huh.NewMultiSelect[string]().
		Title("Features").
		Options(featureOptions...).
		Value(&state.Features)

	branchPatternField := huh.NewSelect[string]().
		Title("Branch Pattern").
		Options(branchPatternOptions...).
		Value(&state.BranchPattern)

	commitPatternField := huh.NewSelect[string]().
		Title("Commit Pattern").
		Options(commitPatternOptions...).
		Value(&state.CommitPattern)

	chatModesField := huh.NewMultiSelect[string]().
		Title("Copilot Chat Modes").
		Options(
			huh.NewOption("Architect mode (architect)", string(types.ChatModeIdArchitect)),
			huh.NewOption("Reviewer mode (reviewer)", string(types.ChatModeIdReviewer)),
		).
		Value(&state.ChatModes)

	groups := []*huh.Group{
		// Phase 1
		huh.NewGroup(selectFooterDescription(scopeField, func() string {
			return selectHoverDescription(scopeField, scopeDescriptions, defaultHoverHint)
		})),
		huh.NewGroup(multiSelectFooterDescription(toolsField, func() string {
			return multiSelectHoverDescription(toolsField, toolDescriptions, defaultHoverHint)
		})),
		huh.NewGroup(multiSelectFooterDescription(skillsField, func() string {
			return multiSelectHoverDescription(skillsField, skillDescriptions, defaultHoverHint)
		})),
		huh.NewGroup(multiSelectFooterDescription(agentsField, func() string {
			return multiSelectHoverDescription(agentsField, agentDescriptions, defaultHoverHint)
		})),
		huh.NewGroup(selectFooterDescription(mcpPresetField, func() string {
			return selectHoverDescription(mcpPresetField, mcpPresetDescriptions, defaultHoverHint)
		})),
		huh.NewGroup(multiSelectFooterDescription(mcpServersField, func() string {
			return multiSelectHoverDescription(mcpServersField, catalogServerDescriptions(), defaultHoverHint)
		})),
		huh.NewGroup(
			huh.NewInput().
				Title("Project Name").
				Description("Used in generated project identity and local config names.").
				Placeholder(defaultPhase1ProjectName()).
				Value(&state.ProjectName).
				Validate(validateProjectName),
		).WithHideFunc(func() bool {
			return state.Scope == "global"
		}),
		huh.NewGroup(multiSelectFooterDescription(cliToolsField, func() string {
			return multiSelectHoverDescription(cliToolsField, catalogCliToolDescriptions(), defaultHoverHint)
		})),
		huh.NewGroup(
			huh.NewInput().
				Title("Organization Name").
				Description("Leave blank to skip — stays as <!-- fill-in --> in AGENTS.md").
				Value(&state.Organization),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Team Name").
				Description("Optional team identifier").
				Value(&state.Team),
		),

		// Phase 2
		huh.NewGroup(selectFooterDescription(presetField, func() string {
			return selectHoverDescription(presetField, presetDescriptions, defaultHoverHint)
		})),
		huh.NewGroup(multiSelectFooterDescription(featuresField, func() string {
			return multiSelectHoverDescription(featuresField, featureDescriptions, defaultHoverHint)
		})).WithHideFunc(func() bool { return state.Preset != "custom" }),
		huh.NewGroup(selectFooterDescription(branchPatternField, func() string {
			return selectHoverDescription(branchPatternField, branchPatternDescriptions, defaultHoverHint)
		})),
		huh.NewGroup(
			huh.NewInput().
				Title("Custom branch pattern (use {type}, {ticket}, {description}):").
				Description("Template for new branch names; supported placeholders are shown in the title.").
				Placeholder(types.DefaultGitConventions().BranchPattern).
				Value(&state.CustomBranch),
		).WithHideFunc(func() bool { return state.BranchPattern != "custom" }),
		huh.NewGroup(selectFooterDescription(commitPatternField, func() string {
			return selectHoverDescription(commitPatternField, commitPatternDescriptions, defaultHoverHint)
		})),
		huh.NewGroup(
			huh.NewInput().
				Title("Custom commit pattern (use {type}, {scope}, {ticket}, {description}):").
				Description("Template for commit messages; supported placeholders are shown in the title.").
				Placeholder(types.DefaultGitConventions().CommitPattern).
				Value(&state.CustomCommit),
		).WithHideFunc(func() bool { return state.CommitPattern != "custom" }),
		huh.NewGroup(
			huh.NewConfirm().
				Title("Require Ticket").
				Description("When enabled, branch and commit guidance expects a ticket placeholder.").
				Value(&state.RequireTicket),
		),
		huh.NewGroup(
			huh.NewConfirm().
				Title("Analyze existing code to auto-populate project details?").
				Description("Runs deterministic Scout/Reversa analysis against this project before scaffolding.").
				Value(&state.AnalyzeExistingCode),
		).Title("Project Profile"),
		huh.NewGroup(multiSelectFooterDescription(chatModesField, func() string {
			return multiSelectHoverDescription(chatModesField, chatModeDescriptions, defaultHoverHint)
		})).WithHideFunc(func() bool { return state.Preset != "custom" }),
		// Phase 5
		huh.NewGroup(
			huh.NewInput().
				Title("Memory Path").
				Description("Project-local default for bootstrap and housekeeping.").
				Placeholder(".specify/memory").
				Value(&state.MemoryPath),
		),
	}

	return theme.NewForm(groups...)
}

func extractResults(state *WizardState) (*Phase1Result, *Phase2Result, *Phase5Result) {
	// Phase 1
	validTools := filterToolsByScope(stringsToToolIDs(state.Tools), types.SetupScope(state.Scope))

	p1 := buildPhase1Result(
		types.SetupScope(state.Scope),
		validTools,
		stringsToSkillIDs(state.Skills),
		stringsToAgentIDs(state.Agents),
		McpPreset(state.McpPreset),
		state.ProjectName,
		state.CliTools,
		state.McpServers,
		state.Organization,
		state.Team,
	)

	// Phase 2
	branch := state.BranchPattern
	if branch == "custom" {
		branch = state.CustomBranch
	}
	commit := state.CommitPattern
	if commit == "custom" {
		commit = state.CustomCommit
	}

	p2 := buildPhase2Result(
		p1.Scope,
		types.PresetLevel(state.Preset),
		buildFeaturesFromSelection(state.Features),
		branch,
		commit,
		state.RequireTicket,
		stringsToChatModeIds(state.ChatModes),
		stringsToOpenCodeCommandIds(state.OpenCodeCommands),
		stringsToOpenCodeModeIds(state.OpenCodeModes),
	)
	p2.UseReversa = boolPtr(state.AnalyzeExistingCode)

	// Phase 5
	p5 := buildPhase5Result(
		state.MemoryPath,
		state.EnableObsidian,
		"",
		state.EnableCodegraph,
		state.CodegraphDataPath,
		state.OpenCodePlugins,
	)
	p5.OpenCodeProviders = state.OpenCodeProviders

	return p1, p2, p5
}

func askWizardMode() (WizardMode, error) {
	mode := string(WizardModePersonalized)
	group := huh.NewGroup(
		huh.NewSelect[string]().
			Title("Setup mode").
			Description("Choose an initialization flow").
			Options(
				huh.NewOption("Express", string(WizardModeExpress)),
				huh.NewOption("Personalized", string(WizardModePersonalized)),
			).
			Value(&mode),
	)

	if err := theme.NewForm(group).Run(); err != nil {
		return "", err
	}
	return WizardMode(mode), nil
}
