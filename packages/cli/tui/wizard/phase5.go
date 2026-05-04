package wizard

import (
	"fmt"
	"os/exec"

	"charm.land/huh/v2"
)

type phase5StepInfo struct {
	Current   int
	Total     int
	StepTitle string
}

func (s phase5StepInfo) Title() string {
	return fmt.Sprintf("Optional Tooling — %d/%d: %s", s.Current, s.Total, s.StepTitle)
}

type Phase5Result struct {
	MemoryPath        string
	EnableObsidian    bool
	ObsidianVaultPath string
	EnableQmd         bool
	QmdIndexPath      string
	EnableCodegraph   bool
	CodegraphDataPath string
	EnableGraphify    bool
	GraphifyDataPath  string
	OpenCodePlugins   []string
}

// RunPhase5 runs the optional tooling phase.
// opencodeSelected is true when opencode is in the user's tool list — it gates
// the plugin install step (which also requires the binary to be on PATH).
func RunPhase5(defaults *Phase5Result, nonInteractive bool, opencodeSelected ...bool) (*Phase5Result, PhaseAction, error) {
	selected := len(opencodeSelected) > 0 && opencodeSelected[0]
	if nonInteractive {
		return runPhase5NonInteractive(defaults)
	}
	return runPhase5Interactive(defaults, selected)
}

func runPhase5NonInteractive(defaults *Phase5Result) (*Phase5Result, PhaseAction, error) {
	result := defaultPhase5Result()
	if defaults != nil {
		result = *defaults
	}
	return buildPhase5Result(
		result.MemoryPath,
		result.EnableObsidian,
		result.ObsidianVaultPath,
		result.EnableQmd,
		result.QmdIndexPath,
		result.EnableCodegraph,
		result.CodegraphDataPath,
		result.EnableGraphify,
		result.GraphifyDataPath,
		result.OpenCodePlugins,
	), PhaseContinue, nil
}

func runPhase5Interactive(defaults *Phase5Result, opencodeSelected bool) (*Phase5Result, PhaseAction, error) {
	state := defaultPhase5Result()
	if defaults != nil {
		state = *defaults
	}

	// Plugin step is shown only when opencode is selected AND binary is on PATH.
	showPlugins := opencodeSelected && opencodeBinaryPresent()

	currentStep := 1
	maxStep := 9
	if showPlugins {
		maxStep = 10
	}
	for currentStep >= 1 && currentStep <= maxStep {
		switch currentStep {
		case 1:
			memoryPath, action, err := askMemoryPath(state.MemoryPath, phase5MemoryPathStepInfo(state))
			if err != nil {
				return nil, action, err
			}
			state.MemoryPath = memoryPath
			currentStep++
		case 2:
			enableObsidian, action, err := askEnableObsidian(state.EnableObsidian, phase5EnableObsidianStepInfo(state))
			if err != nil {
				return nil, action, err
			}
			state.EnableObsidian = enableObsidian
			if !state.EnableObsidian {
				currentStep = 4
				continue
			}
			currentStep++
		case 3:
			obsidianVaultPath, action, err := askObsidianVaultPath(state.ObsidianVaultPath, phase5ObsidianVaultPathStepInfo(state))
			if err != nil {
				return nil, action, err
			}
			state.ObsidianVaultPath = obsidianVaultPath
			currentStep++
		case 4:
			enableQmd, action, err := askEnableQmd(state.EnableQmd, phase5EnableQmdStepInfo(state))
			if err != nil {
				return nil, action, err
			}
			state.EnableQmd = enableQmd
			if !state.EnableQmd {
				currentStep = 6
				continue
			}
			currentStep++
		case 5:
			qmdIndexPath, action, err := askQmdIndexPath(state.QmdIndexPath, phase5QmdIndexPathStepInfo(state))
			if err != nil {
				return nil, action, err
			}
			state.QmdIndexPath = qmdIndexPath
			currentStep++
		case 6:
			enableCodegraph, action, err := askEnableCodegraph(state.EnableCodegraph, phase5EnableCodegraphStepInfo(state))
			if err != nil {
				return nil, action, err
			}
			state.EnableCodegraph = enableCodegraph
			if !state.EnableCodegraph {
				currentStep = 8
				continue
			}
			currentStep++
		case 7:
			codegraphDataPath, action, err := askCodegraphDataPath(state.CodegraphDataPath, phase5CodegraphDataPathStepInfo(state, showPlugins))
			if err != nil {
				return nil, action, err
			}
			state.CodegraphDataPath = codegraphDataPath
			currentStep++
		case 8:
			enableGraphify, action, err := askEnableGraphify(state.EnableGraphify, phase5EnableGraphifyStepInfo(state))
			if err != nil {
				return nil, action, err
			}
			state.EnableGraphify = enableGraphify
			if !state.EnableGraphify {
				currentStep = 10
				continue
			}
			currentStep++
		case 9:
			graphifyDataPath, action, err := askGraphifyDataPath(state.GraphifyDataPath, phase5GraphifyDataPathStepInfo(state, showPlugins))
			if err != nil {
				return nil, action, err
			}
			state.GraphifyDataPath = graphifyDataPath
			currentStep++
		case 10:
			plugins, action, err := askOpenCodePlugins(state.OpenCodePlugins, phase5OpenCodePluginsStepInfo(state, showPlugins))
			if err != nil {
				return nil, action, err
			}
			if action == PhaseBack {
				currentStep--
				continue
			}
			state.OpenCodePlugins = plugins
			currentStep++
		}
	}

	return buildPhase5Result(
		state.MemoryPath,
		state.EnableObsidian,
		state.ObsidianVaultPath,
		state.EnableQmd,
		state.QmdIndexPath,
		state.EnableCodegraph,
		state.CodegraphDataPath,
		state.EnableGraphify,
		state.GraphifyDataPath,
		state.OpenCodePlugins,
	), PhaseContinue, nil
}

func buildPhase5Result(memoryPath string, enableObsidian bool, obsidianVaultPath string, enableQmd bool, qmdIndexPath string, enableCodegraph bool, codegraphDataPath string, enableGraphify bool, graphifyDataPath string, opencodePlugins []string) *Phase5Result {
	if memoryPath == "" {
		memoryPath = "specs/memory"
	}
	if enableQmd && qmdIndexPath == "" {
		qmdIndexPath = ".qmd-index"
	}
	if enableCodegraph && codegraphDataPath == "" {
		codegraphDataPath = ".codegraph"
	}
	if enableGraphify && graphifyDataPath == "" {
		graphifyDataPath = "graphify-out"
	}

	return &Phase5Result{
		MemoryPath:        memoryPath,
		EnableObsidian:    enableObsidian,
		ObsidianVaultPath: obsidianVaultPath,
		EnableQmd:         enableQmd,
		QmdIndexPath:      qmdIndexPath,
		EnableCodegraph:   enableCodegraph,
		CodegraphDataPath: codegraphDataPath,
		EnableGraphify:    enableGraphify,
		GraphifyDataPath:  graphifyDataPath,
		OpenCodePlugins:   opencodePlugins,
	}
}

func defaultPhase5Result() Phase5Result {
	return Phase5Result{
		MemoryPath:        "specs/memory",
		QmdIndexPath:      ".qmd-index",
		CodegraphDataPath: ".codegraph",
		GraphifyDataPath:  "graphify-out",
	}
}

func askMemoryPath(defaultValue string, info phase5StepInfo) (string, PhaseAction, error) {
	memoryPath := defaultValue
	group := huh.NewGroup(
		huh.NewInput().Title(info.Title()).Description("Project-local default for bootstrap and housekeeping.").Placeholder("specs/memory").Value(&memoryPath),
	)

	if err := huh.NewForm(group).Run(); err != nil {
		return "", PhaseCancel, fmt.Errorf("phase 5 cancelled: %w", err)
	}

	return memoryPath, PhaseContinue, nil
}

func askEnableObsidian(defaultValue bool, info phase5StepInfo) (bool, PhaseAction, error) {
	enabled := defaultValue
	group := huh.NewGroup(
		huh.NewConfirm().Title(info.Title()).Description("Read-only discovery only by default; future config writes remain explicit.").Value(&enabled),
	)

	if err := huh.NewForm(group).Run(); err != nil {
		return false, PhaseCancel, fmt.Errorf("phase 5 cancelled: %w", err)
	}

	return enabled, PhaseContinue, nil
}

func askObsidianVaultPath(defaultValue string, info phase5StepInfo) (string, PhaseAction, error) {
	vaultPath := defaultValue
	group := huh.NewGroup(
		huh.NewInput().Title(info.Title()).Value(&vaultPath),
	)

	if err := huh.NewForm(group).Run(); err != nil {
		return "", PhaseCancel, fmt.Errorf("phase 5 cancelled: %w", err)
	}

	return vaultPath, PhaseContinue, nil
}

func askEnableQmd(defaultValue bool, info phase5StepInfo) (bool, PhaseAction, error) {
	enabled := defaultValue
	group := huh.NewGroup(
		huh.NewConfirm().Title(info.Title()).Description("Read-only retrieval allowed; sync/index writes remain approval-gated.").Value(&enabled),
	)

	if err := huh.NewForm(group).Run(); err != nil {
		return false, PhaseCancel, fmt.Errorf("phase 5 cancelled: %w", err)
	}

	return enabled, PhaseContinue, nil
}

func askQmdIndexPath(defaultValue string, info phase5StepInfo) (string, PhaseAction, error) {
	indexPath := defaultValue
	group := huh.NewGroup(
		huh.NewInput().Title(info.Title()).Placeholder(".qmd-index").Value(&indexPath),
	)

	if err := huh.NewForm(group).Run(); err != nil {
		return "", PhaseCancel, fmt.Errorf("phase 5 cancelled: %w", err)
	}

	return indexPath, PhaseContinue, nil
}

func askEnableCodegraph(defaultValue bool, info phase5StepInfo) (bool, PhaseAction, error) {
	enabled := defaultValue
	group := huh.NewGroup(
		huh.NewConfirm().Title(info.Title()).Description("Read-only drift checks allowed; sync/index writes remain approval-gated.").Value(&enabled),
	)

	if err := huh.NewForm(group).Run(); err != nil {
		return false, PhaseCancel, fmt.Errorf("phase 5 cancelled: %w", err)
	}

	return enabled, PhaseContinue, nil
}

func askCodegraphDataPath(defaultValue string, info phase5StepInfo) (string, PhaseAction, error) {
	dataPath := defaultValue
	group := huh.NewGroup(
		huh.NewInput().Title(info.Title()).Placeholder(".codegraph").Value(&dataPath),
	)

	if err := huh.NewForm(group).Run(); err != nil {
		return "", PhaseCancel, fmt.Errorf("phase 5 cancelled: %w", err)
	}

	return dataPath, PhaseContinue, nil
}

func askEnableGraphify(defaultValue bool, info phase5StepInfo) (bool, PhaseAction, error) {
	enabled := defaultValue
	group := huh.NewGroup(
		huh.NewConfirm().Title(info.Title()).Description("Read-only graph inspection allowed; graph rebuilds remain approval-gated.").Value(&enabled),
	)

	if err := huh.NewForm(group).Run(); err != nil {
		return false, PhaseCancel, fmt.Errorf("phase 5 cancelled: %w", err)
	}

	return enabled, PhaseContinue, nil
}

func askGraphifyDataPath(defaultValue string, info phase5StepInfo) (string, PhaseAction, error) {
	dataPath := defaultValue
	group := huh.NewGroup(
		huh.NewInput().Title(info.Title()).Placeholder("graphify-out").Value(&dataPath),
	)

	if err := huh.NewForm(group).Run(); err != nil {
		return "", PhaseCancel, fmt.Errorf("phase 5 cancelled: %w", err)
	}

	return dataPath, PhaseContinue, nil
}

func phase5MemoryPathStepInfo(state Phase5Result) phase5StepInfo {
	return phase5StepInfo{Current: 1, Total: phase5TotalSteps(state), StepTitle: "Memory Path"}
}

func phase5EnableObsidianStepInfo(state Phase5Result) phase5StepInfo {
	return phase5StepInfo{Current: 2, Total: phase5TotalSteps(state), StepTitle: "Enable Obsidian"}
}

func phase5ObsidianVaultPathStepInfo(state Phase5Result) phase5StepInfo {
	return phase5StepInfo{Current: 3, Total: phase5TotalSteps(state), StepTitle: "Obsidian Vault Path"}
}

func phase5EnableQmdStepInfo(state Phase5Result) phase5StepInfo {
	return phase5StepInfo{Current: 3 + boolToInt(state.EnableObsidian), Total: phase5TotalSteps(state), StepTitle: "Enable qmd"}
}

func phase5QmdIndexPathStepInfo(state Phase5Result) phase5StepInfo {
	return phase5StepInfo{Current: 4 + boolToInt(state.EnableObsidian), Total: phase5TotalSteps(state), StepTitle: "qmd Index Path"}
}

func phase5EnableCodegraphStepInfo(state Phase5Result) phase5StepInfo {
	return phase5StepInfo{Current: 4 + boolToInt(state.EnableObsidian) + boolToInt(state.EnableQmd), Total: phase5TotalSteps(state), StepTitle: "Enable Codegraph"}
}

func phase5CodegraphDataPathStepInfo(state Phase5Result, showPlugins bool) phase5StepInfo {
	return phase5StepInfo{Current: 5 + boolToInt(state.EnableObsidian) + boolToInt(state.EnableQmd), Total: phase5TotalSteps(state, showPlugins), StepTitle: "Codegraph Data Path"}
}

func phase5EnableGraphifyStepInfo(state Phase5Result) phase5StepInfo {
	return phase5StepInfo{Current: 5 + boolToInt(state.EnableObsidian) + boolToInt(state.EnableQmd) + boolToInt(state.EnableCodegraph), Total: phase5TotalSteps(state), StepTitle: "Enable Graphify"}
}

func phase5GraphifyDataPathStepInfo(state Phase5Result, showPlugins bool) phase5StepInfo {
	return phase5StepInfo{Current: 6 + boolToInt(state.EnableObsidian) + boolToInt(state.EnableQmd) + boolToInt(state.EnableCodegraph), Total: phase5TotalSteps(state, showPlugins), StepTitle: "Graphify Data Path"}
}

func phase5OpenCodePluginsStepInfo(state Phase5Result, showPlugins bool) phase5StepInfo {
	return phase5StepInfo{Current: phase5TotalSteps(state, showPlugins), Total: phase5TotalSteps(state, showPlugins), StepTitle: "OpenCode Plugins"}
}

func phase5TotalSteps(state Phase5Result, showPlugins ...bool) int {
	extra := 0
	if len(showPlugins) > 0 && showPlugins[0] {
		extra = 1
	}
	return 5 + boolToInt(state.EnableObsidian) + boolToInt(state.EnableQmd) + boolToInt(state.EnableCodegraph) + boolToInt(state.EnableGraphify) + extra
}

func opencodeBinaryPresent() bool {
	_, err := exec.LookPath("opencode")
	return err == nil
}

func askOpenCodePlugins(current []string, info phase5StepInfo) ([]string, PhaseAction, error) {
	selected := append([]string(nil), current...)

	field := huh.NewMultiSelect[string]().
		Title(info.Title()).
		Description("Select OpenCode plugins to install via `opencode plugin`. Deselect to skip.").
		Options(append(opencodePluginOptions(), huh.NewOption("↩ Back", "__phase5_back__"))...).
		Value(&selected)

	if err := huh.NewForm(huh.NewGroup(field)).Run(); err != nil {
		return nil, PhaseCancel, fmt.Errorf("phase 5 cancelled: %w", err)
	}

	filtered := selected[:0]
	for _, s := range selected {
		if s != "__phase5_back__" {
			filtered = append(filtered, s)
		}
	}
	for _, s := range selected {
		if s == "__phase5_back__" {
			return nil, PhaseBack, nil
		}
	}
	return filtered, PhaseContinue, nil
}

func opencodePluginOptions() []huh.Option[string] {
	options := make([]huh.Option[string], 0, len(opencodePluginURLs))
	for _, url := range opencodePluginURLs {
		options = append(options, huh.NewOption(url, url))
	}
	return options
}

var opencodePluginURLs = []string{
	"https://github.com/Opencode-DCP/opencode-dynamic-context-pruning",
	"https://github.com/spoons-and-mirrors/subtask2",
	"https://github.com/JRedeker/opencode-shell-strategy",
	"https://github.com/boxpositron/envsitter-guard",
	"https://github.com/kdcokenny/opencode-background-agents",
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
