package scaffold

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/ricardoborges-teachable/ai-setup/internal/conflict"
	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// ConstitutionFiles lists the constitution template files to install.
var ConstitutionFiles = []string{
	"constitution",
	"constraints",
	"quality-gates",
	"uncertainty",
}

// ScaffoldConstitution installs constitution files to .ai/constitution/.
// Ported from src/scaffold/constitution.ts.
func ScaffoldConstitution(targetDir, libraryDir, projectName string, fileRecords *[]types.TrackedFile, strategy types.ConflictStrategy, perFileOverrides map[string]types.ConflictStrategy) error {
	constitutionDir := filepath.Join(targetDir, ".ai", "constitution")
	if err := files.EnsureDir(constitutionDir); err != nil {
		return err
	}

	for _, file := range ConstitutionFiles {
		templatePath := filepath.Join(libraryDir, "constitution", file+".template.md")
		if !files.FileExists(templatePath) {
			continue
		}

		data, err := files.ReadFile(templatePath)
		if err != nil {
			log.Printf("Warning: could not read template %s: %v", templatePath, err)
			continue
		}

		content := strings.ReplaceAll(string(data), "[YOUR_PROJECT_NAME]", projectName)
		dest := filepath.Join(constitutionDir, file+".md")
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

		if err := files.WriteFile(dest, []byte(content), 0o644); err != nil {
			return err
		}

		hash, _ := files.FileHash(dest)
		*fileRecords = append(*fileRecords, types.TrackedFile{
			Path:   relPath,
			Hash:   hash,
			Source: "constitution/" + file + ".template.md",
			Owner:  types.FileOwnerLibrary,
		})
	}

	return nil
}
