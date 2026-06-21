package adapter

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// --- Test: CopyWithRecord with fs.FS ---

func TestCopyWithRecord_FromFS(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)
	dest := filepath.Join(targetDir, ".opencode", "agents", "implementer.md")

	err := CopyWithRecord("agents/implementer.md", dest, ctx, true, nil, 0o644)
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
	expected, _ := ctx.LibraryFS.Open("agents/implementer.md")
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
	if ctx.FileRecords[0].Source != "agents/implementer.md" {
		t.Errorf("expected source 'agents/implementer.md', got %q", ctx.FileRecords[0].Source)
	}
}

// --- Test: CopyWithRecord with transform ---

func TestCopyWithRecord_WithTransform(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)
	dest := filepath.Join(targetDir, ".opencode", "agents", "implementer.md")

	transformCalled := false
	transform := func(content []byte) []byte {
		transformCalled = true
		return append([]byte("<!-- transformed -->\n"), content...)
	}

	err := CopyWithRecord("agents/implementer.md", dest, ctx, true, transform, 0o644)
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

	// researcher.md should exist (it's in our selection).
	researcherDest := filepath.Join(agentsDir, "researcher.md")
	if _, err := os.Stat(researcherDest); os.IsNotExist(err) {
		t.Error("researcher.md was not copied")
	}

	// extra.md should NOT exist (not in our selection).
	extraDest := filepath.Join(agentsDir, "extra.md")
	if _, err := os.Stat(extraDest); !os.IsNotExist(err) {
		t.Error("extra.md should not have been copied (not selected)")
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

	// Both implementer.md and extra.md should exist.
	for _, name := range []string{"implementer.md", "extra.md"} {
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
	dest := filepath.Join(targetDir, "implementer.md")

	// Write an existing file.
	existingContent := []byte("existing content")
	_ = files.EnsureDir(filepath.Dir(dest))
	_ = os.WriteFile(dest, existingContent, 0o644)

	// CopyWithRecord with skip strategy should skip.
	err := CopyWithRecord("agents/implementer.md", dest, ctx, true, nil, 0o644)
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
	dest := filepath.Join(targetDir, "implementer.md")

	// Write an existing file.
	existingContent := []byte("existing content")
	_ = files.EnsureDir(filepath.Dir(dest))
	_ = os.WriteFile(dest, existingContent, 0o644)

	err := CopyWithRecord("agents/implementer.md", dest, ctx, true, nil, 0o644)
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

// --- Test: Disk fallback mode ---

func TestCopyLibraryDirectory_DiskFallback(t *testing.T) {
	// Create a temp library directory on disk.
	libDir := t.TempDir()
	agentsDir := filepath.Join(libDir, "agents")
	_ = os.MkdirAll(agentsDir, 0o755)
	_ = os.WriteFile(filepath.Join(agentsDir, "researcher.md"), []byte("---\nname: Researcher\n---\n\nResearcher content."), 0o644)

	targetDir := t.TempDir()
	destAgentsDir := filepath.Join(targetDir, ".opencode", "agents")
	_ = files.EnsureDir(destAgentsDir)

	ctx := &AdapterContext{
		TargetDir:  targetDir,
		LibraryDir: libDir,
		LibraryFS:  nil, // nil = disk mode
		Strategy:   types.ConflictStrategyAlign,
		Selections: AdapterSelections{
			Agents: []types.AgentId{"researcher"},
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

	destFile := filepath.Join(destAgentsDir, "researcher.md")
	if _, err := os.Stat(destFile); os.IsNotExist(err) {
		t.Error("researcher.md was not copied from disk")
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
