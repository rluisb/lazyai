// Package adapter provides the OpenCode adapter implementation.
// Ported from the TypeScript OpenCodeAdapter.
package adapter

import (
	"encoding/json"
	"log"
	"path/filepath"

	"github.com/ricardoborges-teachable/ai-setup/internal/configmerge"
	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// OpenCodeAdapter implements ToolAdapter for OpenCode (opencode CLI).
type OpenCodeAdapter struct{}

func (a *OpenCodeAdapter) ID() types.ToolId  { return types.ToolIdOpenCode }
func (a *OpenCodeAdapter) Name() string      { return "OpenCode" }
func (a *OpenCodeAdapter) ConfigDir() string { return ".opencode" }

func (a *OpenCodeAdapter) Install(ctx *AdapterContext) ([]types.TrackedFile, error) {
	ocDir, err := ResolveToolRoot(types.ToolIdOpenCode, ctx.SetupScope, ctx)
	if err != nil {
		return nil, err
	}

	_ = files.EnsureDir(ocDir)
	_ = files.EnsureDir(filepath.Join(ocDir, "agents"))
	_ = files.EnsureDir(filepath.Join(ocDir, "skills"))
	_ = files.EnsureDir(filepath.Join(ocDir, "commands"))

	log.Println("Installing OpenCode tools...")

	// Merge default opencode.json (preserves user-authored keys). Skip the
	// merge only if the user-preferred JSONC variant already exists.
	configPath := filepath.Join(ocDir, "opencode.json")
	jsoncConfigPath := filepath.Join(ocDir, "opencode.jsonc")
	if !files.FileExists(jsoncConfigPath) {
		defaultConfig := map[string]any{
			"$schema":      "https://opencode.ai/config.json",
			"instructions": []any{"AGENTS.md"},
			"permission": map[string]any{
				"edit": "ask",
				"bash": "ask",
			},
		}
		preExisted := files.FileExists(configPath)
		if _, err := configmerge.MergeJSONFile(configPath, defaultConfig); err != nil {
			return nil, err
		}
		if !preExisted {
			hash, _ := files.FileHash(configPath)
			ctx.FileRecords = append(ctx.FileRecords, types.TrackedFile{
				Path:   configPath,
				Hash:   hash,
				Source: "generated",
				Owner:  types.FileOwnerLibrary,
			})
		}
	}

	// Copy agents (excluding orchestrator unless enabled).
	if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "agents",
		SelectionKey: "agents",
		ToDestPath: func(file string) string {
			return filepath.Join(ocDir, "agents", file)
		},
		WarnOnSkip: true,
		Transform:  StripFrontmatterAndInjectModel,
		IncludeFile: func(file string) bool {
			name := fileID(file)
			return name != "orchestrator"
		},
	}); err != nil {
		return nil, err
	}

	// Orchestrator agent if enabled.
	if IsOrchestratorEnabled(ctx) {
		content := GetOrchestratorAgentContent(ctx)
		if err := CopyWithRecord("agents/orchestrator.md",
			filepath.Join(ocDir, "agents", "orchestrator.md"),
			ctx, true,
			func([]byte) []byte { return content },
		); err != nil {
			return nil, err
		}
	}

	// Copy skills (directory-per-skill layout).
	if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "skills",
		SelectionKey: "skills",
		ToDestPath: func(file string) string {
			name := fileID(file)
			return filepath.Join(ocDir, "skills", name, "SKILL.md")
		},
		WarnOnSkip: true,
	}); err != nil {
		return nil, err
	}

	// Install tool context files (AGENTS.md in each directory).
	if err := InstallToolContextFiles(InstallToolContextFilesOption{
		Ctx:             ctx,
		ToolDir:         ocDir,
		ContextFileName: "AGENTS.md",
		AgentsDestDir:   "agents",
		SkillsDestDir:   "skills",
		WarnOnSkip:      true,
	}); err != nil {
		return nil, err
	}

	return ctx.FileRecords, nil
}

func (a *OpenCodeAdapter) CompileMCP(ctx CompileContext) ([]types.TrackedFile, error) {
	return CompileMCPForTool(types.ToolIdOpenCode, ctx)
}

func (a *OpenCodeAdapter) CanRunHeadless() bool { return false }

func (a *OpenCodeAdapter) RunHeadlessValidation(ctx *AdapterContext) error { return nil }

// fileID extracts the filename without extension.
func fileID(filename string) string {
	ext := filepath.Ext(filename)
	return filename[:len(filename)-len(ext)]
}

// MarshalJSON marshals v to indented JSON bytes.
func MarshalJSON(v any) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}
