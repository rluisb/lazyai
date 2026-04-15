// Package adapter provides the GitHub Copilot adapter implementation.
// Ported from the TypeScript CopilotAdapter.
package adapter

import (
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
	githubDir := filepath.Join(ctx.TargetDir, ".github")
	_ = files.EnsureDir(githubDir)

	instructionsDir := filepath.Join(githubDir, "instructions")
	_ = files.EnsureDir(instructionsDir)
	promptsDir := filepath.Join(githubDir, "prompts")
	_ = files.EnsureDir(promptsDir)

	log.Println("Installing GitHub Copilot tools...")

	selectedSkills := selectionSet(ctx.Selections.Skills)
	selectedPrompts := selectionSet(ctx.Selections.Prompts)

	// Copy prompts as Copilot .prompt.md files.
	promptTemplatesDir := filepath.Join(ctx.LibraryDir, "prompts")
	if files.DirExists(promptTemplatesDir) {
		for _, file := range files.ListDir(promptTemplatesDir) {
			fileIDVal := fileID(file)
			if selectedPrompts != nil && !selectedPrompts[types.PromptId(fileIDVal)] {
				continue
			}
			srcPath := filepath.Join(promptTemplatesDir, file)
			if files.IsDirectory(srcPath) {
				continue
			}
			destFile := fileIDVal + ".prompt.md"
			if err := a.copyFileWithRecord(srcPath, filepath.Join(promptsDir, destFile), ctx); err != nil {
				return nil, err
			}
		}
	}

	// Copy skills as Copilot prompt files.
	skillsDir := filepath.Join(ctx.LibraryDir, "skills")
	if files.DirExists(skillsDir) {
		for _, file := range files.ListDir(skillsDir) {
			srcPath := filepath.Join(skillsDir, file)
			if files.IsDirectory(srcPath) {
				continue
			}
			fileIDVal := fileID(file)
			if selectedSkills != nil && !selectedSkills[types.SkillId(fileIDVal)] {
				continue
			}
			destFile := fileIDVal + ".prompt.md"
			dest := filepath.Join(promptsDir, destFile)
			if err := a.copySkillAsPromptWithRecord(srcPath, dest, ctx); err != nil {
				return nil, err
			}
		}
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

	// Install root AGENTS.md template.
	if err := InstallRootTemplateIfMissing(ctx, "AGENTS.md",
		filepath.Join(ctx.TargetDir, "AGENTS.md"),
		"root/AGENTS.template.md"); err != nil {
		return nil, err
	}

	// Install copilot-instructions.md.
	if err := InstallRootTemplateIfMissing(ctx, ".github/copilot-instructions.md",
		filepath.Join(githubDir, "copilot-instructions.md"),
		"root/copilot-instructions.template.md"); err != nil {
		return nil, err
	}

	return ctx.FileRecords, nil
}

func (a *CopilotAdapter) CompileMCP(targetDir string, fileRecords []types.TrackedFile) ([]types.TrackedFile, error) {
	return CompileMCPForTool(types.ToolIdCopilot, targetDir, fileRecords)
}

// copyFileWithRecord copies a file with conflict resolution and tracking.
func (a *CopilotAdapter) copyFileWithRecord(src, dest string, ctx *AdapterContext) error {
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
	source, _ := filepath.Rel(ctx.LibraryDir, src)
	ctx.FileRecords = append(ctx.FileRecords, types.TrackedFile{
		Path: relPath, Hash: hash, Source: source, Owner: types.FileOwnerLibrary,
	})
	return nil
}

// copySkillAsPromptWithRecord reads a skill file, transforms it to a Copilot
// prompt format (mode: agent frontmatter), and writes it.
func (a *CopilotAdapter) copySkillAsPromptWithRecord(src, dest string, ctx *AdapterContext) error {
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
	source, _ := filepath.Rel(ctx.LibraryDir, src)
	ctx.FileRecords = append(ctx.FileRecords, types.TrackedFile{
		Path: relPath, Hash: hash, Source: source, Owner: types.FileOwnerLibrary,
	})
	return nil
}
