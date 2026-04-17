package scaffold

import (
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	"github.com/ricardoborges-teachable/ai-setup/internal/conflict"
	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// FeatureFlagsForTemplate converts types.FeatureFlags into a map suitable for
// template conditional rendering, including legacy snake_case aliases.
type FeatureFlagsForTemplate struct {
	ContextEngineering bool
	RPIWorkflow        bool
	ChainOfThought     bool
	TreeOfThoughts     bool
	ADREnforcement     bool
	QualityGates       bool
	AgentHarness       bool
	BugResolution      bool
	PivotHandling      bool
	GitConventions     bool

	// Legacy snake_case aliases.
	ContextEngineering_ bool
	RPIWorkflow_        bool
	ChainOfThought_     bool
	TreeOfThoughts_     bool
	ADREnforcement_     bool
	QualityGates_       bool
	AgentHarness_       bool
	BugResolution_      bool
	PivotHandling_      bool
	GitConventions_     bool
}

// FragmentContext holds the template rendering context.
// Ported from the TypeScript FragmentContext used in compiled-root.ts.
type FragmentContext struct {
	ProjectName         string
	PlanningDir         string
	PrimaryLanguage     string
	Framework           string
	WorkspaceType       string
	ProjectInstructions string
	TestFramework       string
	PackageManager      string
	TestCommand         string
	LintCommand         string
	BuildCommand        string
	DevCommand          string
	InstallCommand      string
	ProjectDescription  string
	Features            FeatureFlagsForTemplate
}

// ScaffoldCompiledRoot compiles and writes root AI tool configuration files.
// This is a simplified version that reads a root template from the library and
// performs basic variable substitution. The full template compiler (with fragment
// assembly) will be ported separately.
// Ported from src/scaffold/compiled-root.ts.
func ScaffoldCompiledRoot(opts ScaffoldCompiledRootOptions) error {
	effectiveFeatures := types.DefaultFeatureFlags()
	if opts.Features != nil {
		effectiveFeatures = *opts.Features
	}

	// Build fragment context from options.
	ctx := FragmentContext{
		ProjectName:         opts.ProjectName,
		PlanningDir:         opts.PlanningDir,
		PrimaryLanguage:     opts.PrimaryLanguage,
		Framework:           opts.Framework,
		WorkspaceType:       opts.WorkspaceType,
		ProjectInstructions: opts.ProjectInstructions,
		Features: FeatureFlagsForTemplate{
			ContextEngineering: effectiveFeatures.ContextEngineering,
			RPIWorkflow:        effectiveFeatures.RPIWorkflow,
			ChainOfThought:     effectiveFeatures.ChainOfThought,
			TreeOfThoughts:     effectiveFeatures.TreeOfThoughts,
			ADREnforcement:     effectiveFeatures.ADREnforcement,
			QualityGates:       effectiveFeatures.QualityGates,
			AgentHarness:       effectiveFeatures.AgentHarness,
			BugResolution:      effectiveFeatures.BugResolution,
			PivotHandling:      effectiveFeatures.PivotHandling,
			GitConventions:     opts.GitConventions != nil,
			// Legacy aliases.
			ContextEngineering_: effectiveFeatures.ContextEngineering,
			RPIWorkflow_:        effectiveFeatures.RPIWorkflow,
			ChainOfThought_:     effectiveFeatures.ChainOfThought,
			TreeOfThoughts_:     effectiveFeatures.TreeOfThoughts,
			ADREnforcement_:     effectiveFeatures.ADREnforcement,
			QualityGates_:       effectiveFeatures.QualityGates,
			AgentHarness_:       effectiveFeatures.AgentHarness,
			BugResolution_:      effectiveFeatures.BugResolution,
			PivotHandling_:      effectiveFeatures.PivotHandling,
			GitConventions_:     opts.GitConventions != nil,
		},
	}

	// Build workspace repos section if applicable.
	var workspaceReposSection string
	if len(opts.Repos) > 0 {
		var lines []string
		lines = append(lines, "", "## Workspace Repos", "",
			"This workspace contains the following repositories:", "")
		for _, repo := range opts.Repos {
			lines = append(lines, fmt.Sprintf("### %s", repo.Name), "")
			lines = append(lines, fmt.Sprintf("- **Path**: `%s`", repo.Path))
			if repo.Type != "" && repo.Type != "unknown" {
				lines = append(lines, fmt.Sprintf("- **Type**: %s", repo.Type))
			}
			if repo.Description != "" {
				lines = append(lines, fmt.Sprintf("- **Description**: %s", repo.Description))
			}
			lines = append(lines, "")
		}
		lines = append(lines, "When working in a repo, refer to its README or package.json for repo-specific details.", "")
		workspaceReposSection = strings.Join(lines, "\n")
	}

	// Compile for each tool.
	for _, tool := range opts.Tools {
		outputFile, ok := RootFileByTool[tool]
		if !ok {
			continue
		}

		// Read the tool-specific root template from the library FS.
		templateRelPath := "root/" + outputFile + ".template.md"
		if !files.ExistsFS(opts.LibraryFS, templateRelPath) {
			// Try alternative naming: AGENTS.template.md, CLAUDE.template.md, etc.
			baseName := strings.TrimSuffix(outputFile, filepath.Ext(outputFile))
			templateRelPath = "root/" + baseName + ".template.md"
			if !files.ExistsFS(opts.LibraryFS, templateRelPath) {
				continue
			}
		}

		data, err := files.ReadFS(opts.LibraryFS, templateRelPath)
		if err != nil {
			log.Printf("Warning: could not read root template %s: %v", templateRelPath, err)
			continue
		}

		content := string(data)
		// Perform basic substitutions.
		content = strings.ReplaceAll(content, "[YOUR_PROJECT_NAME]", opts.ProjectName)
		content = strings.ReplaceAll(content, "{{projectName}}", ctx.ProjectName)
		content = strings.ReplaceAll(content, "{{planningDir}}", ctx.PlanningDir)
		if ctx.PrimaryLanguage != "" {
			content = strings.ReplaceAll(content, "{{primaryLanguage}}", ctx.PrimaryLanguage)
		}
		if ctx.Framework != "" {
			content = strings.ReplaceAll(content, "{{framework}}", ctx.Framework)
		}

		// Append workspace repos section if applicable.
		if workspaceReposSection != "" {
			content = content + workspaceReposSection
		}

		destPath := filepath.Join(opts.TargetDir, outputFile)
		if strings.HasPrefix(outputFile, ".github/") {
			if err := files.EnsureDir(filepath.Dir(destPath)); err != nil {
				return err
			}
		}

		action, err := conflict.ApplyStrategy(destPath, opts.Strategy, opts.PerFileOverrides, opts.TargetDir)
		if err != nil {
			return err
		}
		if action == "skip" {
			relPath, _ := filepath.Rel(opts.TargetDir, destPath)
			log.Printf("Skipping existing file: %s", relPath)
			continue
		}

		if err := files.WriteFile(destPath, []byte(content), 0o644); err != nil {
			return err
		}

		hash, _ := files.FileHash(destPath)
		relPath, _ := filepath.Rel(opts.TargetDir, destPath)
		*opts.FileRecords = append(*opts.FileRecords, types.TrackedFile{
			Path:   relPath,
			Hash:   hash,
			Source: "compiled:" + string(tool),
			Owner:  types.FileOwnerLibrary,
		})
	}

	return nil
}

// ScaffoldCompiledRootOptions holds the options for compiling root files.
type ScaffoldCompiledRootOptions struct {
	TargetDir        string
	LibraryFS        fs.FS
	Tools            []types.ToolId
	ProjectName      string
	PlanningDir      string
	Features         *types.FeatureFlags
	GitConventions   *types.GitConventions
	FileRecords      *[]types.TrackedFile
	Strategy         types.ConflictStrategy
	PerFileOverrides map[string]types.ConflictStrategy
	SetupScope       types.SetupScope
	// Optional context overrides.
	PrimaryLanguage     string
	Framework           string
	WorkspaceType       string
	ProjectInstructions string
	// Referenced repos for workspace scope.
	Repos []types.RepoInfo
}