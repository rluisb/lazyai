package scaffold

import (
	"log"
	"path/filepath"

	"github.com/ricardoborges-teachable/ai-setup/internal/conflict"
	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// ScaffoldOrchestration installs orchestration definitions (chains, teams,
// workflows, skills) to .ai/orchestration/.
// Ported from src/scaffold/orchestration.ts.
func ScaffoldOrchestration(targetDir, libraryDir string, fileRecords *[]types.TrackedFile, strategy types.ConflictStrategy, perFileOverrides map[string]types.ConflictStrategy) error {
	sourceRoot := filepath.Join(libraryDir, "orchestration")
	if !files.FileExists(sourceRoot) {
		return nil
	}

	targetRoot := filepath.Join(targetDir, ".ai", "orchestration")
	if err := files.EnsureDir(targetRoot); err != nil {
		return err
	}

	return copyOrchestrationTree(sourceRoot, targetRoot, targetDir, libraryDir, fileRecords, strategy, perFileOverrides)
}

// copyOrchestrationTree recursively copies files from sourceRoot to targetRoot.
func copyOrchestrationTree(currentSourceDir, currentTargetDir, targetDir, libraryDir string, fileRecords *[]types.TrackedFile, strategy types.ConflictStrategy, perFileOverrides map[string]types.ConflictStrategy) error {
	entries := files.ListDir(currentSourceDir)

	for _, entry := range entries {
		sourcePath := filepath.Join(currentSourceDir, entry)
		targetPath := filepath.Join(currentTargetDir, entry)

		if files.IsDirectory(sourcePath) {
			if err := files.EnsureDir(targetPath); err != nil {
				return err
			}
			if err := copyOrchestrationTree(sourcePath, targetPath, targetDir, libraryDir, fileRecords, strategy, perFileOverrides); err != nil {
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

		if err := files.CopyFile(sourcePath, targetPath); err != nil {
			return err
		}

		hash, _ := files.FileHash(targetPath)
		source, _ := filepath.Rel(libraryDir, sourcePath)
		*fileRecords = append(*fileRecords, types.TrackedFile{
			Path:   relPath,
			Hash:   hash,
			Source: source,
			Owner:  types.FileOwnerLibrary,
		})
	}

	return nil
}
