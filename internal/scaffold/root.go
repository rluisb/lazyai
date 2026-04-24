package scaffold

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	"github.com/ricardoborges-teachable/ai-setup/internal/conflict"
	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/globalpaths"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// errMemoryDocScopeUnsupported is returned by memoryDocDestPath for Copilot × global.
var errMemoryDocScopeUnsupported = errors.New("memory doc not supported at this scope")

// memoryDocDestPath returns the absolute path where the tool's memory doc
// (AGENTS.md / CLAUDE.md / GEMINI.md / .github/copilot-instructions.md) should
// land for the given scope. Returns errMemoryDocScopeUnsupported when the
// combination is not supported (e.g. Copilot × global).
//
// Placement rules:
//   - project / workspace: <targetDir>/<outputFile> (Copilot's .github/ prefix
//     preserved).
//   - global: under the tool's global root, using the bare basename of
//     outputFile — Copilot × global is unsupported.
func memoryDocDestPath(tool types.ToolId, scope types.SetupScope, targetDir, homeDir, outputFile string) (string, error) {
	switch scope {
	case types.SetupScopeProject, types.SetupScopeWorkspace, "":
		return filepath.Join(targetDir, outputFile), nil
	case types.SetupScopeGlobal:
		root, err := globalpaths.ResolveGlobalToolTargetDir(tool, homeDir)
		if err != nil {
			return "", err
		}
		if root == "" {
			return "", errMemoryDocScopeUnsupported
		}
		return filepath.Join(root, filepath.Base(outputFile)), nil
	}
	return "", fmt.Errorf("%w: unknown scope %q", errMemoryDocScopeUnsupported, scope)
}

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
	if opts.HomeDir == "" && opts.SetupScope == types.SetupScopeGlobal {
		return fmt.Errorf("ScaffoldCompiledRoot: HomeDir must be set when scope is global")
	}

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
		// Perform hybrid [YOUR_*] placeholder fill (spec 010 wave C):
		// mechanical fields get concrete values; subjective fields get
		// HTML-comment <!-- fill-in --> markers.
		content = fillClaudeMdPlaceholders(content, opts)
		// Templating handlebars-style substitutions.
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

		homeDir := opts.HomeDir
		destPath, err := memoryDocDestPath(tool, opts.SetupScope, opts.TargetDir, homeDir, outputFile)
		if err != nil {
			if errors.Is(err, errMemoryDocScopeUnsupported) {
				log.Printf("Skipping memory doc for %s at scope %q: not supported", tool, opts.SetupScope)
				continue
			}
			return err
		}
		if err := files.EnsureDir(filepath.Dir(destPath)); err != nil {
			return err
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
		if relPath == "" || strings.HasPrefix(relPath, "..") {
			// Global-scope destinations live outside TargetDir — record the
			// absolute path so downstream tooling can find the file.
			relPath = destPath
		}
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
	HomeDir          string
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
	ProjectDescription  string // optional; substituted into [YOUR_PROJECT_DESCRIPTION]
	Organization        string // optional; substituted into [YOUR_ORG]
	Team                string // optional; substituted into [YOUR_TEAM]
	// Referenced repos for workspace scope.
	Repos []types.RepoInfo
}

// fillClaudeMdPlaceholders replaces [YOUR_*] placeholders in a CLAUDE.md /
// AGENTS.md / GEMINI.md template with values derived from opts, or with
// <!-- fill-in: hint --> markers when the value is unknown or subjective.
// Applying this centrally means each tool's root file gets consistent
// substitutions.
func fillClaudeMdPlaceholders(content string, opts ScaffoldCompiledRootOptions) string {
	// Project name — always known; already substituted upstream but kept here
	// for robustness when this helper runs on its own.
	if opts.ProjectName != "" {
		content = strings.ReplaceAll(content, "[YOUR_PROJECT_NAME]", opts.ProjectName)
	}

	// Project description — derived from opts or a gentle default.
	desc := opts.ProjectDescription
	if desc == "" {
		desc = "AI-assisted development project"
	}
	content = strings.ReplaceAll(content, "[YOUR_PROJECT_DESCRIPTION]", desc)

	// Tech stack — derived from primary language + framework, else fill-in.
	techStack := "<!-- fill-in: tech stack -->"
	if opts.PrimaryLanguage != "" {
		techStack = opts.PrimaryLanguage
		if opts.Framework != "" {
			techStack += " · " + opts.Framework
		}
	}
	content = strings.ReplaceAll(content, "[YOUR_TECH_STACK]", techStack)

	// Organization and Team — from wizard/flags, else fill-in.
	content = strings.ReplaceAll(content, "[YOUR_ORG]",
		fallbackMarker(opts.Organization, "your org"))
	content = strings.ReplaceAll(content, "[YOUR_TEAM]",
		fallbackMarker(opts.Team, "your team"))

	// Dev/test/build commands — inferred by primary language.
	dev, test, build := commandsForLanguage(opts.PrimaryLanguage)
	content = strings.ReplaceAll(content, "[YOUR_DEV_COMMAND]", dev)
	content = strings.ReplaceAll(content, "[YOUR_DEV_DESCRIPTION]", "Run the app from source")
	content = strings.ReplaceAll(content, "[YOUR_TEST_COMMAND]", test)
	content = strings.ReplaceAll(content, "[YOUR_TEST_DESCRIPTION]", "Run the test suite")
	content = strings.ReplaceAll(content, "[YOUR_BUILD_COMMAND]", build)
	content = strings.ReplaceAll(content, "[YOUR_BUILD_DESCRIPTION]", "Build the project")

	// Subjective fields — always fill-in markers.
	subjectiveMarkers := map[string]string{
		"[YOUR_ARCHITECTURE_NOTES]":       "<!-- fill-in: architecture and key patterns -->",
		"[YOUR_CODE_STYLE]":               "<!-- fill-in: code style -->",
		"[YOUR_NAMING_CONVENTIONS]":       "<!-- fill-in: naming conventions -->",
		"[YOUR_TESTING_STRATEGY]":         "<!-- fill-in: testing strategy -->",
		"[YOUR_GIT_WORKFLOW]":             "<!-- fill-in: git workflow -->",
		"[YOUR_RULE_1]":                   "<!-- fill-in: rule 1 -->",
		"[YOUR_RULE_2]":                   "<!-- fill-in: rule 2 -->",
		"[YOUR_DO_NOT_1]":                 "<!-- fill-in: project-specific don't -->",
		"[YOUR_DO_NOT_2]":                 "<!-- fill-in: project-specific don't -->",
		"[YOUR_UNIT_TESTING_STRATEGY]":    "<!-- fill-in: unit testing strategy -->",
		"[YOUR_INTEGRATION_TESTING_STRATEGY]": "<!-- fill-in: integration testing strategy -->",
		"[YOUR_E2E_TESTING_STRATEGY]":     "<!-- fill-in: e2e testing strategy -->",
		"[YOUR_SESSION_CHECK]":            "<!-- fill-in: team-specific session check -->",
		"[YOUR_COMPONENT_1]":              "<!-- fill-in: component -->",
		"[YOUR_COMPONENT_2]":              "<!-- fill-in: component -->",
		"[YOUR_RESPONSIBILITY_1]":         "<!-- fill-in: responsibility -->",
		"[YOUR_RESPONSIBILITY_2]":         "<!-- fill-in: responsibility -->",
		"[YOUR_PATH_1]":                   "<!-- fill-in: path -->",
		"[YOUR_PATH_2]":                   "<!-- fill-in: path -->",
	}
	for placeholder, marker := range subjectiveMarkers {
		content = strings.ReplaceAll(content, placeholder, marker)
	}

	return content
}

// fallbackMarker returns value if non-empty, else an HTML-comment fill-in hint.
func fallbackMarker(value, hint string) string {
	if value != "" {
		return value
	}
	return "<!-- fill-in: " + hint + " -->"
}

// commandsForLanguage returns canonical (dev, test, build) commands per
// primary language. Returns fill-in markers when the language is unknown.
func commandsForLanguage(lang string) (dev, test, build string) {
	switch strings.ToLower(lang) {
	case "go":
		return "go run .", "go test ./...", "go build ./..."
	case "node", "nodejs", "node.js", "javascript", "typescript":
		return "npm run dev", "npm test", "npm run build"
	case "python":
		return "python -m app", "pytest", "python -m build"
	case "rust":
		return "cargo run", "cargo test", "cargo build --release"
	case "ruby":
		return "bundle exec rails s", "bundle exec rspec", "bundle exec rake build"
	case "java":
		return "./gradlew bootRun", "./gradlew test", "./gradlew build"
	}
	return "<!-- fill-in: dev command -->",
		"<!-- fill-in: test command -->",
		"<!-- fill-in: build command -->"
}