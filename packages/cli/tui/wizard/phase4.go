package wizard

import (
	"fmt"
	"strings"

	"charm.land/huh/v2"

	"github.com/rluisb/lazyai/packages/cli/internal/theme"
)

// Phase4Result holds the confirmation outcome from the final wizard phase.
type Phase4Result struct {
	Confirmed              bool
	InstallConsentsAccepted bool
}

// RunPhase4 runs the confirmation phase, showing a summary of what will be
// installed and asking the user to confirm or cancel.
func RunPhase4(plan *InstallPlan, nonInteractive bool, installConsents []string) (*Phase4Result, PhaseAction, error) {
	if nonInteractive {
		return &Phase4Result{Confirmed: true, InstallConsentsAccepted: false}, PhaseContinue, nil
	}
	return runPhase4Interactive(plan, installConsents)
}

func runPhase4Interactive(plan *InstallPlan, installConsents []string) (*Phase4Result, PhaseAction, error) {
	// Build a summary string.
	var sb strings.Builder

	sb.WriteString(theme.SectionHeader("Setup Summary"))
	sb.WriteString("\n\n")

	// File counts by category.
	newCount := 0
	updateCount := 0
	categories := make(map[string]*categoryCount)

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

	// Category display names.
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

	for cat, counts := range categories {
		name := categoryNames[cat]
		if name == "" {
			name = cat
		}
		var parts []string
		if counts.new > 0 {
			parts = append(parts, fmt.Sprintf("%d new", counts.new))
		}
		if counts.existing > 0 {
			parts = append(parts, fmt.Sprintf("%d existing", counts.existing))
		}
		sb.WriteString(fmt.Sprintf("  %s: %s\n", name, strings.Join(parts, ", ")))
	}

	sb.WriteString(fmt.Sprintf("\n  Total: %d new, %d updates\n", newCount, updateCount))

	if len(plan.Conflicts) > 0 {
		sb.WriteString(fmt.Sprintf("  Conflicts: %d\n", len(plan.Conflicts)))
	}

	// Show the summary.
	fmt.Println(sb.String())

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

	// Confirm.
	var confirmValue string
	confirmSelect := huh.NewSelect[string]().
		Title("Proceed with installation?").
		Options(
			huh.NewOption("Yes, install", "yes"),
			huh.NewOption("No, cancel", "no"),
		).
		Value(&confirmValue)

	form := theme.NewForm(huh.NewGroup(confirmSelect).Title("Review & Confirm"))
	if err := form.Run(); err != nil {
		return nil, PhaseCancel, fmt.Errorf("phase 4 cancelled: %w", err)
	}

	if confirmValue == "yes" {
		return &Phase4Result{Confirmed: true, InstallConsentsAccepted: installConsentsAccepted}, PhaseContinue, nil
	}

	return &Phase4Result{Confirmed: false, InstallConsentsAccepted: installConsentsAccepted}, PhaseCancel, nil
}

type categoryCount struct {
	new      int
	existing int
}
