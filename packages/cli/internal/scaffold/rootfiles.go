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
