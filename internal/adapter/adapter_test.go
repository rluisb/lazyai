package adapter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/jsonc"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// createTestFS creates a memo FS with the minimum files needed for adapter tests.
func createTestFS() fstest.MapFS {
	return fstest.MapFS{
		"agents/builder.md": &fstest.MapFile{
			Data: []byte("---\nname: Builder\nmodel: sonnet\n---\n\n# Builder\n\nYou are a builder."),
		},
		"agents/orchestrator.md": &fstest.MapFile{
			Data: []byte("---\nname: Orchestrator\nmodel: opus\ntools: list_catalog compose_agent start_chain\n---\n\n# Orchestrator\n\nYou coordinate agents."),
		},
		"skills/implement.md": &fstest.MapFile{
			Data: []byte("---\nname: implement\ndescription: Implementation skill\n---\n\n# Implement\n\nImplement features."),
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
			Skills: []types.SkillId{"implement"},
		},
	}, targetDir
}

// --- Test: CopyWithRecord with fs.FS ---

func TestCopyWithRecord_FromFS(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)
	dest := filepath.Join(targetDir, ".opencode", "agents", "builder.md")

	err := CopyWithRecord("agents/builder.md", dest, ctx, true, nil)
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

	err := CopyWithRecord("agents/builder.md", dest, ctx, true, transform)
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
	err := CopyWithRecord("agents/builder.md", dest, ctx, true, nil)
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

	err := CopyWithRecord("agents/builder.md", dest, ctx, true, nil)
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

// --- Test: readOrchestratorAgentSource from fs.FS ---

func TestReadOrchestratorAgentSource_FromFS(t *testing.T) {
	ctx, _ := createTestAdapterContext(t)
	result := readOrchestratorAgentSource(ctx)

	if result == "" {
		t.Fatal("readOrchestratorAgentSource returned empty string")
	}
	if len(result) < 20 {
		t.Errorf("orchestrator agent source seems too short: %q", result[:min(50, len(result))])
	}
}

// --- Test: readOrchestratorAgentSource fallback ---

func TestReadOrchestratorAgentSource_Fallback(t *testing.T) {
	ctx := &AdapterContext{
		LibraryFS: nil, // No FS
	}
	result := readOrchestratorAgentSource(ctx)

	if result == "" {
		t.Fatal("fallback should return hardcoded content")
	}
	// Verify fallback content includes key markers.
	if len(result) < 10 {
		t.Errorf("fallback content seems too short")
	}
}

// --- Test: OpenCode adapter with fs.FS ---

func TestOpenCodeAdapter_Install_FromFS(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)
	ctx.Selections = AdapterSelections{
		Agents: types.ALL_AGENTS[:],
		Skills: types.ALL_SKILLS[:],
	}
	ctx.EnableServers = []string{"orchestrator"}

	adapter := &OpenCodeAdapter{}
	records, err := adapter.Install(ctx)
	if err != nil {
		t.Fatalf("OpenCode Install failed: %v", err)
	}

	if len(records) == 0 {
		t.Fatal("expected at least one tracked file record")
	}

	// Check that key files were created. The root AGENTS.md at .opencode/
	// is the target of the `instructions: ["AGENTS.md"]` key and must exist
	// at every scope (see TestOpenCodeAdapter_InstructionsKeyResolves).
	keyFiles := []string{
		".opencode/opencode.jsonc",
		".opencode/agents/builder.md",
		".opencode/agents/orchestrator.md",
		".opencode/skills/implement/SKILL.md",
		".opencode/agents/AGENTS.md",
		".opencode/AGENTS.md",
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

	// --- Skill transformed with mode: agent frontmatter ---
	skillPromptFile := filepath.Join(promptsDir, "implement.prompt.md")
	if _, err := os.Stat(skillPromptFile); os.IsNotExist(err) {
		t.Error("skill prompt .prompt.md was not created")
	}
	data, _ := os.ReadFile(skillPromptFile)
	content := string(data)
	if !strings.Contains(content, "mode: agent") {
		t.Error("skill prompt missing mode: agent frontmatter")
	}

	// Root AGENTS.md and .github/copilot-instructions.md are emitted by
	// scaffold.ScaffoldCompiledRoot (scope-aware) rather than the adapter;
	// asserting them here would test the wrong layer.

	// --- Tracked file records created (prompts only) ---
	if len(ctx.FileRecords) < 2 {
		t.Errorf("expected at least 2 tracked file records, got %d", len(ctx.FileRecords))
	}
	hasPreFlight := false
	hasImplement := false
	for _, rec := range ctx.FileRecords {
		switch rec.Path {
		case ".github/prompts/preflight-task-framing.prompt.md":
			hasPreFlight = true
		case ".github/prompts/implement.prompt.md":
			hasImplement = true
		}
	}
	if !hasPreFlight {
		t.Error("no tracked record for preflight-task-framing.prompt.md")
	}
	if !hasImplement {
		t.Error("no tracked record for implement.prompt.md")
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
			Skills: []types.SkillId{"implement"},
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
			Skills: []types.SkillId{"implement"},
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
		Skills: []types.SkillId{"implement"},
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

// --- Test: instructions key resolves to a real file at every scope ---
//
// opencode resolves entries in opencode.jsonc's `instructions` array
// relative to the config file's directory. So `instructions: ["AGENTS.md"]`
// in `<root>/opencode.jsonc` must point at `<root>/AGENTS.md` — and that
// file must exist after install for each of project/workspace/global.

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
			rootFn: func(_ , homeDir string) string {
				return filepath.Join(homeDir, ".config", "opencode")
			},
		},
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
					Skills: []types.SkillId{"implement"},
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
