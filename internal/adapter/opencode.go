// Package adapter provides the OpenCode adapter implementation.
// Ported from the TypeScript OpenCodeAdapter.
package adapter

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ricardoborges-teachable/ai-setup/internal/configmerge"
	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// OpenCodeConfigFilename is the canonical opencode config filename. Both the
// install-time default-config writer and the compile-time MCP writer target
// this single file so users never end up with both opencode.json and
// opencode.jsonc side-by-side.
const OpenCodeConfigFilename = "opencode.jsonc"

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

	jsonPath := filepath.Join(ocDir, "opencode.json")
	jsoncPath := filepath.Join(ocDir, OpenCodeConfigFilename)

	// One-shot migration: collapse any pre-existing opencode.json onto
	// opencode.jsonc. The original .json is preserved as a .bak sidecar so
	// users can recover if the rename surprises them. If both files exist
	// (rare), .jsonc is authoritative.
	if files.FileExists(jsonPath) {
		bakPath := jsonPath + ".bak"
		if !files.FileExists(bakPath) {
			if err := files.CopyFile(jsonPath, bakPath); err != nil {
				return nil, fmt.Errorf("backup opencode.json: %w", err)
			}
		}
		if !files.FileExists(jsoncPath) {
			if err := files.CopyFile(jsonPath, jsoncPath); err != nil {
				return nil, fmt.Errorf("migrate opencode.json to opencode.jsonc: %w", err)
			}
		}
		if err := os.Remove(jsonPath); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("remove opencode.json after migration: %w", err)
		}
	}

	// Install the default config only when no config is present. Skipping the
	// merge on pre-existing configs preserves user customizations (e.g., a
	// hand-tuned `permission.edit` value) across re-runs.
	if !files.FileExists(jsoncPath) {
		defaultConfig := map[string]any{
			"$schema":      "https://opencode.ai/config.json",
			"instructions": []any{"AGENTS.md"},
			"permission": map[string]any{
				"edit": "ask",
				"bash": "ask",
			},
		}
		if _, err := configmerge.MergeJSONFile(jsoncPath, defaultConfig); err != nil {
			return nil, err
		}
		hash, _ := files.FileHash(jsoncPath)
		ctx.FileRecords = append(ctx.FileRecords, types.TrackedFile{
			Path:   jsoncPath,
			Hash:   hash,
			Source: "generated",
			Owner:  types.FileOwnerLibrary,
		})
	}

	// Copy agents (excluding orchestrator unless enabled). The transform
	// rewrites each source agent with opencode-schema frontmatter so
	// `opencode debug agent <name>` can parse it.
	if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "agents",
		SelectionKey: "agents",
		ToDestPath: func(file string) string {
			return filepath.Join(ocDir, "agents", file)
		},
		WarnOnSkip: true,
		Transform: func(content []byte) []byte {
			return BuildOpenCodeAgentFrontmatter(content, OpenCodeAgentOpts{})
		},
		IncludeFile: func(file string) bool {
			name := fileID(file)
			return name != "orchestrator"
		},
	}); err != nil {
		return nil, err
	}

	// Orchestrator agent if enabled. It is the primary entry point
	// (opencode's default_agent typically points here), so mode=primary.
	if IsOrchestratorEnabled(ctx) {
		raw := GetOrchestratorAgentContent(ctx)
		content := BuildOpenCodeAgentFrontmatter(raw, OpenCodeAgentOpts{Mode: "primary"})
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
