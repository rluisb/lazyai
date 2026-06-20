package adapter

import (
	"path/filepath"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// KiroAdapter installs the Kiro IDE setup surface.
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

	_ = files.EnsureDir(kiroDir)
	_ = files.EnsureDir(filepath.Join(kiroDir, "skills"))

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

	return ctx.FileRecords, nil
}

func (a *KiroAdapter) CompileMCP(ctx CompileContext) ([]types.TrackedFile, error) {
	return CompileMCPForTool(types.ToolIdKiro, ctx)
}

func (a *KiroAdapter) CanRunHeadless() bool { return false }

func (a *KiroAdapter) RunHeadlessValidation(ctx *AdapterContext) error { return nil }

func (a *KiroAdapter) RunHeadlessInit(ctx *AdapterContext, prompt string) error { return nil }
