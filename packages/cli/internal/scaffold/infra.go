package scaffold

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/rluisb/lazyai/packages/cli/internal/conflict"
	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// ScaffoldInfra installs infrastructure files (pre-commit hook, compliance,
// CODEOWNERS, KNOWLEDGE_MAP). Ported from src/scaffold/infra.ts.
func ScaffoldInfra(targetDir string, libFS fs.FS, projectName string, infra []types.InfraId, fileRecords *[]types.TrackedFile, strategy types.ConflictStrategy, perFileOverrides map[string]types.ConflictStrategy) error {
	if len(infra) == 0 {
		return nil
	}

	// Process pre-commit hook.
	for _, id := range infra {
		if id == types.InfraIdPreCommit {
			if err := scaffoldPreCommit(targetDir, libFS, fileRecords, strategy, perFileOverrides); err != nil {
				return err
			}
		}
	}

	// Process compliance.md.
	for _, id := range infra {
		if id == types.InfraIdCompliance {
			if err := scaffoldCompliance(targetDir, libFS, fileRecords, strategy, perFileOverrides); err != nil {
				return err
			}
		}
	}

	// Process KNOWLEDGE_MAP.
	for _, id := range infra {
		if id == types.InfraIdKnowledgeMap {
			if err := scaffoldKnowledgeMap(targetDir, libFS, projectName, fileRecords, strategy, perFileOverrides); err != nil {
				return err
			}
		}
	}

	// Process CODEOWNERS.
	for _, id := range infra {
		if id == types.InfraIdCodeowners {
			if err := scaffoldCodeowners(targetDir, libFS, fileRecords, strategy, perFileOverrides); err != nil {
				return err
			}
		}
	}

	return nil
}

func scaffoldPreCommit(targetDir string, libFS fs.FS, fileRecords *[]types.TrackedFile, strategy types.ConflictStrategy, perFileOverrides map[string]types.ConflictStrategy) error {
	gitDir := filepath.Join(targetDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return nil
	}

	hookDir := filepath.Join(gitDir, "hooks")
	if err := files.EnsureDir(hookDir); err != nil {
		return err
	}

	srcRelPath := "infra/pre-commit.hook"
	dest := filepath.Join(hookDir, "pre-commit")

	return copyInfraFileFromFS(libFS, srcRelPath, dest, targetDir, fileRecords, strategy, perFileOverrides)
}

func scaffoldCompliance(targetDir string, libFS fs.FS, fileRecords *[]types.TrackedFile, strategy types.ConflictStrategy, perFileOverrides map[string]types.ConflictStrategy) error {
	// In speckit mode (.specify/ exists), compliance data lives in
	// .specify/ rather than specs/compliance.md. Skip to avoid
	// duplicate and misplaced files.
	if HasSpecKitStructure(targetDir) {
		return nil
	}

	specsDir := filepath.Join(targetDir, "specs")
	if err := files.EnsureDir(specsDir); err != nil {
		return err
	}

	srcRelPath := "infra/compliance.md"
	dest := filepath.Join(specsDir, "compliance.md")

	return copyInfraFileFromFS(libFS, srcRelPath, dest, targetDir, fileRecords, strategy, perFileOverrides)
}

func scaffoldKnowledgeMap(targetDir string, libFS fs.FS, projectName string, fileRecords *[]types.TrackedFile, strategy types.ConflictStrategy, perFileOverrides map[string]types.ConflictStrategy) error {
	srcRelPath := "infra/KNOWLEDGE_MAP.template.md"
	if !files.ExistsFS(libFS, srcRelPath) {
		return nil
	}

	data, err := files.ReadFS(libFS, srcRelPath)
	if err != nil {
		return err
	}

	content := strings.ReplaceAll(string(data), "[YOUR_PROJECT_NAME]", projectName)
	dest := filepath.Join(targetDir, "KNOWLEDGE_MAP.md")
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
		return nil
	}

	if err := files.WriteFile(dest, []byte(content), 0o644); err != nil {
		return err
	}

	hash, _ := files.FileHash(dest)
	*fileRecords = append(*fileRecords, types.TrackedFile{
		Path:   relPath,
		Hash:   hash,
		Source: srcRelPath,
		Owner:  types.FileOwnerLibrary,
	})
	return nil
}

func scaffoldCodeowners(targetDir string, libFS fs.FS, fileRecords *[]types.TrackedFile, strategy types.ConflictStrategy, perFileOverrides map[string]types.ConflictStrategy) error {
	srcRelPath := "infra/CODEOWNERS.template"
	dest := filepath.Join(targetDir, "CODEOWNERS")

	return copyInfraFileFromFS(libFS, srcRelPath, dest, targetDir, fileRecords, strategy, perFileOverrides)
}

// copyInfraFileFromFS copies an infrastructure file from the library FS with conflict resolution and tracking.
func copyInfraFileFromFS(libFS fs.FS, srcRelPath, dest, targetDir string, fileRecords *[]types.TrackedFile, strategy types.ConflictStrategy, perFileOverrides map[string]types.ConflictStrategy) error {
	if !files.ExistsFS(libFS, srcRelPath) {
		return nil
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
		return nil
	}

	data, err := files.ReadFS(libFS, srcRelPath)
	if err != nil {
		log.Printf("Warning: could not read %s: %v", srcRelPath, err)
		return nil
	}

	if err := files.WriteFile(dest, data, 0o644); err != nil {
		return err
	}

	hash, _ := files.FileHash(dest)
	*fileRecords = append(*fileRecords, types.TrackedFile{
		Path:   relPath,
		Hash:   hash,
		Source: srcRelPath,
		Owner:  types.FileOwnerLibrary,
	})
	return nil
}
