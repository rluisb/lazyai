package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ricardoborges-teachable/ai-setup/internal/preset"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
	"github.com/ricardoborges-teachable/ai-setup/tui/wizard"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize AI development environment",
	Long:  "Initialize AI development environment with selected tools, agents, and skills.",
	RunE:  runInit,
}

func init() {
	initCmd.Flags().String("scope", "", "Setup scope (global, workspace, project)")
	initCmd.Flags().StringSlice("tools", []string{}, "Tools to configure (opencode, claude-code, gemini, copilot, codex)")
	initCmd.Flags().String("preset", "", "Preset configuration name (minimal, standard, full, custom)")
	initCmd.Flags().StringSlice("features", []string{}, "Features to enable")
	initCmd.Flags().StringSlice("disable-features", []string{}, "Features to disable")
	initCmd.Flags().String("name", "", "Project name")
	initCmd.Flags().String("branch-pattern", "", "Git branch naming pattern")
	initCmd.Flags().String("commit-pattern", "", "Git commit message pattern")
	initCmd.Flags().Bool("non-interactive", false, "Run without interactive prompts")
	initCmd.Flags().Bool("force", false, "Overwrite existing files")
	initCmd.Flags().Bool("dry-run", false, "Show what would be done without making changes")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	// Parse flags.
	scopeStr, _ := cmd.Flags().GetString("scope")
	toolsStr, _ := cmd.Flags().GetStringSlice("tools")
	presetStr, _ := cmd.Flags().GetString("preset")
	featuresStr, _ := cmd.Flags().GetStringSlice("features")
	nameStr, _ := cmd.Flags().GetString("name")
	branchPattern, _ := cmd.Flags().GetString("branch-pattern")
	commitPattern, _ := cmd.Flags().GetString("commit-pattern")
	nonInteractive, _ := cmd.Flags().GetBool("non-interactive")
	force, _ := cmd.Flags().GetBool("force")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Build CLI tools from flags.
	var tools []types.ToolId
	for _, t := range toolsStr {
		tools = append(tools, types.ToolId(t))
	}

	var preset types.PresetLevel
	if presetStr != "" {
		preset = types.PresetLevel(presetStr)
	}

	homeDir, _ := os.UserHomeDir()
	targetDir, _ := os.Getwd()

	config := &wizard.WizardConfig{
		Interactive: !nonInteractive,
		Force:       force,
		DryRun:      dryRun,
		HomeDir:     homeDir,
		TargetDir:   targetDir,
		CLIScope:    types.SetupScope(scopeStr),
		CLITools:    tools,
		CLIName:     nameStr,
		CLIPreset:   preset,
		CLIFeatures: featuresStr,
		CLIBranch:   branchPattern,
		CLICommit:   commitPattern,
	}

	if nonInteractive {
		return runInitNonInteractive(config)
	}

	return runInitInteractive(config)
}

func runInitInteractive(config *wizard.WizardConfig) error {
	result, err := wizard.RunWizard(config)
	if err != nil {
		if err == wizard.ErrUserCancelled {
			fmt.Println("\nSetup cancelled.")
			return nil
		}
		return fmt.Errorf("wizard failed: %w", err)
	}

	if result == nil || result.Phase4 == nil || !result.Phase4.Confirmed {
		fmt.Println("\nSetup cancelled.")
		return nil
	}

	// TODO: Implement the actual scaffolding once the scaffold packages are ported.
	// This is where we would call the scaffold functions based on the wizard results.
	fmt.Println("\n✅ Setup confirmed! (Scaffolding not yet implemented in Go)")
	fmt.Printf("  Scope: %s\n", result.Phase1.Scope)
	fmt.Printf("  Tools: %v\n", result.Phase1.Tools)
	fmt.Printf("  Project: %s\n", result.Phase1.ProjectName)
	fmt.Printf("  Preset: %s\n", result.Phase2.Preset)

	return nil
}

func runInitNonInteractive(config *wizard.WizardConfig) error {
	// Non-interactive mode: use CLI flags directly.
	// Validate required flags.
	if config.CLIScope == "" {
		return fmt.Errorf("--scope is required in non-interactive mode (global | workspace | project)")
	}
	if len(config.CLITools) == 0 {
		return fmt.Errorf("--tools is required in non-interactive mode (opencode, claude-code, gemini, copilot, codex)")
	}
	if config.CLIName == "" {
		// Default project name to directory name.
		dir, _ := os.Getwd()
		if dir != "" {
			config.CLIName = filepath.Base(dir)
		} else {
			config.CLIName = "my-project"
		}
	}

	// Determine preset.
	presetLevel := config.CLIPreset
	if presetLevel == "" {
		presetLevel = types.PresetLevelStandard
	}

	// Create Phase1 result from CLI flags.
	phase1 := &wizard.Phase1Result{
		Scope:       config.CLIScope,
		Tools:       config.CLITools,
		ProjectName: config.CLIName,
	}

	// Create Phase2 result from preset + features.
	features := types.DefaultFeatureFlags()
	resolved := preset.ResolvePreset(presetLevel)
	if resolved != nil {
		features = *resolved
	}
	// Apply CLI feature overrides.
	for _, f := range config.CLIFeatures {
		switch f {
		case "contextEngineering":
			features.ContextEngineering = true
		case "rpiWorkflow":
			features.RPIWorkflow = true
		case "chainOfThought":
			features.ChainOfThought = true
		case "treeOfThoughts":
			features.TreeOfThoughts = true
		case "adrEnforcement":
			features.ADREnforcement = true
		case "qualityGates":
			features.QualityGates = true
		case "agentHarness":
			features.AgentHarness = true
		case "bugResolution":
			features.BugResolution = true
		case "pivotHandling":
			features.PivotHandling = true
		}
	}
	gitConvs := types.DefaultGitConventions()
	if config.CLIBranch != "" {
		gitConvs.BranchPattern = config.CLIBranch
	}
	if config.CLICommit != "" {
		gitConvs.CommitPattern = config.CLICommit
	}

	phase2 := &wizard.Phase2Result{
		Preset:   presetLevel,
		Features: &features,
		GitConv:  &gitConvs,
	}

	// Build WizardResult with pre-computed phases.
	wizardDefaults := &wizard.WizardResult{
		Phase1: phase1,
		Phase2: phase2,
	}

	// Run the wizard in non-interactive mode — it will use the defaults we provided.
	result, err := wizard.RunWizardWithDefaults(config, wizardDefaults)
	if err != nil {
		return fmt.Errorf("non-interactive setup failed: %w", err)
	}

	if result == nil || result.Phase4 == nil || !result.Phase4.Confirmed {
		// In non-interactive mode, confirmation is automatic.
		// This shouldn't happen, but handle it gracefully.
		return nil
	}

	// TODO: Implement the actual scaffolding once the scaffold packages are ported.
	fmt.Println("Setup complete! (Scaffolding not yet implemented in Go)")

	return nil
}
