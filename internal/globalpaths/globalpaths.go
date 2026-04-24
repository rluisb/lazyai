// Package globalpaths resolves global and project-level paths for ai-setup.
// Ported from the TypeScript utilities in src/utils/global-paths.ts.
package globalpaths

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ricardoborges-teachable/ai-setup/internal/types"
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
func ProjectSetupDir(projectDir string) string {
	return projectDir
}

// WorkspaceSetupDir returns the workspace directory.
func WorkspaceSetupDir(workspaceDir string) string {
	return workspaceDir
}

// ResolveGlobalToolTargetDir returns the target directory for global-scope
// configuration of OpenCode.
func ResolveGlobalToolTargetDir(tool types.ToolId, homeDir string) (string, error) {
	switch tool {
	case types.ToolIdOpenCode:
		return filepath.Join(homeDir, ".config", "opencode"), nil
	default:
		return "", nil
	}
}

// IsGlobalSupportedTool reports whether a tool supports file-based global config.
func IsGlobalSupportedTool(tool types.ToolId) bool {
	return tool == types.ToolIdOpenCode
}
