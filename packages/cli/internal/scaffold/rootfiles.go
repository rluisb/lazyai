package scaffold

import (
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/rluisb/lazyai/packages/cli/internal/conflict"
	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// deprecatedRootTemplateByFile maps output filenames to their template sources.
// Kept for backward compatibility. Ported from src/scaffold/root-files.ts.
var deprecatedRootTemplateByFile = map[string]string{
	"AGENTS.md":                       "root/AGENTS.template.md",
	".github/copilot-instructions.md": "root/copilot-instructions.template.md",
}

// ScaffoldRootFiles creates root-level AI tool configuration files using the
// deprecated template approach. Prefer ScaffoldCompiledRoot instead.
// Ported from src/scaffold/root-files.ts.
func ScaffoldRootFiles(targetDir string, libFS fs.FS, projectName string, tools []types.ToolId, fileRecords *[]types.TrackedFile, strategy types.ConflictStrategy, perFileOverrides map[string]types.ConflictStrategy) error {
	for _, tool := range tools {
		outputName, ok := RootFileByTool[tool]
		if !ok {
			continue
		}

		templateSource, ok := deprecatedRootTemplateByFile[outputName]
		if !ok {
			continue
		}

		templateRelPath := "root/" + filepath.Base(templateSource)
		if !files.ExistsFS(libFS, templateRelPath) {
			continue
		}

		data, err := files.ReadFS(libFS, templateRelPath)
		if err != nil {
			continue
		}

		content := strings.ReplaceAll(string(data), "[YOUR_PROJECT_NAME]", projectName)
		destPath := filepath.Join(targetDir, outputName)

		if strings.HasPrefix(outputName, ".github/") {
			if err := files.EnsureDir(filepath.Dir(destPath)); err != nil {
				return err
			}
		}

		relPath, err := filepath.Rel(targetDir, destPath)
		if err != nil {
			relPath = destPath
		}

		action, err := conflict.ApplyStrategy(destPath, strategy, perFileOverrides, targetDir)
		if err != nil {
			return err
		}
		if action == "skip" {
			continue
		}

		if err := files.WriteFile(destPath, []byte(content), 0o644); err != nil {
			return err
		}

		hash, _ := files.FileHash(destPath)
		*fileRecords = append(*fileRecords, types.TrackedFile{
			Path:   relPath,
			Hash:   hash,
			Source: templateSource,
			Owner:  types.FileOwnerLibrary,
		})
	}

	return nil
}

// writeRootFile is an internal helper to write a root file with conflict resolution.
func writeRootFile(dest, content, targetDir, source string, fileRecords *[]types.TrackedFile, strategy types.ConflictStrategy, perFileOverrides map[string]types.ConflictStrategy) {
	relPath, err := filepath.Rel(targetDir, dest)
	if err != nil {
		relPath = dest
	}

	action, err := conflict.ApplyStrategy(dest, strategy, perFileOverrides, targetDir)
	if err != nil || action == "skip" {
		scaffoldLog.Info("skipping root file", "path", relPath)
		return
	}

	if err := files.WriteFile(dest, []byte(content), 0o644); err != nil {
		scaffoldLog.Error("error writing root file", "path", dest, "error", err)
		return
	}

	hash, _ := files.FileHash(dest)
	*fileRecords = append(*fileRecords, types.TrackedFile{
		Path:   relPath,
		Hash:   hash,
		Source: source,
		Owner:  types.FileOwnerLibrary,
	})
}
