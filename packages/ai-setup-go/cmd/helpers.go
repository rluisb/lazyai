package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ricardoborges-teachable/ai-setup/internal/db"
	"github.com/ricardoborges-teachable/ai-setup/internal/library"
	"github.com/ricardoborges-teachable/ai-setup/internal/preset"
	reversa "github.com/ricardoborges-teachable/ai-setup/internal/reversa/scout"
	"github.com/ricardoborges-teachable/ai-setup/internal/scaffold"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
	"github.com/ricardoborges-teachable/ai-setup/tui/wizard"
)

// getLibraryDir resolves the library directory using the library package.
func getLibraryDir() string {
	dir, err := library.FindLibraryDir()
	if err != nil {
		return ""
	}
	return dir
}

func validateToolFlag(tool string) error {
	if tool == "" || types.IsValidToolId(types.ToolId(tool)) {
		return nil
	}
	return fmt.Errorf("unsupported tool %q (supported tools: opencode, claude-code, copilot)", tool)
}

// openStore opens the SQLite database for the given target directory.
// It also auto-imports from .ai-setup.json if the DB doesn't exist yet.
func openStore(targetDir string) (*db.DB, error) {
	dbPath := filepath.Join(targetDir, ".ai-setup.db")
	database, err := db.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// Run migrations.
	if err := db.RunMigrations(database); err != nil {
		database.Close()
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	// Auto-import from JSON if DB is new.
	imported, err := db.AutoImportJSON(targetDir, database)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  Warning: JSON import failed: %v\n", err)
	}
	if imported {
		fmt.Println("  Imported existing .ai-setup.json → SQLite")
	}

	return database, nil
}

// buildScaffoldContext creates a ScaffoldContext from wizard results and config.
func buildScaffoldContext(result *wizard.WizardResult, config *wizard.WizardConfig) (*scaffold.ScaffoldContext, error) {
	if result == nil || result.Phase1 == nil || result.Phase2 == nil {
		return nil, fmt.Errorf("wizard result is incomplete")
	}

	// Resolve preset features.
	presetLevel := result.Phase2.Preset
	if presetLevel == "" {
		presetLevel = preset.DefaultPresetForScope(result.Phase1.Scope)
	}

	features := types.DefaultFeatureFlags()
	if result.Phase2.Features != nil {
		features = *result.Phase2.Features
	} else {
		resolved := preset.ResolvePreset(presetLevel)
		if resolved != nil {
			features = *resolved
		}
	}

	gitConvs := types.DefaultGitConventions()
	if result.Phase2.GitConv != nil {
		gitConvs = *result.Phase2.GitConv
	}

	// Resolve selections based on preset.
	specsDirs := preset.SpecsDirsForPreset(presetLevel)
	templates := preset.TemplatesForPreset(presetLevel)
	rules := preset.RulesForPreset(presetLevel)

	// Resolve agents/skills/prompts from explicit wizard selections when present;
	// otherwise fall back to preset-driven defaults for compatibility with
	// non-interactive flows that do not yet expose these selectors.
	var agents []types.AgentId
	var skills []types.SkillId
	var prompts []types.PromptId
	var chatmodes []types.ChatModeId
	var opencodeCommands []types.OpenCodeCommandId
	var opencodeModes []types.OpenCodeModeId
	var opencodePlugins []string
	var infra []types.InfraId

	if presetLevel != types.PresetLevelMinimal {
		agents = types.ALL_AGENTS[:]
		skills = types.ALL_SKILLS[:]
		prompts = types.ALL_PROMPTS[:]
		chatmodes = types.ALL_CHATMODES[:]
		opencodeCommands = types.ALL_OPENCODE_COMMANDS[:]
		opencodeModes = types.ALL_OPENCODE_MODES[:]
		infra = types.ALL_INFRA[:]
	}

	// When the user picked the custom preset AND went through the wizard's
	// commands/chatmodes selection, honour their explicit choice instead of
	// the ALL_* defaults.
	if presetLevel == types.PresetLevelCustom {
		if result.Phase2.ChatModes != nil {
			chatmodes = result.Phase2.ChatModes
		}
		if result.Phase2.OpenCodeCommands != nil {
			opencodeCommands = result.Phase2.OpenCodeCommands
		}
		if result.Phase2.OpenCodeModes != nil {
			opencodeModes = result.Phase2.OpenCodeModes
		}
	}

	if result.Phase1 != nil {
		if result.Phase1.Agents != nil {
			agents = append([]types.AgentId(nil), result.Phase1.Agents...)
		}
		if result.Phase1.Skills != nil {
			skills = append([]types.SkillId(nil), result.Phase1.Skills...)
		}
	}

	// Plugin selections come from Phase5 regardless of preset.
	if result.Phase5 != nil {
		opencodePlugins = result.Phase5.OpenCodePlugins
	}

	// Resolve library directory and FS.
	libDir := getLibraryDir()
	libFS := library.GetLibraryFS()

	// Convert ToolId slice to string slice for CLITools.
	cliTools := result.Phase1.CliTools
	if len(cliTools) == 0 && len(config.CLICliTools) > 0 {
		cliTools = config.CLICliTools
	}

	enableServers := result.Phase1.EnableServers
	if len(enableServers) == 0 && len(config.CLIEnableServers) > 0 {
		enableServers = config.CLIEnableServers
	}

	existingSetupPolicy := config.CLIExistingSetupPolicy
	if existingSetupPolicy == "" {
		existingSetupPolicy = types.SetupPolicyAbsorb
		config.CLIExistingSetupPolicy = existingSetupPolicy
	}
	strategy := conflictStrategyForSetupPolicy(existingSetupPolicy)
	if existingSetupPolicy == types.SetupPolicyAdapt {
		warnAdaptFallbackForUnsupportedTools(result.Phase1.Tools)
	}

	ctx := &scaffold.ScaffoldContext{
		TargetDir:           config.TargetDir,
		LibraryDir:          libDir,
		LibraryFS:           libFS,
		Tools:               result.Phase1.Tools,
		CLITools:            cliTools,
		EnableServers:       enableServers,
		ProjectName:         result.Phase1.ProjectName,
		PlanningDir:         "specs",
		SetupScope:          result.Phase1.Scope,
		ExistingSetupPolicy: existingSetupPolicy,
		HomeDir:             config.HomeDir,
		Features:            &features,
		GitConventions:      &gitConvs,
		Strategy:            strategy,
		Force:               config.Force,
		DryRun:              config.DryRun,
		DriveCLI:            config.CLIDriveCLI,
		LocalSecrets:        config.CLILocalSecrets,
		Organization:        firstNonEmpty(result.Phase1.Organization, config.CLIOrg),
		Team:                firstNonEmpty(result.Phase1.Team, config.CLITeam),
		ProjectOverview:     result.Phase2.ProjectOverview,
		NamingConventions:   result.Phase2.NamingConventions,
		ErrorHandling:       result.Phase2.ErrorHandling,
		APIConventions:      result.Phase2.APIConventions,
		ImportOrder:         result.Phase2.ImportOrder,
		ProtectedBranch:     result.Phase2.ProtectedBranch,
		TestCommand:         result.Phase2.TestCommand,
		LintCommand:         result.Phase2.LintCommand,
		BuildCommand:        result.Phase2.BuildCommand,
		CoverageThreshold:   normalizeCoverageThreshold(result.Phase2.CoverageThreshold),
		Agents:              agents,
		Skills:              skills,
		Prompts:             prompts,
		ChatModes:           chatmodes,
		OpenCodeCommands:    opencodeCommands,
		OpenCodeModes:       opencodeModes,
		OpenCodePlugins:     opencodePlugins,
		Templates:           templates,
		Rules:               rules,
		Infra:               infra,
		SpecsDirs:           specsDirs,
		Housekeeping:        housekeepingFromResult(result),
		PlanningRepoPath:    config.HomeDir,
	}

	if config.HomeDir == "" {
		ctx.PlanningRepoPath, _ = os.UserHomeDir()
	}

	// Run deterministic Scout to populate mechanical placeholders.
	if config.TargetDir != "" {
		surface, err := reversa.RunScout(config.TargetDir)
		if err == nil && surface != nil {
			// Fill in what Scout detected.
			if surface.PrimaryLanguage != "" {
				ctx.PrimaryLanguage = surface.PrimaryLanguage
			}
			if len(surface.Frameworks) > 0 {
				ctx.Framework = surface.Frameworks[0].Name
			}
			if surface.PackageManager != "" {
				ctx.PackageManager = surface.PackageManager
			}
			if surface.TestFramework != "" {
				ctx.TestFramework = surface.TestFramework
			}
			if len(surface.DatabaseHints) > 0 {
				db := reversa.InferDatabase(surface.DatabaseHints)
				if db != "" {
					ctx.Database = db
				}
				orm := reversa.InferORM(surface.DatabaseHints)
				if orm != "" {
					ctx.ORM = orm
				}
				mPath := reversa.InferMigrationsPath(surface.DatabaseHints)
				if mPath != "" {
					ctx.MigrationsPath = mPath
				}
			}
			if len(surface.Modules) > 0 || len(surface.EntryPoints) > 0 {
				ctx.CodebaseMap = reversa.BuildCodebaseMapEntries(surface.Modules, surface.EntryPoints)
			}
			// Infer commands from language + package manager.
			if ctx.InstallCommand == "" {
				ic := reversa.InferInstallCommandFromSurface(surface)
				if ic != "" {
					ctx.InstallCommand = ic
				}
			}
			if ctx.LintCommand == "" {
				lc := reversa.InferLintCommandFromSurface(surface)
				if lc != "" {
					ctx.LintCommand = lc
				}
			}
			if ctx.TestCommand == "" {
				// Derive from Scout only if the wizard didn't set it.
				tc := reversa.InferTestCommandFromSurface(surface)
				if tc != "" {
					ctx.TestCommand = tc
				}
			}
			if ctx.TestPath == "" {
				tp := reversa.InferTestPathFromSurface(surface)
				if tp != "" {
					ctx.TestPath = tp
				}
			}
			if ctx.StrictMode == "" {
				sm := reversa.InferStrictMode(config.TargetDir)
				if sm != "" {
					ctx.StrictMode = sm
				}
			}
			if ctx.ProtectedBranch == "" {
				branch := reversa.InferProtectedBranch(config.TargetDir)
				if branch != "" {
					ctx.ProtectedBranchGit = branch
					ctx.ProtectedBranch = branch
				}
			}
			// Store surface data for downstream use.
			ctx.SurfaceData = surface
		}
	}

	return ctx, nil
}

func conflictStrategyForSetupPolicy(policy types.SetupPolicy) types.ConflictStrategy {
	switch policy {
	case types.SetupPolicyBackupOnly:
		return types.ConflictStrategyAlign
	case types.SetupPolicyAbsorb, types.SetupPolicyAdapt:
		return types.ConflictStrategySkip
	default:
		return types.ConflictStrategySkip
	}
}

func warnAdaptFallbackForUnsupportedTools(tools []types.ToolId) {
	for _, tool := range tools {
		if tool == types.ToolIdOpenCode {
			continue
		}
		fmt.Fprintf(os.Stderr, "WARN: --existing-setup-policy=adapt is MVP-supported for OpenCode only; preserving existing %s files via absorb behavior\n", tool)
	}
}

// writeStoreFromScaffoldResult writes the scaffold results to the SQLite database.
func writeStoreFromScaffoldResult(database *db.DB, ctx *scaffold.ScaffoldContext, presetLevel types.PresetLevel, result *scaffold.ScaffoldResult) error {
	store := db.NewStore(database)

	// Build store data from scaffold context.
	storeData := types.DefaultStoreData()
	storeData.Meta.CLIVersion = Version
	storeData.Config.SetupScope = ctx.SetupScope
	storeData.Config.Tools = ctx.Tools
	storeData.Config.CLITools = ctx.CLITools
	storeData.Config.EnableServers = ctx.EnableServers
	storeData.Config.ProjectName = ctx.ProjectName
	storeData.Config.TargetDir = ctx.TargetDir
	storeData.Config.PlanningDir = ctx.PlanningDir
	storeData.Config.Repos = ctx.Repos
	storeData.Config.Housekeeping = ctx.Housekeeping
	storeData.Config.ProjectOverview = ctx.ProjectOverview
	storeData.Config.NamingConventions = ctx.NamingConventions
	storeData.Config.ErrorHandling = ctx.ErrorHandling
	storeData.Config.ApiConventions = ctx.APIConventions
	storeData.Config.ImportOrder = ctx.ImportOrder
	storeData.Config.ProtectedBranch = ctx.ProtectedBranch
	storeData.Config.TestCommand = ctx.TestCommand
	storeData.Config.LintCommand = ctx.LintCommand
	storeData.Config.BuildCommand = ctx.BuildCommand
	storeData.Config.CoverageThreshold = normalizeCoverageThreshold(ctx.CoverageThreshold)

	// Selections.
	storeData.Selections.Templates = ctx.Templates
	storeData.Selections.Rules = ctx.Rules
	storeData.Selections.Agents = ctx.Agents
	storeData.Selections.Skills = ctx.Skills
	storeData.Selections.Prompts = ctx.Prompts
	storeData.Selections.ChatModes = ctx.ChatModes
	storeData.Selections.OpenCodeCommands = ctx.OpenCodeCommands
	storeData.Selections.OpenCodeModes = ctx.OpenCodeModes
	storeData.Selections.OpenCodePlugins = ctx.OpenCodePlugins
	storeData.Selections.Infra = ctx.Infra
	storeData.Selections.Features = ctx.Features
	storeData.Selections.GitConventions = ctx.GitConventions

	// Tracked files from scaffold result.
	if result != nil {
		storeData.Files = result.Files
	}

	// Write to database.
	return store.WriteStoreData(&storeData)
}

func housekeepingFromResult(result *wizard.WizardResult) *types.HousekeepingConfig {
	if result == nil || result.Phase5 == nil {
		return nil
	}

	return &types.HousekeepingConfig{
		MemoryPath:        result.Phase5.MemoryPath,
		EnableObsidian:    result.Phase5.EnableObsidian,
		ObsidianVaultPath: result.Phase5.ObsidianVaultPath,
		EnableQmd:         result.Phase5.EnableQmd,
		QmdIndexPath:      result.Phase5.QmdIndexPath,
		EnableCodegraph:   result.Phase5.EnableCodegraph,
		CodegraphDataPath: result.Phase5.CodegraphDataPath,
	}
}

// projectNameFromDir returns the directory name as a project name fallback.
func projectNameFromDir(dir string) string {
	base := filepath.Base(dir)
	if base == "" || base == "." || base == "/" {
		return "my-project"
	}
	return strings.ReplaceAll(strings.ReplaceAll(base, " ", "-"), "_", "-")
}

// firstNonEmpty returns the first non-empty string from values.
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func normalizeCoverageThreshold(value int) int {
	if value >= 1 && value <= 100 {
		return value
	}
	return 80
}
