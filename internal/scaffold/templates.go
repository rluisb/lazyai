package scaffold

import (
	"log"
	"path/filepath"

	"github.com/ricardoborges-teachable/ai-setup/internal/conflict"
	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// ScaffoldTemplatesRules installs template and rule files into the specs directory.
// Ported from src/scaffold/templates-rules.ts.
func ScaffoldTemplatesRules(targetDir, libraryDir string, templates []types.TemplateId, rules []types.RuleId, fileRecords *[]types.TrackedFile, strategy types.ConflictStrategy, perFileOverrides map[string]types.ConflictStrategy) error {
	specsDir := filepath.Join(targetDir, "specs")

	// Copy selected templates.
	if len(templates) > 0 {
		templatesDir := filepath.Join(specsDir, "templates")
		if err := files.EnsureDir(templatesDir); err != nil {
			return err
		}

		for _, templateId := range templates {
			src := filepath.Join(libraryDir, "templates", string(templateId)+".md")
			dest := filepath.Join(templatesDir, string(templateId)+".md")
			if err := copyLibraryFileWithRecord(src, dest, targetDir, libraryDir, fileRecords, strategy, perFileOverrides); err != nil {
				return err
			}
		}
	}

	// Copy selected rules.
	if len(rules) > 0 {
		rulesDir := filepath.Join(specsDir, "rules")
		if err := files.EnsureDir(rulesDir); err != nil {
			return err
		}

		for _, ruleId := range rules {
			src := filepath.Join(libraryDir, "rules", string(ruleId)+".md")
			dest := filepath.Join(rulesDir, string(ruleId)+".md")
			if err := copyLibraryFileWithRecord(src, dest, targetDir, libraryDir, fileRecords, strategy, perFileOverrides); err != nil {
				return err
			}
		}
	}

	// Always copy prompts/local-examples directory.
	srcDir := filepath.Join(libraryDir, "prompts/local-examples")
	destDir := filepath.Join(specsDir, "prompts/local-examples")
	return copyLibraryDir(srcDir, destDir, targetDir, libraryDir, fileRecords, strategy, perFileOverrides)
}

// copyLibraryFileWithRecord copies a single file with conflict resolution and tracking.
func copyLibraryFileWithRecord(src, dest, targetDir, libraryDir string, fileRecords *[]types.TrackedFile, strategy types.ConflictStrategy, perFileOverrides map[string]types.ConflictStrategy) error {
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
	source, _ := filepath.Rel(libraryDir, src)
	*fileRecords = append(*fileRecords, types.TrackedFile{
		Path:   relPath,
		Hash:   hash,
		Source: source,
		Owner:  types.FileOwnerLibrary,
	})
	return nil
}

// copyLibraryDir recursively copies a directory with conflict resolution and tracking.
func copyLibraryDir(src, dest, targetDir, libraryDir string, fileRecords *[]types.TrackedFile, strategy types.ConflictStrategy, perFileOverrides map[string]types.ConflictStrategy) error {
	if !files.FileExists(src) {
		return nil
	}

	if err := files.EnsureDir(dest); err != nil {
		return err
	}

	entries := files.ListDir(src)
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry)
		destPath := filepath.Join(dest, entry)

		if files.IsDirectory(srcPath) {
			if err := copyLibraryDir(srcPath, destPath, targetDir, libraryDir, fileRecords, strategy, perFileOverrides); err != nil {
				return err
			}
			continue
		}

		if err := copyLibraryFileWithRecord(srcPath, destPath, targetDir, libraryDir, fileRecords, strategy, perFileOverrides); err != nil {
			return err
		}
	}

	return nil
}
