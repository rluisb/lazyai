package scaffold

import (
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rluisb/lazyai/packages/cli/internal/conflict"
	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// ScaffoldTemplatesRules installs template and rule files into the specs directory.
// Ported from src/scaffold/templates-rules.ts.
// When .specify/ exists (speckit mode), templates go into .specify/templates/
// instead (handled by ScaffoldSpecs) — suppress the specs/ copies to avoid
// duplicate template locations.
func ScaffoldTemplatesRules(targetDir string, libFS fs.FS, templates []types.TemplateId, rules []types.RuleId, fileRecords *[]types.TrackedFile, strategy types.ConflictStrategy, perFileOverrides map[string]types.ConflictStrategy, coverageThreshold int) error {
	// In speckit mode (.specify/ exists), all specs content lives under
	// .specify/ — suppress duplicate templates while still honoring selected
	// rules such as specs/rules/testing.md.
	if HasSpecKitStructure(targetDir) {
		templates = nil
	}

	specsDir := filepath.Join(targetDir, "specs")

	// Copy selected templates.
	if len(templates) > 0 {
		templatesDir := filepath.Join(specsDir, "templates")
		if err := files.EnsureDir(templatesDir); err != nil {
			return err
		}

		for _, templateId := range templates {
			srcRelPath := "templates/" + string(templateId) + ".md"
			dest := filepath.Join(templatesDir, string(templateId)+".md")
			if err := copyLibraryFileFromFS(libFS, srcRelPath, dest, targetDir, fileRecords, strategy, perFileOverrides, coverageThreshold); err != nil {
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
			srcRelPath := "rules/" + string(ruleId) + ".md"
			dest := filepath.Join(rulesDir, string(ruleId)+".md")
			if err := copyLibraryFileFromFS(libFS, srcRelPath, dest, targetDir, fileRecords, strategy, perFileOverrides, coverageThreshold); err != nil {
				return err
			}
		}
	}

	// Always copy prompts/local-examples directory.
	srcRelDir := "prompts/local-examples"
	destDir := filepath.Join(specsDir, "prompts/local-examples")
	return copyLibraryDirFromFS(libFS, srcRelDir, destDir, targetDir, fileRecords, strategy, perFileOverrides)
}

// copyLibraryFileFromFS copies a single file from the library FS with conflict resolution and tracking.
func copyLibraryFileFromFS(libFS fs.FS, srcRelPath, dest, targetDir string, fileRecords *[]types.TrackedFile, strategy types.ConflictStrategy, perFileOverrides map[string]types.ConflictStrategy, coverageThreshold int) error {
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
		scaffoldLog.Info("skipping existing file", "path", relPath)
		return nil
	}

	data, err := files.ReadFS(libFS, srcRelPath)
	if err != nil {
		scaffoldLog.Warn("could not read library file", "path", srcRelPath, "error", err)
		return nil
	}

	data = applyTemplateRuleSubstitutions(data, coverageThreshold)
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

// copyLibraryDirFromFS recursively copies a directory from the library FS with conflict resolution and tracking.
func copyLibraryDirFromFS(libFS fs.FS, srcRelDir, dest, targetDir string, fileRecords *[]types.TrackedFile, strategy types.ConflictStrategy, perFileOverrides map[string]types.ConflictStrategy) error {
	if !files.ExistsFS(libFS, srcRelDir) {
		return nil
	}

	if err := files.EnsureDir(dest); err != nil {
		return err
	}

	entries, err := files.ReadDirFS(libFS, srcRelDir)
	if err != nil {
		return nil // directory doesn't exist, skip
	}

	for _, entry := range entries {
		srcPath := srcRelDir + "/" + entry.Name()
		destPath := filepath.Join(dest, entry.Name())

		if entry.IsDir() {
			if err := copyLibraryDirFromFS(libFS, srcPath, destPath, targetDir, fileRecords, strategy, perFileOverrides); err != nil {
				return err
			}
			continue
		}

		if err := copyLibraryFileFromFS(libFS, srcPath, destPath, targetDir, fileRecords, strategy, perFileOverrides, 80); err != nil {
			return err
		}
	}

	return nil
}

func applyTemplateRuleSubstitutions(data []byte, coverageThreshold int) []byte {
	threshold := coverageThreshold
	if threshold < 1 || threshold > 100 {
		threshold = 80
	}
	content := strings.ReplaceAll(string(data), "{{COVERAGE_THRESHOLD}}", strconv.Itoa(threshold))
	return []byte(content)
}
