// Package adapter provides the Claude Code adapter implementation.
// Ported from the TypeScript ClaudeCodeAdapter.
package adapter

import (
	"log"
	"path/filepath"

	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// ClaudeCodeAdapter implements ToolAdapter for Claude Code (claude CLI).
type ClaudeCodeAdapter struct{}

func (a *ClaudeCodeAdapter) ID() types.ToolId  { return types.ToolIdClaudeCode }
func (a *ClaudeCodeAdapter) Name() string      { return "Claude Code" }
func (a *ClaudeCodeAdapter) ConfigDir() string { return ".claude" }

func (a *ClaudeCodeAdapter) Install(ctx *AdapterContext) ([]types.TrackedFile, error) {
	isGlobal := ctx.SetupScope == types.SetupScopeGlobal
	claudeDir := ctx.TargetDir
	if !isGlobal {
		claudeDir = filepath.Join(ctx.TargetDir, ".claude")
	}

	settingsPath := filepath.Join(claudeDir, "settings.json")
	rulesDir := filepath.Join(claudeDir, "rules")
	sampleRulePath := filepath.Join(rulesDir, "typescript.md")

	_ = files.EnsureDir(claudeDir)
	_ = files.EnsureDir(rulesDir)
	_ = files.EnsureDir(filepath.Join(claudeDir, "skills"))
	if !isGlobal {
		_ = files.EnsureDir(filepath.Join(claudeDir, "agents"))
	}

	// Write default settings.json if it doesn't exist.
	if !files.FileExists(settingsPath) {
		defaultSettings := map[string]any{
			"permissions": map[string]any{
				"allow": []any{},
				"deny":  []any{},
			},
		}
		if err := WriteJSONFile(settingsPath, defaultSettings); err != nil {
			return nil, err
		}
		relPath, _ := filepath.Rel(ctx.TargetDir, settingsPath)
		hash, _ := files.FileHash(settingsPath)
		ctx.FileRecords = append(ctx.FileRecords, types.TrackedFile{
			Path: relPath, Hash: hash, Source: "generated", Owner: types.FileOwnerLibrary,
		})
	}

	// Write sample rule if it doesn't exist.
	if !files.FileExists(sampleRulePath) {
		content := "---\npaths:\n  - \"src/**/*.ts\"\n---\n\n# TypeScript Rules\n\n- Use strict TypeScript\n- Prefer interfaces over types for objects\n"
		if err := files.WriteFile(sampleRulePath, []byte(content), 0o644); err != nil {
			return nil, err
		}
		relPath, _ := filepath.Rel(ctx.TargetDir, sampleRulePath)
		hash, _ := files.FileHash(sampleRulePath)
		ctx.FileRecords = append(ctx.FileRecords, types.TrackedFile{
			Path: relPath, Hash: hash, Source: "generated", Owner: types.FileOwnerLibrary,
		})
	}

	log.Println("Installing Claude Code tools...")

	// Copy agents.
	if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "agents",
		SelectionKey: "agents",
		ToDestPath: func(file string) string {
			if isGlobal {
				return filepath.Join(claudeDir, file)
			}
			return filepath.Join(claudeDir, "agents", file)
		},
		IncludeFile: func(file string) bool {
			return fileID(file) != "orchestrator"
		},
	}); err != nil {
		return nil, err
	}

	// Orchestrator agent if enabled.
	if !isGlobal && IsOrchestratorEnabled(ctx) {
		orchestratorSource := filepath.Join(ctx.LibraryDir, "agents", "orchestrator.md")
		content := GetOrchestratorAgentContent(ctx)
		if err := CopyWithRecord(orchestratorSource,
			filepath.Join(claudeDir, "agents", "orchestrator.md"),
			ctx, false,
			func([]byte) []byte { return content },
		); err != nil {
			return nil, err
		}
	}

	// Copy skills.
	if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "skills",
		SelectionKey: "skills",
		ToDestPath: func(file string) string {
			name := fileID(file)
			return filepath.Join(claudeDir, "skills", name, "SKILL.md")
		},
	}); err != nil {
		return nil, err
	}

	// Install tool context files.
	if err := InstallToolContextFiles(InstallToolContextFilesOption{
		Ctx:             ctx,
		ToolDir:         claudeDir,
		ContextFileName: "CLAUDE.md",
		AgentsDestDir: func() string {
			if isGlobal {
				return "."
			}
			return "agents"
		}(),
		SkillsDestDir: "skills",
	}); err != nil {
		return nil, err
	}

	// Install root CLAUDE.md template.
	destPath := filepath.Join(ctx.TargetDir, "CLAUDE.md")
	if isGlobal {
		destPath = filepath.Join(claudeDir, "CLAUDE.md")
	}
	if err := InstallRootTemplateIfMissing(ctx, "CLAUDE.md", destPath, "root/CLAUDE.template.md"); err != nil {
		return nil, err
	}

	return ctx.FileRecords, nil
}

func (a *ClaudeCodeAdapter) CompileMCP(targetDir string, fileRecords []types.TrackedFile) ([]types.TrackedFile, error) {
	return CompileMCPForTool(types.ToolIdClaudeCode, targetDir, fileRecords)
}
