package wizard

import (
	"fmt"

	"charm.land/huh/v2"

	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// Phase1Result holds the collected context from the first wizard phase.
type Phase1Result struct {
	Scope       types.SetupScope
	Tools       []types.ToolId
	ProjectName string
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
	result := &Phase1Result{}
	if defaults != nil {
		*result = *defaults
	}

	// Pre-fill defaults for fields that have defaults.
	var scopeValue string
	var toolsValue []string
	var nameValue string

	scopeDefault := "project"
	nameDefault := "my-project"
	if defaults != nil {
		if defaults.Scope != "" {
			scopeDefault = string(defaults.Scope)
		}
		if len(defaults.Tools) > 0 {
			toolStrs := make([]string, len(defaults.Tools))
			for i, t := range defaults.Tools {
				toolStrs[i] = string(t)
			}
			toolsValue = toolStrs
		}
		if defaults.ProjectName != "" {
			nameDefault = defaults.ProjectName
		}
	}
	scopeValue = scopeDefault
	nameValue = nameDefault

	// --- Scope selection ---
	scopeSelect := huh.NewSelect[string]().
		Title("Setup scope:").
		Options(
			huh.NewOption("Global  — Install to ~/.ai/ + native tool global paths", "global"),
			huh.NewOption("Workspace  — Planning repo with multi-project management", "workspace"),
			huh.NewOption("Project (recommended)  — Self-contained single repository", "project"),
		).
		Value(&scopeValue)

	// --- Tools multi-select ---
	toolOptions := []huh.Option[string]{
		huh.NewOption("OpenCode", "opencode"),
		huh.NewOption("Claude Code", "claude-code"),
		huh.NewOption("Gemini CLI", "gemini"),
		huh.NewOption("GitHub Copilot", "copilot"),
		huh.NewOption("Codex (OpenAI)", "codex"),
	}
	// Pre-select previously chosen tools.
	selectedTools := toolsValue
	toolsMulti := huh.NewMultiSelect[string]().
		Title("Which AI tools are you using?").
		Options(toolOptions...).
		Value(&selectedTools)

	// --- Project name input ---
	nameInput := huh.NewInput().
		Title("Project name:").
		Placeholder(nameDefault).
		Value(&nameValue).
		Validate(func(s string) error {
			if s == "" {
				return fmt.Errorf("project name is required")
			}
			return nil
		})

	group := huh.NewGroup(scopeSelect, toolsMulti, nameInput).Title("Phase 1/4: Setup Context")

	form := huh.NewForm(group)
	if err := form.Run(); err != nil {
		return nil, PhaseCancel, fmt.Errorf("phase 1 cancelled: %w", err)
	}

	// Map results.
	result.Scope = types.SetupScope(scopeValue)
	result.ProjectName = nameValue
	result.Tools = make([]types.ToolId, len(selectedTools))
	for i, t := range selectedTools {
		result.Tools[i] = types.ToolId(t)
	}

	// If scope is global, force project name to "global"
	if result.Scope == types.SetupScopeGlobal {
		result.ProjectName = "global"
	}

	return result, PhaseContinue, nil
}
