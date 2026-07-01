package sidecar

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

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
