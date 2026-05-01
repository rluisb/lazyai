package wizard

import (
	"charm.land/huh/v2"

	"github.com/ricardoborges-teachable/ai-setup/internal/preset"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
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
	Preset           string
	Features         []string
	BranchPattern    string
	CustomBranch     string
	CommitPattern    string
	CustomCommit     string
	RequireTicket    bool
	ChatModes        []string
	OpenCodeCommands []string
	OpenCodeModes    []string

	// Phase 5
	MemoryPath        string
	EnableObsidian    bool
	ObsidianVaultPath string
	EnableQmd         bool
	QmdIndexPath      string
	EnableCodegraph   bool
	CodegraphDataPath string
	OpenCodePlugins   []string
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
	}

	// Set Phase 5 Defaults
	s.MemoryPath = "specs/memory"
	s.QmdIndexPath = ".qmd-index"
	s.CodegraphDataPath = ".codegraph"

	if defaults != nil && defaults.Phase5 != nil {
		if defaults.Phase5.MemoryPath != "" {
			s.MemoryPath = defaults.Phase5.MemoryPath
		}
		s.EnableObsidian = defaults.Phase5.EnableObsidian
		if defaults.Phase5.ObsidianVaultPath != "" {
			s.ObsidianVaultPath = defaults.Phase5.ObsidianVaultPath
		}
		s.EnableQmd = defaults.Phase5.EnableQmd
		if defaults.Phase5.QmdIndexPath != "" {
			s.QmdIndexPath = defaults.Phase5.QmdIndexPath
		}
		s.EnableCodegraph = defaults.Phase5.EnableCodegraph
		if defaults.Phase5.CodegraphDataPath != "" {
			s.CodegraphDataPath = defaults.Phase5.CodegraphDataPath
		}
		if len(defaults.Phase5.OpenCodePlugins) > 0 {
			s.OpenCodePlugins = defaults.Phase5.OpenCodePlugins
		}
	}

	return s
}

func buildInteractiveForm(state *WizardState) *huh.Form {
	groups := []*huh.Group{
		// Phase 1
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Scope").
				Options(
					huh.NewOption("Global  — Install to ~/.config/opencode/ + native tool global paths", "global"),
					huh.NewOption("Workspace  — Planning repo with multi-project management", "workspace"),
					huh.NewOption("Project (recommended)  — Self-contained single repository", "project"),
				).
				Value(&state.Scope),
		),
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("AI Tools").
				OptionsFunc(func() []huh.Option[string] {
					return toolOptionsForScope(types.SetupScope(state.Scope))
				}, &state.Scope).
				Value(&state.Tools),
		),
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Skills").
				Options(skillOptions()...).
				Value(&state.Skills),
		),
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Agents").
				Options(agentOptions()...).
				Value(&state.Agents),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("MCP Preset").
				Description("Choose a starting MCP set, then refine individual servers next.").
				Options(
					huh.NewOption("Minimal — core local setup tools", string(McpPresetMinimal)),
					huh.NewOption("Recommended — balanced default", string(McpPresetRecommended)),
					huh.NewOption("Full — all catalog servers", string(McpPresetFull)),
				).
				Value(&state.McpPreset),
		),
		huh.NewGroup(
			NewMcpServersSelect(state.McpServers).
				Title("MCP Servers").
				Value(&state.McpServers),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Project Name").
				Placeholder(defaultPhase1ProjectName()).
				Value(&state.ProjectName).
				Validate(validateProjectName),
		).WithHideFunc(func() bool {
			return state.Scope == "global"
		}),
		huh.NewGroup(
			NewCliToolsSelect(state.CliTools, detectInstalledCliToolsFromCatalog()).
				Title("CLI Tools").
				Value(&state.CliTools),
		),
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
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Preset").
				Options(
					huh.NewOption("Minimal — Quality gates + git only", "minimal"),
					huh.NewOption("Standard (recommended) — +RPI, reasoning, bug resolution", "standard"),
					huh.NewOption("Full — All features enabled", "full"),
					huh.NewOption("Custom — Pick features individually", "custom"),
				).
				Value(&state.Preset),
		),
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Features").
				Options(featureOptions...).
				Value(&state.Features),
		).WithHideFunc(func() bool { return state.Preset != "custom" }),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Branch Pattern").
				Options(branchPatternOptions...).
				Value(&state.BranchPattern),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Custom branch pattern (use {type}, {ticket}, {description}):").
				Placeholder(types.DefaultGitConventions().BranchPattern).
				Value(&state.CustomBranch),
		).WithHideFunc(func() bool { return state.BranchPattern != "custom" }),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Commit Pattern").
				Options(commitPatternOptions...).
				Value(&state.CommitPattern),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Custom commit pattern (use {type}, {scope}, {ticket}, {description}):").
				Placeholder(types.DefaultGitConventions().CommitPattern).
				Value(&state.CustomCommit),
		).WithHideFunc(func() bool { return state.CommitPattern != "custom" }),
		huh.NewGroup(
			huh.NewConfirm().
				Title("Require Ticket").
				Value(&state.RequireTicket),
		),
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Copilot Chat Modes").
				Description("Select Copilot chat modes to install. Deselect to skip.").
				Options(
					huh.NewOption("Architect mode (architect)", string(types.ChatModeIdArchitect)),
					huh.NewOption("Reviewer mode (reviewer)", string(types.ChatModeIdReviewer)),
				).
				Value(&state.ChatModes),
		).WithHideFunc(func() bool { return state.Preset != "custom" }),
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("OpenCode Commands").
				Description("Select OpenCode slash commands to install. Deselect to skip.").
				Options(
					huh.NewOption("Review branch (review)", string(types.OpenCodeCommandIdReview)),
					huh.NewOption("Run tests (test)", string(types.OpenCodeCommandIdTest)),
					huh.NewOption("Draft commit (commit)", string(types.OpenCodeCommandIdCommit)),
				).
				Value(&state.OpenCodeCommands),
		).WithHideFunc(func() bool { return state.Preset != "custom" }),
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("OpenCode Modes").
				Description("Select OpenCode chat modes to install. Deselect to skip.").
				Options(
					huh.NewOption("Plan mode (plan)", string(types.OpenCodeModeIdPlan)),
					huh.NewOption("Audit mode (audit)", string(types.OpenCodeModeIdAudit)),
				).
				Value(&state.OpenCodeModes),
		).WithHideFunc(func() bool { return state.Preset != "custom" }),

		// Phase 5
		huh.NewGroup(
			huh.NewInput().
				Title("Memory Path").
				Description("Project-local default for bootstrap and housekeeping.").
				Placeholder("specs/memory").
				Value(&state.MemoryPath),
		),
		huh.NewGroup(
			huh.NewConfirm().
				Title("Enable Obsidian").
				Description("Read-only discovery only by default; future config writes remain explicit.").
				Value(&state.EnableObsidian),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Obsidian Vault Path").
				Value(&state.ObsidianVaultPath),
		).WithHideFunc(func() bool { return !state.EnableObsidian }),
		huh.NewGroup(
			huh.NewConfirm().
				Title("Enable qmd").
				Description("Read-only retrieval allowed; sync/index writes remain approval-gated.").
				Value(&state.EnableQmd),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("qmd Index Path").
				Placeholder(".qmd-index").
				Value(&state.QmdIndexPath),
		).WithHideFunc(func() bool { return !state.EnableQmd }),
		huh.NewGroup(
			huh.NewConfirm().
				Title("Enable Codegraph").
				Description("Read-only drift checks allowed; sync/index writes remain approval-gated.").
				Value(&state.EnableCodegraph),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Codegraph Data Path").
				Placeholder(".codegraph").
				Value(&state.CodegraphDataPath),
		).WithHideFunc(func() bool { return !state.EnableCodegraph }),
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("OpenCode Plugins").
				Description("Select OpenCode plugins to install via `opencode plugin`. Deselect to skip.").
				Options(
					huh.NewOption("Desktop Commander (@opencode/desktop-commander)", "@opencode/desktop-commander"),
					huh.NewOption("Context Files (@opencode/context-files)", "@opencode/context-files"),
					huh.NewOption("Git Tools (@opencode/git-tools)", "@opencode/git-tools"),
				).
				Value(&state.OpenCodePlugins),
		).WithHideFunc(func() bool {
			return !containsString(state.Tools, "opencode") || !opencodeBinaryPresent()
		}),
	}

	return huh.NewForm(groups...)
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

	// Phase 5
	p5 := buildPhase5Result(
		state.MemoryPath,
		state.EnableObsidian,
		state.ObsidianVaultPath,
		state.EnableQmd,
		state.QmdIndexPath,
		state.EnableCodegraph,
		state.CodegraphDataPath,
		state.OpenCodePlugins,
	)

	return p1, p2, p5
}
