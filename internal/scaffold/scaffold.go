package scaffold

import (
	"fmt"
	"log"

	"github.com/ricardoborges-teachable/ai-setup/internal/adapter"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
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

	fileRecords := &result.Files
	libFS := ctx.LibraryFS
	if libFS == nil {
		return nil, fmt.Errorf("scaffold context has no LibraryFS — call library.GetLibraryFS() to set it")
	}

	// Step 1: Constitution files.
	if err := ScaffoldConstitution(ctx.TargetDir, libFS, ctx.ProjectName, fileRecords, ctx.Strategy, ctx.PerFileOverrides); err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("constitution: %w", err))
	}

	// Step 2: MCP configuration.
	if err := ScaffoldMcp(ctx.TargetDir, ctx.LibraryDir, libFS, ctx.CLITools, ctx.EnableServers, fileRecords, ctx.Strategy, ctx.PerFileOverrides); err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("mcp: %w", err))
	}

	// Step 3: Orchestration definitions.
	if err := ScaffoldOrchestration(ctx.TargetDir, libFS, fileRecords, ctx.Strategy, ctx.PerFileOverrides); err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("orchestration: %w", err))
	}

	// Step 4: Specs directory structure.
	if err := ScaffoldSpecs(ctx.TargetDir, ctx.SetupScope, libFS, ctx.SpecsDirs, fileRecords, ctx.Strategy, ctx.PerFileOverrides); err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("specs: %w", err))
	}

	if err := ScaffoldHousekeeping(ctx.TargetDir, ctx.Housekeeping); err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("housekeeping: %w", err))
	}

	// Step 5: Templates and rules.
	if err := ScaffoldTemplatesRules(ctx.TargetDir, libFS, ctx.Templates, ctx.Rules, fileRecords, ctx.Strategy, ctx.PerFileOverrides); err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("templates-rules: %w", err))
	}

	// Step 6: Infrastructure files.
	if err := ScaffoldInfra(ctx.TargetDir, libFS, ctx.ProjectName, ctx.Infra, fileRecords, ctx.Strategy, ctx.PerFileOverrides); err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("infra: %w", err))
	}

	// Step 7: Compiled root files.
	if err := ScaffoldCompiledRoot(ScaffoldCompiledRootOptions{
		TargetDir:           ctx.TargetDir,
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
		PrimaryLanguage:     ctx.PrimaryLanguage,
		Framework:           ctx.Framework,
		WorkspaceType:       ctx.WorkspaceType,
		ProjectInstructions: ctx.ProjectInstructions,
		ProjectDescription:  ctx.ProjectDescription,
		Organization:        ctx.Organization,
		Team:                ctx.Team,
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
	for _, tool := range ctx.Tools {
		if tool == types.ToolIdOpenCode {
			adapterCtx := &adapter.AdapterContext{
				TargetDir:  ctx.TargetDir,
				HomeDir:    ctx.HomeDir,
				SetupScope: ctx.SetupScope,
				LibraryFS:  ctx.LibraryFS,
			}
			warnings, _ := adapter.ValidateOpenCodeInstall(adapterCtx)
			for _, w := range warnings {
				log.Printf("WARN %s", w)
			}
			break
		}
	}

	// Step 9: .env.example (depends on MCP config).
	if err := ScaffoldEnvExample(ctx.TargetDir, fileRecords, ctx.Strategy, ctx.PerFileOverrides); err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("env-example: %w", err))
	}

	// Step 10: Workspace-specific: repo roots and ledgers.
	if ctx.SetupScope == types.SetupScopeWorkspace && len(ctx.Repos) > 0 {
		planningPath := ctx.PlanningRepoPath
		if planningPath == "" {
			planningPath = ctx.TargetDir
		}

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

	log.Printf("Scaffold complete: %d files, %d errors", len(result.Files), len(result.Errors))

	if len(result.Errors) > 0 {
		return result, fmt.Errorf("scaffold completed with %d errors", len(result.Errors))
	}

	return result, nil
}
