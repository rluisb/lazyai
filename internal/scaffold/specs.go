package scaffold

import (
	"log"
	"path/filepath"

	"github.com/ricardoborges-teachable/ai-setup/internal/conflict"
	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// ScaffoldSpecs creates the specs/ directory structure and copies AGENTS.md files
// for selected subdirectories. Ported from src/scaffold/specs.ts.
func ScaffoldSpecs(targetDir string, setupScope types.SetupScope, libraryDir string, specsDirs, specsAgents []string, fileRecords *[]types.TrackedFile, strategy types.ConflictStrategy, perFileOverrides map[string]types.ConflictStrategy) error {
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

	// 2. Copy AGENTS.md files for selected specs directories.
	for _, dir := range specsAgents {
		src := filepath.Join(libraryDir, "specs-agents", dir+".md")
		dest := filepath.Join(specsDir, dir, "AGENTS.md")

		if !files.FileExists(src) {
			continue
		}

		relPath, err := filepath.Rel(targetDir, dest)
		if err != nil {
			relPath = dest
		}

		action, err := conflict.ApplyStrategy(dest, strategy, perFileOverrides, targetDir)
		if err != nil {
			return err
		}
		if action == "skip" {
			log.Printf("Skipping existing file: %s", relPath)
			continue
		}

		if err := files.CopyFile(src, dest); err != nil {
			return err
		}

		hash, _ := files.FileHash(dest)
		source, _ := filepath.Rel(targetDir, src)
		*fileRecords = append(*fileRecords, types.TrackedFile{
			Path:   relPath,
			Hash:   hash,
			Source: source,
			Owner:  types.FileOwnerLibrary,
		})
	}

	return nil
}
