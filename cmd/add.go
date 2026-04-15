package cmd

import (
	"fmt"

	"charm.land/huh/v2"

	"github.com/spf13/cobra"

	"github.com/ricardoborges-teachable/ai-setup/internal/types"
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
	addCmd.Flags().Bool("non-interactive", false, "Run without interactive prompts")
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	nonInteractive, _ := cmd.Flags().GetBool("non-interactive")

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
				huh.NewOption("Gemini CLI", "gemini"),
				huh.NewOption("GitHub Copilot", "copilot"),
				huh.NewOption("Codex (OpenAI)", "codex"),
			).
			Value(&toolStrs)

		if err := huh.NewForm(huh.NewGroup(toolSelect)).Run(); err != nil {
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

		if err := huh.NewForm(huh.NewGroup(agentSelect)).Run(); err != nil {
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

		if err := huh.NewForm(huh.NewGroup(skillSelect)).Run(); err != nil {
			return fmt.Errorf("cancelled: %w", err)
		}
		skills = skillStrs
	}

	// TODO: Implement actual add logic once scaffold packages are ported.
	fmt.Printf("Would add: tools=%v, agents=%v, skills=%v\n", tools, agents, skills)
	return nil
}

func runAddNonInteractive(tools []types.ToolId, agents, skills []string) error {
	// TODO: Implement actual add logic once scaffold packages are ported.
	fmt.Printf("Would add: tools=%v, agents=%v, skills=%v\n", tools, agents, skills)
	return nil
}
