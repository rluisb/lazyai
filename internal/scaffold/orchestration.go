package scaffold

import (
	"io/fs"
	"log"
	"path/filepath"

	"github.com/ricardoborges-teachable/ai-setup/internal/conflict"
	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// ScaffoldOrchestration installs orchestration definitions (chains, teams,
// workflows, skills) to .ai/orchestration/.
// Ported from src/scaffold/orchestration.ts.
func ScaffoldOrchestration(targetDir string, libFS fs.FS, fileRecords *[]types.TrackedFile, strategy types.ConflictStrategy, perFileOverrides map[string]types.ConflictStrategy) error {
	if !files.ExistsFS(libFS, "orchestration") {
		return nil
	}

	targetRoot := filepath.Join(targetDir, ".ai", "orchestration")
	if err := files.EnsureDir(targetRoot); err != nil {
		return err
	}

	return copyOrchestrationTreeFS(libFS, "orchestration", targetRoot, targetDir, fileRecords, strategy, perFileOverrides)
}

// copyOrchestrationTreeFS recursively copies files from the library FS to the target directory.
func copyOrchestrationTreeFS(libFS fs.FS, sourceRelDir, currentTargetDir, targetDir string, fileRecords *[]types.TrackedFile, strategy types.ConflictStrategy, perFileOverrides map[string]types.ConflictStrategy) error {
	entries, err := files.ReadDirFS(libFS, sourceRelDir)
	if err != nil {
		return nil // directory doesn't exist in FS, skip
	}

	for _, entry := range entries {
		sourceRelPath := sourceRelDir + "/" + entry.Name()
		targetPath := filepath.Join(currentTargetDir, entry.Name())

		if entry.IsDir() {
			if err := files.EnsureDir(targetPath); err != nil {
				return err
			}
			if err := copyOrchestrationTreeFS(libFS, sourceRelPath, targetPath, targetDir, fileRecords, strategy, perFileOverrides); err != nil {
				return err
			}
			continue
		}

		relPath, err := filepath.Rel(targetDir, targetPath)
		if err != nil {
			relPath = targetPath
		}

		action, err := conflict.ApplyStrategy(targetPath, strategy, perFileOverrides, targetDir)
		if err != nil {
			return err
		}
		if action == "skip" {
			log.Printf("Skipping existing file: %s", relPath)
			continue
		}

		// Read from library FS and write to target.
		data, err := files.ReadFS(libFS, sourceRelPath)
		if err != nil {
			log.Printf("Warning: could not read %s: %v", sourceRelPath, err)
			continue
		}

		if err := files.WriteFile(targetPath, data, 0o644); err != nil {
			return err
		}

		hash, _ := files.FileHash(targetPath)
		*fileRecords = append(*fileRecords, types.TrackedFile{
			Path:   relPath,
			Hash:   hash,
			Source: sourceRelPath,
			Owner:  types.FileOwnerLibrary,
		})
	}

	return nil
}