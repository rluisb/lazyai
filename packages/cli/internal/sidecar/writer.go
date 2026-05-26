package sidecar

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// WriteWorkspaceSidecar adds or updates the sidecar block on the active workspace entry.
func WriteWorkspaceSidecar(cfg *SidecarConfig) error {
	wsCfg, err := LoadWorkspaceConfig()
	if err != nil {
		return err
	}

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

	return saveWorkspaceConfig(wsCfg)
}

// WriteProjectSidecar writes .lazyai-sidecar.yaml in projectRoot.
func WriteProjectSidecar(projectRoot string, cfg *SidecarConfig) error {
	path := filepath.Join(projectRoot, ".lazyai-sidecar.yaml")
	if err := os.MkdirAll(projectRoot, 0755); err != nil {
		return fmt.Errorf("creating project root: %w", err)
	}

	// Backup existing file if present.
	if _, err := os.Stat(path); err == nil {
		backupPath := path + ".bak"
		if err := os.Rename(path, backupPath); err != nil {
			return fmt.Errorf("backing up existing sidecar: %w", err)
		}
	}

	data, err := yaml.Marshal(ProjectSidecarConfig{Sidecar: cfg})
	if err != nil {
		return fmt.Errorf("marshaling project sidecar: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
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

	// Backup existing file if present.
	if _, err := os.Stat(path); err == nil {
		backupPath := path + ".bak"
		if err := os.Rename(path, backupPath); err != nil {
			return fmt.Errorf("backing up existing global sidecar: %w", err)
		}
	}

	data, err := yaml.Marshal(GlobalSidecarConfig{Sidecar: cfg})
	if err != nil {
		return fmt.Errorf("marshaling global sidecar: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing global sidecar: %w", err)
	}
	return nil
}

// saveWorkspaceConfig writes the workspace config back to disk.
func saveWorkspaceConfig(cfg *WorkspaceConfig) error {
	path, err := getWorkspacesConfigPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	// Backup existing file if present.
	if _, err := os.Stat(path); err == nil {
		backupPath := path + ".bak"
		if err := os.Rename(path, backupPath); err != nil {
			return fmt.Errorf("backing up existing workspace config: %w", err)
		}
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling workspace config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing workspace config: %w", err)
	}
	return nil
}
