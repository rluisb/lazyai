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
	// Skills placement is scope-dependent. Workspace/project installs use the
	// Antigravity workspace skills dir (.agents/skills). Global installs MUST use
	// the documented global skills root (~/.gemini/config/skills), not
	// ~/.agents/skills, which Antigravity does not discover (#486 gap 1).
	skillsDir := filepath.Join(agentsRoot, "skills")
	if ctx.SetupScope == types.SetupScopeGlobal {
		skillsDir = filepath.Join(geminiDir, "config", "skills")
	}
	if err := files.EnsureDir(geminiDir); err != nil {
		return nil, err
	}
	if err := files.EnsureDir(hooksDir); err != nil {
		return nil, err
	}
	if err := files.EnsureDir(skillsDir); err != nil {
		return nil, err
	}

	defaultSettings, err := readJSONAsset(ctx, "antigravity/settings.json")
	if err != nil {
		return nil, err
	}
	if _, err := configmerge.MergeJSONFile(settingsPath, defaultSettings); err != nil {
		return nil, err
	}

	relPath, err := filepath.Rel(ctx.TargetDir, settingsPath)
	if err != nil {
		return nil, fmt.Errorf("record antigravity/settings.json path: %w", err)
	}
	relPath = filepath.ToSlash(relPath)
	hash, err := files.FileHash(settingsPath)
	if err != nil {
		return nil, fmt.Errorf("hash antigravity/settings.json: %w", err)
	}
	ctx.FileRecords = append(ctx.FileRecords, types.TrackedFile{
		Path: relPath, Hash: hash, Source: "antigravity/settings.json", Owner: types.FileOwnerLibrary,
	})

	hooksJSONPath := filepath.Join(agentsRoot, "hooks.json")

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

	hooksConfig, err := readJSONAsset(ctx, "antigravity/hooks.json")
	if err != nil {
		return nil, err
	}
	hooksPayload, err := json.MarshalIndent(hooksConfig, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal antigravity/hooks.json: %w", err)
	}
	if err := files.WriteFile(hooksJSONPath, hooksPayload, 0o644); err != nil {
		return nil, err
	}

	relHooksPath, err := filepath.Rel(ctx.TargetDir, hooksJSONPath)
	if err != nil {
		return nil, fmt.Errorf("record antigravity/hooks.json path: %w", err)
	}
	relHooksPath = filepath.ToSlash(relHooksPath)
	hash, err = files.FileHash(hooksJSONPath)
	if err != nil {
		return nil, fmt.Errorf("hash antigravity/hooks.json: %w", err)
	}
	ctx.FileRecords = append(ctx.FileRecords, types.TrackedFile{
		Path: relHooksPath, Hash: hash, Source: "antigravity/hooks.json", Owner: types.FileOwnerLibrary,
	})

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

	// Antigravity IDE discovers workspace rules from .agents/rules/*.md, not from
	// a bare root AGENTS.md. Emit a thin rules file that imports the canonical
	// AGENTS.md via the repo-relative @ mention so IDE workspaces pick up project
	// instructions (#486 gap 2). Global rules live in ~/.gemini/GEMINI.md and are
	// user-managed, so this is skipped at global scope.
	if ctx.SetupScope != types.SetupScopeGlobal {
		rulesDir := filepath.Join(agentsRoot, "rules")
		if err := files.EnsureDir(rulesDir); err != nil {
			return nil, err
		}
		rulesPath := filepath.Join(rulesDir, "lazyai.md")
		rulesBody := []byte("# Project Rules\n\n" +
			"Canonical agent instructions for this project live in AGENTS.md at the\n" +
			"repository root. They are imported below so the Antigravity Agent applies\n" +
			"them as a workspace rule.\n\n" +
			"@/AGENTS.md\n")
		if err := files.WriteFile(rulesPath, rulesBody, 0o644); err != nil {
			return nil, err
		}
		relRulesPath, err := filepath.Rel(ctx.TargetDir, rulesPath)
		if err != nil {
			return nil, fmt.Errorf("record antigravity rules path: %w", err)
		}
		relRulesPath = filepath.ToSlash(relRulesPath)
		rulesHash, err := files.FileHash(rulesPath)
		if err != nil {
			return nil, fmt.Errorf("hash antigravity rules: %w", err)
		}
		ctx.FileRecords = append(ctx.FileRecords, types.TrackedFile{
			Path: relRulesPath, Hash: rulesHash, Source: "compiled:antigravity-rules", Owner: types.FileOwnerLibrary,
		})
	}
	return ctx.FileRecords, nil
}

func (a *AntigravityAdapter) CompileMCP(ctx CompileContext) ([]types.TrackedFile, error) {
	return CompileMCPForTool(types.ToolIdAntigravity, ctx)
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
