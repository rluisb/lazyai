package sidecar

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// getGlobalConfigDir returns the directory for global LazyAI configuration.
func getGlobalConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to determine home directory: %w", err)
	}
	return filepath.Join(home, ".lazyai"), nil
}

// getWorkspacesConfigPath returns the path to workspaces.yaml.
func getWorkspacesConfigPath() (string, error) {
	dir, err := getGlobalConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "workspaces.yaml"), nil
}

// LoadWorkspaceConfig reads the full workspace config from disk.
// Missing file returns an empty config, not an error.
func LoadWorkspaceConfig() (*WorkspaceConfig, error) {
	path, err := getWorkspacesConfigPath()
	if err != nil {
		return nil, err
	}

	return loadWorkspaceConfigFromPath(path)
}

// LoadWorkspaceSidecar reads the active workspace's sidecar block.
// Returns nil, nil when there is no active workspace or no sidecar block.
func LoadWorkspaceSidecar() (*SidecarConfig, error) {
	cfg, err := LoadWorkspaceConfig()
	if err != nil {
		return nil, err
	}
	if cfg.Active == "" {
		return nil, nil
	}

	for _, w := range cfg.Workspaces {
		if w.Name == cfg.Active {
			if w.Sidecar == nil {
				return nil, nil
			}
			return w.Sidecar, nil
		}
	}
	return nil, nil
}

// LoadSidecarAt reads <scopeRoot>/.lazyai/sidecar.yaml.
// Returns (nil, nil) when the file does not exist — a missing sidecar at any
// scope root is not an error.
func LoadSidecarAt(scopeRoot string) (*SidecarConfig, error) {
	path := filepath.Join(scopeRoot, ".lazyai", "sidecar.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading sidecar at %s: %w", scopeRoot, err)
	}
	var file SidecarFile
	if err := yaml.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("parsing sidecar at %s: %w", scopeRoot, err)
	}
	return file.Sidecar, nil
}
