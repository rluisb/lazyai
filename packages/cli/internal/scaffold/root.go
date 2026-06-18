package scaffold

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rluisb/lazyai/packages/cli/internal/compiler"
	"github.com/rluisb/lazyai/packages/cli/internal/conflict"
	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/globalpaths"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// errMemoryDocScopeUnsupported is returned by memoryDocDestPath for Copilot × global.
var errMemoryDocScopeUnsupported = errors.New("memory doc not supported at this scope")

const claudeAgentsReference = "<!-- ai-setup: AGENTS.md reference -->\nThis project uses [AGENTS.md](./AGENTS.md) as the canonical AI agent instruction file."

// TargetedUpdatePatch is the audit contract emitted by targeted AGENTS.md
// updates. It records only the slots that were safely patched and any slots
// skipped to preserve hand-authored content.
type TargetedUpdatePatch struct {
	File                         string                      `json:"file"`
	Replacements                 []TargetedUpdateReplacement `json:"replacements"`
	Warnings                     []string                    `json:"warnings"`
	PreservedUnrecognizedContent bool                        `json:"preservedUnrecognizedContent"`
}

type TargetedUpdateReplacement struct {
	Field    string                 `json:"field"`
	OldText  string                 `json:"oldText"`
	NewText  string                 `json:"newText"`
	Location TargetedUpdateLocation `json:"location"`
}

type TargetedUpdateLocation struct {
	Section   *string `json:"section"`
	LineStart *int    `json:"lineStart"`
	LineEnd   *int    `json:"lineEnd"`
}

type targetedFieldSpec struct {
	field        string
	newText      string
	section      string
	placeholders []string
	linePrefixes []string
}

// memoryDocDestPath returns the absolute path where the tool's memory doc
// (AGENTS.md / GEMINI.md / .github/copilot-instructions.md, or existing
// CLAUDE.md compatibility reference updates) should land for the given scope.
// Returns errMemoryDocScopeUnsupported when the
// combination is not supported (e.g. Copilot × global).
//
// Placement rules:
//   - project: <targetDir>/<outputFile> (Copilot's .github/ prefix preserved).
//   - workspace: <workspaceRoot>/<outputFile> when provided, otherwise targetDir
//     for backward compatibility.
//   - global: under the tool's global root, using the bare basename of
//     outputFile — Copilot × global is unsupported.
func memoryDocDestPath(tool types.ToolId, scope types.SetupScope, targetDir, workspaceRoot, homeDir, outputFile string) (string, error) {
	switch scope {
	case types.SetupScopeProject, "":
		return filepath.Join(targetDir, outputFile), nil
	case types.SetupScopeWorkspace:
		root := targetDir
		if workspaceRoot != "" {
			root = workspaceRoot
		}
		return filepath.Join(root, outputFile), nil
	case types.SetupScopeGlobal:
		if tool == types.ToolIdCopilot {
			return "", errMemoryDocScopeUnsupported
		}
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
	ctx := buildRootFragmentContext(opts, effectiveFeatures)

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
		if tool == types.ToolIdClaudeCode {
			if err := appendClaudeAgentsReference(opts); err != nil {
				return err
			}
		}

		outputFile, ok := RootFileByTool[tool]
		if !ok {
			continue
		}

		// Read the tool-specific root template from the library FS.
		templateRelPath := deprecatedRootTemplateByFile[outputFile]
		if templateRelPath == "" {
			templateRelPath = "root/" + outputFile + ".template.md"
		}
		if !files.ExistsFS(opts.LibraryFS, templateRelPath) {
			// Try alternative naming: AGENTS.template.md, copilot-instructions.template.md, etc.
			baseName := strings.TrimSuffix(outputFile, filepath.Ext(outputFile))
			templateRelPath = "root/" + baseName + ".template.md"
			if !files.ExistsFS(opts.LibraryFS, templateRelPath) {
				continue
			}
		}

		data, err := files.ReadFS(opts.LibraryFS, templateRelPath)
		if err != nil {
			scaffoldLog.Warn("could not read root template", "path", templateRelPath, "error", err)
			continue
		}

		content := string(data)
		// Perform hybrid [YOUR_*] placeholder fill (spec 010 wave C):
		// mechanical fields get concrete values; subjective fields get
		// HTML-comment <!-- fill-in --> markers.
		content = fillClaudeMdPlaceholders(content, opts)
		content = compiler.NewFragmentResolver("", opts.LibraryFS).Resolve(content, ctx)
		// Templating handlebars-style substitutions.
		content = strings.ReplaceAll(content, "{{projectName}}", ctx.ProjectName)
		content = strings.ReplaceAll(content, "{{planningDir}}", ctx.PlanningDir)
		if ctx.PrimaryLanguage != "" {
			content = strings.ReplaceAll(content, "{{primaryLanguage}}", ctx.PrimaryLanguage)
		}
		if ctx.Framework != "" {
			content = strings.ReplaceAll(content, "{{framework}}", ctx.Framework)
		}
		// Resolve any fallback [YOUR_*] placeholders emitted by the fragment
		// resolver after {{VARIABLE}} substitution.
		content = fillClaudeMdPlaceholders(content, opts)

		// Append workspace repos section if applicable.
		if workspaceReposSection != "" {
			content = content + workspaceReposSection
		}

		homeDir := opts.HomeDir
		destPath, err := memoryDocDestPath(tool, opts.SetupScope, opts.TargetDir, opts.WorkspaceRoot, homeDir, outputFile)
		if err != nil {
			if errors.Is(err, errMemoryDocScopeUnsupported) {
				scaffoldLog.Warn("skipping memory doc for unsupported scope", "tool", tool, "scope", opts.SetupScope)
				continue
			}
			return err
		}
		if err := files.EnsureDir(filepath.Dir(destPath)); err != nil {
			return err
		}

		recordRoot := opts.recordRoot()
		action, err := conflict.ApplyStrategy(destPath, opts.Strategy, opts.PerFileOverrides, recordRoot)
		if err != nil {
			return err
		}
		if action == "skip" {
			relPath, _ := filepath.Rel(recordRoot, destPath)
			scaffoldLog.Info("skipping existing file", "path", relPath)
			continue
		}
		if outputFile == "AGENTS.md" && files.FileExists(destPath) {
			existing, err := files.ReadFile(destPath)
			if err != nil {
				return err
			}
			patched, patch := BuildTargetedAgentsUpdatePatch(outputFile, string(existing), ctx)
			for _, warning := range patch.Warnings {
				scaffoldLog.Warn("targeted update warning", "path", outputFile, "warning", warning)
			}
			content = patched
		}

		if err := files.WriteFile(destPath, []byte(content), 0o644); err != nil {
			return err
		}
		// Signal populate if placeholders remain.
		if err := writePopulateSignal(recordRoot, content); err != nil {
			scaffoldLog.Warn("failed to write populate signal", "error", err)
		}

		hash, _ := files.FileHash(destPath)
		relPath, _ := filepath.Rel(recordRoot, destPath)
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

// BuildTargetedAgentsUpdatePatch applies the W1.A targeted update policy for an
// existing AGENTS.md: exact fallback placeholders are replaced, simple generated
// value slots are patched only when still safely recognizable, and all other
// content is preserved byte-for-byte with warnings for unsafe known slots.
func BuildTargetedAgentsUpdatePatch(file, existing string, ctx compiler.FragmentContext) (string, TargetedUpdatePatch) {
	patch := TargetedUpdatePatch{
		File:                         file,
		Replacements:                 []TargetedUpdateReplacement{},
		Warnings:                     []string{},
		PreservedUnrecognizedContent: true,
	}
	content := existing

	for _, spec := range targetedAgentsFieldSpecs(ctx) {
		if spec.newText == "" {
			continue
		}
		for _, placeholder := range spec.placeholders {
			content = replaceTargetedExact(content, placeholder, spec, &patch)
		}
	}

	for _, spec := range targetedAgentsFieldSpecs(ctx) {
		if spec.newText == "" || len(spec.linePrefixes) == 0 {
			continue
		}
		content = replaceTargetedLineSlots(content, spec, &patch)
	}
	warnUnsafeProjectOverview(content, ctx, &patch)

	return content, patch
}

func targetedAgentsFieldSpecs(ctx compiler.FragmentContext) []targetedFieldSpec {
	c := ctx.Constitution
	if c == nil {
		c = &compiler.ConstitutionContext{}
	}
	coverage := ""
	if c.CoverageThreshold != nil {
		coverage = strconv.Itoa(*c.CoverageThreshold)
	}
	return []targetedFieldSpec{
		{field: "PROJECT_OVERVIEW", newText: strings.TrimSpace(c.ProjectOverview), section: "Project Overview", placeholders: []string{"[YOUR_PROJECT_OVERVIEW]"}},
		{field: "LANGUAGE", newText: strings.TrimSpace(c.Stack.Language), section: "Project Overview", placeholders: []string{"[YOUR_LANGUAGE]"}, linePrefixes: []string{"- Language: "}},
		{field: "FRAMEWORK", newText: strings.TrimSpace(c.Stack.Framework), section: "Project Overview", placeholders: []string{"[YOUR_FRAMEWORK]"}, linePrefixes: []string{"- Framework: "}},
		{field: "DATABASE", newText: strings.TrimSpace(c.Stack.Database), section: "Project Overview", placeholders: []string{"[YOUR_DATABASE]"}, linePrefixes: []string{"- Database: "}},
		{field: "ORM", newText: strings.TrimSpace(c.Stack.ORM), section: "Project Overview", placeholders: []string{"[YOUR_ORM]"}, linePrefixes: []string{"- ORM/Query: "}},
		{field: "TEST_FRAMEWORK", newText: strings.TrimSpace(c.Stack.Testing), section: "Project Overview", placeholders: []string{"[YOUR_TEST_FRAMEWORK]"}, linePrefixes: []string{"- Testing: "}},
		{field: "PACKAGE_MANAGER", newText: strings.TrimSpace(c.Stack.PackageManager), section: "Project Overview", placeholders: []string{"[YOUR_PACKAGE_MANAGER]"}, linePrefixes: []string{"- Package manager: "}},
		{field: "NAMING_CONVENTIONS", newText: strings.TrimSpace(c.Conventions.Naming), section: "Conventions", placeholders: []string{"[YOUR_NAMING_CONVENTION]"}},
		{field: "ERROR_HANDLING", newText: strings.TrimSpace(c.Conventions.ErrorHandling), section: "Conventions", placeholders: []string{"[YOUR_ERROR_PATTERN]"}},
		{field: "API_CONVENTIONS", newText: strings.TrimSpace(c.Conventions.APIResponses), section: "Conventions", placeholders: []string{"[YOUR_API_CONVENTION]"}},
		{field: "IMPORT_ORDER", newText: strings.TrimSpace(c.Conventions.ImportOrder), section: "Conventions", placeholders: []string{"[YOUR_IMPORT_ORDER]"}},
		{field: "PROTECTED_BRANCH", newText: strings.TrimSpace(c.ProtectedBranch), section: "Do NOT", placeholders: []string{"[YOUR_PROTECTED_BRANCH]"}},
		{field: "TEST_COMMAND", newText: strings.TrimSpace(c.Commands.Test), section: "Key Commands", placeholders: []string{"<!-- fill-in: test command -->"}},
		{field: "LINT_COMMAND", newText: strings.TrimSpace(c.Commands.Lint), section: "Key Commands", placeholders: []string{"[YOUR_LINT_COMMAND]"}},
		{field: "BUILD_COMMAND", newText: strings.TrimSpace(c.Commands.Build), section: "Key Commands", placeholders: []string{"<!-- fill-in: build command -->"}},
		{field: "COVERAGE_THRESHOLD", newText: coverage, section: "Testing", placeholders: []string{"[YOUR_COVERAGE_THRESHOLD]"}, linePrefixes: []string{"- Minimum coverage: ", "- Minimum coverage threshold: "}},
	}
}

func replaceTargetedExact(content, oldText string, spec targetedFieldSpec, patch *TargetedUpdatePatch) string {
	if oldText == "" || spec.newText == "" || oldText == spec.newText {
		return content
	}
	searchStart := 0
	for searchStart <= len(content) {
		relativeIdx := strings.Index(content[searchStart:], oldText)
		if relativeIdx < 0 {
			return content
		}
		idx := searchStart + relativeIdx
		line := lineNumberAt(content, idx)
		patch.Replacements = append(patch.Replacements, TargetedUpdateReplacement{
			Field:    spec.field,
			OldText:  oldText,
			NewText:  spec.newText,
			Location: targetedLocation(spec.section, line, line),
		})
		content = content[:idx] + spec.newText + content[idx+len(oldText):]
		searchStart = idx + len(spec.newText)
	}
	return content
}

func replaceTargetedLineSlots(content string, spec targetedFieldSpec, patch *TargetedUpdatePatch) string {
	lines := strings.SplitAfter(content, "\n")
	changed := false
	for idx, line := range lines {
		body, ending := splitLineEnding(line)
		for _, prefix := range spec.linePrefixes {
			if !strings.HasPrefix(body, prefix) {
				continue
			}
			oldValue := strings.TrimSpace(strings.TrimPrefix(body, prefix))
			if normalizeSlotValue(oldValue) == spec.newText {
				continue
			}
			if !isSafeTargetedSlot(oldValue, spec) {
				patch.Warnings = append(patch.Warnings, fmt.Sprintf("left %s unchanged at line %d because existing value is not a recognized placeholder/value slot", spec.field, idx+1))
				continue
			}
			newBody := prefix + preserveSlotDelimiters(oldValue, spec.newText)
			patch.Replacements = append(patch.Replacements, TargetedUpdateReplacement{
				Field:    spec.field,
				OldText:  oldValue,
				NewText:  spec.newText,
				Location: targetedLocation(spec.section, idx+1, idx+1),
			})
			lines[idx] = newBody + ending
			changed = true
		}
	}
	if !changed {
		return content
	}
	return strings.Join(lines, "")
}

func warnUnsafeProjectOverview(content string, ctx compiler.FragmentContext, patch *TargetedUpdatePatch) {
	if ctx.Constitution == nil || strings.TrimSpace(ctx.Constitution.ProjectOverview) == "" {
		return
	}
	if strings.Contains(content, ctx.Constitution.ProjectOverview) || strings.Contains(content, "[YOUR_PROJECT_OVERVIEW]") {
		return
	}
	lines := strings.Split(content, "\n")
	for idx, line := range lines {
		if strings.TrimSpace(line) != "## Project Overview" {
			continue
		}
		for next := idx + 1; next < len(lines); next++ {
			trimmed := strings.TrimSpace(lines[next])
			if trimmed == "" || strings.HasPrefix(trimmed, "<!--") {
				continue
			}
			if strings.HasPrefix(trimmed, "## ") || strings.HasPrefix(trimmed, "**Stack:**") {
				return
			}
			patch.Warnings = append(patch.Warnings, fmt.Sprintf("left PROJECT_OVERVIEW unchanged at line %d because existing value is not a recognized placeholder/value slot", next+1))
			return
		}
	}
}

func splitLineEnding(line string) (body string, ending string) {
	if strings.HasSuffix(line, "\n") {
		ending = "\n"
		body = strings.TrimSuffix(line, "\n")
		if strings.HasSuffix(body, "\r") {
			body = strings.TrimSuffix(body, "\r")
			ending = "\r\n"
		}
		return body, ending
	}
	return line, ""
}

func isSafeTargetedSlot(oldValue string, spec targetedFieldSpec) bool {
	normalized := normalizeSlotValue(oldValue)
	if normalized == "" || strings.Contains(normalized, "[YOUR_") || strings.Contains(normalized, "{{") || strings.Contains(normalized, "fill-in:") {
		return true
	}
	for _, placeholder := range spec.placeholders {
		if normalized == placeholder {
			return true
		}
	}
	return spec.field == "COVERAGE_THRESHOLD" && normalized == "80"
}

func normalizeSlotValue(value string) string {
	trimmed := strings.TrimSpace(value)
	trimmed = strings.TrimSuffix(trimmed, "%")
	trimmed = strings.TrimPrefix(trimmed, "`")
	trimmed = strings.TrimSuffix(trimmed, "`")
	return strings.TrimSpace(trimmed)
}

func preserveSlotDelimiters(oldValue, newText string) string {
	trimmed := strings.TrimSpace(oldValue)
	if strings.HasPrefix(trimmed, "`") && strings.HasSuffix(trimmed, "`") {
		return "`" + newText + "`"
	}
	return newText
}

func targetedLocation(section string, lineStart, lineEnd int) TargetedUpdateLocation {
	sectionCopy := section
	return TargetedUpdateLocation{Section: &sectionCopy, LineStart: &lineStart, LineEnd: &lineEnd}
}

func lineNumberAt(content string, idx int) int {
	if idx <= 0 {
		return 1
	}
	return strings.Count(content[:idx], "\n") + 1
}

func appendClaudeAgentsReference(opts ScaffoldCompiledRootOptions) error {
	if opts.SetupScope != types.SetupScopeProject && opts.SetupScope != types.SetupScopeWorkspace && opts.SetupScope != "" {
		return nil
	}

	claudePath := filepath.Join(opts.recordRoot(), "CLAUDE.md")
	if !files.FileExists(claudePath) {
		return nil
	}

	data, err := files.ReadFile(claudePath)
	if err != nil {
		return err
	}
	content := string(data)
	if strings.Contains(content, claudeAgentsReference) {
		return nil
	}

	separator := "\n\n"
	if strings.HasSuffix(content, "\n") {
		separator = "\n"
	}
	updated := content + separator + claudeAgentsReference + "\n"
	return files.WriteFile(claudePath, []byte(updated), 0o644)
}

// ScaffoldCompiledRootOptions holds the options for compiling root files.
type ScaffoldCompiledRootOptions struct {
	TargetDir        string
	WorkspaceRoot    string
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
	StoreData        *types.StoreData
	// Optional context overrides.
	PrimaryLanguage     string
	Framework           string
	WorkspaceType       string
	ProjectInstructions string
	ProjectDescription  string // optional; substituted into [YOUR_PROJECT_DESCRIPTION]
	Organization        string // optional; substituted into [YOUR_ORG]
	Team                string // optional; substituted into [YOUR_TEAM]
	ProjectOverview     string
	Database            string
	ORM                 string
	TestFramework       string
	PackageManager      string
	MigrationsPath      string
	TestPath            string
	StrictMode          string
	InstallCommand      string
	ProtectedBranchGit  string // git-detected default branch
	NamingConventions   string
	ErrorHandling       string
	APIConventions      string
	ImportOrder         string
	ProtectedBranch     string
	TestCommand         string
	LintCommand         string
	BuildCommand        string
	CoverageThreshold   int
	CodebaseMap         []compiler.CodebaseMapEntry
	// Referenced repos for workspace scope.
	Repos []types.RepoInfo
}

func (opts ScaffoldCompiledRootOptions) recordRoot() string {
	if opts.SetupScope == types.SetupScopeWorkspace && opts.WorkspaceRoot != "" {
		return opts.WorkspaceRoot
	}
	return opts.TargetDir
}

func buildRootFragmentContext(opts ScaffoldCompiledRootOptions, features types.FeatureFlags) compiler.FragmentContext {
	config := types.Config{
		ProjectOverview:   opts.ProjectOverview,
		NamingConventions: opts.NamingConventions,
		ErrorHandling:     opts.ErrorHandling,
		ApiConventions:    opts.APIConventions,
		ImportOrder:       opts.ImportOrder,
		ProtectedBranch:   opts.ProtectedBranch,
		TestCommand:       opts.TestCommand,
		LintCommand:       opts.LintCommand,
		BuildCommand:      opts.BuildCommand,
		CoverageThreshold: opts.CoverageThreshold,
	}
	if opts.StoreData != nil {
		config = opts.StoreData.Config
		if opts.StoreData.Selections.Features != nil {
			features = *opts.StoreData.Selections.Features
		}
	}

	devCommand, testCommand, buildCommand := commandsForRoot(opts.PrimaryLanguage, opts.PackageManager)
	if config.TestCommand == "" {
		config.TestCommand = testCommand
	}
	if config.LintCommand == "" {
		config.LintCommand = lintCommandForRoot(opts.PackageManager, opts.PrimaryLanguage)
	}
	if config.BuildCommand == "" {
		config.BuildCommand = buildCommand
	}
	if config.ProtectedBranch == "" {
		config.ProtectedBranch = opts.ProtectedBranchGit
	}
	installCommand := firstRootNonEmpty(opts.InstallCommand, installCommandForPackageManager(opts.PackageManager), fallbackMarker("", "install command"))

	coverageThreshold := config.CoverageThreshold
	var coverageThresholdPtr *int
	if coverageThreshold > 0 {
		coverageThresholdPtr = &coverageThreshold
	}

	gitConventions := opts.GitConventions != nil
	return compiler.FragmentContext{
		ProjectName:         opts.ProjectName,
		PlanningDir:         opts.PlanningDir,
		PrimaryLanguage:     opts.PrimaryLanguage,
		Framework:           opts.Framework,
		WorkspaceType:       opts.WorkspaceType,
		ProjectInstructions: opts.ProjectInstructions,
		TestFramework:       opts.TestFramework,
		PackageManager:      opts.PackageManager,
		TestCommand:         config.TestCommand,
		LintCommand:         config.LintCommand,
		BuildCommand:        config.BuildCommand,
		DevCommand:          devCommand,
		InstallCommand:      installCommand,
		ProjectDescription:  opts.ProjectDescription,
		Features: &compiler.FeatureFlags{
			ContextEngineering: templateBoolPtr(features.ContextEngineering),
			RPIWorkflow:        templateBoolPtr(features.RPIWorkflow),
			ChainOfThought:     templateBoolPtr(features.ChainOfThought),
			TreeOfThoughts:     templateBoolPtr(features.TreeOfThoughts),
			ADREnforcement:     templateBoolPtr(features.ADREnforcement),
			QualityGates:       templateBoolPtr(features.QualityGates),
			AgentHarness:       templateBoolPtr(features.AgentHarness),
			BugResolution:      templateBoolPtr(features.BugResolution),
			PivotHandling:      templateBoolPtr(features.PivotHandling),
			GitConventions:     templateBoolPtr(gitConventions),
			AdversarialDesign:  templateBoolPtr(features.AdversarialDesign),
			// Legacy aliases.
			ContextEngineering_: templateBoolPtr(features.ContextEngineering),
			RPIWorkflow_:        templateBoolPtr(features.RPIWorkflow),
			ChainOfThought_:     templateBoolPtr(features.ChainOfThought),
			TreeOfThoughts_:     templateBoolPtr(features.TreeOfThoughts),
			ADREnforcement_:     templateBoolPtr(features.ADREnforcement),
			QualityGates_:       templateBoolPtr(features.QualityGates),
			AgentHarness_:       templateBoolPtr(features.AgentHarness),
			BugResolution_:      templateBoolPtr(features.BugResolution),
			PivotHandling_:      templateBoolPtr(features.PivotHandling),
			GitConventions_:     templateBoolPtr(gitConventions),
			AdversarialDesign_:  templateBoolPtr(features.AdversarialDesign),
		},
		Constitution: &compiler.ConstitutionContext{
			ProjectOverview: config.ProjectOverview,
			Stack: compiler.ConstitutionStack{
				Language:       opts.PrimaryLanguage,
				Framework:      opts.Framework,
				Database:       opts.Database,
				ORM:            opts.ORM,
				Testing:        opts.TestFramework,
				PackageManager: opts.PackageManager,
			},
			Conventions: compiler.ConstitutionConventions{
				Naming:        config.NamingConventions,
				ErrorHandling: config.ErrorHandling,
				APIResponses:  config.ApiConventions,
				ImportOrder:   config.ImportOrder,
			},
			Commands: compiler.ConstitutionCommands{
				Test:  config.TestCommand,
				Lint:  config.LintCommand,
				Build: config.BuildCommand,
			},
			ProtectedBranch:   config.ProtectedBranch,
			CoverageThreshold: coverageThresholdPtr,
			CodebaseMap:       buildRootCodebaseMap(opts),
		},
	}
}

func templateBoolPtr(value bool) *bool {
	return &value
}

func buildRootCodebaseMap(opts ScaffoldCompiledRootOptions) []compiler.CodebaseMapEntry {
	if len(opts.CodebaseMap) > 0 {
		return opts.CodebaseMap
	}
	if len(opts.Repos) == 0 {
		return detectTopLevelCodebaseMap(opts.TargetDir)
	}

	entries := make([]compiler.CodebaseMapEntry, 0, len(opts.Repos))
	for _, repo := range opts.Repos {
		entryPath := repo.Path
		if entryPath == "" {
			entryPath = repo.Name
		}
		entries = append(entries, compiler.CodebaseMapEntry{Path: entryPath})
	}
	return entries
}

func detectTopLevelCodebaseMap(targetDir string) []compiler.CodebaseMapEntry {
	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return nil
	}

	codebaseMap := make([]compiler.CodebaseMapEntry, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() || isIgnoredTopLevelCodebasePath(entry.Name()) {
			continue
		}
		codebaseMap = append(codebaseMap, compiler.CodebaseMapEntry{Path: entry.Name()})
	}
	return codebaseMap
}

func isIgnoredTopLevelCodebasePath(path string) bool {
	switch path {
	case "node_modules", "dist", ".git", "vendor":
		return true
	default:
		return false
	}
}

// fillClaudeMdPlaceholders replaces [YOUR_*] placeholders in an AGENTS.md /
// GEMINI.md template with values derived from opts, or with
// <!-- fill-in: hint --> markers when the value is unknown or subjective.
// Applying this centrally means each tool's root file gets consistent
// substitutions.
func fillClaudeMdPlaceholders(content string, opts ScaffoldCompiledRootOptions) string {
	dev, inferredTest, inferredBuild := commandsForRoot(opts.PrimaryLanguage, opts.PackageManager)
	testCommand := firstRootNonEmpty(opts.TestCommand, inferredTest)
	buildCommand := firstRootNonEmpty(opts.BuildCommand, inferredBuild)
	lintCommand := firstRootNonEmpty(opts.LintCommand, lintCommandForRoot(opts.PackageManager, opts.PrimaryLanguage))
	installCommand := firstRootNonEmpty(opts.InstallCommand, installCommandForPackageManager(opts.PackageManager))
	protectedBranch := firstRootNonEmpty(opts.ProtectedBranch, opts.ProtectedBranchGit)
	coverageThreshold := "80"
	if opts.CoverageThreshold > 0 {
		coverageThreshold = strconv.Itoa(opts.CoverageThreshold)
	}

	techStack := ""
	if opts.PrimaryLanguage != "" {
		techStack = opts.PrimaryLanguage
		if opts.Framework != "" {
			techStack += " · " + opts.Framework
		}
	}

	mechanicalMarkers := map[string]string{
		"[YOUR_PROJECT_NAME]":        fallbackMarker(opts.ProjectName, "project name"),
		"[YOUR_PROJECT_DESCRIPTION]": fallbackMarker(opts.ProjectDescription, "project description"),
		"[YOUR_PROJECT_OVERVIEW]":    fallbackMarker(opts.ProjectOverview, "project overview"),
		"[YOUR_LANGUAGE]":            fallbackMarker(opts.PrimaryLanguage, "language"),
		"[YOUR_FRAMEWORK]":           fallbackMarker(opts.Framework, "framework"),
		"[YOUR_DATABASE]":            fallbackMarker(opts.Database, "database"),
		"[YOUR_ORM]":                 fallbackMarker(opts.ORM, "ORM or query layer"),
		"[YOUR_TEST_FRAMEWORK]":      fallbackMarker(opts.TestFramework, "test framework"),
		"[YOUR_PACKAGE_MANAGER]":     fallbackMarker(opts.PackageManager, "package manager"),
		"[YOUR_TECH_STACK]":          fallbackMarker(techStack, "tech stack"),
		"[YOUR_ORG]":                 fallbackMarker(opts.Organization, "your org"),
		"[YOUR_TEAM]":                fallbackMarker(opts.Team, "your team"),
		"[YOUR_INSTALL_COMMAND]":     fallbackMarker(installCommand, "install command"),
		"[YOUR_TEST_COMMAND]":        fallbackMarker(testCommand, "test command"),
		"[YOUR_LINT_COMMAND]":        fallbackMarker(lintCommand, "lint command"),
		"[YOUR_DEV_COMMAND]":         fallbackMarker(dev, "dev command"),
		"[YOUR_BUILD_COMMAND]":       fallbackMarker(buildCommand, "build command"),
		"[YOUR_COVERAGE_THRESHOLD]":  coverageThreshold,
		"[YOUR_NAMING_CONVENTION]":   fallbackMarker(opts.NamingConventions, "naming convention"),
		"[YOUR_NAMING_CONVENTIONS]":  fallbackMarker(opts.NamingConventions, "naming conventions"),
		"[YOUR_ERROR_PATTERN]":       fallbackMarker(opts.ErrorHandling, "error handling pattern"),
		"[YOUR_API_CONVENTION]":      fallbackMarker(opts.APIConventions, "API response convention"),
		"[YOUR_IMPORT_ORDER]":        fallbackMarker(opts.ImportOrder, "import order"),
		"[YOUR_PROTECTED_BRANCH]":    fallbackMarker(protectedBranch, "protected branch"),
		"[YOUR_MIGRATIONS_PATH]":     fallbackMarker(opts.MigrationsPath, "migrations path"),
		"[YOUR_STRICT_MODE]":         fallbackMarker(opts.StrictMode, "strict mode"),
		"[YOUR_SHARED_PATH]":         fallbackMarker(sharedPathFromCodebaseMap(opts.CodebaseMap), "shared path"),
		"[YOUR_TEST_PATH]":           fallbackMarker(opts.TestPath, "test path"),
	}
	for placeholder, value := range mechanicalMarkers {
		content = strings.ReplaceAll(content, placeholder, value)
	}

	content = strings.ReplaceAll(content, "[YOUR_DEV_DESCRIPTION]", "Run the app from source")
	content = strings.ReplaceAll(content, "[YOUR_TEST_DESCRIPTION]", "Run the test suite")
	content = strings.ReplaceAll(content, "[YOUR_BUILD_DESCRIPTION]", "Build the project")

	// Subjective fields — always fill-in markers.
	subjectiveMarkers := map[string]string{
		"[YOUR_ARCHITECTURE_NOTES]":           "<!-- fill-in: architecture and key patterns -->",
		"[YOUR_CODE_STYLE]":                   "<!-- fill-in: code style -->",
		"[YOUR_NAMING_CONVENTIONS]":           "<!-- fill-in: naming conventions -->",
		"[YOUR_TESTING_STRATEGY]":             "<!-- fill-in: testing strategy -->",
		"[YOUR_GIT_WORKFLOW]":                 "<!-- fill-in: git workflow -->",
		"[YOUR_RULE_1]":                       "<!-- fill-in: rule 1 -->",
		"[YOUR_RULE_2]":                       "<!-- fill-in: rule 2 -->",
		"[YOUR_DO_NOT_1]":                     "<!-- fill-in: project-specific don't -->",
		"[YOUR_DO_NOT_2]":                     "<!-- fill-in: project-specific don't -->",
		"[YOUR_UNIT_TESTING_STRATEGY]":        "<!-- fill-in: unit testing strategy -->",
		"[YOUR_INTEGRATION_TESTING_STRATEGY]": "<!-- fill-in: integration testing strategy -->",
		"[YOUR_E2E_TESTING_STRATEGY]":         "<!-- fill-in: e2e testing strategy -->",
		"[YOUR_SESSION_CHECK]":                "<!-- fill-in: team-specific session check -->",
		"[YOUR_COMPONENT_1]":                  "<!-- fill-in: component -->",
		"[YOUR_COMPONENT_2]":                  "<!-- fill-in: component -->",
		"[YOUR_RESPONSIBILITY_1]":             "<!-- fill-in: responsibility -->",
		"[YOUR_RESPONSIBILITY_2]":             "<!-- fill-in: responsibility -->",
		"[YOUR_PATH_1]":                       "<!-- fill-in: path -->",
		"[YOUR_PATH_2]":                       "<!-- fill-in: path -->",
		"[YOUR_PATH_3]":                       "<!-- fill-in: path -->",
		"[YOUR_INFRA_PATH]":                   "<!-- fill-in: infra path -->",
	}
	for placeholder, marker := range subjectiveMarkers {
		content = strings.ReplaceAll(content, placeholder, marker)
	}

	return content
}

// fallbackMarker returns value if non-empty, else an HTML-comment fill-in hint.
func fallbackMarker(value, hint string) string {
	if strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return "<!-- fill-in: " + hint + " -->"
}

func firstRootNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func commandsForRoot(lang, pkgManager string) (dev, test, build string) {
	if isJSLanguage(lang) {
		switch strings.ToLower(pkgManager) {
		case "pnpm":
			return "pnpm dev", "pnpm test", "pnpm build"
		case "yarn":
			return "yarn dev", "yarn test", "yarn build"
		case "npm":
			return "npm run dev", "npm test", "npm run build"
		}
	}
	return commandsForLanguage(lang)
}

func installCommandForPackageManager(pkgManager string) string {
	switch strings.ToLower(pkgManager) {
	case "pnpm":
		return "pnpm install"
	case "yarn":
		return "yarn install"
	case "npm":
		return "npm install"
	case "go modules":
		return "go mod tidy"
	case "cargo":
		return "cargo build"
	case "bundler":
		return "bundle install"
	case "poetry":
		return "poetry install"
	case "pipenv":
		return "pipenv install"
	case "composer":
		return "composer install"
	}
	return ""
}

func lintCommandForRoot(pkgManager, lang string) string {
	switch strings.ToLower(lang) {
	case "go":
		return "go vet ./..."
	case "python":
		return "ruff check ."
	case "rust":
		return "cargo clippy"
	case "ruby":
		return "bundle exec rubocop"
	}
	switch strings.ToLower(pkgManager) {
	case "pnpm":
		return "pnpm lint"
	case "yarn":
		return "yarn lint"
	case "npm":
		return "npm run lint"
	}
	return ""
}

func isJSLanguage(lang string) bool {
	switch strings.ToLower(lang) {
	case "node", "nodejs", "node.js", "javascript", "typescript":
		return true
	}
	return false
}

func sharedPathFromCodebaseMap(entries []compiler.CodebaseMapEntry) string {
	for _, entry := range entries {
		if strings.Contains(strings.ToLower(entry.Responsibility), "shared utilities") &&
			strings.TrimSpace(entry.Path) != "" &&
			!strings.Contains(entry.Path, "fill-in") {
			return entry.Path
		}
	}
	return ""
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
