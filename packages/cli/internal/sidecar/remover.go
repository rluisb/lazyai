package sidecar

import (
	"fmt"
	"os"
	"path/filepath"
)

// RemoveWorkspaceSidecar removes the sidecar block from the active workspace entry.
// Returns nil if there is no active workspace or no sidecar block.
func RemoveWorkspaceSidecar() error {
	return UpdateWorkspaceConfig(func(wsCfg *WorkspaceConfig) error {
		if wsCfg.Active == "" {
			return nil
		}

		found := false
		for i := range wsCfg.Workspaces {
			if wsCfg.Workspaces[i].Name == wsCfg.Active {
				if wsCfg.Workspaces[i].Sidecar == nil {
					return nil
				}
				wsCfg.Workspaces[i].Sidecar = nil
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

// RemoveProjectSidecar deletes .lazyai-sidecar.yaml from projectRoot.
// Returns nil if the file does not exist.
func RemoveProjectSidecar(projectRoot string) error {
	path := filepath.Join(projectRoot, ".lazyai-sidecar.yaml")
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("removing project sidecar: %w", err)
	}
	return nil
}
