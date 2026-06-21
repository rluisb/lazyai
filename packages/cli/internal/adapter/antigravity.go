package adapter

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/rluisb/lazyai/packages/cli/internal/configmerge"
	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// AntigravityAdapter installs the verified minimal .gemini surface plus
// official local Agent Skills at .agents/skills/<name>/SKILL.md.
type AntigravityAdapter struct{}

func (a *AntigravityAdapter) ID() types.ToolId  { return types.ToolIdAntigravity }
func (a *AntigravityAdapter) Name() string      { return "Antigravity" }
func (a *AntigravityAdapter) ConfigDir() string { return ".gemini" }

func (a *AntigravityAdapter) Install(ctx *AdapterContext) ([]types.TrackedFile, error) {
	if !IsScopeSupported(types.ToolIdAntigravity, ctx.SetupScope) {
		return ctx.FileRecords, nil
	}
	geminiDir, err := ResolveToolRoot(types.ToolIdAntigravity, ctx.SetupScope, ctx)
	if err != nil {
		return nil, err
	}

	settingsPath := filepath.Join(geminiDir, "settings.json")
	hooksDir := filepath.Join(geminiDir, "hooks", "lazyai")
	agentsRoot := filepath.Join(filepath.Dir(geminiDir), ".agents")
	skillsDir := filepath.Join(agentsRoot, "skills")
	_ = files.EnsureDir(geminiDir)
	_ = files.EnsureDir(hooksDir)
	_ = files.EnsureDir(skillsDir)

	defaultSettings, err := readJSONAsset(ctx, "antigravity/settings.json")
	if err != nil {
		return nil, err
	}
	if _, err := configmerge.MergeJSONFile(settingsPath, defaultSettings); err != nil {
		return nil, err
	}

	relPath, _ := filepath.Rel(ctx.TargetDir, settingsPath)
	hash, _ := files.FileHash(settingsPath)
	ctx.FileRecords = append(ctx.FileRecords, types.TrackedFile{
		Path: relPath, Hash: hash, Source: "antigravity/settings.json", Owner: types.FileOwnerLibrary,
	})

	if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "antigravity/hooks",
		Recursive:    true,
		ToDestPath: func(file string) string {
			return filepath.Join(geminiDir, "hooks", file)
		},
		Mode: 0o755,
	}); err != nil {
		return nil, err
	}

	if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "skills",
		SelectionKey: "skills",
		ToDestPath: func(file string) string {
			name := fileID(file)
			return filepath.Join(skillsDir, name, "SKILL.md")
		},
	}); err != nil {
		return nil, err
	}

	return ctx.FileRecords, nil
}

func (a *AntigravityAdapter) CompileMCP(ctx CompileContext) ([]types.TrackedFile, error) {
	return ctx.FileRecords, nil
}

func (a *AntigravityAdapter) CanRunHeadless() bool { return false }

func (a *AntigravityAdapter) RunHeadlessValidation(ctx *AdapterContext) error { return nil }

func (a *AntigravityAdapter) RunHeadlessInit(ctx *AdapterContext, prompt string) error { return nil }

func readJSONAsset(ctx *AdapterContext, sourcePath string) (map[string]any, error) {
	var data []byte
	var err error
	if ctx.LibraryFS != nil {
		data, err = fs.ReadFile(ctx.LibraryFS, sourcePath)
	} else {
		data, err = files.ReadFile(filepath.Join(ctx.LibraryDir, sourcePath))
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", sourcePath, err)
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("parse %s: %w", sourcePath, err)
	}
	return payload, nil
}
