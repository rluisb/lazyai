package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rluisb/lazyai/packages/cli/internal/adapter"
	"github.com/rluisb/lazyai/packages/cli/internal/preset"
	"github.com/rluisb/lazyai/packages/cli/internal/scaffold"
	"github.com/rluisb/lazyai/packages/cli/internal/theme"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
	"github.com/rluisb/lazyai/packages/cli/tui/components"
	"github.com/rluisb/lazyai/packages/cli/tui/wizard"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize AI development environment",
	Long:  "Initialize AI development environment with selected tools, agents, and skills.",
	RunE:  runInit,
}

func init() {
	initCmd.Flags().String("scope", "", "Setup scope (global, workspace, project)")
	initCmd.Flags().String("workspace-root", "", "Workspace root directory for AI tool configs (workspace scope)")
	initCmd.Flags().StringSlice("tools", []string{}, "Tools to configure (opencode, claude-code, copilot, pi, omp, kiro, antigravity, codex)")
	initCmd.Flags().StringSlice("enable-servers", []string{}, "MCP servers to enable (for example: filesystem, ai-memory, ripgrep)")
	initCmd.Flags().String("preset", "", "Preset configuration name (minimal, standard, full, custom)")
	initCmd.Flags().StringSlice("features", []string{}, "Features to enable")
	initCmd.Flags().StringSlice("disable-features", []string{}, "Features to disable")
	initCmd.Flags().String("name", "", "Project name")
	initCmd.Flags().String("branch-pattern", "", "Git branch naming pattern")
	initCmd.Flags().String("commit-pattern", "", "Git commit message pattern")
	initCmd.Flags().String("existing-setup-policy", string(types.SetupPolicyAbsorb), "How to handle existing setup (absorb, adapt, backup-only)")
	initCmd.Flags().Bool("no-interactive", false, "Run without interactive prompts")
	initCmd.Flags().Bool("drive-cli", false, "Delegate scaffolding to the tool's own CLI when available (Claude Code)")
	initCmd.Flags().Bool("local-secrets", false, "Route Claude Code MCP/settings writes to gitignored .claude/settings.local.json instead of committed surfaces")
	initCmd.Flags().String("org", "", "Organization name (populates [YOUR_ORG] in AGENTS.md)")
	initCmd.Flags().String("team", "", "Team name (populates [YOUR_TEAM] in AGENTS.md)")
	initCmd.Flags().Bool("force", false, "Overwrite existing files")
	initCmd.Flags().Bool("dry-run", false, "Show what would be done without making changes")
	initCmd.Flags().String("memory-path", "", "Project memory path (default: .specify/memory)")
	initCmd.Flags().Bool("reversa", false, "Analyze existing code with Scout/Reversa to auto-populate project details")
	initCmd.Flags().Bool("no-reversa", false, "Skip Scout/Reversa analysis and leave project details explicit/manual")
	initCmd.Flags().Bool("express", false, "Run interactive wizard with Express mode")
	initCmd.Flags().Bool("custom", false, "Run interactive wizard with Personalized mode")
	rootCmd.AddCommand(initCmd)
	initCmd.GroupID = "lifecycle"
}

func runInit(cmd *cobra.Command, args []string) error {
	// Parse flags.
	scopeStr, _ := cmd.Flags().GetString("scope")
	workspaceRoot, _ := cmd.Flags().GetString("workspace-root")
	toolsStr, _ := cmd.Flags().GetStringSlice("tools")
	presetStr, _ := cmd.Flags().GetString("preset")
	featuresStr, _ := cmd.Flags().GetStringSlice("features")
	nameStr, _ := cmd.Flags().GetString("name")
	branchPattern, _ := cmd.Flags().GetString("branch-pattern")
	commitPattern, _ := cmd.Flags().GetString("commit-pattern")
	memoryPath, _ := cmd.Flags().GetString("memory-path")
	nonInteractive, _ := cmd.Flags().GetBool("no-interactive")
	expressMode, _ := cmd.Flags().GetBool("express")
	customMode, _ := cmd.Flags().GetBool("custom")
	force, _ := cmd.Flags().GetBool("force")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	driveCLI, _ := cmd.Flags().GetBool("drive-cli")
	localSecrets, _ := cmd.Flags().GetBool("local-secrets")
	orgName, _ := cmd.Flags().GetString("org")
	teamName, _ := cmd.Flags().GetString("team")
	enableServersStr, _ := cmd.Flags().GetStringSlice("enable-servers")
	existingSetupPolicyStr, _ := cmd.Flags().GetString("existing-setup-policy")
	useReversa, err := resolveReversaFlagChoice(cmd)
	if err != nil {
		return err
	}
	existingSetupPolicy, err := parseExistingSetupPolicy(existingSetupPolicyStr)
	if err != nil {
		return err
	}

	// Build CLI tools from flags.
	var tools []types.ToolId
	for _, t := range toolsStr {
		tools = append(tools, types.ToolId(t))
	}

	var preset types.PresetLevel
	if presetStr != "" {
		preset = types.PresetLevel(presetStr)
	}

	if expressMode && customMode {
		return fmt.Errorf("--express and --custom cannot be used together")
	}

	var wizardMode wizard.WizardMode
	if expressMode {
		wizardMode = wizard.WizardModeExpress
	} else if customMode {
		wizardMode = wizard.WizardModePersonalized
	}

	homeDir, _ := os.UserHomeDir()
	targetDir, _ := os.Getwd()

	config := &wizard.WizardConfig{
		Interactive:            !nonInteractive,
		Force:                  force,
		DryRun:                 dryRun,
		CLIDriveCLI:            driveCLI,
		CLILocalSecrets:        localSecrets,
		CLIOrg:                 orgName,
		CLITeam:                teamName,
		CLIEnableServers:       enableServersStr,
		HomeDir:                homeDir,
		TargetDir:              targetDir,
		CLIScope:               types.SetupScope(scopeStr),
		CLITools:               tools,
		CLIName:                nameStr,
		CLIPreset:              preset,
		CLIFeatures:            featuresStr,
		CLIBranch:              branchPattern,
		CLICommit:              commitPattern,
		CLIWorkspaceRoot:       workspaceRoot,
		CLIMemoryPath:          memoryPath,
		CLIExistingSetupPolicy: existingSetupPolicy,
		CLIUseReversa:          useReversa,
		CLIWizardMode:          wizardMode,
	}

	if nonInteractive {
		return runInitNonInteractive(config)
	}

	return runInitInteractive(config)
}

func parseExistingSetupPolicy(value string) (types.SetupPolicy, error) {
	if value == "" {
		return types.SetupPolicyAbsorb, nil
	}

	policy := types.SetupPolicy(value)
	if !types.IsValidSetupPolicy(policy) {
		return "", fmt.Errorf("invalid --existing-setup-policy %q (expected absorb, adapt, or backup-only)", value)
	}
	return policy, nil
}

func resolveReversaFlagChoice(cmd *cobra.Command) (*bool, error) {
	flags := cmd.Flags()
	reversaChanged := flags.Changed("reversa")
	noReversaChanged := flags.Changed("no-reversa")
	if reversaChanged && noReversaChanged {
		return nil, fmt.Errorf("--reversa and --no-reversa cannot be used together")
	}

	if reversaChanged {
		value, _ := flags.GetBool("reversa")
		return &value, nil
	}
	if noReversaChanged {
		value, _ := flags.GetBool("no-reversa")
		enabled := !value
		return &enabled, nil
	}
	return nil, nil
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

	// Build scaffold context from wizard results.
	ctx, err := buildScaffoldContext(result, config)
	if err != nil {
		return fmt.Errorf("building scaffold context: %w", err)
	}
	if ctx.DryRun {
		fmt.Println("[dry-run] Would create LazyAI files for:")
		fmt.Printf("  • Scope: %s\n", ctx.SetupScope)
		fmt.Printf("  • Tools: %v\n", ctx.Tools)
		fmt.Printf("  • Project: %s\n", ctx.ProjectName)
		fmt.Println("Dry run complete. No files written.")
		return nil
	}

	// Run the scaffold pipeline.
	scaffoldResult, err := scaffold.ScaffoldAll(ctx)
	if err != nil {
		return fmt.Errorf("scaffold failed: %w", err)
	}
	if err := compileMCPForInit(ctx, scaffoldResult); err != nil {
		return err
	}

	fmt.Println()
	theme.Infof(os.Stdout, "AGENTS.md placeholders left for your AI tool to fill.")
	theme.Infof(os.Stdout, "Headless init skipped: this CLI cannot run your AI tool to fill the placeholders automatically.")
	theme.Infof(os.Stdout, "Next: open the project in your AI tool and run /init or /populate, or edit AGENTS.md directly to replace each <!-- fill-in: ... --> marker.")
	theme.Infof(os.Stdout, "Validate after populate: lazyai-cli validate agents && lazyai-cli doctor")

	// Persist results to the SQLite store.
	database, err := openStore(config.TargetDir)
	if err != nil {
		return fmt.Errorf("opening store: %w", err)
	}
	defer database.Close()

	presetLevel := result.Phase2.Preset
	if presetLevel == "" {
		presetLevel = preset.DefaultPresetForScope(result.Phase1.Scope)
	}

	if err := writeStoreFromScaffoldResult(database, ctx, presetLevel, scaffoldResult); err != nil {
		return fmt.Errorf("writing store data: %w", err)
	}

	// Print summary.
	fmt.Println()
	printScaffoldSummary(scaffoldResult, ctx)
	fmt.Println()

	if catalog := adapter.ReadCanonicalMcp(ctx.TargetDir); catalog != nil {
		PrintMcpNextSteps(adapter.GetEnabledServers(catalog))
	}

	printInitNextSteps(ctx)

	return nil
}

func runInitNonInteractive(config *wizard.WizardConfig) error {
	// Non-interactive mode: use CLI flags directly.
	// Validate required flags.
	if config.CLIScope == "" {
		return fmt.Errorf("--scope is required in non-interactive mode (global | workspace | project)")
	}
	if len(config.CLITools) == 0 {
		return fmt.Errorf("--tools is required in non-interactive mode (opencode, claude-code, copilot, pi, antigravity, omp, kiro, codex)")
	}

	// Drop tools that don't support the chosen scope (e.g. copilot × global).
	// Warn per drop; abort only if nothing is left.
	kept := make([]types.ToolId, 0, len(config.CLITools))
	for _, t := range config.CLITools {
		if adapter.IsScopeSupported(t, config.CLIScope) {
			kept = append(kept, t)
			continue
		}
		cmdLog.Warn("skipping unsupported tool for scope", "tool", t, "scope", config.CLIScope)
	}
	if len(kept) == 0 {
		return fmt.Errorf("no tools remain after filtering for scope %q", config.CLIScope)
	}
	config.CLITools = kept
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
		Scope:         config.CLIScope,
		Tools:         config.CLITools,
		ProjectName:   config.CLIName,
		Organization:  config.CLIOrg,
		Team:          config.CLITeam,
		CliTools:      config.CLICliTools,
		EnableServers: config.CLIEnableServers,
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
		Preset:     presetLevel,
		Features:   &features,
		GitConv:    &gitConvs,
		UseReversa: config.CLIUseReversa,
	}

	// Build WizardResult with pre-computed phases.
	wizardDefaults := &wizard.WizardResult{
		Phase1: phase1,
		Phase2: phase2,
		Phase5: &wizard.Phase5Result{
			MemoryPath:        config.CLIMemoryPath,
			EnableObsidian:    true,
			ObsidianVaultPath: "",
			EnableCodegraph:   true,
			CodegraphDataPath: ".codegraph/",
		},
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

	// Build scaffold context from wizard results.
	ctx, err := buildScaffoldContext(result, config)
	if err != nil {
		return fmt.Errorf("building scaffold context: %w", err)
	}
	if ctx.DryRun {
		fmt.Println("[dry-run] Would create LazyAI files for:")
		fmt.Printf("  • Scope: %s\n", ctx.SetupScope)
		fmt.Printf("  • Tools: %v\n", ctx.Tools)
		fmt.Printf("  • Project: %s\n", ctx.ProjectName)
		fmt.Println("Dry run complete. No files written.")
		return nil
	}

	// Run the scaffold pipeline.
	scaffoldResult, err := scaffold.ScaffoldAll(ctx)
	if err != nil {
		return fmt.Errorf("scaffold failed: %w", err)
	}
	if err := compileMCPForInit(ctx, scaffoldResult); err != nil {
		return err
	}

	fmt.Println()
	theme.Infof(os.Stdout, "AGENTS.md placeholders left for your AI tool to fill.")
	theme.Infof(os.Stdout, "Headless init skipped: this CLI cannot run your AI tool to fill the placeholders automatically.")
	theme.Infof(os.Stdout, "Next: open the project in your AI tool and run /init or /populate, or edit AGENTS.md directly to replace each <!-- fill-in: ... --> marker.")
	theme.Infof(os.Stdout, "Validate after populate: lazyai-cli validate agents && lazyai-cli doctor")

	// Persist results to the SQLite store.
	database, err := openStore(config.TargetDir)
	if err != nil {
		return fmt.Errorf("opening store: %w", err)
	}
	defer database.Close()

	if err := writeStoreFromScaffoldResult(database, ctx, presetLevel, scaffoldResult); err != nil {
		return fmt.Errorf("writing store data: %w", err)
	}

	// Print summary.
	fmt.Println()
	printScaffoldSummary(scaffoldResult, ctx)
	fmt.Println()

	if catalog := adapter.ReadCanonicalMcp(ctx.TargetDir); catalog != nil {
		PrintMcpNextSteps(adapter.GetEnabledServers(catalog))
	}

	printInitNextSteps(ctx)

	return nil
}

// printScaffoldSummary prints a human-readable summary of the scaffold results.
func printScaffoldSummary(result *scaffold.ScaffoldResult, ctx *scaffold.ScaffoldContext) {
	box := components.NewSummaryBox("Setup complete!")
	box.AddSuccess(fmt.Sprintf("Scope: %s", ctx.SetupScope))
	box.AddSuccess(fmt.Sprintf("Tools: %v", ctx.Tools))
	box.AddSuccess(fmt.Sprintf("Project: %s", ctx.ProjectName))
	box.SetStats(len(result.Files), 0, 0)

	for _, e := range result.Errors {
		box.AddWarning(fmt.Sprintf("%v", e))
	}

	fmt.Println(box.Render())
}

func compileMCPForInit(ctx *scaffold.ScaffoldContext, result *scaffold.ScaffoldResult) error {
	if ctx == nil || result == nil || ctx.DryRun {
		return nil
	}

	reg := adapter.NewRegistry()
	for _, tool := range ctx.Tools {
		adapt, err := reg.Get(tool)
		if err != nil {
			return fmt.Errorf("compile mcp for %s: %w", tool, err)
		}

		compileCtx := adapter.CompileContext{
			TargetDir:     ctx.TargetDir,
			HomeDir:       ctx.HomeDir,
			SetupScope:    ctx.SetupScope,
			LocalSecrets:  ctx.LocalSecrets,
			WorkspaceRoot: ctx.WorkspaceRoot,
			Repos:         ctx.Repos,
		}
		records, err := adapt.CompileMCP(compileCtx)
		if err != nil {
			return fmt.Errorf("compile mcp for %s: %w", tool, err)
		}
		result.Files = append(result.Files, records...)
	}

	return nil
}

// printInitNextSteps prints a next-step message about filling AGENTS.md placeholders.
func printInitNextSteps(ctx *scaffold.ScaffoldContext) {
	agentsPath := filepath.Join(ctx.TargetDir, "AGENTS.md")
	data, err := os.ReadFile(agentsPath)
	if err != nil {
		return
	}
	remaining := strings.Count(string(data), "\x3c!-- fill-in:")
	if remaining > 0 {
		theme.Successf(os.Stdout, "Next: Run /init or /populate in your AI tool to fill %d remaining placeholder(s).", remaining)
		fmt.Println()
		fmt.Println("Why placeholders remain:")
		fmt.Println("  Headless init cannot run your AI tool, so <!-- fill-in: ... --> markers")
		fmt.Println("  are left in AGENTS.md for the host tool or a human to complete.")
		fmt.Println()
		fmt.Println("What the host tool should do next:")
		fmt.Println("  1. Open this project in your AI tool (Claude Code, OpenCode, etc.).")
		fmt.Println("  2. Run /init or /populate so the AI tool fills the placeholders from code evidence.")
		fmt.Println()
		fmt.Println("To complete setup manually:")
		fmt.Println("  Edit AGENTS.md and replace each <!-- fill-in: ... --> marker with a concrete")
		fmt.Println("  value, or remove the marker if the section does not apply to this project.")
		fmt.Println()
		fmt.Println("Validate after populate:")
		fmt.Println("  lazyai-cli validate agents   # confirm agent frontmatter is still valid")
		fmt.Println("  lazyai-cli doctor            # full setup health check")
	} else {
		theme.Successf(os.Stdout, "All placeholders filled! Your AGENTS.md is ready.")
		fmt.Println()
		fmt.Println("Validate after populate:")
		fmt.Println("  lazyai-cli validate agents   # confirm agent frontmatter is still valid")
		fmt.Println("  lazyai-cli doctor            # full setup health check")
	}

	// Hint for workspace + sidecar flow
	fmt.Println()
	fmt.Println("Workspace & Sidecar:")
	fmt.Println("  lazyai-cli workspace add <path> --name <name>  # Register a project")
	fmt.Println("  lazyai-cli workspace switch <name>               # Set active workspace")
	fmt.Println("  lazyai-cli sidecar init --path <kb-path>         # Attach a sidecar for docs/specs/plans")
	fmt.Println("  lazyai-cli sidecar status                          # Verify resolved paths")
}
