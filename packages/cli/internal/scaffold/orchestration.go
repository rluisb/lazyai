package scaffold

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/rluisb/lazyai/packages/cli/internal/conflict"
	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// ScaffoldOrchestration installs orchestration definitions (chains, teams,
// workflows, skills) to .ai/orchestration/.
// Ported from src/scaffold/orchestration.ts.
func ScaffoldOrchestration(targetDir string, libFS fs.FS, fileRecords *[]types.TrackedFile, strategy types.ConflictStrategy, perFileOverrides map[string]types.ConflictStrategy, features *types.FeatureFlags) error {
	if !files.ExistsFS(libFS, "orchestration") {
		return nil
	}

	targetRoot := filepath.Join(targetDir, ".ai", "orchestration")
	if err := files.EnsureDir(targetRoot); err != nil {
		return err
	}

	if err := copyOrchestrationTreeFS(libFS, "orchestration", targetRoot, targetDir, fileRecords, strategy, perFileOverrides, features); err != nil {
		return err
	}

	for _, extDir := range discoverExtensionDirs(targetDir) {
		extOrchestrationDir := filepath.Join(extDir, "orchestration")
		for _, category := range []string{"chains", "teams", "workflows"} {
			sourceDir := filepath.Join(extOrchestrationDir, category)
			if !files.DirExists(sourceDir) {
				continue
			}

			targetCategoryDir := filepath.Join(targetRoot, category)
			if err := files.EnsureDir(targetCategoryDir); err != nil {
				return err
			}

			if err := copyOrchestrationTreeDir(sourceDir, targetCategoryDir, targetDir, extDir, fileRecords, strategy, perFileOverrides); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyOrchestrationTreeFS recursively copies files from the library FS to the target directory.
func copyOrchestrationTreeFS(libFS fs.FS, sourceRelDir, currentTargetDir, targetDir string, fileRecords *[]types.TrackedFile, strategy types.ConflictStrategy, perFileOverrides map[string]types.ConflictStrategy, features *types.FeatureFlags) error {
	entries, err := files.ReadDirFS(libFS, sourceRelDir)
	if err != nil {
		return nil // directory doesn't exist in FS, skip
	}

	for _, entry := range entries {
		sourceRelPath := sourceRelDir + "/" + entry.Name()
		if sourceRelPath == "orchestration/chains/feature-adversarial.json" {
			continue
		}
		targetPath := filepath.Join(currentTargetDir, entry.Name())

		if entry.IsDir() {
			if err := files.EnsureDir(targetPath); err != nil {
				return err
			}
			if err := copyOrchestrationTreeFS(libFS, sourceRelPath, targetPath, targetDir, fileRecords, strategy, perFileOverrides, features); err != nil {
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
			scaffoldLog.Info("skipping existing file", "path", relPath)
			continue
		}

		// Read from library FS and write to target.
		readRelPath := selectedLibraryOrchestrationSource(libFS, sourceRelPath, features)
		data, err := files.ReadFS(libFS, readRelPath)
		if err != nil {
			scaffoldLog.Warn("could not read library file", "path", readRelPath, "error", err)
			continue
		}

		if err := files.WriteFile(targetPath, data, 0o644); err != nil {
			return err
		}

		hash, _ := files.FileHash(targetPath)
		*fileRecords = append(*fileRecords, types.TrackedFile{
			Path:   relPath,
			Hash:   hash,
			Source: readRelPath,
			Owner:  types.FileOwnerLibrary,
		})
	}

	return nil
}

func selectedLibraryOrchestrationSource(libFS fs.FS, sourceRelPath string, features *types.FeatureFlags) string {
	if sourceRelPath != "orchestration/chains/feature.json" || features == nil || !features.AdversarialDesign {
		return sourceRelPath
	}

	adversarialRelPath := "orchestration/chains/feature-adversarial.json"
	if files.ExistsFS(libFS, adversarialRelPath) {
		return adversarialRelPath
	}
	return sourceRelPath
}

func copyOrchestrationTreeDir(sourceDir, currentTargetDir, targetDir, recordSourceRoot string, fileRecords *[]types.TrackedFile, strategy types.ConflictStrategy, perFileOverrides map[string]types.ConflictStrategy) error {
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		sourcePath := filepath.Join(sourceDir, entry.Name())
		targetPath := filepath.Join(currentTargetDir, entry.Name())

		if entry.IsDir() {
			if err := files.EnsureDir(targetPath); err != nil {
				return err
			}
			if err := copyOrchestrationTreeDir(sourcePath, targetPath, targetDir, recordSourceRoot, fileRecords, strategy, perFileOverrides); err != nil {
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
			scaffoldLog.Info("skipping existing file", "path", relPath)
			continue
		}

		data, err := files.ReadFile(sourcePath)
		if err != nil {
			scaffoldLog.Warn("could not read file", "path", sourcePath, "error", err)
			continue
		}

		if err := files.WriteFile(targetPath, data, 0o644); err != nil {
			return err
		}

		hash, _ := files.FileHash(targetPath)
		sourceRecordPath, err := filepath.Rel(recordSourceRoot, sourcePath)
		if err != nil {
			sourceRecordPath = sourcePath
		}
		*fileRecords = append(*fileRecords, types.TrackedFile{
			Path:   relPath,
			Hash:   hash,
			Source: filepath.ToSlash(sourceRecordPath),
			Owner:  types.FileOwnerLibrary,
		})
	}

	return nil
}

type extensionConfig struct {
	Extensions map[string]extensionConfigEntry `toml:"extensions"`
}

type extensionConfigEntry struct {
	Path string `toml:"path"`
}

func discoverExtensionDirs(targetDir string) []string {
	configDirsByName := map[string]string{}

	for _, configPath := range extensionConfigPaths(targetDir) {
		var cfg extensionConfig
		if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
			continue
		}

		for name, entry := range cfg.Extensions {
			if entry.Path == "" {
				continue
			}

			resolved := resolveExtensionPath(targetDir, entry.Path)
			if resolved == "" || !files.DirExists(resolved) || !hasDiscoverableExtensionContent(resolved) {
				continue
			}

			configDirsByName[name] = resolved
		}
	}

	results := make([]string, 0, len(configDirsByName))
	seenNames := map[string]struct{}{}
	for name, dir := range configDirsByName {
		results = append(results, dir)
		seenNames[name] = struct{}{}
	}

	localExtRoot := filepath.Join(targetDir, ".ai", "extensions")
	entries, err := os.ReadDir(localExtRoot)
	if err != nil {
		return results
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if _, exists := seenNames[entry.Name()]; exists {
			continue
		}

		extDir := filepath.Join(localExtRoot, entry.Name())
		if hasDiscoverableExtensionContent(extDir) {
			results = append(results, extDir)
		}
	}

	return results
}

func extensionConfigPaths(targetDir string) []string {
	paths := []string{}
	if homeDir, err := os.UserHomeDir(); err == nil && homeDir != "" {
		paths = append(paths, filepath.Join(homeDir, ".config", "ai-setup", "config.toml"))
	}
	paths = append(paths, filepath.Join(targetDir, ".ai-setup.toml"))
	return paths
}

func resolveExtensionPath(targetDir, configuredPath string) string {
	if configuredPath == "" {
		return ""
	}
	if configuredPath[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		return filepath.Join(homeDir, strings.TrimPrefix(configuredPath[1:], string(filepath.Separator)))
	}
	if filepath.IsAbs(configuredPath) {
		return configuredPath
	}
	return filepath.Join(targetDir, configuredPath)
}

func hasDiscoverableExtensionContent(extDir string) bool {
	for _, dir := range []string{"agents", "skills", "prompts", "rules", "templates"} {
		dirPath := filepath.Join(extDir, dir)
		entries, err := os.ReadDir(dirPath)
		if err == nil && len(entries) > 0 {
			return true
		}
	}

	if files.FileExists(filepath.Join(extDir, "mcp.json")) {
		return true
	}

	agentsDir := filepath.Join(extDir, "agents")
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if entry.IsDir() && files.FileExists(filepath.Join(agentsDir, entry.Name(), "mcp.json")) {
			return true
		}
	}

	return false
}
