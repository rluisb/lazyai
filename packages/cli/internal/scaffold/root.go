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

// claudeImportToken is the functional @import token written into CLAUDE.md so
// Claude Code pulls in the canonical AGENTS.md instructions. The idempotency
// guard checks for this token (NOT a markdown comment) so a freshly-generated
// file is a true no-op on recompile (#496).
const claudeImportToken = "@AGENTS.md"

// claudeContextAppend is appended to a user-owned CLAUDE.md that lacks the
// AGENTS.md import. It leads with a short human-readable comment and ends with
// the functional @import so the guard-matched token is the import itself.
const claudeContextAppend = "<!-- ai-setup: import canonical AGENTS.md for Claude Code -->\n@AGENTS.md"

// claudeContextDoc is the body written to a generated CLAUDE.md when the Claude
// Code target is selected and no CLAUDE.md exists yet. Claude Code reads
// CLAUDE.md natively (AGENTS.md alone is not sufficient — see FR-012); the
// `@AGENTS.md` import pulls in the canonical instructions so there is a single
// source of truth.
const claudeContextDoc = "# CLAUDE.md\n\n" +
	"Claude Code reads this file as the canonical project instruction file.\n" +
	"The full agent instructions for this project live in AGENTS.md and are\n" +
	"imported below.\n\n" +
	"@AGENTS.md\n"

// geminiImportToken is the functional @import token written into GEMINI.md so
// Gemini CLI pulls in the canonical AGENTS.md instructions. The idempotency
// guard checks for this token (NOT a markdown comment) so a freshly-generated
// file is a true no-op on recompile (#496).
const geminiImportToken = "@./AGENTS.md"

// geminiContextAppend is appended to a user-owned GEMINI.md that lacks the
// AGENTS.md import. It leads with a short human-readable comment and ends with
// the functional @import so the guard-matched token is the import itself.
const geminiContextAppend = "<!-- ai-setup: import canonical AGENTS.md for Gemini -->\n@./AGENTS.md"

// geminiContextDoc is the body written to a generated GEMINI.md when the
// Antigravity/Gemini target is selected and no GEMINI.md exists yet. Gemini CLI
// reads GEMINI.md natively as its context file (a bare root AGENTS.md is not
// discovered — #486 gap 2); the `@./AGENTS.md` memory import pulls in the
// canonical instructions so there is a single source of truth.
const geminiContextDoc = "# GEMINI.md\n\n" +
	"Gemini CLI reads this file as the canonical project context file.\n" +
	"The full agent instructions for this project live in AGENTS.md and are\n" +
	"imported below.\n\n" +
	"@./AGENTS.md\n"

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
		dest := outputFile
		// OMP at project scope: emit to .omp/AGENTS.md (native provider, priority 100)
		// instead of bare AGENTS.md (agents-md provider, priority 10) — M8 fix.
		if tool == types.ToolIdOmp {
			dest = filepath.Join(".omp", filepath.Base(outputFile))
		}
		return filepath.Join(targetDir, dest), nil
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
			if err := ensureClaudeContextDoc(opts); err != nil {
				return err
			}
		}
		if tool == types.ToolIdAntigravity {
			if err := ensureGeminiContextDoc(opts); err != nil {
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

// ensureClaudeContextDoc guarantees a native CLAUDE.md exists for the Claude
// Code target (FR-012): AGENTS.md alone is not sufficient for Claude Code. When
// no CLAUDE.md exists, a minimal one importing AGENTS.md (`@AGENTS.md`) is
// generated and tracked. When the user already has a CLAUDE.md, an AGENTS.md
// reference is appended (idempotently) instead of clobbering their content.
// Only project/workspace scopes are touched — the user's personal
// ~/.claude/CLAUDE.md at global scope is never created or modified.
func ensureClaudeContextDoc(opts ScaffoldCompiledRootOptions) error {
	return ensureToolContextDoc(opts, "CLAUDE.md", claudeContextDoc, claudeImportToken, claudeContextAppend, types.ToolIdClaudeCode)
}

// ensureToolContextDoc is the shared implementation behind ensureClaudeContextDoc
// and ensureGeminiContextDoc. It guarantees a native context doc (CLAUDE.md or
// GEMINI.md) exists for tools that need a per-tool memory file. When the file
// is missing, a minimal body importing AGENTS.md is generated and tracked. When
// the user already owns the file, a functional import is appended (idempotently)
// instead of clobbering their content. Only project/workspace scopes are
// touched — personal config at global scope is never created or modified.
//
// importToken is the functional @mention that triggers the tool to pull in
// AGENTS.md (e.g. "@AGENTS.md" for Claude, "@./AGENTS.md" for Gemini). It is
// also the idempotency guard so a freshly-generated file is a no-op on
// recompile (#496). appendBody is the full text appended to a user-owned file
// that lacks the import — it leads with a short human-readable comment and
// ends with importToken so the guard-matched token is the import itself.
func ensureToolContextDoc(opts ScaffoldCompiledRootOptions, filename, body, importToken, appendBody string, toolID types.ToolId) error {
	if opts.SetupScope != types.SetupScopeProject && opts.SetupScope != types.SetupScopeWorkspace && opts.SetupScope != "" {
		return nil
	}

	recordRoot := opts.recordRoot()
	docPath := filepath.Join(recordRoot, filename)
	if !files.FileExists(docPath) {
		if err := files.EnsureDir(filepath.Dir(docPath)); err != nil {
			return err
		}
		if err := files.WriteFile(docPath, []byte(body), 0o644); err != nil {
			return err
		}
		return recordContextDoc(opts, recordRoot, docPath, toolID)
	}

	data, err := files.ReadFile(docPath)
	if err != nil {
		return err
	}
	content := string(data)
	// Guard on the functional @import token, not a markdown comment, so a
	// freshly-generated file (already containing importToken) is a no-op on
	// recompile and a user-owned file only gets one functional import (#496).
	if strings.Contains(content, importToken) {
		// No content change, but still (re)record the TrackedFile so the
		// lockfile stays in sync on every compile — no drift (#496).
		return recordContextDoc(opts, recordRoot, docPath, toolID)
	}

	separator := "\n\n"
	if strings.HasSuffix(content, "\n") {
		separator = "\n"
	}
	updated := content + separator + appendBody + "\n"
	if err := files.WriteFile(docPath, []byte(updated), 0o644); err != nil {
		return err
	}
	return recordContextDoc(opts, recordRoot, docPath, toolID)
}

// recordContextDoc appends a TrackedFile (relative path + current hash) for a
// generated/updated context doc so the lockfile reflects the on-disk file on
// every compile, including the idempotent no-op path. The FileHash error is
// surfaced rather than swallowed (#496).
func recordContextDoc(opts ScaffoldCompiledRootOptions, recordRoot, absPath string, tool types.ToolId) error {
	if opts.FileRecords == nil {
		return nil
	}
	hash, err := files.FileHash(absPath)
	if err != nil {
		return fmt.Errorf("hash %s: %w", absPath, err)
	}
	relPath, _ := filepath.Rel(recordRoot, absPath)
	if relPath == "" || strings.HasPrefix(relPath, "..") {
		relPath = absPath
	}
	*opts.FileRecords = append(*opts.FileRecords, types.TrackedFile{
		Path:   relPath,
		Hash:   hash,
		Source: "compiled:" + string(tool),
		Owner:  types.FileOwnerLibrary,
	})
	return nil
}

// ensureGeminiContextDoc guarantees a native GEMINI.md exists for the
// Antigravity/Gemini target (#486 gap 2): Gemini CLI reads GEMINI.md as its
// context file and neither Gemini CLI nor Antigravity discovers a bare root
// AGENTS.md. When no GEMINI.md exists, a minimal one importing AGENTS.md
// (`@./AGENTS.md`) is generated and tracked. When the user already has a
// GEMINI.md, an AGENTS.md reference is appended (idempotently) instead of
// clobbering their content. Only project/workspace scopes are touched — the
// user's personal ~/.gemini/GEMINI.md global rules are never created or modified.
func ensureGeminiContextDoc(opts ScaffoldCompiledRootOptions) error {
	return ensureToolContextDoc(opts, "GEMINI.md", geminiContextDoc, geminiImportToken, geminiContextAppend, types.ToolIdAntigravity)
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
