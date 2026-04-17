package scaffold

import (
	"io/fs"
	"path/filepath"

	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// ScaffoldSpecs creates the specs/ directory structure without subdirectory AGENTS.md files.
func ScaffoldSpecs(targetDir string, setupScope types.SetupScope, libFS fs.FS, specsDirs []string, fileRecords *[]types.TrackedFile, strategy types.ConflictStrategy, perFileOverrides map[string]types.ConflictStrategy) error {
	_ = libFS
	_ = fileRecords
	_ = strategy
	_ = perFileOverrides
	specsDir := filepath.Join(targetDir, "specs")

	// 1. Create selected specs directories.
	for _, dir := range specsDirs {
		dirPath := filepath.Join(specsDir, dir)
		if err := files.EnsureDir(dirPath); err != nil {
			return err
		}

		// Special case: memory also needs handoffs subdirectory.
		if dir == "memory" {
			if setupScope == types.SetupScopeWorkspace {
				_ = files.EnsureDir(filepath.Join(specsDir, "memory", "decisions"))
				_ = files.EnsureDir(filepath.Join(specsDir, "memory", "patterns"))
				_ = files.EnsureDir(filepath.Join(specsDir, "memory", "projects"))
			}
			_ = files.EnsureDir(filepath.Join(specsDir, "memory", "handoffs"))
		}
	}

	return nil
}
