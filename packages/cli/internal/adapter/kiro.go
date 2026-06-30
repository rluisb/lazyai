package adapter

import (
	"path/filepath"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// KiroAdapter installs the Kiro IDE/CLI setup surface. Kiro CLI v3 discovers
// custom agent profiles from .kiro/agents/<name>.json (.json required; .md is
// not recognized), skills from .kiro/skills/<name>/SKILL.md, prompt templates
// from .kiro/prompts/*.md, and native hooks from .kiro/hooks/<name>.json.
type KiroAdapter struct{}

func (a *KiroAdapter) ID() types.ToolId  { return types.ToolIdKiro }
func (a *KiroAdapter) Name() string      { return "Kiro" }
func (a *KiroAdapter) ConfigDir() string { return ".kiro" }

func (a *KiroAdapter) Install(ctx *AdapterContext) ([]types.TrackedFile, error) {
	if !IsScopeSupported(types.ToolIdKiro, ctx.SetupScope) {
		return ctx.FileRecords, nil
	}
	kiroDir, err := ResolveToolRoot(types.ToolIdKiro, ctx.SetupScope, ctx)
	if err != nil {
		return nil, err
	}

	if err := files.EnsureDir(filepath.Join(kiroDir, "agents")); err != nil {
		return nil, err
	}
	if err := files.EnsureDir(filepath.Join(kiroDir, "prompts")); err != nil {
		return nil, err
	}
	if err := files.EnsureDir(filepath.Join(kiroDir, "skills")); err != nil {
		return nil, err
	}

	// Default agent: emit as JSON per official Kiro docs (kiro.dev/docs/cli/custom-agents/).
	if err := CopyWithRecord(
		"canonical/agents/"+defaultAgentID+".md",
		filepath.Join(kiroDir, "agents", defaultAgentID+".json"),
		ctx, true,
		func(content []byte) []byte {
			out, err := RewriteAgentForKiro(content, ctx)
			if err != nil {
				adapterLog.Warn("kiro default agent rewrite failed", "adapter", "kiro", "error", err)
				return content
			}
			return out
		},
		0o644,
	); err != nil {
		return nil, err
	}

	if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "canonical/agents",
		SelectionKey: "agents",
		ToDestPath: func(file string) string {
			return filepath.Join(kiroDir, "agents", fileID(file)+".json")
		},
		IncludeFile: func(file string) bool {
			return !isDefaultAgentFile(file) && isCanonicalAgentFile(file)
		},
		Transform: func(content []byte) []byte {
			out, err := RewriteAgentForKiro(content, ctx)
			if err != nil {
				adapterLog.Warn("kiro agent rewrite fell back to verbatim copy", "adapter", "kiro", "error", err)
				return content
			}
			return out
		},
	}); err != nil {
		return nil, err
	}

	if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "skills",
		SelectionKey: "skills",
		ToDestPath: func(file string) string {
			name := fileID(file)
			return filepath.Join(kiroDir, "skills", name, "SKILL.md")
		},
	}); err != nil {
		return nil, err
	}

	if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "prompts",
		SelectionKey: "prompts",
		ToDestPath: func(file string) string {
			return filepath.Join(kiroDir, "prompts", filepath.Base(file))
		},
	}); err != nil {
		return nil, err
	}

	if err := files.EnsureDir(filepath.Join(kiroDir, "hooks")); err != nil {
		return nil, err
	}

	if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "kiro/hooks",
		ToDestPath: func(file string) string {
			return filepath.Join(kiroDir, "hooks", file)
		},
	}); err != nil {
		return nil, err
	}

	return ctx.FileRecords, nil
}

func (a *KiroAdapter) CompileMCP(ctx CompileContext) ([]types.TrackedFile, error) {
	return CompileMCPForTool(types.ToolIdKiro, ctx)
}

func (a *KiroAdapter) CanRunHeadless() bool { return false }

func (a *KiroAdapter) RunHeadlessValidation(ctx *AdapterContext) error { return nil }

func (a *KiroAdapter) RunHeadlessInit(ctx *AdapterContext, prompt string) error { return nil }
