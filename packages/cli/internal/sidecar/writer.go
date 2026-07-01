package sidecar

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"gopkg.in/yaml.v3"
)

// WriteSidecarAt writes <scopeRoot>/.lazyai/sidecar.yaml, creating
// <scopeRoot>/.lazyai/ (0o755) if needed, via the existing
// files.AtomicWriteFile (temp+fsync+rename+single-slot .bak).
func WriteSidecarAt(scopeRoot string, cfg *SidecarConfig) error {
	dir := filepath.Join(scopeRoot, ".lazyai")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating sidecar directory: %w", err)
	}

	path := filepath.Join(dir, "sidecar.yaml")
	data, err := yaml.Marshal(SidecarFile{Sidecar: cfg})
	if err != nil {
		return fmt.Errorf("marshaling sidecar: %w", err)
	}

	if _, err := files.AtomicWriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing sidecar: %w", err)
	}
	return nil
}
