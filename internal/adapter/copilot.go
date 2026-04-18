// Package adapter provides the GitHub Copilot adapter implementation.
// Ported from the TypeScript CopilotAdapter.
package adapter

import (
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	"github.com/ricardoborges-teachable/ai-setup/internal/conflict"
	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/frontmatter"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// CopilotAdapter implements ToolAdapter for GitHub Copilot.
type CopilotAdapter struct{}

func (a *CopilotAdapter) ID() types.ToolId  { return types.ToolIdCopilot }
func (a *CopilotAdapter) Name() string      { return "GitHub Copilot" }
func (a *CopilotAdapter) ConfigDir() string { return ".github" }

func (a *CopilotAdapter) Install(ctx *AdapterContext) ([]types.TrackedFile, error) {
	if !IsScopeSupported(types.ToolIdCopilot, ctx.SetupScope) {
		log.Printf("[copilot] scope %q not supported; skipping install", ctx.SetupScope)
		return ctx.FileRecords, nil
	}
	githubDir, err := ResolveToolRoot(types.ToolIdCopilot, ctx.SetupScope, ctx)
	if err != nil {
		return nil, err
	}
	_ = files.EnsureDir(githubDir)

	instructionsDir := filepath.Join(githubDir, "instructions")
	_ = files.EnsureDir(instructionsDir)
	promptsDir := filepath.Join(githubDir, "prompts")
	_ = files.EnsureDir(promptsDir)

	log.Println("Installing GitHub Copilot tools...")

	selectedSkills := selectionSet(ctx.Selections.Skills)
	selectedPrompts := selectionSet(ctx.Selections.Prompts)

	// Copy prompts as Copilot .prompt.md files.
	if err := a.copyLibrarySubdirAsPrompts(ctx, "prompts", selectedPrompts, promptsDir); err != nil {
		return nil, err
	}

	// Copy skills as Copilot prompt files.
	if err := a.copyLibrarySubdirAsSkillPrompts(ctx, "skills", selectedSkills, promptsDir); err != nil {
		return nil, err
	}

	// Orchestrator as a prompt file if enabled.
	if IsOrchestratorEnabled(ctx) {
		content := GetOrchestratorPromptContent(ctx)
		if err := WriteContentWithRecord(
			filepath.Join(promptsDir, "orchestrator.prompt.md"),
			content, ctx, "generated:orchestrator-prompt", false,
		); err != nil {
			return nil, err
		}
	}

	// Copy chat modes to .github/chatmodes/<name>.chatmode.md.
	chatmodesDir := filepath.Join(githubDir, "chatmodes")
	if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "chatmodes",
		SelectionKey: "chatmodes",
		ToDestPath: func(file string) string {
			return filepath.Join(chatmodesDir, filepath.Base(file))
		},
	}); err != nil {
		return nil, err
	}

	// Memory docs (AGENTS.md and .github/copilot-instructions.md) are placed
	// by scaffold/root.go, which knows the scope-correct destination.

	return ctx.FileRecords, nil
}

// copyLibrarySubdirAsPrompts copies files from a library subdirectory as prompt files.
func (a *CopilotAdapter) copyLibrarySubdirAsPrompts(ctx *AdapterContext, subdir string, selected map[types.PromptId]bool, promptsDir string) error {
	libFS := ctx.LibraryFS
	if libFS != nil {
		return a.copySubdirAsPromptsFromFS(ctx, libFS, subdir, selected, promptsDir)
	}
	// Fallback: disk mode
	srcDir := filepath.Join(ctx.LibraryDir, subdir)
	if !files.DirExists(srcDir) {
		return nil
	}
	for _, file := range files.ListDir(srcDir) {
		srcPath := filepath.Join(srcDir, file)
		if files.IsDirectory(srcPath) {
			continue
		}
		fileIDVal := fileID(file)
		if selected != nil && !selected[types.PromptId(fileIDVal)] {
			continue
		}
		destFile := fileIDVal + ".prompt.md"
		if err := a.copyFileWithRecord(srcPath, filepath.Join(promptsDir, destFile), ctx, subdir+"/"+file); err != nil {
			return err
		}
	}
	return nil
}

// copySubdirAsPromptsFromFS copies files from the library FS as prompt files.
func (a *CopilotAdapter) copySubdirAsPromptsFromFS(ctx *AdapterContext, libFS fs.FS, subdir string, selected map[types.PromptId]bool, promptsDir string) error {
	entries, err := fs.ReadDir(libFS, subdir)
	if err != nil {
		return nil // directory doesn't exist
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		file := entry.Name()
		fileIDVal := fileID(file)
		if selected != nil && !selected[types.PromptId(fileIDVal)] {
			continue
		}
		destFile := fileIDVal + ".prompt.md"
		dest := filepath.Join(promptsDir, destFile)
		if err := CopyWithRecord(subdir+"/"+file, dest, ctx, false, nil); err != nil {
			return err
		}
	}
	return nil
}

// copyLibrarySubdirAsSkillPrompts copies skill files as Copilot prompt files.
func (a *CopilotAdapter) copyLibrarySubdirAsSkillPrompts(ctx *AdapterContext, subdir string, selected map[types.SkillId]bool, promptsDir string) error {
	libFS := ctx.LibraryFS
	if libFS != nil {
		return a.copySubdirAsSkillPromptsFromFS(ctx, libFS, subdir, selected, promptsDir)
	}
	// Fallback: disk mode
	srcDir := filepath.Join(ctx.LibraryDir, subdir)
	if !files.DirExists(srcDir) {
		return nil
	}
	for _, file := range files.ListDir(srcDir) {
		srcPath := filepath.Join(srcDir, file)
		if files.IsDirectory(srcPath) {
			continue
		}
		fileIDVal := fileID(file)
		if selected != nil && !selected[types.SkillId(fileIDVal)] {
			continue
		}
		destFile := fileIDVal + ".prompt.md"
		dest := filepath.Join(promptsDir, destFile)
		libRelPath := subdir + "/" + file
		if err := a.copySkillAsPromptWithRecord(ctx, srcPath, dest, libRelPath); err != nil {
			return err
		}
	}
	return nil
}

// copySubdirAsSkillPromptsFromFS copies skill files from FS as Copilot prompts.
func (a *CopilotAdapter) copySubdirAsSkillPromptsFromFS(ctx *AdapterContext, libFS fs.FS, subdir string, selected map[types.SkillId]bool, promptsDir string) error {
	entries, err := fs.ReadDir(libFS, subdir)
	if err != nil {
		return nil
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		file := entry.Name()
		fileIDVal := fileID(file)
		if selected != nil && !selected[types.SkillId(fileIDVal)] {
			continue
		}
		destFile := fileIDVal + ".prompt.md"
		dest := filepath.Join(promptsDir, destFile)
		libRelPath := subdir + "/" + file
		if err := a.copySkillAsPromptFromFS(ctx, libFS, libRelPath, dest); err != nil {
			return err
		}
	}
	return nil
}

// copyFileWithRecord copies a file from disk with conflict resolution and tracking.
// src is an absolute filesystem path; sourcePath is the library-relative path for tracking.
func (a *CopilotAdapter) copyFileWithRecord(src, dest string, ctx *AdapterContext, sourcePath string) error {
	relPath, _ := filepath.Rel(ctx.TargetDir, dest)

	effectiveStrategy := ctx.Strategy
	if override, ok := ctx.PerFileOverrides[dest]; ok {
		effectiveStrategy = override
	}

	action, err := conflict.ResolveConflictWithOptions(dest, relPath, conflict.ConflictOptions{
		Force:    ctx.Force,
		Strategy: effectiveStrategy,
	})
	if err != nil {
		return err
	}
	if action == conflict.ActionSkip {
		return nil
	}
	if action == conflict.ActionReplace && files.FileExists(dest) {
		if _, err := files.BackupFile(dest, ctx.TargetDir); err != nil {
			return err
		}
	}

	if err := files.CopyFile(src, dest); err != nil {
		return err
	}

	hash, _ := files.FileHash(dest)
	ctx.FileRecords = append(ctx.FileRecords, types.TrackedFile{
		Path: relPath, Hash: hash, Source: sourcePath, Owner: types.FileOwnerLibrary,
	})
	return nil
}

// copySkillAsPromptWithRecord reads a skill file from disk, transforms it to a Copilot
// prompt format (mode: agent frontmatter), and writes it.
func (a *CopilotAdapter) copySkillAsPromptWithRecord(ctx *AdapterContext, src, dest string, sourcePath string) error {
	relPath, _ := filepath.Rel(ctx.TargetDir, dest)

	effectiveStrategy := ctx.Strategy
	if override, ok := ctx.PerFileOverrides[dest]; ok {
		effectiveStrategy = override
	}

	action, err := conflict.ResolveConflictWithOptions(dest, relPath, conflict.ConflictOptions{
		Force:    ctx.Force,
		Strategy: effectiveStrategy,
	})
	if err != nil {
		return err
	}
	if action == conflict.ActionSkip {
		return nil
	}
	if action == conflict.ActionReplace && files.FileExists(dest) {
		if _, err := files.BackupFile(dest, ctx.TargetDir); err != nil {
			return err
		}
	}

	data, err := files.ReadFile(src)
	if err != nil {
		return err
	}

	// Strip frontmatter, then add mode: agent frontmatter.
	_, body := frontmatter.SplitYamlFrontmatter(string(data))
	transformed := EnsureModeAgentFrontmatter(strings.TrimSpace(body))

	if err := files.EnsureDir(filepath.Dir(dest)); err != nil {
		return err
	}
	if err := files.WriteFile(dest, []byte(transformed), 0o644); err != nil {
		return err
	}

	hash, _ := files.FileHash(dest)
	ctx.FileRecords = append(ctx.FileRecords, types.TrackedFile{
		Path: relPath, Hash: hash, Source: sourcePath, Owner: types.FileOwnerLibrary,
	})
	return nil
}

// copySkillAsPromptFromFS reads a skill from the library FS, transforms to Copilot prompt.
func (a *CopilotAdapter) copySkillAsPromptFromFS(ctx *AdapterContext, libFS fs.FS, src, dest string) error {
	relPath, _ := filepath.Rel(ctx.TargetDir, dest)

	effectiveStrategy := ctx.Strategy
	if override, ok := ctx.PerFileOverrides[dest]; ok {
		effectiveStrategy = override
	}

	action, err := conflict.ResolveConflictWithOptions(dest, relPath, conflict.ConflictOptions{
		Force:    ctx.Force,
		Strategy: effectiveStrategy,
	})
	if err != nil {
		return err
	}
	if action == conflict.ActionSkip {
		return nil
	}
	if action == conflict.ActionReplace && files.FileExists(dest) {
		if _, err := files.BackupFile(dest, ctx.TargetDir); err != nil {
			return err
		}
	}

	data, err := fs.ReadFile(libFS, src)
	if err != nil {
		return fmt.Errorf("read FS %s: %w", src, err)
	}

	// Strip frontmatter, then add mode: agent frontmatter.
	_, body := frontmatter.SplitYamlFrontmatter(string(data))
	transformed := EnsureModeAgentFrontmatter(strings.TrimSpace(body))

	if err := files.EnsureDir(filepath.Dir(dest)); err != nil {
		return err
	}
	if err := files.WriteFile(dest, []byte(transformed), 0o644); err != nil {
		return err
	}

	hash, _ := files.FileHash(dest)
	ctx.FileRecords = append(ctx.FileRecords, types.TrackedFile{
		Path: relPath, Hash: hash, Source: src, Owner: types.FileOwnerLibrary,
	})
	return nil
}

func (a *CopilotAdapter) CompileMCP(ctx CompileContext) ([]types.TrackedFile, error) {
	return CompileMCPForTool(types.ToolIdCopilot, ctx)
}

func (a *CopilotAdapter) CanRunHeadless() bool { return false }

func (a *CopilotAdapter) RunHeadlessValidation(ctx *AdapterContext) error { return nil }
