// Package adapter provides the Gemini adapter implementation.
// Ported from the TypeScript GeminiAdapter.
package adapter

import (
	"log"
	"path/filepath"

	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// GeminiAdapter implements ToolAdapter for Gemini CLI.
type GeminiAdapter struct{}

func (a *GeminiAdapter) ID() types.ToolId  { return types.ToolIdGemini }
func (a *GeminiAdapter) Name() string      { return "Gemini CLI" }
func (a *GeminiAdapter) ConfigDir() string { return ".gemini" }

func (a *GeminiAdapter) Install(ctx *AdapterContext) ([]types.TrackedFile, error) {
	geminiDir := filepath.Join(ctx.TargetDir, ".gemini")
	_ = files.EnsureDir(geminiDir)
	_ = files.EnsureDir(filepath.Join(geminiDir, "skills"))

	// Write default settings.json if it doesn't exist.
	settingsPath := filepath.Join(geminiDir, "settings.json")
	if !files.FileExists(settingsPath) {
		defaultSettings := map[string]any{
			"general": map[string]any{
				"defaultApprovalMode": "default",
			},
			"model": map[string]any{
				"name": "gemini-2.5-pro",
			},
			"context": map[string]any{
				"fileName":             "GEMINI.md",
				"includeDirectoryTree": true,
			},
		}
		if err := WriteJSONFile(settingsPath, defaultSettings); err != nil {
			return nil, err
		}
		hash, _ := files.FileHash(settingsPath)
		ctx.FileRecords = append(ctx.FileRecords, types.TrackedFile{
			Path: ".gemini/settings.json", Hash: hash, Source: "generated", Owner: types.FileOwnerLibrary,
		})
	}

	log.Println("Installing Gemini CLI tools...")

	// Gemini has no agents concept — skip agents entirely.
	// Copy skills.
	if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "skills",
		SelectionKey: "skills",
		ToDestPath: func(file string) string {
			name := fileID(file)
			return filepath.Join(geminiDir, "skills", name, "SKILL.md")
		},
	}); err != nil {
		return nil, err
	}

	// Orchestrator as a skill if enabled.
	if IsOrchestratorEnabled(ctx) {
		content := GetOrchestratorSkillContent(ctx)
		if err := WriteContentWithRecord(
			filepath.Join(geminiDir, "skills", "orchestrator", "SKILL.md"),
			content, ctx, "generated:orchestrator-skill", false,
		); err != nil {
			return nil, err
		}
	}

	// Install root GEMINI.md template.
	if err := InstallRootTemplateIfMissing(ctx, "GEMINI.md",
		filepath.Join(ctx.TargetDir, "GEMINI.md"),
		"root/GEMINI.template.md"); err != nil {
		return nil, err
	}

	return ctx.FileRecords, nil
}

func (a *GeminiAdapter) CompileMCP(targetDir string, fileRecords []types.TrackedFile) ([]types.TrackedFile, error) {
	return CompileMCPForTool(types.ToolIdGemini, targetDir, fileRecords)
}
