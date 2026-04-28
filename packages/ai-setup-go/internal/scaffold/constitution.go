package scaffold

import (
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	"github.com/ricardoborges-teachable/ai-setup/internal/conflict"
	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// ScaffoldConstitution installs the constitution to .specify/memory/constitution.md.
// Spec 022 merged the 4 separate files into a single constitution.template.md.
func ScaffoldConstitution(targetDir string, libFS fs.FS, projectName string, fileRecords *[]types.TrackedFile, strategy types.ConflictStrategy, perFileOverrides map[string]types.ConflictStrategy) error {
	// Write to .specify/memory/ (speckit-compatible path)
	memoryDir := filepath.Join(targetDir, ".specify", "memory")
	if err := files.EnsureDir(memoryDir); err != nil {
		return err
	}

	templateRelPath := "constitution/constitution.template.md"
	if !files.ExistsFS(libFS, templateRelPath) {
		log.Printf("Warning: constitution template not found at %s, skipping", templateRelPath)
		return nil
	}

	data, err := files.ReadFS(libFS, templateRelPath)
	if err != nil {
		return err
	}

	content := strings.ReplaceAll(string(data), "[YOUR_PROJECT_NAME]", projectName)
	dest := filepath.Join(memoryDir, "constitution.md")
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
		Path:   filepath.ToSlash(relPath),
		Hash:   hash,
		Source: templateRelPath,
		Owner:  types.FileOwnerLibrary,
	})

	return nil
}
