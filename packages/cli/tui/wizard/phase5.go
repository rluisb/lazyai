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
	maxStep := phase5TotalSteps(showPlugins)
	for currentStep >= 1 && currentStep <= maxStep {
		switch currentStep {
		case 1:
			memoryPath, action, err := askMemoryPath(state.MemoryPath, phase5MemoryPathStepInfo(state, showPlugins))
			if err != nil {
				return nil, action, err
			}
			state.MemoryPath = memoryPath
			currentStep++
		case 2:
			if !showPlugins {
				currentStep++
				continue
			}
			plugins, action, err := askOpenCodePlugins(state.OpenCodePlugins, phase5OpenCodePluginsStepInfo(state, showPlugins))
			if err != nil {
				return nil, action, err
			}
			if action == PhaseBack {
				currentStep = 1
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
		memoryPath = ".specify/memory"
	}
	if enableCodegraph && codegraphDataPath == "" {
		codegraphDataPath = ".codegraph/"
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
		MemoryPath:        ".specify/memory",
		EnableObsidian:    true,
		EnableQmd:         true,
		EnableCodegraph:   true,
		CodegraphDataPath: ".codegraph/",
		EnableGraphify:    true,
		GraphifyDataPath:  "graphify-out",
	}
}

func askMemoryPath(defaultValue string, info phase5StepInfo) (string, PhaseAction, error) {
	memoryPath := defaultValue
	group := huh.NewGroup(
		huh.NewInput().Title(info.Title()).Description("Project-local default for bootstrap and housekeeping.").Placeholder(".specify/memory").Value(&memoryPath),
	)

	if err := huh.NewForm(group).Run(); err != nil {
		return "", PhaseCancel, fmt.Errorf("phase 5 cancelled: %w", err)
	}

	return memoryPath, PhaseContinue, nil
}

func phase5MemoryPathStepInfo(_ Phase5Result, showPlugins ...bool) phase5StepInfo {
	return phase5StepInfo{Current: 1, Total: phase5TotalSteps(showPlugins...), StepTitle: "Memory Path"}
}

func phase5OpenCodePluginsStepInfo(_ Phase5Result, showPlugins bool) phase5StepInfo {
	return phase5StepInfo{Current: phase5TotalSteps(showPlugins), Total: phase5TotalSteps(showPlugins), StepTitle: "OpenCode Plugins"}
}

func phase5TotalSteps(showPlugins ...bool) int {
	if len(showPlugins) > 0 && showPlugins[0] {
		return 2
	}
	return 1
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
