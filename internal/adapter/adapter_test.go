package adapter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/ricardoborges-teachable/ai-setup/internal/files"
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

	// Check that key files were created.
	keyFiles := []string{
		".opencode/opencode.json",
		".opencode/agents/builder.md",
		".opencode/agents/orchestrator.md",
		".opencode/skills/implement/SKILL.md",
		".opencode/agents/AGENTS.md",
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

	// --- Root template AGENTS.md installed ---
	agentsFile := filepath.Join(targetDir, "AGENTS.md")
	if _, err := os.Stat(agentsFile); os.IsNotExist(err) {
		t.Error("AGENTS.md root template was not created")
	}

	// --- Root template .github/copilot-instructions.md installed ---
	copilotInstructionsFile := filepath.Join(targetDir, ".github", "copilot-instructions.md")
	if _, err := os.Stat(copilotInstructionsFile); os.IsNotExist(err) {
		t.Error(".github/copilot-instructions.md was not created")
	}

	// --- Tracked file records created ---
	if len(ctx.FileRecords) < 3 {
		t.Errorf("expected at least 3 tracked file records, got %d", len(ctx.FileRecords))
	}
	hasPreFlight := false
	hasImplement := false
	hasAGENTS := false
	hasCopilotInstructions := false
	for _, rec := range ctx.FileRecords {
		switch rec.Path {
		case ".github/prompts/preflight-task-framing.prompt.md":
			hasPreFlight = true
		case ".github/prompts/implement.prompt.md":
			hasImplement = true
		case "AGENTS.md":
			hasAGENTS = true
		case ".github/copilot-instructions.md":
			hasCopilotInstructions = true
		}
	}
	if !hasPreFlight {
		t.Error("no tracked record for preflight-task-framing.prompt.md")
	}
	if !hasImplement {
		t.Error("no tracked record for implement.prompt.md")
	}
	if !hasAGENTS {
		t.Error("no tracked record for AGENTS.md")
	}
	if !hasCopilotInstructions {
		t.Error("no tracked record for .github/copilot-instructions.md")
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

// --- Test: isGlobalOpenCodeDir helper ---

func TestIsGlobalOpenCodeDir(t *testing.T) {
	tests := []struct {
		dir      string
		expected bool
	}{
		{"/home/user/.config/opencode", true},
		{"/Users/ricardo/.config/opencode", true},
		{"/some/path/.config/opencode", true},
		{"/tmp/project/.opencode", false},
		{"/tmp/project", false},
		{"/home/user/.config/other", false},
		{"/home/user/opencode", false},
		{"opencode", false},
	}

	for _, tc := range tests {
		result := isGlobalOpenCodeDir(tc.dir)
		if result != tc.expected {
			t.Errorf("isGlobalOpenCodeDir(%q) = %v, want %v", tc.dir, result, tc.expected)
		}
	}
}

// --- Test: isGlobalOpenCodeDir helper with MCP compilation paths ---

func TestIsGlobalOpenCodeDir_MCPPaths(t *testing.T) {
	// When targetDir is ~/.config/opencode, the config should be written to opencode.jsonc (not .opencode/opencode.jsonc).
	configOpencode := filepath.Join(t.TempDir(), ".config", "opencode")
	_ = os.MkdirAll(configOpencode, 0o755)
	if !isGlobalOpenCodeDir(configOpencode) {
		t.Fatalf("isGlobalOpenCodeDir(%q) should return true", configOpencode)
	}
}

func TestIsGlobalOpenCodeDir_ProjectPaths(t *testing.T) {
	projectDir := t.TempDir()

	// Project scope target dir should NOT be detected as global.
	if isGlobalOpenCodeDir(projectDir) {
		t.Fatalf("isGlobalOpenCodeDir(%q) should return false", projectDir)
	}

	opencodeDir := filepath.Join(projectDir, ".opencode")
	if isGlobalOpenCodeDir(opencodeDir) {
		t.Fatalf("isGlobalOpenCodeDir(%q) should return false", opencodeDir)
	}
}
