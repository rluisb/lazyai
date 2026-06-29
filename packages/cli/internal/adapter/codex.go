package adapter

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// CodexAdapter installs OpenAI Codex CLI's native surfaces. Codex reads
// project instructions from AGENTS.md (emitted by the scaffold root layer),
// MCP servers from .codex/config.toml [mcp_servers.<name>] (compiled by
// CompileMCP), custom subagents from .codex/agents/<name>.toml, lifecycle
// hooks from .codex/hooks.json, and Agent Skills from .agents/skills/<name>/
// SKILL.md (the open agent-skills standard, read natively).
// Refs: https://developers.openai.com/codex/config-reference,
// /codex/subagents, /codex/hooks, /codex/skills.
type CodexAdapter struct{}

func (a *CodexAdapter) ID() types.ToolId  { return types.ToolIdCodex }
func (a *CodexAdapter) Name() string      { return "Codex" }
func (a *CodexAdapter) ConfigDir() string { return ".codex" }

func (a *CodexAdapter) Install(ctx *AdapterContext) ([]types.TrackedFile, error) {
	if !IsScopeSupported(types.ToolIdCodex, ctx.SetupScope) {
		return ctx.FileRecords, nil
	}
	codexDir, err := ResolveToolRoot(types.ToolIdCodex, ctx.SetupScope, ctx)
	if err != nil {
		return nil, err
	}

	if err := files.EnsureDir(filepath.Join(codexDir, "agents")); err != nil {
		return nil, err
	}

	// Custom subagents: canonical markdown agents become Codex TOML config
	// layers at .codex/agents/<name>.toml. The default "guide" agent is the
	// session itself (AGENTS.md), so it is excluded from the subagent set.
	if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "canonical/agents",
		SelectionKey: "agents",
		ToDestPath: func(file string) string {
			return filepath.Join(codexDir, "agents", fileID(file)+".toml")
		},
		IncludeFile: func(file string) bool {
			return !isDefaultAgentFile(file) && isCanonicalAgentFile(file)
		},
		Transform: func(content []byte) []byte {
			out, err := RewriteAgentForCodex(content, ctx)
			if err != nil {
				adapterLog.Warn("codex agent rewrite fell back to verbatim copy", "adapter", "codex", "error", err)
				return content
			}
			return out
		},
	}); err != nil {
		return nil, err
	}

	// Agent Skills: Codex discovers .agents/skills/<name>/SKILL.md from the
	// repo root (project/workspace) or $HOME/.agents/skills (global). Both
	// resolve as <parent-of-.codex>/.agents/skills.
	skillsDir := filepath.Join(filepath.Dir(codexDir), ".agents", "skills")
	if err := files.EnsureDir(skillsDir); err != nil {
		return nil, err
	}
	if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "skills",
		SelectionKey: "skills",
		ToDestPath: func(file string) string {
			return filepath.Join(skillsDir, fileID(file), "SKILL.md")
		},
	}); err != nil {
		return nil, err
	}

	// Lifecycle hooks: copy hook scripts under .codex/hooks/lazyai/ and write
	// the Codex-native .codex/hooks.json event configuration.
	if err := CopyLibraryDirectory(CopyLibraryDirectoryOption{
		Ctx:          ctx,
		SourceSubdir: "codex/hooks",
		Recursive:    true,
		ToDestPath: func(file string) string {
			return filepath.Join(codexDir, "hooks", file)
		},
		Mode: 0o755,
	}); err != nil {
		return nil, err
	}

	hooksConfig, err := readJSONAsset(ctx, "codex/hooks.json")
	if err != nil {
		return nil, err
	}
	hooksPayload, err := json.MarshalIndent(hooksConfig, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal codex/hooks.json: %w", err)
	}
	hooksJSONPath := filepath.Join(codexDir, "hooks.json")
	if err := files.WriteFile(hooksJSONPath, hooksPayload, 0o644); err != nil {
		return nil, err
	}
	if err := trackFile(ctx, hooksJSONPath, "codex/hooks.json"); err != nil {
		return nil, err
	}

	return ctx.FileRecords, nil
}

func (a *CodexAdapter) CompileMCP(ctx CompileContext) ([]types.TrackedFile, error) {
	return CompileMCPForTool(types.ToolIdCodex, ctx)
}

func (a *CodexAdapter) CanRunHeadless() bool { return false }

func (a *CodexAdapter) RunHeadlessValidation(ctx *AdapterContext) error { return nil }

func (a *CodexAdapter) RunHeadlessInit(ctx *AdapterContext, prompt string) error { return nil }
