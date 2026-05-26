package cmd

import (
	"fmt"
	"os"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"

	"github.com/rluisb/lazyai/packages/cli/internal/db"
	"github.com/rluisb/lazyai/packages/cli/internal/library"
	"github.com/rluisb/lazyai/packages/cli/internal/scaffold"
	setupsvc "github.com/rluisb/lazyai/packages/cli/internal/setup"
	"github.com/rluisb/lazyai/packages/cli/internal/theme"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add artifacts to an existing setup",
	Long:  "Add agents, skills, or tool configurations to an existing AI setup.",
	RunE:  runAdd,
}

func init() {
	addCmd.Flags().StringSlice("tools", []string{}, "Tools to add configuration for")
	addCmd.Flags().StringSlice("agents", []string{}, "Agents to add")
	addCmd.Flags().StringSlice("skills", []string{}, "Skills to add")
	addCmd.Flags().Bool("no-interactive", false, "Run without interactive prompts")
	rootCmd.AddCommand(addCmd)
	addCmd.GroupID = "scaffold"
}

func runAdd(cmd *cobra.Command, args []string) error {
	nonInteractive, _ := cmd.Flags().GetBool("no-interactive")

	// Parse flags.
	toolsStr, _ := cmd.Flags().GetStringSlice("tools")
	agentsStr, _ := cmd.Flags().GetStringSlice("agents")
	skillsStr, _ := cmd.Flags().GetStringSlice("skills")

	// Convert to domain types.
	var tools []types.ToolId
	for _, t := range toolsStr {
		tools = append(tools, types.ToolId(t))
	}

	if nonInteractive {
		return runAddNonInteractive(tools, agentsStr, skillsStr)
	}
	return runAddInteractive(tools, agentsStr, skillsStr)
}

func runAddInteractive(tools []types.ToolId, agents, skills []string) error {
	// If no tools specified, ask interactively.
	if len(tools) == 0 {
		var toolStrs []string
		toolSelect := huh.NewMultiSelect[string]().
			Title("Which AI tools to add configuration for?").
			Options(
				huh.NewOption("OpenCode", "opencode"),
				huh.NewOption("Claude Code", "claude-code"),
				huh.NewOption("GitHub Copilot", "copilot"),
			).
			Value(&toolStrs)

		if err := theme.NewForm(huh.NewGroup(toolSelect)).Run(); err != nil {
			return fmt.Errorf("cancelled: %w", err)
		}
		tools = make([]types.ToolId, len(toolStrs))
		for i, t := range toolStrs {
			tools[i] = types.ToolId(t)
		}
	}

	// If no agents specified, ask interactively.
	if len(agents) == 0 {
		var agentStrs []string
		agentSelect := huh.NewMultiSelect[string]().
			Title("Which agents to add?").
			Options(
				huh.NewOption("Builder", "builder"),
				huh.NewOption("Documenter", "documenter"),
				huh.NewOption("Orchestrator", "orchestrator"),
				huh.NewOption("Planner", "planner"),
				huh.NewOption("Red Team", "red-team"),
				huh.NewOption("Reviewer", "reviewer"),
				huh.NewOption("Scout", "scout"),
			).
			Value(&agentStrs)

		if err := theme.NewForm(huh.NewGroup(agentSelect)).Run(); err != nil {
			return fmt.Errorf("cancelled: %w", err)
		}
		agents = agentStrs
	}

	// If no skills specified, ask interactively.
	if len(skills) == 0 {
		var skillStrs []string
		skillSelect := huh.NewMultiSelect[string]().
			Title("Which skills to add?").
			Options(
				huh.NewOption("Anti-Speculation", "anti-speculation"),
				huh.NewOption("Extract Standards", "extract-standards"),
				huh.NewOption("Implement", "implement"),
				huh.NewOption("Iterate", "iterate"),
				huh.NewOption("Memory Write", "memory-write"),
				huh.NewOption("Orchestrate", "orchestrate"),
				huh.NewOption("Parallel Execution", "parallel-execution"),
				huh.NewOption("Plan", "plan"),
				huh.NewOption("Research", "research"),
				huh.NewOption("TDD Loop", "tdd-loop"),
			).
			Value(&skillStrs)

		if err := theme.NewForm(huh.NewGroup(skillSelect)).Run(); err != nil {
			return fmt.Errorf("cancelled: %w", err)
		}
		skills = skillStrs
	}

	return runAddWithSelections(tools, agents, skills)
}

func runAddNonInteractive(tools []types.ToolId, agents, skills []string) error {
	if len(tools) == 0 && len(agents) == 0 && len(skills) == 0 {
		return fmt.Errorf("at least one of --tools, --agents, or --skills is required in non-interactive mode")
	}
	return runAddWithSelections(tools, agents, skills)
}

// runAddWithSelections adds new artifacts to an existing setup by merging
// selections and re-running the scaffold pipeline.
func runAddWithSelections(newTools []types.ToolId, newAgents, newSkills []string) error {
	targetDir, _ := os.Getwd()

	// Open store and read existing data.
	database, err := openStore(targetDir)
	if err != nil {
		return fmt.Errorf("opening store: %w", err)
	}
	defer database.Close()

	store := db.NewStore(database)
	storeData, err := store.ReadStoreData()
	if err != nil {
		return fmt.Errorf("reading store data: %w", err)
	}

	ctx, presetLevel, err := setupsvc.BuildAddScaffoldContext(targetDir, setupsvc.Library{
		Dir: getLibraryDir(),
		FS:  library.GetLibraryFS(),
	}, storeData, setupsvc.AddSelections{
		Tools:  newTools,
		Agents: newAgents,
		Skills: newSkills,
	})
	if err != nil {
		return err
	}

	// Run the scaffold pipeline.
	result, err := scaffold.ScaffoldAll(ctx)
	if err != nil {
		return fmt.Errorf("scaffold failed: %w", err)
	}

	// Update the store with merged data and new file records.
	if err := writeStoreFromScaffoldResult(database, ctx, presetLevel, result); err != nil {
		return fmt.Errorf("writing store data: %w", err)
	}

	// Print summary.
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Primary)
	bulletStyle := lipgloss.NewStyle().Foreground(theme.Success)

	fmt.Println()
	// Replaces the prior `✅ Artifacts added!` (✅ is an emoji — off-design;
	// canonical glyph is ✓ per the lazyai-design-system skill).
	fmt.Println(headerStyle.Render(theme.GlyphSuccess + " Artifacts added!"))
	fmt.Println()
	if len(newTools) > 0 {
		fmt.Printf("  %s Tools: %v\n", bulletStyle.Render(theme.GlyphBullet), newTools)
	}
	if len(newAgents) > 0 {
		fmt.Printf("  %s Agents: %v\n", bulletStyle.Render(theme.GlyphBullet), newAgents)
	}
	if len(newSkills) > 0 {
		fmt.Printf("  %s Skills: %v\n", bulletStyle.Render(theme.GlyphBullet), newSkills)
	}
	fmt.Printf("  %s Files updated: %d\n", bulletStyle.Render(theme.GlyphBullet), len(result.Files))

	if len(result.Errors) > 0 {
		fmt.Println()
		theme.Warnf(os.Stderr, "%d warnings:", len(result.Errors))
		for _, e := range result.Errors {
			fmt.Printf("  %s %v\n", theme.GlyphBullet, e)
		}
	}
	fmt.Println()
	return nil
}
