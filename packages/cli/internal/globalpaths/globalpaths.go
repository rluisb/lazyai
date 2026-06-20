// Package globalpaths resolves global and project-level paths for ai-setup.
// Ported from the TypeScript utilities in src/utils/global-paths.ts.
package globalpaths

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// GlobalConfigDir returns the directory for global ai-setup configuration.
// Uses XDG_CONFIG_HOME if set, otherwise falls back to ~/.config/opencode/.
func GlobalConfigDir() (string, error) {
	xdgConfig := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfig != "" {
		return filepath.Join(xdgConfig, "opencode"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}

	return filepath.Join(home, ".config", "opencode"), nil
}

// GlobalSetupDir returns the directory for global-scope ai-setup installs.
// This is the same as GlobalConfigDir.
func GlobalSetupDir() (string, error) {
	return GlobalConfigDir()
}

// ProjectSetupDir returns the project root directory.
// For project scope, this is simply projectDir.
func ProjectSetupDir(projectDir string) string {
	return projectDir
}

// WorkspaceSetupDir returns the workspace directory.
// For workspace scope, this is simply workspaceDir.
func WorkspaceSetupDir(workspaceDir string) string {
	return workspaceDir
}

// ResolveGlobalToolTargetDir returns the target directory for global-scope
// configuration of a given tool, or empty string if the tool doesn't support global config.
func ResolveGlobalToolTargetDir(tool types.ToolId, homeDir string) (string, error) {
	switch tool {
	case types.ToolIdOpenCode:
		return filepath.Join(homeDir, ".config", "opencode"), nil
	case types.ToolIdOmp:
		return filepath.Join(homeDir, ".omp", "agent"), nil
	case types.ToolIdClaudeCode:
		return filepath.Join(homeDir, ".claude"), nil
	case types.ToolIdCopilot:
		return filepath.Join(homeDir, ".copilot"), nil
	default:
		return "", nil
	}
}

// IsGlobalSupportedTool reports whether a tool supports file-based global config.
// Supported tools all support global config. Use adapter.IsScopeSupported for
// probe-aware gating (e.g., Copilot requires the copilot CLI or ~/.copilot/ presence).
func IsGlobalSupportedTool(tool types.ToolId) bool {
	switch tool {
	case types.ToolIdOpenCode, types.ToolIdClaudeCode, types.ToolIdCopilot, types.ToolIdOmp:
		return true
	default:
		return false
	}
}

// LogUnsupportedGlobalTool prints a message for tools that don't support global config.
// Deprecated: prefer adapter.ErrScopeUnsupported and the wizard/non-interactive scope
// filters; this helper remains for callers that still need a simple stderr warning.
func LogUnsupportedGlobalTool(tool types.ToolId) {
	// All supported tools support global scope (some conditionally).
}
