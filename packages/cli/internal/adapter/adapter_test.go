package adapter

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/frontmatter"
	"github.com/rluisb/lazyai/packages/cli/internal/jsonc"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// createTestFS creates a memo FS with the minimum files needed for adapter tests.
func createTestFS() fstest.MapFS {
	return fstest.MapFS{
		"agents/builder.md": &fstest.MapFile{
			Data: []byte("---\nname: Builder\ndescription: Test builder agent.\nmodel: sonnet\n---\n\n# Builder\n\nYou are a builder."),
		},
		"canonical/agents/primary-agent.md": &fstest.MapFile{
			Data: []byte("# Primary Agent\n\nDefault LazyAI runtime entry point."),
		},
		"canonical/agents/builder.md": &fstest.MapFile{
			Data: canonicalAgentFixture("builder", "Test builder agent."),
		},
		"canonical/agents/planner.md": &fstest.MapFile{
			Data: canonicalAgentFixture("planner", "Test planner agent."),
		},
		"canonical/agents/reviewer.md": &fstest.MapFile{
			Data: canonicalAgentFixture("reviewer", "Test reviewer agent."),
		},
		"canonical/agents/scout.md": &fstest.MapFile{
			Data: canonicalAgentFixture("scout", "Test scout agent."),
		},
		"agents/orchestrator.md": &fstest.MapFile{
			Data: []byte("---\nname: Orchestrator\ndescription: Test orchestrator agent.\nmodel: opus\ntools: list_catalog compose_agent start_chain\n---\n\n# Orchestrator\n\nYou coordinate agents."),
		},
		"skills/implement.md": &fstest.MapFile{
			Data: []byte("---\nname: implement\ndescription: Implementation skill\n---\n\n# Implement\n\nImplement features."),
		},
		"canonical/skills/codebase-exploration.md": &fstest.MapFile{
			Data: canonicalSkillFixture("codebase-exploration", "Explore code paths."),
		},
		"canonical/skills/test-first-change.md": &fstest.MapFile{
			Data: canonicalSkillFixture("test-first-change", "Drive changes through tests."),
		},
		"canonical/skills/diagnose.md": &fstest.MapFile{
			Data: canonicalSkillFixture("diagnose", "Diagnose failures."),
		},
		"canonical/skills/pr-review.md": &fstest.MapFile{
			Data: canonicalSkillFixture("pr-review", "Review pull requests."),
		},
		"tool-agents/agents-dir.md": &fstest.MapFile{
			Data: []byte("# Agents Directory\n\nThis directory contains agent definitions."),
		},
		"tool-agents/skills-dir.md": &fstest.MapFile{
			Data: []byte("# Skills Directory\n\nThis directory contains skill definitions."),
		},
		"tool-agents/root-dir.md": &fstest.MapFile{
			Data: []byte("# Root Directory\n\nProject context at root level."),
		},
		"root/AGENTS.template.md": &fstest.MapFile{
			Data: []byte("# AGENTS\n\n{{PROJECT_NAME}} project agents."),
		},
		"root/copilot-instructions.template.md": &fstest.MapFile{
			Data: []byte("# Copilot Instructions\n\nUse these instructions with Copilot."),
		},
		"prompts/preflight-task-framing.md": &fstest.MapFile{
			Data: []byte("---\nname: preflight-task-framing\n---\n\n# Task Framing\n\nFrame tasks before starting."),
		},
		"rules/typescript.md": &fstest.MapFile{
			Data: []byte("---\npaths:\n  - \"src/**/*.ts\"\n---\n\n# TypeScript Rules\n\n- Use strict TypeScript\n- Prefer interfaces over types for objects\n"),
		},
		"commands/rpi.toml": &fstest.MapFile{
			Data: []byte("name = \"rpi\"\ndescription = \"Start RPI\"\nprompt = \"Begin RPI\"\n"),
		},
		"commands/review.toml": &fstest.MapFile{
			Data: []byte("name = \"review\"\ndescription = \"Review work\"\nprompt = \"Do review\"\n"),
		},
		"commands/plan.toml": &fstest.MapFile{
			Data: []byte("name = \"plan\"\ndescription = \"Plan work\"\nprompt = \"Make plan\"\n"),
		},
		"chatmodes/architect.chatmode.md": &fstest.MapFile{
			Data: []byte("---\ndescription: Architect mode\n---\nArchitect instructions."),
		},
		"chatmodes/reviewer.chatmode.md": &fstest.MapFile{
			Data: []byte("---\ndescription: Reviewer mode\n---\nReviewer instructions."),
		},
		"opencode/commands/review.md": &fstest.MapFile{
			Data: []byte("---\ndescription: Review branch\n---\n\nReview body."),
		},
		"opencode/commands/test.md": &fstest.MapFile{
			Data: []byte("---\ndescription: Run tests\n---\n\nTest body."),
		},
		"opencode/commands/commit.md": &fstest.MapFile{
			Data: []byte("---\ndescription: Draft commit\n---\n\nCommit body."),
		},
		"opencode/modes/plan.md": &fstest.MapFile{
			Data: []byte("---\ndescription: Plan mode\ntools:\n  write: false\n  read: true\n---\n\nPlan body."),
		},
		"opencode/modes/audit.md": &fstest.MapFile{
			Data: []byte("---\ndescription: Audit mode\ntools:\n  write: false\n  read: true\n---\n\nAudit body."),
		},
		"claudecode/commands/review.md": &fstest.MapFile{
			Data: []byte("---\ndescription: Review changes\nargument-hint: \"[pr]\"\nallowed-tools: Bash Read\n---\n\nReview body."),
		},
		"claudecode/commands/test.md": &fstest.MapFile{
			Data: []byte("---\ndescription: Run tests\nargument-hint: \"[target]\"\nallowed-tools: Bash Read\n---\n\nTest body."),
		},
		"claudecode/commands/commit.md": &fstest.MapFile{
			Data: []byte("---\ndescription: Draft commit\nargument-hint: \"\"\nallowed-tools: Bash Read\n---\n\nCommit body."),
		},
		"claudecode/output-styles/terse.md": &fstest.MapFile{
			Data: []byte("---\nname: Terse\ndescription: Short responses\nkeep-coding-instructions: true\n---\n\nTerse style body."),
		},
		"claudecode/output-styles/explanatory.md": &fstest.MapFile{
			Data: []byte("---\nname: Explanatory\ndescription: Detailed responses\nkeep-coding-instructions: true\n---\n\nExplanatory style body."),
		},
	}
}

// createTestAdapterContext creates an AdapterContext for testing with a temp target dir.
func createTestAdapterContext(t *testing.T) (*AdapterContext, string) {
	t.Helper()
	targetDir := t.TempDir()
	libFS := createTestFS()

	return &AdapterContext{
		TargetDir:  targetDir,
		SetupScope: types.SetupScopeProject,
		LibraryDir: "", // empty = production mode, use LibraryFS
		LibraryFS:  libFS,
		Strategy:   types.ConflictStrategyAlign,
		Selections: AdapterSelections{
			Agents: []types.AgentId{"builder"},
			Skills: []types.SkillId{types.SkillIdDiagnose},
		},
	}, targetDir
}

// --- Test: CopyWithRecord with fs.FS ---

func TestCopyWithRecord_FromFS(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)
	dest := filepath.Join(targetDir, ".opencode", "agents", "builder.md")

	err := CopyWithRecord("agents/builder.md", dest, ctx, true, nil, 0o644)
	if err != nil {
		t.Fatalf("CopyWithRecord failed: %v", err)
	}

	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("failed to read destination file: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("destination file is empty")
	}

	// Check that the content matches what's in the FS.
	expected, _ := ctx.LibraryFS.Open("agents/builder.md")
	expectedData := make([]byte, len(data))
	n, _ := expected.Read(expectedData)
	expectedData = expectedData[:n]

	if string(data) != string(expectedData) {
		t.Errorf("content mismatch:\ngot: %q\nwant: %q", string(data), string(expectedData))
	}

	// Check tracked file.
	if len(ctx.FileRecords) != 1 {
		t.Fatalf("expected 1 file record, got %d", len(ctx.FileRecords))
	}
	if ctx.FileRecords[0].Source != "agents/builder.md" {
		t.Errorf("expected source 'agents/builder.md', got %q", ctx.FileRecords[0].Source)
	}
}

// --- Test: CopyWithRecord with transform ---

func TestCopyWithRecord_WithTransform(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)
	dest := filepath.Join(targetDir, ".opencode", "agents", "builder.md")

	transformCalled := false
	transform := func(content []byte) []byte {
		transformCalled = true
		return append([]byte("<!-- transformed -->\n"), content...)
	}

	err := CopyWithRecord("agents/builder.md", dest, ctx, true, transform, 0o644)
	if err != nil {
		t.Fatalf("CopyWithRecord with transform failed: %v", err)
	}

	if !transformCalled {
		t.Fatal("transform function was not called")
	}

	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("failed to read destination file: %v", err)
	}

	if string(data[:21]) != "<!-- transformed -->\n" {
		t.Errorf("expected transform prefix, got %q", string(data[:min(30, len(data))]))
	}
}

// --- Test: CopyLibraryDirectory with fs.FS ---

func TestCopyLibraryDirectory_FromFS(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)
	agentsDir := filepath.Join(targetDir, ".opencode", "agents")
	_ = files.EnsureDir(agentsDir)

	err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "agents",
		SelectionKey: "agents",
		ToDestPath: func(file string) string {
			return filepath.Join(agentsDir, file)
		},
		WarnOnSkip: true,
	})
	if err != nil {
		t.Fatalf("CopyLibraryDirectory failed: %v", err)
	}

	// builder.md should exist (it's in our selection).
	builderDest := filepath.Join(agentsDir, "builder.md")
	if _, err := os.Stat(builderDest); os.IsNotExist(err) {
		t.Error("builder.md was not copied")
	}

	// orchestrator.md should NOT exist (not in our selection).
	orchDest := filepath.Join(agentsDir, "orchestrator.md")
	if _, err := os.Stat(orchDest); !os.IsNotExist(err) {
		t.Error("orchestrator.md should not have been copied (not selected)")
	}
}

// --- Test: CopyLibraryDirectory - install all when no selection ---

func TestCopyLibraryDirectory_InstallAll(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)
	// Empty selections = install everything.
	ctx.Selections = AdapterSelections{}
	agentsDir := filepath.Join(targetDir, ".opencode", "agents")
	_ = files.EnsureDir(agentsDir)

	err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "agents",
		SelectionKey: "agents",
		ToDestPath: func(file string) string {
			return filepath.Join(agentsDir, file)
		},
		WarnOnSkip: true,
	})
	if err != nil {
		t.Fatalf("CopyLibraryDirectory failed: %v", err)
	}

	// Both builder.md and orchestrator.md should exist.
	for _, name := range []string{"builder.md", "orchestrator.md"} {
		path := filepath.Join(agentsDir, name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("%s was not copied (should install all)", name)
		}
	}
}

// --- Test: InstallToolContextFiles with fs.FS ---

func TestInstallToolContextFiles_FromFS(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)
	toolDir := filepath.Join(targetDir, ".opencode")
	_ = files.EnsureDir(toolDir)
	_ = files.EnsureDir(filepath.Join(toolDir, "agents"))

	err := InstallToolContextFiles(InstallToolContextFilesOption{
		Ctx:             ctx,
		ToolDir:         toolDir,
		ContextFileName: "AGENTS.md",
		AgentsDestDir:   "agents",
		SkillsDestDir:   "skills",
	})
	if err != nil {
		t.Fatalf("InstallToolContextFiles failed: %v", err)
	}

	// Check that tool-agents files were copied.
	for _, file := range []string{"agents/AGENTS.md", "skills/AGENTS.md", "AGENTS.md"} {
		path := filepath.Join(toolDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("%s was not created", file)
		}
	}
}

// --- Test: InstallRootTemplateIfMissing with fs.FS ---

func TestInstallRootTemplateIfMissing_FromFS(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)

	err := InstallRootTemplateIfMissing(ctx, "AGENTS.md",
		filepath.Join(targetDir, "AGENTS.md"),
		"root/AGENTS.template.md")
	if err != nil {
		t.Fatalf("InstallRootTemplateIfMissing failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(targetDir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("failed to read AGENTS.md: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("AGENTS.md is empty")
	}
}

// --- Test: CopyWithRecord skip when file exists (skip strategy) ---

func TestCopyWithRecord_SkipExisting(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)
	ctx.Strategy = types.ConflictStrategySkip
	dest := filepath.Join(targetDir, "builder.md")

	// Write an existing file.
	existingContent := []byte("existing content")
	_ = files.EnsureDir(filepath.Dir(dest))
	_ = os.WriteFile(dest, existingContent, 0o644)

	// CopyWithRecord with skip strategy should skip.
	err := CopyWithRecord("agents/builder.md", dest, ctx, true, nil, 0o644)
	if err != nil {
		t.Fatalf("CopyWithRecord failed: %v", err)
	}

	// Existing content should be unchanged.
	data, _ := os.ReadFile(dest)
	if string(data) != "existing content" {
		t.Error("existing file was overwritten when it should have been skipped")
	}
}

// --- Test: CopyWithRecord force overwrite with backup ---

func TestCopyWithRecord_ForceOverwrite(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)
	ctx.Force = true
	ctx.Strategy = types.ConflictStrategyBackupAndReplace
	dest := filepath.Join(targetDir, "builder.md")

	// Write an existing file.
	existingContent := []byte("existing content")
	_ = files.EnsureDir(filepath.Dir(dest))
	_ = os.WriteFile(dest, existingContent, 0o644)

	err := CopyWithRecord("agents/builder.md", dest, ctx, true, nil, 0o644)
	if err != nil {
		t.Fatalf("CopyWithRecord force failed: %v", err)
	}

	// Content should be overwritten.
	data, _ := os.ReadFile(dest)
	if string(data) == "existing content" {
		t.Error("file was not overwritten with force")
	}

	// Check a backup was created (conflict package puts backups in .ai-setup-backup/).
	backupDir := filepath.Join(targetDir, ".ai-setup-backup")
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		t.Errorf("no backup directory was created: %v", err)
	} else if len(entries) == 0 {
		t.Error("backup directory is empty")
	}
}

// --- Test: OpenCode adapter with fs.FS ---

func TestOpenCodeAdapter_Install_FromFS(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)
	ctx.Selections = AdapterSelections{
		Agents: types.ALL_AGENTS[:],
		Skills: types.ALL_SKILLS[:],
	}

	adapter := &OpenCodeAdapter{}
	records, err := adapter.Install(ctx)
	if err != nil {
		t.Fatalf("OpenCode Install failed: %v", err)
	}

	if len(records) == 0 {
		t.Fatal("expected at least one tracked file record")
	}

	// Check that key files were created. Adapter install preserves actual agent
	// and skill definitions, but must not create reserved context docs inside the
	// tool directory; root context docs are handled elsewhere.
	keyFiles := []string{
		".opencode/opencode.jsonc",
		".opencode/agents/primary-agent.md",
		".opencode/agents/builder.md",
		".opencode/skills/diagnose/SKILL.md",
	}
	// The pre-unification .json variant must never be produced on a fresh install.
	if _, err := os.Stat(filepath.Join(targetDir, ".opencode", "opencode.json")); err == nil {
		t.Error("opencode.json should not exist; install must target opencode.jsonc only")
	}
	for _, f := range keyFiles {
		path := filepath.Join(targetDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s was not created", f)
		}
	}
	if _, err := os.Stat(filepath.Join(targetDir, ".opencode", "agents", "orchestrator.md")); err == nil {
		t.Error("orchestrator.md should not be installed")
	}
	for _, f := range []string{
		".opencode/AGENTS.md",
		".opencode/agents/AGENTS.md",
		".opencode/skills/AGENTS.md",
	} {
		if _, err := os.Stat(filepath.Join(targetDir, f)); err == nil {
			t.Errorf("reserved context doc %s should not be created by adapter install", f)
		}
	}

	// Every installed agent file must carry opencode-schema-valid
	// frontmatter: at minimum `description` and `mode`. This closes the gap
	// where the old shared stripper emitted HTML comments that opencode
	// could not parse.
	agentsDir := filepath.Join(targetDir, ".opencode", "agents")
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		t.Fatalf("read agents dir: %v", err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") || e.Name() == "AGENTS.md" {
			continue
		}
		data, _ := os.ReadFile(filepath.Join(agentsDir, e.Name()))
		fm, _, err := frontmatter.ExtractFrontmatter(data)
		if err != nil {
			t.Errorf("%s: frontmatter does not parse: %v", e.Name(), err)
			continue
		}
		if fm["description"] == nil || fm["description"] == "" {
			t.Errorf("%s: missing description key", e.Name())
		}
		if fm["mode"] == nil || fm["mode"] == "" {
			t.Errorf("%s: missing mode key", e.Name())
		}
	}
}

// --- Test: Copilot adapter with fs.FS ---

func TestCopilotAdapter_Install_FromFS(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)
	ctx.Selections = AdapterSelections{
		Agents:  types.ALL_AGENTS[:],
		Skills:  types.ALL_SKILLS[:],
		Prompts: []types.PromptId{"preflight-task-framing"},
	}

	adapter := &CopilotAdapter{}
	records, err := adapter.Install(ctx)
	if err != nil {
		t.Fatalf("Copilot Install failed: %v", err)
	}

	if len(records) == 0 {
		t.Fatal("expected at least one tracked file record")
	}

	// --- Prompts copied to .github/prompts/*.prompt.md ---
	promptsDir := filepath.Join(targetDir, ".github", "prompts")
	promptFile := filepath.Join(promptsDir, "preflight-task-framing.prompt.md")
	if _, err := os.Stat(promptFile); os.IsNotExist(err) {
		t.Error("prompt .prompt.md was not created in .github/prompts/")
	}

	// --- Skill transformed to agent YAML format ---
	agentsDir := filepath.Join(targetDir, ".github", "agents")
	skillAgentFile := filepath.Join(agentsDir, "diagnose.agent.yaml")
	if _, err := os.Stat(skillAgentFile); os.IsNotExist(err) {
		t.Error("skill agent .agent.yaml was not created")
	}
	data, _ := os.ReadFile(skillAgentFile)
	content := string(data)
	if !strings.Contains(content, "name: diagnose") {
		t.Error("skill agent missing 'name: diagnose' in YAML")
	}

	// Root AGENTS.md and .github/copilot-instructions.md are emitted by
	// scaffold.ScaffoldCompiledRoot (scope-aware) rather than the adapter;
	// asserting them here would test the wrong layer.

	// --- Tracked file records created (prompts + agents) ---
	if len(ctx.FileRecords) < 2 {
		t.Errorf("expected at least 2 tracked file records, got %d", len(ctx.FileRecords))
	}
	hasPreFlight := false
	hasDiagnose := false
	for _, rec := range ctx.FileRecords {
		switch rec.Path {
		case ".github/prompts/preflight-task-framing.prompt.md":
			hasPreFlight = true
		case ".github/agents/diagnose.agent.yaml":
			hasDiagnose = true
		}
	}
	if !hasPreFlight {
		t.Error("no tracked record for preflight-task-framing.prompt.md")
	}
	if !hasDiagnose {
		t.Error("no tracked record for diagnose.agent.yaml")
	}
}

// --- Test: Disk fallback mode ---

func TestCopyLibraryDirectory_DiskFallback(t *testing.T) {
	// Create a temp library directory on disk.
	libDir := t.TempDir()
	agentsDir := filepath.Join(libDir, "agents")
	_ = os.MkdirAll(agentsDir, 0o755)
	_ = os.WriteFile(filepath.Join(agentsDir, "builder.md"), []byte("---\nname: Builder\n---\n\nBuilder content."), 0o644)

	targetDir := t.TempDir()
	destAgentsDir := filepath.Join(targetDir, ".opencode", "agents")
	_ = files.EnsureDir(destAgentsDir)

	ctx := &AdapterContext{
		TargetDir:  targetDir,
		LibraryDir: libDir,
		LibraryFS:  nil, // nil = disk mode
		Strategy:   types.ConflictStrategyAlign,
		Selections: AdapterSelections{
			Agents: []types.AgentId{"builder"},
		},
	}

	err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "agents",
		SelectionKey: "agents",
		ToDestPath: func(file string) string {
			return filepath.Join(destAgentsDir, file)
		},
		WarnOnSkip: true,
	})
	if err != nil {
		t.Fatalf("CopyLibraryDirectory disk fallback failed: %v", err)
	}

	destFile := filepath.Join(destAgentsDir, "builder.md")
	if _, err := os.Stat(destFile); os.IsNotExist(err) {
		t.Error("builder.md was not copied from disk")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// --- Test: OpenCode adapter global scope path resolution ---

func TestOpenCodeAdapter_GlobalScope_UsesGlobalPath(t *testing.T) {
	targetDir := t.TempDir()
	libFS := createTestFS()

	// Create a temp dir that mimics ~/.config/opencode structure.
	homeDir := t.TempDir()
	expectedDir := filepath.Join(homeDir, ".config", "opencode")

	ctx := &AdapterContext{
		TargetDir:  targetDir,
		SetupScope: types.SetupScopeGlobal,
		HomeDir:    homeDir,
		LibraryDir: "",
		LibraryFS:  libFS,
		Strategy:   types.ConflictStrategyAlign,
		Selections: AdapterSelections{
			Agents: []types.AgentId{"builder"},
			Skills: []types.SkillId{types.SkillIdDiagnose},
		},
	}

	adapter := &OpenCodeAdapter{}
	records, err := adapter.Install(ctx)
	if err != nil {
		t.Fatalf("OpenCode Install in global scope failed: %v", err)
	}

	if len(records) == 0 {
		t.Fatal("expected at least one tracked file record")
	}

	// Verify the expected directory structure was created in the global dir.
	keyDirs := []string{
		filepath.Join(expectedDir, "agents"),
		filepath.Join(expectedDir, "skills"),
		filepath.Join(expectedDir, "commands"),
	}
	for _, d := range keyDirs {
		if _, err := os.Stat(d); os.IsNotExist(err) {
			t.Errorf("expected directory %s was not created", d)
		}
	}

	// Verify no .opencode directory was created in the project target dir.
	projectOpencode := filepath.Join(targetDir, ".opencode")
	if _, err := os.Stat(projectOpencode); !os.IsNotExist(err) {
		t.Error("global scope should not create .opencode in project target dir")
	}
}

func TestOpenCodeAdapter_GlobalScope_FallbackHomeDir(t *testing.T) {
	targetDir := t.TempDir()
	libFS := createTestFS()

	ctx := &AdapterContext{
		TargetDir:  targetDir,
		SetupScope: types.SetupScopeGlobal,
		HomeDir:    "", // empty — should fall back to os.UserHomeDir()
		LibraryDir: "",
		LibraryFS:  libFS,
		Strategy:   types.ConflictStrategyAlign,
		Selections: AdapterSelections{
			Agents: []types.AgentId{"builder"},
			Skills: []types.SkillId{types.SkillIdDiagnose},
		},
	}

	adapter := &OpenCodeAdapter{}
	_, err := adapter.Install(ctx)
	if err != nil {
		t.Fatalf("OpenCode Install with empty HomeDir failed: %v", err)
	}

	// Verify files were written to the real home directory.
	realHome, _ := os.UserHomeDir()
	expectedDir := filepath.Join(realHome, ".config", "opencode")
	if _, err := os.Stat(expectedDir); os.IsNotExist(err) {
		t.Errorf("expected global dir %s was not created", expectedDir)
	}
}

// --- Test: OpenCode adapter migrates pre-existing opencode.json to .jsonc ---

func TestOpenCodeAdapter_Install_MigratesJsonToJsonc(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)
	ctx.Selections = AdapterSelections{
		Agents: []types.AgentId{"builder"},
		Skills: []types.SkillId{types.SkillIdDiagnose},
	}

	// Seed a pre-existing opencode.json with a user-authored key that must
	// survive the migration unchanged.
	ocDir := filepath.Join(targetDir, ".opencode")
	if err := files.EnsureDir(ocDir); err != nil {
		t.Fatalf("EnsureDir: %v", err)
	}
	jsonPath := filepath.Join(ocDir, "opencode.json")
	jsoncPath := filepath.Join(ocDir, "opencode.jsonc")
	original := []byte(`{
  "$schema": "https://opencode.ai/config.json",
  "permission": { "edit": "allow" },
  "user_key": "preserved"
}
`)
	if err := os.WriteFile(jsonPath, original, 0o644); err != nil {
		t.Fatalf("seed opencode.json: %v", err)
	}

	adapter := &OpenCodeAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("OpenCode Install failed: %v", err)
	}

	// .jsonc must exist; .json must be gone; .json.bak must preserve original.
	if _, err := os.Stat(jsoncPath); os.IsNotExist(err) {
		t.Fatal("opencode.jsonc was not created by migration")
	}
	if _, err := os.Stat(jsonPath); err == nil {
		t.Error("opencode.json should have been removed after migration")
	}
	bakPath := jsonPath + ".bak"
	bakContents, err := os.ReadFile(bakPath)
	if err != nil {
		t.Fatalf("opencode.json.bak was not created: %v", err)
	}
	if string(bakContents) != string(original) {
		t.Errorf(".bak sidecar content mismatch.\nwant: %q\ngot:  %q", original, bakContents)
	}

	// The migrated .jsonc must carry the user-authored key forward verbatim
	// (no default-config merge on top of the migrated file — that would
	// silently clobber customizations like permission.edit == "allow").
	jsoncContents, err := os.ReadFile(jsoncPath)
	if err != nil {
		t.Fatalf("read migrated .jsonc: %v", err)
	}
	if !strings.Contains(string(jsoncContents), `"user_key": "preserved"`) {
		t.Errorf("migrated .jsonc dropped user_key:\n%s", jsoncContents)
	}
	if !strings.Contains(string(jsoncContents), `"edit": "allow"`) {
		t.Errorf("migrated .jsonc did not preserve user-authored permission.edit:\n%s", jsoncContents)
	}
}

// --- Test: opencode commands + modes install at every scope ---

func TestOpenCodeAdapter_InstallsCommandsAndModes(t *testing.T) {
	type scopeCase struct {
		name   string
		scope  types.SetupScope
		rootFn func(targetDir, homeDir string) string
	}
	cases := []scopeCase{
		{"project", types.SetupScopeProject, func(t, _ string) string { return filepath.Join(t, ".opencode") }},
		{"workspace", types.SetupScopeWorkspace, func(t, _ string) string { return filepath.Join(t, ".opencode") }},
		{"global", types.SetupScopeGlobal, func(_, h string) string { return filepath.Join(h, ".config", "opencode") }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			targetDir := t.TempDir()
			homeDir := t.TempDir()
			ctx := &AdapterContext{
				TargetDir:  targetDir,
				SetupScope: tc.scope,
				HomeDir:    homeDir,
				LibraryFS:  createTestFS(),
				Strategy:   types.ConflictStrategyAlign,
				Selections: AdapterSelections{
					Agents: []types.AgentId{"builder"},
					Skills: []types.SkillId{types.SkillIdDiagnose},
					// Leaving OpenCodeCommands / OpenCodeModes unset means
					// "install all" — the wizard will populate these later.
				},
			}

			if _, err := (&OpenCodeAdapter{}).Install(ctx); err != nil {
				t.Fatalf("Install (%s): %v", tc.name, err)
			}

			root := tc.rootFn(targetDir, homeDir)
			for _, want := range []string{
				"commands/review.md",
				"commands/test.md",
				"commands/commit.md",
				"modes/plan.md",
				"modes/audit.md",
			} {
				path := filepath.Join(root, want)
				if _, err := os.Stat(path); os.IsNotExist(err) {
					t.Errorf("%s: missing %s after install", tc.name, want)
				}
			}
		})
	}
}

// TestOpenCodeAdapter_SelectionFiltersCommandsAndModes verifies that
// ctx.Selections.OpenCodeCommands narrows the install set. An explicit
// selection of ["review"] must leave test.md and commit.md uninstalled.
func TestOpenCodeAdapter_SelectionFiltersCommandsAndModes(t *testing.T) {
	targetDir := t.TempDir()
	ctx := &AdapterContext{
		TargetDir:  targetDir,
		SetupScope: types.SetupScopeProject,
		LibraryFS:  createTestFS(),
		Strategy:   types.ConflictStrategyAlign,
		Selections: AdapterSelections{
			Agents:           []types.AgentId{"builder"},
			Skills:           []types.SkillId{types.SkillIdDiagnose},
			OpenCodeCommands: []types.OpenCodeCommandId{types.OpenCodeCommandIdReview},
			OpenCodeModes:    []types.OpenCodeModeId{types.OpenCodeModeIdPlan},
		},
	}
	if _, err := (&OpenCodeAdapter{}).Install(ctx); err != nil {
		t.Fatalf("Install: %v", err)
	}

	root := filepath.Join(targetDir, ".opencode")
	mustExist := []string{"commands/review.md", "modes/plan.md"}
	mustNotExist := []string{"commands/test.md", "commands/commit.md", "modes/audit.md"}
	for _, p := range mustExist {
		if _, err := os.Stat(filepath.Join(root, p)); os.IsNotExist(err) {
			t.Errorf("selection did not include %s: missing", p)
		}
	}
	for _, p := range mustNotExist {
		if _, err := os.Stat(filepath.Join(root, p)); err == nil {
			t.Errorf("selection leaked: %s should not exist", p)
		}
	}
}

// --- Test: instructions key resolves to a real file at every scope ---
//
// opencode resolves entries in opencode.jsonc's `instructions` array
// relative to the config file's directory. Project/workspace configs live in
// `.opencode/`, so they must point up to the root AGENTS.md. Global configs
// live alongside the global AGENTS.md and can use `AGENTS.md` directly.

func TestOpenCodeAdapter_InstructionsKeyResolves(t *testing.T) {
	type scopeCase struct {
		name  string
		scope types.SetupScope
		// rootFn returns the expected .opencode/ root given a fresh (targetDir, homeDir) pair.
		rootFn func(targetDir, homeDir string) string
	}

	cases := []scopeCase{
		{
			name:  "project",
			scope: types.SetupScopeProject,
			rootFn: func(targetDir, _ string) string {
				return filepath.Join(targetDir, ".opencode")
			},
		},
		{
			name:  "workspace",
			scope: types.SetupScopeWorkspace,
			rootFn: func(targetDir, _ string) string {
				return filepath.Join(targetDir, ".opencode")
			},
		},
		{
			name:  "global",
			scope: types.SetupScopeGlobal,
			rootFn: func(_, homeDir string) string {
				return filepath.Join(homeDir, ".config", "opencode")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			targetDir := t.TempDir()
			homeDir := t.TempDir()
			rootDoc := filepath.Join(targetDir, "AGENTS.md")
			if tc.scope == types.SetupScopeGlobal {
				rootDoc = filepath.Join(homeDir, ".config", "opencode", "AGENTS.md")
				if err := os.MkdirAll(filepath.Dir(rootDoc), 0o755); err != nil {
					t.Fatalf("create global root doc dir: %v", err)
				}
			}
			if err := os.WriteFile(rootDoc, []byte("# Root instructions\n"), 0o644); err != nil {
				t.Fatalf("write root doc: %v", err)
			}
			ctx := &AdapterContext{
				TargetDir:  targetDir,
				SetupScope: tc.scope,
				HomeDir:    homeDir,
				LibraryFS:  createTestFS(),
				Strategy:   types.ConflictStrategyAlign,
				Selections: AdapterSelections{
					Agents: []types.AgentId{"builder"},
					Skills: []types.SkillId{types.SkillIdDiagnose},
				},
			}

			if _, err := (&OpenCodeAdapter{}).Install(ctx); err != nil {
				t.Fatalf("Install (%s): %v", tc.name, err)
			}

			root := tc.rootFn(targetDir, homeDir)

			// Read back opencode.jsonc and assert instructions shape.
			cfgPath := filepath.Join(root, OpenCodeConfigFilename)
			cfg, err := jsonc.ReadJSONCFile(cfgPath)
			if err != nil {
				t.Fatalf("read %s: %v", cfgPath, err)
			}
			rawInstr, ok := cfg["instructions"]
			if !ok {
				t.Fatalf("opencode.jsonc missing `instructions` key at %s", tc.name)
			}
			instr, ok := rawInstr.([]any)
			if !ok || len(instr) == 0 {
				t.Fatalf("instructions must be a non-empty array, got %T: %v", rawInstr, rawInstr)
			}

			// Each instructions entry, resolved relative to the config dir,
			// must point at an existing, non-empty file.
			cfgDir := filepath.Dir(cfgPath)
			for _, raw := range instr {
				rel, ok := raw.(string)
				if !ok {
					t.Errorf("instructions entry is not a string: %T %v", raw, raw)
					continue
				}
				resolved := rel
				if !filepath.IsAbs(rel) {
					resolved = filepath.Join(cfgDir, rel)
				}
				info, err := os.Stat(resolved)
				if err != nil {
					t.Errorf("instructions entry %q resolves to %q which does not exist: %v", rel, resolved, err)
					continue
				}
				if info.Size() == 0 {
					t.Errorf("instructions entry %q resolves to an empty file at %q", rel, resolved)
				}
			}
		})
	}
}

// Tests for the removed isGlobalOpenCodeDir heuristic were deleted when
// CompileMCP became scope-aware via CompileContext. Scope parity is now
// asserted by TestCompileMCPForTool_ScopeParity.

// TestNormalizeToolsFrontmatter_Delimiters verifies that the space and comma
// delimiter options work correctly (spec 012 task 004).
func TestNormalizeToolsFrontmatter_Delimiters(t *testing.T) {
	input := `---
name: Test Agent
tools: Bash, Read, Edit
---

Test content`

	tests := []struct {
		delimiter string
		wantTools string
	}{
		{"space", "tools: Bash Read Edit"},
		{"comma", "tools: Bash, Read, Edit"},
	}

	for _, tt := range tests {
		t.Run(tt.delimiter, func(t *testing.T) {
			got := NormalizeToolsFrontmatter(input, tt.delimiter)
			if !strings.Contains(got, tt.wantTools) {
				t.Errorf("delimiter %q: expected %q to be in output, got:\n%s",
					tt.delimiter, tt.wantTools, got)
			}
		})
	}
}

// TestClaudeCodeOutputStylesFrontmatter verifies that Claude Code output styles have
// required frontmatter fields (spec 012 task 006).
func TestClaudeCodeOutputStylesFrontmatter(t *testing.T) {
	libFS := createTestFS()
	styles := []string{"terse", "explanatory"}

	for _, style := range styles {
		t.Run(style, func(t *testing.T) {
			path := "claudecode/output-styles/" + style + ".md"
			data, err := fs.ReadFile(libFS, path)
			if err != nil {
				t.Fatalf("read output style: %v", err)
			}

			fm, _, err := frontmatter.ExtractFrontmatter(data)
			if err != nil {
				t.Fatalf("parse frontmatter: %v", err)
			}

			// Check required fields
			if _, ok := fm["name"]; !ok {
				t.Error("missing 'name' field")
			}
			if _, ok := fm["description"]; !ok {
				t.Error("missing 'description' field")
			}
			if keepCoding, ok := fm["keep-coding-instructions"]; !ok {
				t.Error("missing 'keep-coding-instructions' field")
			} else if kb, ok := keepCoding.(bool); !ok || !kb {
				t.Errorf("keep-coding-instructions should be true, got: %v", keepCoding)
			}
		})
	}
}

// TestClaudeCodeCommandsFrontmatter verifies that Claude Code commands have
// required frontmatter fields (spec 012 task 005).
func TestClaudeCodeCommandsFrontmatter(t *testing.T) {
	libFS := createTestFS()
	commands := []string{"review", "test", "commit"}

	for _, cmd := range commands {
		t.Run(cmd, func(t *testing.T) {
			path := "claudecode/commands/" + cmd + ".md"
			data, err := fs.ReadFile(libFS, path)
			if err != nil {
				t.Fatalf("read command: %v", err)
			}

			fm, _, err := frontmatter.ExtractFrontmatter(data)
			if err != nil {
				t.Fatalf("parse frontmatter: %v", err)
			}

			// Check required fields
			if _, ok := fm["description"]; !ok {
				t.Error("missing 'description' field")
			}
			if _, ok := fm["allowed-tools"]; !ok {
				t.Error("missing 'allowed-tools' field")
			}
			if _, ok := fm["argument-hint"]; !ok {
				t.Error("missing 'argument-hint' field")
			}

			// Verify allowed-tools is space-separated, not comma-separated
			toolsVal := fm["allowed-tools"]
			if toolsVal != nil {
				toolsStr := fmt.Sprintf("%v", toolsVal)
				if strings.Contains(toolsStr, ",") && !strings.Contains(toolsStr, "Read") {
					// If there's a comma but it's not part of a proper (Bash(...)) format, it's wrong
					t.Errorf("allowed-tools appears comma-separated: %s", toolsStr)
				}
			}
		})
	}
}

// --- Test: OpenCode adapter primary-agent install ---

func testRepoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Dir(filepath.Dir(filepath.Dir(file)))
}

func TestMergeManagedBlock_CreateNew(t *testing.T) {
	managed := []byte("managed content")
	result := MergeManagedBlock(nil, managed, ManagedBlockStartMarker, ManagedBlockEndMarker)
	if !strings.Contains(string(result), ManagedBlockStartMarker) {
		t.Error("missing start marker")
	}
	if !strings.Contains(string(result), ManagedBlockEndMarker) {
		t.Error("missing end marker")
	}
	if !strings.Contains(string(result), "managed content") {
		t.Error("missing managed content")
	}
}

func TestMergeManagedBlock_AppendToExisting(t *testing.T) {
	existing := []byte("# User Content\n")
	managed := []byte("managed content")
	result := MergeManagedBlock(existing, managed, ManagedBlockStartMarker, ManagedBlockEndMarker)
	if !strings.Contains(string(result), "# User Content") {
		t.Error("existing content was lost")
	}
	if !strings.Contains(string(result), ManagedBlockStartMarker) {
		t.Error("missing start marker")
	}
	if !strings.Contains(string(result), ManagedBlockEndMarker) {
		t.Error("missing end marker")
	}
}

func TestMergeManagedBlock_ReplaceExistingBlock(t *testing.T) {
	existing := []byte("# User Content\n\n" + ManagedBlockStartMarker + "\nold content\n" + ManagedBlockEndMarker + "\n\nmore user content")
	managed := []byte("new managed content")
	result := MergeManagedBlock(existing, managed, ManagedBlockStartMarker, ManagedBlockEndMarker)
	if !strings.Contains(string(result), "# User Content") {
		t.Error("pre-block user content was lost")
	}
	if !strings.Contains(string(result), "more user content") {
		t.Error("post-block user content was lost")
	}
	if !strings.Contains(string(result), "new managed content") {
		t.Error("new managed content not inserted")
	}
	if strings.Contains(string(result), "old content") {
		t.Error("old managed content was not replaced")
	}
}

func TestOpenCodeAdapter_Install_PrimaryAgentMode(t *testing.T) {
	targetDir := t.TempDir()
	repoRoot := testRepoRoot(t)
	libFS := os.DirFS(filepath.Join(repoRoot, "library"))

	ctx := &AdapterContext{
		TargetDir:  targetDir,
		SetupScope: types.SetupScopeProject,
		LibraryFS:  libFS,
		Strategy:   types.ConflictStrategyAlign,
		Selections: AdapterSelections{},
	}

	adapter := &OpenCodeAdapter{}
	_, err := adapter.Install(ctx)
	if err != nil {
		t.Fatalf("OpenCode Install failed: %v", err)
	}

	cfgPath := filepath.Join(targetDir, ".opencode", "opencode.jsonc")
	cfg, err := jsonc.ReadJSONCFile(cfgPath)
	if err != nil {
		t.Fatalf("read opencode.jsonc: %v", err)
	}
	if da, ok := cfg["default_agent"].(string); !ok || da != "primary-agent" {
		t.Errorf("default_agent = %q, want primary-agent", da)
	}
	if _, err := os.Stat(filepath.Join(targetDir, ".opencode", "agents", "primary-agent.md")); os.IsNotExist(err) {
		t.Error("primary-agent.md was not installed")
	}
	if _, err := os.Stat(filepath.Join(targetDir, ".opencode", "agents", "orchestrator.md")); err == nil {
		t.Error("orchestrator.md should not be installed")
	}
	if _, err := os.Stat(filepath.Join(targetDir, ".opencode", "STARTUP.md")); err == nil {
		t.Error("STARTUP.md should not be installed")
	}
}

// --- Test: CopyLibraryDirectory recursive nested paths ---

func TestCopyLibraryDirectory_RecursiveNestedPaths(t *testing.T) {
	libFS := fstest.MapFS{
		"skills/nested/skill-a/SKILL.md": &fstest.MapFile{
			Data: []byte("# Skill A\n"),
		},
		"skills/nested/skill-b/SKILL.md": &fstest.MapFile{
			Data: []byte("# Skill B\n"),
		},
		"skills/nested/skill-b/scripts/run.sh": &fstest.MapFile{
			Data: []byte("#!/bin/sh\n"),
		},
	}

	targetDir := t.TempDir()
	destSkillsDir := filepath.Join(targetDir, ".opencode", "skills")
	_ = files.EnsureDir(destSkillsDir)

	ctx := &AdapterContext{
		TargetDir:  targetDir,
		LibraryFS:  libFS,
		Strategy:   types.ConflictStrategyAlign,
		Selections: AdapterSelections{},
	}

	err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "skills",
		SelectionKey: "skills",
		ToDestPath: func(file string) string {
			return filepath.Join(destSkillsDir, file)
		},
		Recursive:  true,
		WarnOnSkip: true,
	})
	if err != nil {
		t.Fatalf("CopyLibraryDirectory recursive failed: %v", err)
	}

	// Assert nested relative paths preserved
	for _, rel := range []string{
		"nested/skill-a/SKILL.md",
		"nested/skill-b/SKILL.md",
		"nested/skill-b/scripts/run.sh",
	} {
		path := filepath.Join(destSkillsDir, rel)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s was not copied", rel)
		}
	}
}

// --- Test: ChmodScriptsExecutable makes .sh executable ---

func TestChmodScriptsExecutable(t *testing.T) {
	dir := t.TempDir()

	scriptPath := filepath.Join(dir, "test.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/sh\n"), 0o644); err != nil {
		t.Fatalf("write script: %v", err)
	}

	otherPath := filepath.Join(dir, "readme.md")
	if err := os.WriteFile(otherPath, []byte("# readme\n"), 0o644); err != nil {
		t.Fatalf("write readme: %v", err)
	}

	if err := ChmodScriptsExecutable(dir); err != nil {
		t.Fatalf("ChmodScriptsExecutable failed: %v", err)
	}

	scriptInfo, err := os.Stat(scriptPath)
	if err != nil {
		t.Fatalf("stat script: %v", err)
	}
	if scriptInfo.Mode().Perm()&0o111 == 0 {
		t.Errorf("script %s should be executable, got %o", scriptPath, scriptInfo.Mode().Perm())
	}

	otherInfo, err := os.Stat(otherPath)
	if err != nil {
		t.Fatalf("stat readme: %v", err)
	}
	if otherInfo.Mode().Perm() != 0o644 {
		t.Errorf("readme %s should remain 0644, got %o", otherPath, otherInfo.Mode().Perm())
	}
}

func sliceContains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
