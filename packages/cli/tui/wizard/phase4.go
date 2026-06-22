package wizard

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/huh/v2"

	"github.com/rluisb/lazyai/packages/cli/internal/theme"
)

// Phase4Result holds the confirmation outcome from the final wizard phase.
type Phase4Result struct {
	Confirmed               bool
	InstallConsentsAccepted bool
}

// RunPhase4 runs the confirmation phase, showing a summary of what will be
// installed and asking the user to confirm or cancel.
func RunPhase4(plan *InstallPlan, result *WizardResult, nonInteractive bool, installConsents []string) (*Phase4Result, PhaseAction, error) {
	if nonInteractive {
		return &Phase4Result{Confirmed: true, InstallConsentsAccepted: false}, PhaseContinue, nil
	}
	return runPhase4Interactive(plan, result, installConsents)
}

func runPhase4Interactive(plan *InstallPlan, result *WizardResult, installConsents []string) (*Phase4Result, PhaseAction, error) {
	fmt.Println(formatDryRunSummary(plan, result))

	installConsentsAccepted := true
	if len(installConsents) > 0 {
		var consentValue string
		var consentDetails strings.Builder
		consentDetails.WriteString("Some selected MCP servers require CLI installation:\n")
		for _, hint := range installConsents {
			consentDetails.WriteString(fmt.Sprintf("  - %s\n", hint))
		}
		consentDetails.WriteString("\n")
		consentPrompt := huh.NewSelect[string]().
			Title("Approve CLI install prompts?").
			Description(consentDetails.String()).
			Options(
				huh.NewOption("Yes, I have installed or will install these", "yes"),
				huh.NewOption("No, continue without installing", "no"),
			).
			Value(&consentValue)

		consentForm := theme.NewForm(huh.NewGroup(consentPrompt).Title("Optional CLI Installs"))
		if err := consentForm.Run(); err != nil {
			return nil, PhaseCancel, fmt.Errorf("phase 4 cancelled: %w", err)
		}
		installConsentsAccepted = consentValue == "yes"
	}

	var confirmValue string
	confirmSelect := huh.NewSelect[string]().
		Title("Proceed with installation?").
		Description("Review the dry-run above. Choose Edit to change wizard answers before install.").
		Options(
			huh.NewOption("Yes, install", "yes"),
			huh.NewOption("Edit choices", "edit"),
			huh.NewOption("No, cancel", "no"),
		).
		Value(&confirmValue)

	form := theme.NewForm(huh.NewGroup(confirmSelect).Title("Review & Confirm"))
	if err := form.Run(); err != nil {
		return nil, PhaseCancel, fmt.Errorf("phase 4 cancelled: %w", err)
	}

	switch confirmValue {
	case "yes":
		return &Phase4Result{Confirmed: true, InstallConsentsAccepted: installConsentsAccepted}, PhaseContinue, nil
	case "edit":
		return &Phase4Result{Confirmed: false, InstallConsentsAccepted: installConsentsAccepted}, PhaseBack, nil
	default:
		return &Phase4Result{Confirmed: false, InstallConsentsAccepted: installConsentsAccepted}, PhaseCancel, nil
	}
}

func formatDryRunSummary(plan *InstallPlan, result *WizardResult) string {
	var sb strings.Builder
	sb.WriteString(theme.SectionHeader("Installation Dry Run"))
	sb.WriteString("\n\n")

	if result != nil && result.Phase1 != nil {
		p1 := result.Phase1
		sb.WriteString(fmt.Sprintf("  Scope: %s\n", p1.Scope))
		if p1.ProjectName != "" && p1.Scope != "global" {
			sb.WriteString(fmt.Sprintf("  Project: %s\n", p1.ProjectName))
		}
		sb.WriteString(fmt.Sprintf("  AI tools: %s\n", joinOrNone(toolIDsToStrings(p1.Tools))))
		sb.WriteString(fmt.Sprintf("  Skills: %d selected\n", len(p1.Skills)))
		sb.WriteString(fmt.Sprintf("  Agents: %d selected\n", len(p1.Agents)))
		sb.WriteString(fmt.Sprintf("  MCP preset: %s\n", normalizeMcpPreset(p1.McpPreset)))
		sb.WriteString(fmt.Sprintf("  MCP servers: %s\n", joinOrNone(p1.EnableServers)))
		sb.WriteString(fmt.Sprintf("  CLI tools: %s\n", joinOrNone(p1.CliTools)))
	}

	if result != nil && result.Phase2 != nil {
		p2 := result.Phase2
		sb.WriteString(fmt.Sprintf("  Preset: %s\n", p2.Preset))
		if p2.GitConv != nil {
			sb.WriteString(fmt.Sprintf("  Branch pattern: %s\n", p2.GitConv.BranchPattern))
			sb.WriteString(fmt.Sprintf("  Commit pattern: %s\n", p2.GitConv.CommitPattern))
			sb.WriteString(fmt.Sprintf("  Require ticket: %t\n", p2.GitConv.RequireTicket))
		}
		if len(p2.ChatModes) > 0 {
			sb.WriteString(fmt.Sprintf("  Copilot chat modes: %s\n", joinOrNone(chatModeIdsToStrings(p2.ChatModes))))
		}
		if p2.UseReversa == nil || *p2.UseReversa {
			sb.WriteString("  Project analysis: enabled\n")
		} else {
			sb.WriteString("  Project analysis: disabled\n")
		}
	}

	if result != nil && result.Phase5 != nil {
		p5 := result.Phase5
		sb.WriteString(fmt.Sprintf("  Memory path: %s\n", p5.MemoryPath))
		sb.WriteString(fmt.Sprintf("  Codegraph: %t", p5.EnableCodegraph))
		if p5.EnableCodegraph && p5.CodegraphDataPath != "" {
			sb.WriteString(fmt.Sprintf(" (%s)", p5.CodegraphDataPath))
		}
		sb.WriteString("\n")
	}

	newCount, updateCount, categories := countPlannedFiles(plan)
	if len(categories) > 0 {
		sb.WriteString("\n")
		for _, line := range formatCategoryCounts(categories) {
			sb.WriteString("  " + line + "\n")
		}
	}
	sb.WriteString(fmt.Sprintf("\n  Files: %d new, %d updates\n", newCount, updateCount))
	if plan != nil && len(plan.Conflicts) > 0 {
		sb.WriteString(fmt.Sprintf("  Conflicts: %d\n", len(plan.Conflicts)))
	}

	return sb.String()
}

func countPlannedFiles(plan *InstallPlan) (int, int, map[string]*categoryCount) {
	newCount := 0
	updateCount := 0
	categories := make(map[string]*categoryCount)
	if plan == nil {
		return newCount, updateCount, categories
	}
	for _, f := range plan.FilesToInstall {
		counts, ok := categories[f.Type]
		if !ok {
			counts = &categoryCount{}
			categories[f.Type] = counts
		}
		if f.Existing {
			counts.existing++
			updateCount++
		} else {
			counts.new++
			newCount++
		}
	}
	return newCount, updateCount, categories
}

func formatCategoryCounts(categories map[string]*categoryCount) []string {
	categoryNames := map[string]string{
		"constitution": "Constitution files",
		"specs":        "Specs dirs",
		"template":     "Templates",
		"rule":         "Rules",
		"infra":        "Infrastructure",
		"root":         "Root config files",
		"agent":        "Agent definitions",
		"skill":        "Skills",
		"prompt":       "Prompt templates",
		"mcp":          "MCP configuration",
	}
	keys := make([]string, 0, len(categories))
	for key := range categories {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	lines := make([]string, 0, len(keys))
	for _, cat := range keys {
		counts := categories[cat]
		name := categoryNames[cat]
		if name == "" {
			name = cat
		}
		var parts []string
		if counts.new > 0 {
			parts = append(parts, fmt.Sprintf("%d new", counts.new))
		}
		if counts.existing > 0 {
			parts = append(parts, fmt.Sprintf("%d updates", counts.existing))
		}
		lines = append(lines, fmt.Sprintf("%s: %s", name, strings.Join(parts, ", ")))
	}
	return lines
}

func joinOrNone(values []string) string {
	if len(values) == 0 {
		return "none"
	}
	copied := append([]string(nil), values...)
	sort.Strings(copied)
	return strings.Join(copied, ", ")
}

type categoryCount struct {
	new      int
	existing int
}
