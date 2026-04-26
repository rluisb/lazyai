package adapter

import (
	"encoding/json"
	"path/filepath"

	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// PiAdapter implements ToolAdapter for Pi setup targets.
//
// Concrete in-repo evidence for the supported layout comes from the existing
// TypeScript adapter + migration parser and is intentionally conservative:
//   - .pi/settings.json
//   - .pi/skills/<name>/SKILL.md
//   - .pi/prompts/<name>.md
//   - root AGENTS.md
//
// No MCP compile/install behavior is inferred here.
type PiAdapter struct{}

func (a *PiAdapter) ID() types.ToolId  { return types.ToolIdPi }
func (a *PiAdapter) Name() string      { return "Pi" }
func (a *PiAdapter) ConfigDir() string { return ".pi" }

func (a *PiAdapter) Install(ctx *AdapterContext) ([]types.TrackedFile, error) {
	piDir, err := ResolveToolRoot(types.ToolIdPi, ctx.SetupScope, ctx)
	if err != nil {
		return nil, err
	}

	_ = files.EnsureDir(piDir)
	_ = files.EnsureDir(filepath.Join(piDir, "skills"))
	_ = files.EnsureDir(filepath.Join(piDir, "prompts"))

	settingsPath := filepath.Join(piDir, "settings.json")
	if !files.FileExists(settingsPath) {
		content, err := json.MarshalIndent(map[string]any{
			"compaction": map[string]any{"enabled": true},
		}, "", "  ")
		if err != nil {
			return nil, err
		}
		if err := WriteContentWithRecord(settingsPath, content, ctx, "generated", false); err != nil {
			return nil, err
		}
	}

	if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "skills",
		SelectionKey: "skills",
		ToDestPath: func(file string) string {
			return filepath.Join(piDir, "skills", trimExt(file), "SKILL.md")
		},
	}); err != nil {
		return nil, err
	}

	if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "prompts",
		SelectionKey: "prompts",
		ToDestPath: func(file string) string {
			return filepath.Join(piDir, "prompts", file)
		},
	}); err != nil {
		return nil, err
	}

	if err := InstallRootTemplateIfMissing(ctx, "AGENTS.md", filepath.Join(ctx.TargetDir, "AGENTS.md"), "root/AGENTS.template.md"); err != nil {
		return nil, err
	}

	return ctx.FileRecords, nil
}

func (a *PiAdapter) CompileMCP(ctx CompileContext) ([]types.TrackedFile, error) {
	return ctx.FileRecords, nil
}

func (a *PiAdapter) CanRunHeadless() bool { return false }

func (a *PiAdapter) RunHeadlessValidation(ctx *AdapterContext) error { return nil }

func trimExt(file string) string {
	return file[:len(file)-len(filepath.Ext(file))]
}
