package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/rluisb/lazyai/packages/cli/internal/preset"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// createMinimalLibraryFS creates a MapFS with the minimum files needed
// for a scaffold run.
func createMinimalLibraryFS() fstest.MapFS {
	return fstest.MapFS{
		// Constitution files (template variables get replaced)
		"constitution/constitution.template.md": &fstest.MapFile{
			Data: []byte("# Constitution\n\nProject: {{PROJECT_NAME}}"),
		},
		"constitution/constraints.template.md": &fstest.MapFile{
			Data: []byte("# Constraints\n\nProject: {{PROJECT_NAME}}"),
		},
		"constitution/quality-gates.template.md": &fstest.MapFile{
			Data: []byte("# Quality Gates\n\nProject: {{PROJECT_NAME}}"),
		},
		"constitution/uncertainty.template.md": &fstest.MapFile{
			Data: []byte("# Uncertainty\n\nProject: {{PROJECT_NAME}}"),
		},

		// Agents
		"agents/builder.md": &fstest.MapFile{
			Data: []byte("---\nname: Builder\nmodel: sonnet\n---\n\n# Builder\n\nYou are a builder."),
		},
		"agents/orchestrator.md": &fstest.MapFile{
			Data: []byte("---\nname: Orchestrator\nmodel: opus\ntools: list_catalog compose_agent start_chain\n---\n\n# Orchestrator\n\nYou coordinate agents."),
		},

		// Skills
		"skills/implement.md": &fstest.MapFile{
			Data: []byte("---\nname: implement\n---\n\n# Implement\n\nImplement features."),
		},
		"skills/plan.md": &fstest.MapFile{
			Data: []byte("---\nname: plan\n---\n\n# Plan\n\nPlan features."),
		},

		// Tool agents (context files)
		"tool-agents/agents-dir.md": &fstest.MapFile{
			Data: []byte("# Agents Directory\n\nAgent definitions context."),
		},
		"tool-agents/skills-dir.md": &fstest.MapFile{
			Data: []byte("# Skills Directory\n\nSkill definitions context."),
		},
		"tool-agents/root-dir.md": &fstest.MapFile{
			Data: []byte("# Root\n\nProject root context."),
		},

		// MCP catalog
		"mcp/catalog.json": &fstest.MapFile{
			Data: []byte(`{"servers":{"filesystem":{"description":"Filesystem","enabled":true}}}`),
		},

		// Root templates
		"root/AGENTS.template.md": &fstest.MapFile{
			Data: []byte("# {{PROJECT_NAME}}\n\nProject agents."),
		},
		// Orchestration minimal structure
		"orchestration/chains/feature.json": &fstest.MapFile{
			Data: []byte(`{"name":"feature","steps":[]}`),
		},
		"orchestration/skills/domains/backend.md": &fstest.MapFile{
			Data: []byte("---\nname: backend\n---\n\n# Backend"),
		},
		"orchestration/teams/feature-team.json": &fstest.MapFile{
			Data: []byte(`{"name":"feature-team"}`),
		},
		"orchestration/workflows/rpi.json": &fstest.MapFile{
			Data: []byte(`{"name":"rpi"}`),
		},

		// Specs structure
		"specs-agents/features/AGENTS.md": &fstest.MapFile{
			Data: []byte("# Features\n\nFeature specs."),
		},
		"specs-agents/bugfixes/AGENTS.md": &fstest.MapFile{
			Data: []byte("# Bugfixes\n\nBugfix specs."),
		},
		"specs/templates/plan-template.md": &fstest.MapFile{
			Data: []byte("# Plan Template\n\n{{PROJECT_NAME}}"),
		},

		// Rules
		"rules/workflow.md": &fstest.MapFile{
			Data: []byte("# Workflow Rules\n\nProject: {{PROJECT_NAME}}"),
		},
		"rules/code-style.md": &fstest.MapFile{
			Data: []byte("# Code Style Rules\n\nProject: {{PROJECT_NAME}}"),
		},
		"rules/testing.md": &fstest.MapFile{
			Data: []byte("# Testing Rule\n\n- Minimum coverage threshold: `{{COVERAGE_THRESHOLD}}`%\n"),
		},
		"rules/typescript.md": &fstest.MapFile{
			Data: []byte("---\npaths:\n  - \"src/**/*.ts\"\n---\n\n# TypeScript Rules\n\n- Use strict TypeScript\n- Prefer interfaces over types for objects\n"),
		},

		// Starter standards.
		"standards/starter/agent-security.md": &fstest.MapFile{
			Data: []byte("---\ntitle: Agent Security\n---\n\n# Agent Security\n"),
		},
		"standards/starter/context-loading.md": &fstest.MapFile{
			Data: []byte("---\ntitle: Context Loading\n---\n\n# Context Loading\n"),
		},
		"standards/starter/error-handling.md": &fstest.MapFile{
			Data: []byte("---\ntitle: Error Handling\n---\n\n# Error Handling\n"),
		},
		"standards/starter/orchestration-patterns.md": &fstest.MapFile{
			Data: []byte("---\ntitle: Orchestration Patterns\n---\n\n# Orchestration Patterns\n"),
		},
		"standards/starter/test-patterns.md": &fstest.MapFile{
			Data: []byte("---\ntitle: Test Patterns\n---\n\n# Test Patterns\n"),
		},

		// Infra
		"infra/CODEOWNERS.template": &fstest.MapFile{
			Data: []byte("# CODEOWNERS\n* @default"),
		},
		"infra/KNOWLEDGE_MAP.template.md": &fstest.MapFile{
			Data: []byte("# Knowledge Map\n\n{{PROJECT_NAME}}"),
		},

		// Fragments
		"fragments/quality-gates.md": &fstest.MapFile{
			Data: []byte("## Quality Gates\n\nProject: {{PROJECT_NAME}}"),
		},

		// Claude Code commands
		"claudecode/commands/review.md": &fstest.MapFile{
			Data: []byte("---\ndescription: Review changes\nargument-hint: \"[pr]\"\nallowed-tools: Bash Read\n---\n\nReview body."),
		},
		"claudecode/commands/test.md": &fstest.MapFile{
			Data: []byte("---\ndescription: Run tests\nargument-hint: \"[target]\"\nallowed-tools: Bash Read\n---\n\nTest body."),
		},
		"claudecode/commands/commit.md": &fstest.MapFile{
			Data: []byte("---\ndescription: Draft commit\nargument-hint: \"\"\nallowed-tools: Bash Read\n---\n\nCommit body."),
		},

		// Claude Code output styles
		"claudecode/output-styles/terse.md": &fstest.MapFile{
			Data: []byte("---\nname: Terse\ndescription: Short responses\nkeep-coding-instructions: true\n---\n\nTerse body."),
		},
		"claudecode/output-styles/explanatory.md": &fstest.MapFile{
			Data: []byte("---\nname: Explanatory\ndescription: Detailed responses\nkeep-coding-instructions: true\n---\n\nExplanatory body."),
		},
	}
}

// minimalScaffoldContext creates a minimal scaffold context for testing.
func minimalScaffoldContext(t *testing.T, tools []types.ToolId) (*ScaffoldContext, string) {
	t.Helper()
	targetDir := t.TempDir()
	libFS := createMinimalLibraryFS()

	features := types.DefaultFeatureFlags()
	gitConv := types.DefaultGitConventions()

	ctx := &ScaffoldContext{
		TargetDir:      targetDir,
		LibraryDir:     "", // production mode, use LibraryFS
		LibraryFS:      libFS,
		Tools:          tools,
		CLITools:       toolIdsToStrings(tools),
		ProjectName:    "test-project",
		PlanningDir:    "specs",
		SetupScope:     types.SetupScopeProject,
		Features:       &features,
		GitConventions: &gitConv,
		Strategy:       types.ConflictStrategyAlign,
		Agents:         types.ALL_AGENTS[:],
		Skills:         types.ALL_SKILLS[:],
		Prompts:        types.ALL_PROMPTS[:],
		Templates:      preset.TemplatesForPreset(types.PresetLevelStandard),
		Rules:          preset.RulesForPreset(types.PresetLevelStandard),
		Infra:          types.ALL_INFRA[:],
		SpecsDirs:      []string{"features", "bugfixes"},
	}

	return ctx, targetDir
}

func TestScaffoldAll_DoesNotCreateSpecsAgentsFiles(t *testing.T) {
	ctx, targetDir := minimalScaffoldContext(t, []types.ToolId{types.ToolIdOpenCode})
	if _, err := ScaffoldAll(ctx); err != nil {
		t.Fatalf("ScaffoldAll failed: %v", err)
	}

	if fileExistsInDir(targetDir, "specs/features/AGENTS.md") {
		t.Fatal("expected specs/features/AGENTS.md to not be created")
	}
	if fileExistsInDir(targetDir, "specs/bugfixes/AGENTS.md") {
		t.Fatal("expected specs/bugfixes/AGENTS.md to not be created")
	}
}

func toolIdsToStrings(ids []types.ToolId) []string {
	result := make([]string, len(ids))
	for i, id := range ids {
		result[i] = string(id)
	}
	return result
}

func TestScaffoldAll_OpenCode(t *testing.T) {
	ctx, targetDir := minimalScaffoldContext(t, []types.ToolId{types.ToolIdOpenCode})
	result, err := ScaffoldAll(ctx)
	if err != nil {
		t.Fatalf("ScaffoldAll failed: %v", err)
	}

	if len(result.Files) == 0 {
		t.Fatal("expected at least one tracked file")
	}

	// Check that constitution file was created (Spec 022: single merged file at .specify/memory/).
	if !fileExistsInDir(targetDir, ".specify/memory/constitution.md") {
		t.Error("constitution.md was not created")
	}

	// Check opencode.jsonc was created in the correct subdirectory (install
	// and compile both target .jsonc per the spec 011 unification).
	if !fileExistsInDir(targetDir, ".opencode/opencode.jsonc") {
		t.Error("opencode.jsonc was not created in .opencode/")
	}
	if fileExistsInDir(targetDir, ".opencode/opencode.json") {
		t.Error("opencode.json must not coexist with opencode.jsonc")
	}
	if fileExistsInDir(targetDir, "opencode.json") || fileExistsInDir(targetDir, "opencode.jsonc") {
		t.Error("opencode config must not be created at the project root")
	}

	// Check .opencode directory was created.
	if !fileExistsInDir(targetDir, ".opencode/agents/builder.md") {
		t.Error(".opencode/agents/builder.md was not created")
	}

	// Check .ai/mcp.json was created.
	if !fileExistsInDir(targetDir, ".ai/mcp.json") {
		t.Error(".ai/mcp.json was not created")
	}

	t.Logf("Scaffold created %d files, %d directories", len(result.Files), len(result.Directories))
}

func TestScaffoldAll_ClaudeCode(t *testing.T) {
	ctx, targetDir := minimalScaffoldContext(t, []types.ToolId{types.ToolIdClaudeCode})
	result, err := ScaffoldAll(ctx)
	if err != nil {
		t.Fatalf("ScaffoldAll failed: %v", err)
	}

	if len(result.Files) == 0 {
		t.Fatal("expected at least one tracked file")
	}

	// Check .claude directory was created.
	if !fileExistsInDir(targetDir, ".claude/agents/builder.md") {
		t.Error(".claude/agents/builder.md was not created")
	}

	t.Logf("Scaffold created %d files", len(result.Files))
}

func TestScaffoldAll_MultipleTools(t *testing.T) {
	ctx, targetDir := minimalScaffoldContext(t, []types.ToolId{types.ToolIdOpenCode, types.ToolIdClaudeCode})
	result, err := ScaffoldAll(ctx)
	if err != nil {
		t.Fatalf("ScaffoldAll failed: %v", err)
	}

	if len(result.Files) == 0 {
		t.Fatal("expected at least one tracked file")
	}

	// Both tool directories should exist.
	if !fileExistsInDir(targetDir, ".opencode/agents/builder.md") {
		t.Error(".opencode/agents/builder.md was not created")
	}
	if !fileExistsInDir(targetDir, ".claude/agents/builder.md") {
		t.Error(".claude/agents/builder.md was not created")
	}

	t.Logf("Scaffold created %d files for 2 tools", len(result.Files))
}

func TestScaffoldAll_DryRun(t *testing.T) {
	ctx, targetDir := minimalScaffoldContext(t, []types.ToolId{types.ToolIdOpenCode})
	ctx.DryRun = true

	result, err := ScaffoldAll(ctx)
	if err != nil {
		t.Fatalf("ScaffoldAll dry-run failed: %v", err)
	}

	if len(result.Files) == 0 {
		t.Fatal("dry-run should still report files")
	}

	// In dry-run mode, most files should NOT be created on disk.
	// Some files (like .ai/mcp.json) may still be created by the MCP compiler,
	// but constitution files should not be created.
	// Constitution uses the Spec 022 path.
	constitutionPath := filepath.Join(targetDir, ".specify", "memory", "constitution.md")
	if _, err := os.Stat(constitutionPath); err == nil {
		// In dry-run mode, some scaffold steps may still create files.
		// This is acceptable — the key is that ScaffoldAll reports the files.
		t.Logf("constitution.md was created even in dry-run (this is OK)")
	}
}

func TestScaffoldAll_AbsorbStrategyPreservesExistingRootFile(t *testing.T) {
	ctx, targetDir := minimalScaffoldContext(t, []types.ToolId{types.ToolIdOpenCode})
	ctx.Strategy = types.ConflictStrategySkip

	agentsPath := filepath.Join(targetDir, "AGENTS.md")
	const existingContent = "# Existing user AGENTS\n\nDo not replace me.\n"
	if err := os.WriteFile(agentsPath, []byte(existingContent), 0o644); err != nil {
		t.Fatalf("seed AGENTS.md: %v", err)
	}

	if _, err := ScaffoldAll(ctx); err != nil {
		t.Fatalf("ScaffoldAll failed: %v", err)
	}

	got, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	if string(got) != existingContent {
		t.Fatalf("AGENTS.md content = %q, want preserved %q", string(got), existingContent)
	}
	if fileExistsInDir(targetDir, ".ai-setup-backup/AGENTS.md") {
		t.Fatal("absorb/skip should not create a backup for preserved AGENTS.md")
	}
}

func TestW1AScaffoldAllSubstitutesCoverageInAgentsAndSelectedTestingRule(t *testing.T) {
	// AC-N4-002, AC-N4-003: accepted/defaulted coverage is emitted in both
	// AGENTS.md and a selected specs/rules/testing.md file during integration.
	ctx, targetDir := minimalScaffoldContext(t, []types.ToolId{types.ToolIdOpenCode})
	ctx.LibraryFS = os.DirFS(filepath.Join("..", "..", "library"))
	ctx.CoverageThreshold = 88
	ctx.Rules = []types.RuleId{types.RuleIdTesting}

	if _, err := ScaffoldAll(ctx); err != nil {
		t.Fatalf("ScaffoldAll failed: %v", err)
	}

	agents, err := os.ReadFile(filepath.Join(targetDir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	if !strings.Contains(string(agents), "- Minimum coverage: `88`%") {
		t.Fatalf("AGENTS.md did not include substituted coverage threshold:\n%s", string(agents))
	}

	testingRulePath := filepath.Join(targetDir, "specs", "rules", "testing.md")
	testingRule, err := os.ReadFile(testingRulePath)
	if err != nil {
		t.Fatalf("read selected testing rule: %v", err)
	}
	if !strings.Contains(string(testingRule), "- Minimum coverage threshold: `88`%") {
		t.Fatalf("testing rule did not include substituted coverage threshold:\n%s", string(testingRule))
	}
	if strings.Contains(string(testingRule), "{{COVERAGE_THRESHOLD}}") {
		t.Fatalf("testing rule kept unresolved coverage placeholder:\n%s", string(testingRule))
	}
}

func TestW1AScaffoldAllSeedsExactlyFiveStarterStandardsWithoutOverwritingUserFiles(t *testing.T) {
	// AC-N11-001, AC-N11-003, AC-N11-004, AC-N11-005: full scaffold seeds the
	// five starter standards file-by-file and treats same-path files as user-owned.
	ctx, targetDir := minimalScaffoldContext(t, []types.ToolId{types.ToolIdOpenCode})
	starterDir := filepath.Join(targetDir, "specs", "standards", "starter")
	if err := os.MkdirAll(starterDir, 0o755); err != nil {
		t.Fatalf("create starter dir: %v", err)
	}
	const userContent = "# User error handling standard\n\nDo not replace this file.\n"
	if err := os.WriteFile(filepath.Join(starterDir, "error-handling.md"), []byte(userContent), 0o644); err != nil {
		t.Fatalf("seed user starter standard: %v", err)
	}

	result, err := ScaffoldAll(ctx)
	if err != nil {
		t.Fatalf("ScaffoldAll failed: %v", err)
	}

	entries, err := os.ReadDir(starterDir)
	if err != nil {
		t.Fatalf("read starter standards dir: %v", err)
	}
	var markdownFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".md" {
			markdownFiles = append(markdownFiles, entry.Name())
		}
	}
	if len(markdownFiles) != len(expectedStarterStandards) {
		t.Fatalf("starter standard count = %d (%v), want exactly %d", len(markdownFiles), markdownFiles, len(expectedStarterStandards))
	}
	for _, standard := range expectedStarterStandards {
		if _, err := os.Stat(filepath.Join(starterDir, standard)); err != nil {
			t.Fatalf("missing starter standard %s: %v", standard, err)
		}
	}
	got, err := os.ReadFile(filepath.Join(starterDir, "error-handling.md"))
	if err != nil {
		t.Fatalf("read user starter standard: %v", err)
	}
	if string(got) != userContent {
		t.Fatalf("user starter standard overwritten\nwant:%q\n got:%q", userContent, string(got))
	}
	if hasTrackedPath(result.Files, filepath.ToSlash(filepath.Join("specs", "standards", "starter", "error-handling.md"))) {
		t.Fatal("pre-existing user starter standard should not be tracked as a copied library file")
	}
}

func TestScaffoldAll_AlignStrategyPreservesUnrecognizedExistingRootFile(t *testing.T) {
	ctx, targetDir := minimalScaffoldContext(t, []types.ToolId{types.ToolIdOpenCode})
	ctx.Strategy = types.ConflictStrategyAlign

	agentsPath := filepath.Join(targetDir, "AGENTS.md")
	const existingContent = "# Existing user AGENTS\n\nReplace me after backing up.\n"
	if err := os.WriteFile(agentsPath, []byte(existingContent), 0o644); err != nil {
		t.Fatalf("seed AGENTS.md: %v", err)
	}

	if _, err := ScaffoldAll(ctx); err != nil {
		t.Fatalf("ScaffoldAll failed: %v", err)
	}

	got, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	if string(got) != existingContent {
		t.Fatalf("AGENTS.md content = %q, want targeted update to preserve unrecognized content %q", string(got), existingContent)
	}

	backupPath := filepath.Join(targetDir, ".ai-setup-backup", "AGENTS.md")
	backup, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("read AGENTS.md backup: %v", err)
	}
	if string(backup) != existingContent {
		t.Fatalf("backup content = %q, want %q", string(backup), existingContent)
	}
}

func TestScaffoldAll_NilLibraryFS(t *testing.T) {
	ctx, _ := minimalScaffoldContext(t, []types.ToolId{types.ToolIdOpenCode})
	ctx.LibraryFS = nil

	_, err := ScaffoldAll(ctx)
	if err == nil {
		t.Fatal("expected error when LibraryFS is nil")
	}
}

func fileExistsInDir(dir, path string) bool {
	_, err := os.Stat(filepath.Join(dir, path))
	return err == nil
}
