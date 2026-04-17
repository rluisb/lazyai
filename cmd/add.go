package cmd

import (
	"fmt"
	"os"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"

	"github.com/ricardoborges-teachable/ai-setup/internal/db"
	"github.com/ricardoborges-teachable/ai-setup/internal/library"
	"github.com/ricardoborges-teachable/ai-setup/internal/preset"
	"github.com/ricardoborges-teachable/ai-setup/internal/scaffold"
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

	// Merge new selections into existing ones.
	storeData.Config.Tools = mergeToolIds(storeData.Config.Tools, newTools)
	storeData.Selections.Agents = mergeAgentIds(storeData.Selections.Agents, newAgents)
	storeData.Selections.Skills = mergeSkillIds(storeData.Selections.Skills, newSkills)

	// Also merge CLI tools.
	cliTools := mergeStringSlices(storeData.Config.CLITools, toolIdsToStrings(newTools))
	storeData.Config.CLITools = cliTools

	// Determine preset from scope.
	presetLevel := preset.DefaultPresetForScope(storeData.Config.SetupScope)

	// Build scaffold context from merged configuration.
	// LibraryDir may be empty in production mode (embedded FS).
	libDir := getLibraryDir()
	libFS := library.GetLibraryFS()

	ctx := &scaffold.ScaffoldContext{
		TargetDir:      targetDir,
		LibraryDir:     libDir,
		LibraryFS:      libFS,
		Tools:          storeData.Config.Tools,
		CLITools:       storeData.Config.CLITools,
		EnableServers:  storeData.Config.EnableServers,
		ProjectName:    storeData.Config.ProjectName,
		PlanningDir:    storeData.Config.PlanningDir,
		SetupScope:     storeData.Config.SetupScope,
		Features:       storeData.Selections.Features,
		GitConventions: storeData.Selections.GitConventions,
		Strategy:       types.ConflictStrategyAlign,
		Agents:         storeData.Selections.Agents,
		Skills:         storeData.Selections.Skills,
		Prompts:        storeData.Selections.Prompts,
		Templates:      storeData.Selections.Templates,
		Rules:          storeData.Selections.Rules,
		Infra:          storeData.Selections.Infra,
		SpecsDirs:      preset.SpecsDirsForPreset(presetLevel),
		Housekeeping:   storeData.Config.Housekeeping,
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
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	greenStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))

	fmt.Println()
	fmt.Println(headerStyle.Render("✅ Artifacts added!"))
	fmt.Println()
	if len(newTools) > 0 {
		fmt.Printf("  %s Tools: %v\n", greenStyle.Render("•"), newTools)
	}
	if len(newAgents) > 0 {
		fmt.Printf("  %s Agents: %v\n", greenStyle.Render("•"), newAgents)
	}
	if len(newSkills) > 0 {
		fmt.Printf("  %s Skills: %v\n", greenStyle.Render("•"), newSkills)
	}
	fmt.Printf("  %s Files updated: %d\n", greenStyle.Render("•"), len(result.Files))

	if len(result.Errors) > 0 {
		warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))
		fmt.Println()
		fmt.Println(warnStyle.Render(fmt.Sprintf("⚠ %d warnings:", len(result.Errors))))
		for _, e := range result.Errors {
			fmt.Printf("  • %v\n", e)
		}
	}
	fmt.Println()
	return nil
}

// mergeToolIds merges new tool IDs into existing, avoiding duplicates.
func mergeToolIds(existing, new []types.ToolId) []types.ToolId {
	seen := make(map[types.ToolId]bool)
	for _, t := range existing {
		seen[t] = true
	}
	for _, t := range new {
		if !seen[t] {
			existing = append(existing, t)
			seen[t] = true
		}
	}
	return existing
}

// mergeAgentIds merges new agent IDs into existing, avoiding duplicates.
func mergeAgentIds(existing []types.AgentId, new []string) []types.AgentId {
	seen := make(map[types.AgentId]bool)
	for _, a := range existing {
		seen[a] = true
	}
	var result []types.AgentId
	result = append(result, existing...)
	for _, a := range new {
		id := types.AgentId(a)
		if !seen[id] {
			result = append(result, id)
			seen[id] = true
		}
	}
	return result
}

// mergeSkillIds merges new skill IDs into existing, avoiding duplicates.
func mergeSkillIds(existing []types.SkillId, new []string) []types.SkillId {
	seen := make(map[types.SkillId]bool)
	for _, s := range existing {
		seen[s] = true
	}
	var result []types.SkillId
	result = append(result, existing...)
	for _, s := range new {
		id := types.SkillId(s)
		if !seen[id] {
			result = append(result, id)
			seen[id] = true
		}
	}
	return result
}

// mergeStringSlices merges two string slices, removing duplicates.
func mergeStringSlices(existing, new []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range existing {
		if !seen[s] {
			result = append(result, s)
			seen[s] = true
		}
	}
	for _, s := range new {
		if !seen[s] {
			result = append(result, s)
			seen[s] = true
		}
	}
	return result
}

// toolIdsToStrings converts tool IDs to plain strings.
func toolIdsToStrings(ids []types.ToolId) []string {
	result := make([]string, len(ids))
	for i, id := range ids {
		result[i] = string(id)
	}
	return result
}
