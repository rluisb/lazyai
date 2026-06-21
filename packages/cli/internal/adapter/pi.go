package adapter

import (
	"path/filepath"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// PiAdapter installs Pi's native project/user surfaces.
// Pi subagent definitions are markdown files under .pi/agents/<name>.md when
// loaded by Pi's subagent extension; skills stay in .pi/skills/<name>/SKILL.md.
type PiAdapter struct{}

func (a *PiAdapter) ID() types.ToolId  { return types.ToolIdPi }
func (a *PiAdapter) Name() string      { return "OMP/Pi" }
func (a *PiAdapter) ConfigDir() string { return ".pi" }

func (a *PiAdapter) Install(ctx *AdapterContext) ([]types.TrackedFile, error) {
	if !IsScopeSupported(types.ToolIdPi, ctx.SetupScope) {
		return ctx.FileRecords, nil
	}
	piDir, err := ResolveToolRoot(types.ToolIdPi, ctx.SetupScope, ctx)
	if err != nil {
		return nil, err
	}

	_ = files.EnsureDir(piDir)
	_ = files.EnsureDir(filepath.Join(piDir, "skills"))

	if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "canonical/agents",
		SelectionKey: "agents",
		ToDestPath: func(file string) string {
			return filepath.Join(piDir, "agents", filepath.Base(file))
		},
		IncludeFile: isCanonicalAgentFile,
	}); err != nil {
		return nil, err
	}

	if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "skills",
		SelectionKey: "skills",
		ToDestPath: func(file string) string {
			name := fileID(file)
			return filepath.Join(piDir, "skills", name, "SKILL.md")
		},
	}); err != nil {
		return nil, err
	}

	return ctx.FileRecords, nil
}

func (a *PiAdapter) CompileMCP(ctx CompileContext) ([]types.TrackedFile, error) {
	return ctx.FileRecords, nil
}

func (a *PiAdapter) CanRunHeadless() bool { return false }

func (a *PiAdapter) RunHeadlessValidation(ctx *AdapterContext) error { return nil }

func (a *PiAdapter) RunHeadlessInit(ctx *AdapterContext, prompt string) error { return nil }
