// Package adapter — scope.go resolves the on-disk root each adapter writes
// into for a given (tool, scope) pair. It is the single source of truth that
// replaces the scattered `if isGlobal` branches previously duplicated across
// every adapter.
package adapter

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ricardoborges-teachable/ai-setup/internal/globalpaths"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// ErrScopeUnsupported is returned when a tool has no meaningful layout for the
// requested scope (e.g. GitHub Copilot at global scope).
var ErrScopeUnsupported = errors.New("scope not supported for this tool")

// IsScopeSupported reports whether the given tool has a defined on-disk layout
// for the given scope. Wizard + non-interactive callers use this to filter the
// tool list before invoking adapters.
func IsScopeSupported(tool types.ToolId, scope types.SetupScope) bool {
	if !types.IsValidSetupScope(scope) || !types.IsValidToolId(tool) {
		return false
	}
	if tool == types.ToolIdCopilot && scope == types.SetupScopeGlobal {
		return false
	}
	return true
}

// ResolveToolRoot returns the primary directory the adapter should write into
// for the given (tool, scope) pair. It reads ctx.TargetDir for project and
// workspace scopes (treated identically — workspace is "project-shaped layout
// rooted at the user-selected workspace directory") and ctx.HomeDir for global
// scope. If ctx.HomeDir is empty, falls back to os.UserHomeDir().
//
// Codex has two logical roots (config vs. skills); callers that need both
// should use ResolveCodexRoots instead.
func ResolveToolRoot(tool types.ToolId, scope types.SetupScope, ctx *AdapterContext) (string, error) {
	if !IsScopeSupported(tool, scope) {
		return "", fmt.Errorf("%w: tool=%s scope=%s", ErrScopeUnsupported, tool, scope)
	}
	if ctx == nil {
		return "", fmt.Errorf("ResolveToolRoot: nil AdapterContext")
	}

	switch scope {
	case types.SetupScopeProject, types.SetupScopeWorkspace:
		return filepath.Join(ctx.TargetDir, projectSubdir(tool)), nil
	case types.SetupScopeGlobal:
		home, err := resolveHomeDir(ctx)
		if err != nil {
			return "", err
		}
		root, err := globalpaths.ResolveGlobalToolTargetDir(tool, home)
		if err != nil {
			return "", err
		}
		if root == "" {
			return "", fmt.Errorf("%w: tool=%s scope=global", ErrScopeUnsupported, tool)
		}
		return root, nil
	}
	return "", fmt.Errorf("%w: unknown scope %q", ErrScopeUnsupported, scope)
}

// ResolveCodexRoots returns the two distinct directories Codex actually reads:
// configRoot holds config.toml + AGENTS.md (or AGENTS.md at the repo root for
// project/workspace scope); skillsRoot holds <name>/SKILL.md subdirs.
//
// Upstream split (locked decision 3):
//   - project/workspace: configRoot=<target>/.codex, skillsRoot=<target>/.agents/skills
//   - global:            configRoot=~/.codex,        skillsRoot=~/.agents/skills
func ResolveCodexRoots(scope types.SetupScope, ctx *AdapterContext) (configRoot, skillsRoot string, err error) {
	if ctx == nil {
		return "", "", fmt.Errorf("ResolveCodexRoots: nil AdapterContext")
	}
	switch scope {
	case types.SetupScopeProject, types.SetupScopeWorkspace:
		configRoot = filepath.Join(ctx.TargetDir, ".codex")
		skillsRoot = filepath.Join(ctx.TargetDir, ".agents", "skills")
		return configRoot, skillsRoot, nil
	case types.SetupScopeGlobal:
		home, herr := resolveHomeDir(ctx)
		if herr != nil {
			return "", "", herr
		}
		configRoot = filepath.Join(home, ".codex")
		skillsRoot = globalpaths.ResolveCodexSkillsGlobalDir(home)
		return configRoot, skillsRoot, nil
	}
	return "", "", fmt.Errorf("%w: unknown scope %q", ErrScopeUnsupported, scope)
}

// projectSubdir returns the directory name a tool expects under the target
// (project or workspace root). Codex is a special case: it returns ".codex"
// (for config.toml only) — callers that need skills as well should use
// ResolveCodexRoots.
func projectSubdir(tool types.ToolId) string {
	switch tool {
	case types.ToolIdClaudeCode:
		return ".claude"
	case types.ToolIdOpenCode:
		return ".opencode"
	case types.ToolIdGemini:
		return ".gemini"
	case types.ToolIdCopilot:
		return ".github"
	case types.ToolIdCodex:
		return ".codex"
	}
	return ""
}

func resolveHomeDir(ctx *AdapterContext) (string, error) {
	if ctx.HomeDir != "" {
		return ctx.HomeDir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("determine home dir: %w", err)
	}
	return home, nil
}
