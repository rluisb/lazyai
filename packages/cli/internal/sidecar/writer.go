package sidecar

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"gopkg.in/yaml.v3"
)

// WriteWorkspaceSidecar adds or updates the sidecar block on the active workspace entry.
func WriteWorkspaceSidecar(cfg *SidecarConfig) error {
	return UpdateWorkspaceConfig(func(wsCfg *WorkspaceConfig) error {
		if wsCfg.Active == "" {
			return fmt.Errorf("no active workspace set")
		}

		found := false
		for i := range wsCfg.Workspaces {
			if wsCfg.Workspaces[i].Name == wsCfg.Active {
				wsCfg.Workspaces[i].Sidecar = cfg
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("active workspace %q not found", wsCfg.Active)
		}

		return nil
	})
}

// SaveWorkspaceConfig writes the workspace config back to disk.
func SaveWorkspaceConfig(cfg *WorkspaceConfig) error {
	path, err := getWorkspacesConfigPath()
	if err != nil {
		return err
	}
	return saveWorkspaceConfigToPath(path, cfg)
}

// UpdateWorkspaceConfig reads, mutates, and saves the workspace config under a shared lock.
func UpdateWorkspaceConfig(update func(*WorkspaceConfig) error) error {
	path, err := getWorkspacesConfigPath()
	if err != nil {
		return err
	}

	return files.WithFileLock(path+".lock", 5*time.Second, 30*time.Second, func() error {
		cfg, err := loadWorkspaceConfigFromPath(path)
		if err != nil {
			return err
		}
		if err := update(cfg); err != nil {
			return err
		}
		return saveWorkspaceConfigToPath(path, cfg)
	})
}

func loadWorkspaceConfigFromPath(path string) (*WorkspaceConfig, error) {
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

func saveWorkspaceConfigToPath(path string, cfg *WorkspaceConfig) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling workspace config: %w", err)
	}

	if _, err := files.AtomicWriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing workspace config: %w", err)
	}
	return nil
}

// WriteProjectSidecar writes .lazyai-sidecar.yaml in projectRoot.
func WriteProjectSidecar(projectRoot string, cfg *SidecarConfig) error {
	info, err := os.Stat(projectRoot)
	if err != nil {
		return fmt.Errorf("project root not accessible: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("project root is not a directory: %s", projectRoot)
	}

	path := filepath.Join(projectRoot, ".lazyai-sidecar.yaml")
	data, err := yaml.Marshal(ProjectSidecarConfig{Sidecar: cfg})
	if err != nil {
		return fmt.Errorf("marshaling project sidecar: %w", err)
	}

	if _, err := files.AtomicWriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing project sidecar: %w", err)
	}
	return nil
}

// WriteGlobalSidecar writes ~/.lazyai/sidecar.yaml.
func WriteGlobalSidecar(cfg *SidecarConfig) error {
	dir, err := getGlobalConfigDir()
	if err != nil {
		return err
	}
	path := filepath.Join(dir, "sidecar.yaml")

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating global config directory: %w", err)
	}

	data, err := yaml.Marshal(GlobalSidecarConfig{Sidecar: cfg})
	if err != nil {
		return fmt.Errorf("marshaling global sidecar: %w", err)
	}

	if _, err := files.AtomicWriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing global sidecar: %w", err)
	}
	return nil
}
