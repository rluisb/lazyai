package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"

	"github.com/rluisb/lazyai/packages/cli/internal/adapter"
	"github.com/rluisb/lazyai/packages/cli/internal/preset"
	"github.com/rluisb/lazyai/packages/cli/internal/scaffold"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
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
	initCmd.Flags().StringSlice("tools", []string{}, "Tools to configure (opencode, claude-code, copilot)")
	initCmd.Flags().StringSlice("enable-servers", []string{}, "MCP servers to enable (orchestrator, filesystem, memory)")
	initCmd.Flags().String("preset", "", "Preset configuration name (minimal, standard, full, custom)")
	initCmd.Flags().StringSlice("features", []string{}, "Features to enable")
	initCmd.Flags().StringSlice("disable-features", []string{}, "Features to disable")
	initCmd.Flags().String("name", "", "Project name")
	initCmd.Flags().String("branch-pattern", "", "Git branch naming pattern")
	initCmd.Flags().String("commit-pattern", "", "Git commit message pattern")
	initCmd.Flags().String("existing-setup-policy", string(types.SetupPolicyAbsorb), "How to handle existing setup (absorb, adapt, backup-only)")
	initCmd.Flags().Bool("non-interactive", false, "Run without interactive prompts")
	initCmd.Flags().Bool("drive-cli", false, "Delegate scaffolding to the tool's own CLI when available (Claude Code)")
	initCmd.Flags().Bool("local-secrets", false, "Route Claude Code MCP/settings writes to gitignored .claude/settings.local.json instead of committed surfaces")
	initCmd.Flags().String("org", "", "Organization name (populates [YOUR_ORG] in AGENTS.md)")
	initCmd.Flags().String("team", "", "Team name (populates [YOUR_TEAM] in AGENTS.md)")
	initCmd.Flags().String("project-overview", "", "Project overview for generated AGENTS.md")
	initCmd.Flags().String("naming-conventions", "", "Naming conventions for generated AGENTS.md")
	initCmd.Flags().String("error-handling", "", "Error handling conventions for generated AGENTS.md")
	initCmd.Flags().String("api-conventions", "", "API response conventions for generated AGENTS.md")
	initCmd.Flags().String("import-order", "", "Import ordering conventions for generated AGENTS.md")
	initCmd.Flags().String("protected-branch", "", "Protected branch name for generated AGENTS.md")
	initCmd.Flags().String("test-command", "", "Test command for generated AGENTS.md")
	initCmd.Flags().String("lint-command", "", "Lint command for generated AGENTS.md")
	initCmd.Flags().String("build-command", "", "Build command for generated AGENTS.md")
	initCmd.Flags().Int("coverage-threshold", 0, "Coverage threshold percentage (1-100; default 80)")
	initCmd.Flags().Bool("force", false, "Overwrite existing files")
	initCmd.Flags().Bool("dry-run", false, "Show what would be done without making changes")
	initCmd.Flags().String("memory-path", "", "Project memory path (default: specs/memory)")
	initCmd.Flags().Bool("enable-obsidian", false, "Enable Obsidian integration")
	initCmd.Flags().String("obsidian-vault-path", "", "Obsidian vault path")
	initCmd.Flags().Bool("enable-qmd", false, "Enable qmd markdown retrieval")
	initCmd.Flags().String("qmd-index-path", "", "Project-local qmd index path")
	initCmd.Flags().Bool("enable-codegraph", false, "Enable codegraph analysis")
	initCmd.Flags().String("codegraph-data-path", "", "Project-local codegraph data path")
	initCmd.Flags().Bool("enable-graphify", false, "Enable graphify knowledge graph analysis")
	initCmd.Flags().String("graphify-data-path", "", "Project-local graphify data path")
	initCmd.Flags().Bool("reversa", false, "Analyze existing code with Scout/Reversa to auto-populate project details")
	initCmd.Flags().Bool("no-reversa", false, "Skip Scout/Reversa analysis and leave project details explicit/manual")
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
	memoryPath, _ := cmd.Flags().GetString("memory-path")
	enableObsidian, _ := cmd.Flags().GetBool("enable-obsidian")
	obsidianVaultPath, _ := cmd.Flags().GetString("obsidian-vault-path")
	enableQmd, _ := cmd.Flags().GetBool("enable-qmd")
	qmdIndexPath, _ := cmd.Flags().GetString("qmd-index-path")
	enableCodegraph, _ := cmd.Flags().GetBool("enable-codegraph")
	codegraphDataPath, _ := cmd.Flags().GetString("codegraph-data-path")
	enableGraphify, _ := cmd.Flags().GetBool("enable-graphify")
	graphifyDataPath, _ := cmd.Flags().GetString("graphify-data-path")
	nonInteractive, _ := cmd.Flags().GetBool("non-interactive")
	force, _ := cmd.Flags().GetBool("force")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	driveCLI, _ := cmd.Flags().GetBool("drive-cli")
	localSecrets, _ := cmd.Flags().GetBool("local-secrets")
	orgName, _ := cmd.Flags().GetString("org")
	teamName, _ := cmd.Flags().GetString("team")
	projectOverview, _ := cmd.Flags().GetString("project-overview")
	namingConventions, _ := cmd.Flags().GetString("naming-conventions")
	errorHandling, _ := cmd.Flags().GetString("error-handling")
	apiConventions, _ := cmd.Flags().GetString("api-conventions")
	importOrder, _ := cmd.Flags().GetString("import-order")
	protectedBranch, _ := cmd.Flags().GetString("protected-branch")
	testCommand, _ := cmd.Flags().GetString("test-command")
	lintCommand, _ := cmd.Flags().GetString("lint-command")
	buildCommand, _ := cmd.Flags().GetString("build-command")
	coverageThreshold, _ := cmd.Flags().GetInt("coverage-threshold")
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
		CLIProjectOverview:     projectOverview,
		CLINamingConventions:   namingConventions,
		CLIErrorHandling:       errorHandling,
		CLIApiConventions:      apiConventions,
		CLIImportOrder:         importOrder,
		CLIProtectedBranch:     protectedBranch,
		CLITestCommand:         testCommand,
		CLILintCommand:         lintCommand,
		CLIBuildCommand:        buildCommand,
		CLICoverageThreshold:   coverageThreshold,
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
		CLIMemoryPath:          memoryPath,
		CLIEnableObsidian:      enableObsidian,
		CLIObsidianVaultPath:   obsidianVaultPath,
		CLIEnableQmd:           enableQmd,
		CLIQmdIndexPath:        qmdIndexPath,
		CLIEnableCodegraph:     enableCodegraph,
		CLICodegraphDataPath:   codegraphDataPath,
		CLIEnableGraphify:      enableGraphify,
		CLIGraphifyDataPath:    graphifyDataPath,
		CLIExistingSetupPolicy: existingSetupPolicy,
		CLIUseReversa:          useReversa,
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
	return nil
}

func runInitNonInteractive(config *wizard.WizardConfig) error {
	// Non-interactive mode: use CLI flags directly.
	// Validate required flags.
	if config.CLIScope == "" {
		return fmt.Errorf("--scope is required in non-interactive mode (global | workspace | project)")
	}
	if len(config.CLITools) == 0 {
		return fmt.Errorf("--tools is required in non-interactive mode (opencode, claude-code, copilot)")
	}

	// Drop tools that don't support the chosen scope (e.g. copilot × global).
	// Warn per drop; abort only if nothing is left.
	kept := make([]types.ToolId, 0, len(config.CLITools))
	for _, t := range config.CLITools {
		if adapter.IsScopeSupported(t, config.CLIScope) {
			kept = append(kept, t)
			continue
		}
		fmt.Fprintf(os.Stderr, "WARN: skipping tool %q for scope %q — not supported\n", t, config.CLIScope)
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
		Preset:            presetLevel,
		Features:          &features,
		GitConv:           &gitConvs,
		ProjectOverview:   config.CLIProjectOverview,
		NamingConventions: config.CLINamingConventions,
		ErrorHandling:     config.CLIErrorHandling,
		APIConventions:    config.CLIApiConventions,
		ImportOrder:       config.CLIImportOrder,
		ProtectedBranch:   config.CLIProtectedBranch,
		TestCommand:       config.CLITestCommand,
		LintCommand:       config.CLILintCommand,
		BuildCommand:      config.CLIBuildCommand,
		CoverageThreshold: config.CLICoverageThreshold,
		UseReversa:        config.CLIUseReversa,
	}

	// Build WizardResult with pre-computed phases.
	wizardDefaults := &wizard.WizardResult{
		Phase1: phase1,
		Phase2: phase2,
		Phase5: &wizard.Phase5Result{
			MemoryPath:        config.CLIMemoryPath,
			EnableObsidian:    config.CLIEnableObsidian,
			ObsidianVaultPath: config.CLIObsidianVaultPath,
			EnableQmd:         config.CLIEnableQmd,
			QmdIndexPath:      config.CLIQmdIndexPath,
			EnableCodegraph:   config.CLIEnableCodegraph,
			CodegraphDataPath: config.CLICodegraphDataPath,
			EnableGraphify:    config.CLIEnableGraphify,
			GraphifyDataPath:  config.CLIGraphifyDataPath,
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
	return nil
}

// printScaffoldSummary prints a human-readable summary of the scaffold results.
func printScaffoldSummary(result *scaffold.ScaffoldResult, ctx *scaffold.ScaffoldContext) {
	style := headerStyle()
	greenStyle := successStyle()

	fmt.Println(style.Render("✅ Setup complete!"))
	fmt.Println()
	fmt.Printf("  %s Scope: %s\n", greenStyle.Render("•"), ctx.SetupScope)
	fmt.Printf("  %s Tools: %v\n", greenStyle.Render("•"), ctx.Tools)
	fmt.Printf("  %s Project: %s\n", greenStyle.Render("•"), ctx.ProjectName)
	fmt.Printf("  %s Files installed: %d\n", greenStyle.Render("•"), len(result.Files))

	if len(result.Errors) > 0 {
		warnStyle := warningStyle()
		fmt.Println()
		fmt.Println(warnStyle.Render(fmt.Sprintf("⚠ %d warnings:", len(result.Errors))))
		for _, e := range result.Errors {
			fmt.Printf("  • %v\n", e)
		}
	}
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
			WorkspaceRoot: ctx.TargetDir,
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

// headerStyle returns a bold styled header with the primary color.
func headerStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
}

// successStyle returns a green colored style for success indicators.
func successStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))
}

// warningStyle returns an orange colored style for warnings.
func warningStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))
}
