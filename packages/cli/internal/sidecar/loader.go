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

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &WorkspaceConfig{}, nil
		}
		return nil, fmt.Errorf("reading workspace config: %w", err)
	}

	var cfg WorkspaceConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing workspace config %s: %w", path, err)
	}
	return &cfg, nil
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

	for i := range cfg.Workspaces {
		if cfg.Workspaces[i].Name == cfg.Active {
			return cfg.Workspaces[i].Sidecar, nil
		}
	}
	return nil, nil
}

// LoadProjectSidecar reads .lazyai-sidecar.yaml from projectRoot.
// Returns nil, nil when the file does not exist.
func LoadProjectSidecar(projectRoot string) (*SidecarConfig, error) {
	path := filepath.Join(projectRoot, ".lazyai-sidecar.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading project sidecar: %w", err)
	}

	var cfg ProjectSidecarConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing project sidecar %s: %w", path, err)
	}
	return cfg.Sidecar, nil
}

// LoadGlobalSidecar reads ~/.lazyai/sidecar.yaml.
// Returns nil, nil when the file does not exist.
func LoadGlobalSidecar() (*SidecarConfig, error) {
	dir, err := getGlobalConfigDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "sidecar.yaml")

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading global sidecar: %w", err)
	}

	var cfg GlobalSidecarConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing global sidecar %s: %w", path, err)
	}
	return cfg.Sidecar, nil
}
