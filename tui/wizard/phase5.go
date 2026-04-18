package wizard

import (
	"fmt"

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
}

func RunPhase5(defaults *Phase5Result, nonInteractive bool) (*Phase5Result, PhaseAction, error) {
	if nonInteractive {
		return runPhase5NonInteractive(defaults)
	}
	return runPhase5Interactive(defaults)
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
	), PhaseContinue, nil
}

func runPhase5Interactive(defaults *Phase5Result) (*Phase5Result, PhaseAction, error) {
	state := defaultPhase5Result()
	if defaults != nil {
		state = *defaults
	}

	currentStep := 1
	for currentStep >= 1 && currentStep <= 7 {
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
			codegraphDataPath, action, err := askCodegraphDataPath(state.CodegraphDataPath, phase5CodegraphDataPathStepInfo(state))
			if err != nil {
				return nil, action, err
			}
			state.CodegraphDataPath = codegraphDataPath
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
	), PhaseContinue, nil
}

func buildPhase5Result(memoryPath string, enableObsidian bool, obsidianVaultPath string, enableQmd bool, qmdIndexPath string, enableCodegraph bool, codegraphDataPath string) *Phase5Result {
	if memoryPath == "" {
		memoryPath = "specs/memory"
	}
	if enableQmd && qmdIndexPath == "" {
		qmdIndexPath = ".qmd-index"
	}
	if enableCodegraph && codegraphDataPath == "" {
		codegraphDataPath = ".codegraph"
	}

	return &Phase5Result{
		MemoryPath:        memoryPath,
		EnableObsidian:    enableObsidian,
		ObsidianVaultPath: obsidianVaultPath,
		EnableQmd:         enableQmd,
		QmdIndexPath:      qmdIndexPath,
		EnableCodegraph:   enableCodegraph,
		CodegraphDataPath: codegraphDataPath,
	}
}

func defaultPhase5Result() Phase5Result {
	return Phase5Result{
		MemoryPath:        "specs/memory",
		QmdIndexPath:      ".qmd-index",
		CodegraphDataPath: ".codegraph",
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

func phase5CodegraphDataPathStepInfo(state Phase5Result) phase5StepInfo {
	return phase5StepInfo{Current: 5 + boolToInt(state.EnableObsidian) + boolToInt(state.EnableQmd), Total: phase5TotalSteps(state), StepTitle: "Codegraph Data Path"}
}

func phase5TotalSteps(state Phase5Result) int {
	return 4 + boolToInt(state.EnableObsidian) + boolToInt(state.EnableQmd) + boolToInt(state.EnableCodegraph)
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
