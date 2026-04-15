package scaffold

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ricardoborges-teachable/ai-setup/internal/conflict"
	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// ScaffoldInfra installs infrastructure files (pre-commit hook, compliance,
// CODEOWNERS, KNOWLEDGE_MAP). Ported from src/scaffold/infra.ts.
func ScaffoldInfra(targetDir, libraryDir, projectName string, infra []types.InfraId, fileRecords *[]types.TrackedFile, strategy types.ConflictStrategy, perFileOverrides map[string]types.ConflictStrategy) error {
	if len(infra) == 0 {
		return nil
	}

	// Process pre-commit hook.
	for _, id := range infra {
		if id == types.InfraIdPreCommit {
			if err := scaffoldPreCommit(targetDir, libraryDir, fileRecords, strategy, perFileOverrides); err != nil {
				return err
			}
		}
	}

	// Process compliance.md.
	for _, id := range infra {
		if id == types.InfraIdCompliance {
			if err := scaffoldCompliance(targetDir, libraryDir, fileRecords, strategy, perFileOverrides); err != nil {
				return err
			}
		}
	}

	// Process KNOWLEDGE_MAP.
	for _, id := range infra {
		if id == types.InfraIdKnowledgeMap {
			if err := scaffoldKnowledgeMap(targetDir, libraryDir, projectName, fileRecords, strategy, perFileOverrides); err != nil {
				return err
			}
		}
	}

	// Process CODEOWNERS.
	for _, id := range infra {
		if id == types.InfraIdCodeowners {
			if err := scaffoldCodeowners(targetDir, libraryDir, fileRecords, strategy, perFileOverrides); err != nil {
				return err
			}
		}
	}

	return nil
}

func scaffoldPreCommit(targetDir, libraryDir string, fileRecords *[]types.TrackedFile, strategy types.ConflictStrategy, perFileOverrides map[string]types.ConflictStrategy) error {
	gitDir := filepath.Join(targetDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return nil
	}

	hookDir := filepath.Join(gitDir, "hooks")
	if err := files.EnsureDir(hookDir); err != nil {
		return err
	}

	src := filepath.Join(libraryDir, "infra", "pre-commit.hook")
	dest := filepath.Join(hookDir, "pre-commit")

	return copyInfraFile(src, dest, targetDir, fileRecords, strategy, perFileOverrides)
}

func scaffoldCompliance(targetDir, libraryDir string, fileRecords *[]types.TrackedFile, strategy types.ConflictStrategy, perFileOverrides map[string]types.ConflictStrategy) error {
	specsDir := filepath.Join(targetDir, "specs")
	if err := files.EnsureDir(specsDir); err != nil {
		return err
	}

	src := filepath.Join(libraryDir, "infra", "compliance.md")
	dest := filepath.Join(specsDir, "compliance.md")

	return copyInfraFile(src, dest, targetDir, fileRecords, strategy, perFileOverrides)
}

func scaffoldKnowledgeMap(targetDir, libraryDir, projectName string, fileRecords *[]types.TrackedFile, strategy types.ConflictStrategy, perFileOverrides map[string]types.ConflictStrategy) error {
	src := filepath.Join(libraryDir, "infra", "KNOWLEDGE_MAP.template.md")
	if !files.FileExists(src) {
		return nil
	}

	data, err := files.ReadFile(src)
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
		Source: "infra/KNOWLEDGE_MAP.template.md",
		Owner:  types.FileOwnerLibrary,
	})
	return nil
}

func scaffoldCodeowners(targetDir, libraryDir string, fileRecords *[]types.TrackedFile, strategy types.ConflictStrategy, perFileOverrides map[string]types.ConflictStrategy) error {
	src := filepath.Join(libraryDir, "infra", "CODEOWNERS.template")
	dest := filepath.Join(targetDir, "CODEOWNERS")

	return copyInfraFile(src, dest, targetDir, fileRecords, strategy, perFileOverrides)
}

// copyInfraFile copies an infrastructure file with conflict resolution and tracking.
func copyInfraFile(src, dest, targetDir string, fileRecords *[]types.TrackedFile, strategy types.ConflictStrategy, perFileOverrides map[string]types.ConflictStrategy) error {
	if !files.FileExists(src) {
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
	return nil
}
