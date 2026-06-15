package scaffold

import (
	"fmt"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/adapter"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// ScaffoldAll runs the full scaffold pipeline in order, calling each scaffold
// function with the provided context. It accumulates all file records and
// collects any errors without stopping early.
// Ported from the TypeScript setup pipeline that orchestrates all scaffold modules.
func ScaffoldAll(ctx *ScaffoldContext) (*ScaffoldResult, error) {
	result := &ScaffoldResult{
		Files:       []types.TrackedFile{},
		Directories: []string{},
		Errors:      []error{},
	}
	planningPath := planningRoot(ctx)
	toolRoot := workspaceToolRoot(ctx)

	fileRecords := &result.Files
	libFS := ctx.LibraryFS
	if libFS == nil {
		return nil, fmt.Errorf("scaffold context has no LibraryFS — call library.GetLibraryFS() to set it")
	}

	// Step 1: Constitution files.
	if planningPath != "" {
		if err := ScaffoldConstitution(planningPath, libFS, ctx.ProjectName, fileRecords, ctx.Strategy, ctx.PerFileOverrides); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("constitution: %w", err))
		}
	}

	// Step 2: MCP configuration.
	if err := ScaffoldMcp(toolRoot, ctx.LibraryDir, libFS, ctx.CLITools, ctx.EnableServers, fileRecords, ctx.Strategy, ctx.PerFileOverrides); err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("mcp: %w", err))
	}

	// Step 3: Specs directory structure.
	if planningPath != "" {
		if err := ScaffoldSpecs(planningPath, ctx.SetupScope, libFS, ctx.SpecsDirs, fileRecords, ctx.Strategy, ctx.PerFileOverrides); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("specs: %w", err))
		}
	}

	// Housekeeping files.
	if planningPath != "" {
		if err := ScaffoldHousekeeping(planningPath, ctx.Housekeeping); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("housekeeping: %w", err))
		}
	}

	// Step 4: Templates and rules.
	if planningPath != "" {
		if err := ScaffoldTemplatesRules(planningPath, libFS, ctx.Templates, ctx.Rules, fileRecords, ctx.Strategy, ctx.PerFileOverrides, ctx.CoverageThreshold); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("templates-rules: %w", err))
		}
	}

	infraPath := ctx.TargetDir
	if planningPath != "" {
		infraPath = planningPath
	}
	if err := ScaffoldInfra(infraPath, ctx.SetupScope, libFS, ctx.ProjectName, ctx.Infra, fileRecords, ctx.Strategy, ctx.PerFileOverrides); err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("infra: %w", err))
	}

	// Step 7: Compiled root files.
	if err := ScaffoldCompiledRoot(ScaffoldCompiledRootOptions{
		TargetDir:           ctx.TargetDir,
		WorkspaceRoot:       ctx.WorkspaceRoot,
		HomeDir:             ctx.HomeDir,
		LibraryFS:           libFS,
		Tools:               ctx.Tools,
		ProjectName:         ctx.ProjectName,
		PlanningDir:         ctx.PlanningDir,
		Features:            ctx.Features,
		GitConventions:      ctx.GitConventions,
		FileRecords:         fileRecords,
		Strategy:            ctx.Strategy,
		PerFileOverrides:    ctx.PerFileOverrides,
		SetupScope:          ctx.SetupScope,
		StoreData:           ctx.StoreData,
		PrimaryLanguage:     ctx.PrimaryLanguage,
		Framework:           ctx.Framework,
		WorkspaceType:       ctx.WorkspaceType,
		ProjectInstructions: ctx.ProjectInstructions,
		ProjectDescription:  ctx.ProjectDescription,
		Organization:        ctx.Organization,
		Team:                ctx.Team,
		ProjectOverview:     ctx.ProjectOverview,
		Database:            ctx.Database,
		ORM:                 ctx.ORM,
		TestFramework:       ctx.TestFramework,
		PackageManager:      ctx.PackageManager,
		MigrationsPath:      ctx.MigrationsPath,
		TestPath:            ctx.TestPath,
		StrictMode:          ctx.StrictMode,
		InstallCommand:      ctx.InstallCommand,
		ProtectedBranchGit:  ctx.ProtectedBranchGit,
		NamingConventions:   ctx.NamingConventions,
		ErrorHandling:       ctx.ErrorHandling,
		APIConventions:      ctx.APIConventions,
		ImportOrder:         ctx.ImportOrder,
		ProtectedBranch:     ctx.ProtectedBranch,
		TestCommand:         ctx.TestCommand,
		LintCommand:         ctx.LintCommand,
		BuildCommand:        ctx.BuildCommand,
		CoverageThreshold:   ctx.CoverageThreshold,
		CodebaseMap:         ctx.CodebaseMap,
		Repos:               ctx.Repos,
	}); err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("compiled-root: %w", err))
	}

	// Step 8: Agents, skills, and prompts via adapter registry.
	artifactRecords, err := ctx.ScaffoldArtifacts()
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("artifacts: %w", err))
	}
	result.Files = append(result.Files, artifactRecords...)

	// Step 8b: Post-install validation for OpenCode (no-op when binary absent).
	// Skipped under go test to avoid hanging on external CLI probes.
	if !testing.Testing() {
		for _, tool := range ctx.Tools {
			if tool == types.ToolIdOpenCode {
				adapterCtx := &adapter.AdapterContext{
					TargetDir:     ctx.TargetDir,
					HomeDir:       ctx.HomeDir,
					SetupScope:    ctx.SetupScope,
					WorkspaceRoot: ctx.WorkspaceRoot,
					LibraryFS:     ctx.LibraryFS,
				}
				warnings, _ := adapter.ValidateOpenCodeInstall(adapterCtx)
				for _, w := range warnings {
					scaffoldLog.Warn("OpenCode install validation warning", "warning", w)
				}
				break
			}
		}
	}

	// Step 9: .env.example (depends on MCP config).
	if err := ScaffoldEnvExample(ctx.TargetDir, fileRecords, ctx.Strategy, ctx.PerFileOverrides); err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("env-example: %w", err))
	}

	// Step 10: Workspace-specific: repo roots and ledgers.
	if ctx.SetupScope == types.SetupScopeWorkspace && len(ctx.Repos) > 0 {
		repoResults := ScaffoldRepoRoots(ctx.Repos, planningPath, ctx.Tools, ctx.Strategy, ctx.PerFileOverrides)
		for _, records := range repoResults {
			result.Files = append(result.Files, records...)
		}

		if err := ScaffoldRepoLedgers(planningPath, ctx.Repos, fileRecords, ctx.Strategy, ctx.PerFileOverrides); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("repo-ledgers: %w", err))
		}
	}

	// Step 11: .gitignore guidance.
	CheckGitignoreGuidance(ctx.TargetDir, ctx.LocalSecrets)

	scaffoldLog.Info("scaffold complete", "files", len(result.Files), "errors", len(result.Errors))

	if len(result.Errors) > 0 {
		return result, fmt.Errorf("scaffold completed with %d errors", len(result.Errors))
	}

	return result, nil
}

func workspaceToolRoot(ctx *ScaffoldContext) string {
	if ctx == nil {
		return ""
	}
	if ctx.SetupScope == types.SetupScopeWorkspace && ctx.WorkspaceRoot != "" {
		return ctx.WorkspaceRoot
	}
	return ctx.TargetDir
}

func planningRoot(ctx *ScaffoldContext) string {
	if ctx == nil || ctx.SetupScope == types.SetupScopeGlobal {
		return ""
	}
	if ctx.PlanningRepoPath != "" {
		return ctx.PlanningRepoPath
	}
	return ctx.TargetDir
}
