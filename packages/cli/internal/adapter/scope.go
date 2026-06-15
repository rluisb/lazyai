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

	"github.com/rluisb/lazyai/packages/cli/internal/globalpaths"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// ErrScopeUnsupported is returned when a tool has no meaningful layout for the
// requested scope (e.g. GitHub Copilot at global scope).
var ErrScopeUnsupported = errors.New("scope not supported for this tool")

// IsScopeSupported reports whether the given tool has a defined on-disk layout
// for the given scope. Wizard + non-interactive callers use this to validate
// tool selections before invoking adapters. Copilot supports global scope; the
// adapter itself probes for CLI/home presence and skips if not available. Pi
// and Antigravity are project/workspace-only surfaces.
func IsScopeSupported(tool types.ToolId, scope types.SetupScope) bool {
	if !types.IsValidSetupScope(scope) || !types.IsValidToolId(tool) {
		return false
	}
	if scope == types.SetupScopeGlobal {
		switch tool {
		case types.ToolIdPi, types.ToolIdAntigravity:
			return false
		}
	}
	return true
}

// ResolveToolRoot returns the primary directory the adapter should write into
// for the given (tool, scope) pair. For project scope, uses ctx.TargetDir.
// For workspace scope with WorkspaceRoot set, uses WorkspaceRoot; otherwise
// falls back to TargetDir (backward compat). For global scope, uses HomeDir.
// Callers should use this instead of hard-coding paths.
func ResolveToolRoot(tool types.ToolId, scope types.SetupScope, ctx *AdapterContext) (string, error) {
	if !IsScopeSupported(tool, scope) {
		return "", fmt.Errorf("%w: tool=%s scope=%s", ErrScopeUnsupported, tool, scope)
	}
	if ctx == nil {
		return "", fmt.Errorf("ResolveToolRoot: nil AdapterContext")
	}

	switch scope {
	case types.SetupScopeProject:
		return filepath.Join(ctx.TargetDir, projectSubdir(tool)), nil
	case types.SetupScopeWorkspace:
		root := ctx.TargetDir
		if ctx.WorkspaceRoot != "" {
			root = ctx.WorkspaceRoot
		}
		return filepath.Join(root, projectSubdir(tool)), nil
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

// projectSubdir returns the directory name a tool expects under the target
// (project or workspace root).
func projectSubdir(tool types.ToolId) string {
	switch tool {
	case types.ToolIdClaudeCode:
		return ".claude"
	case types.ToolIdOpenCode:
		return ".opencode"
	case types.ToolIdCopilot:
		return ".github"
	case types.ToolIdPi:
		return ".pi"
	case types.ToolIdAntigravity:
		return ".gemini"
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
