package adapter

import (
	"path/filepath"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// OmpAdapter installs the verified OMP skills-only surface.
type OmpAdapter struct{}

func (a *OmpAdapter) ID() types.ToolId  { return types.ToolIdOmp }
func (a *OmpAdapter) Name() string      { return "OMP" }
func (a *OmpAdapter) ConfigDir() string { return ".omp" }

func (a *OmpAdapter) Install(ctx *AdapterContext) ([]types.TrackedFile, error) {
	if !IsScopeSupported(types.ToolIdOmp, ctx.SetupScope) {
		return ctx.FileRecords, nil
	}
	ompDir, err := ResolveToolRoot(types.ToolIdOmp, ctx.SetupScope, ctx)
	if err != nil {
		return nil, err
	}

	_ = files.EnsureDir(ompDir)
	_ = files.EnsureDir(filepath.Join(ompDir, "skills"))

	if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "skills",
		SelectionKey: "skills",
		ToDestPath: func(file string) string {
			name := fileID(file)
			return filepath.Join(ompDir, "skills", name, "SKILL.md")
		},
	}); err != nil {
		return nil, err
	}

	return ctx.FileRecords, nil
}

func (a *OmpAdapter) CompileMCP(ctx CompileContext) ([]types.TrackedFile, error) {
	return ctx.FileRecords, nil
}

func (a *OmpAdapter) CanRunHeadless() bool { return false }

func (a *OmpAdapter) RunHeadlessValidation(ctx *AdapterContext) error { return nil }

func (a *OmpAdapter) RunHeadlessInit(ctx *AdapterContext, prompt string) error { return nil }
