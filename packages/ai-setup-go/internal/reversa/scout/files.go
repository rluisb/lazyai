package scout

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// countTotalFiles counts regular files while respecting excluded directories.
func countTotalFiles(targetDir string, excludedDirs []string) int {
	excluded := excludedDirSet(excludedDirs)
	count := 0
	_ = filepath.WalkDir(targetDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if path != targetDir && isExcludedDir(d.Name(), excluded) {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Type().IsRegular() {
			count++
		}
		return nil
	})
	return count
}

// detectConfigFiles detects common project configuration files at the root.
func detectConfigFiles(targetDir string) []string {
	patterns := []string{
		"tsconfig.json",
		".env.example",
		".env.template",
		"next.config.*",
		"vite.config.*",
		"webpack.config.*",
		"tailwind.config.*",
		"postcss.config.*",
		".eslintrc.*",
		".prettierrc*",
		"biome.json",
		".editorconfig",
		"go.work",
		"docker-compose*.yml",
		"Makefile",
		"justfile",
		"Taskfile.yml",
	}
	seen := map[string]struct{}{}
	var configs []string
	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(targetDir, pattern))
		if err != nil {
			continue
		}
		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil || info.IsDir() {
				continue
			}
			rel, relErr := filepath.Rel(targetDir, match)
			if relErr != nil {
				continue
			}
			rel = filepath.ToSlash(rel)
			if _, ok := seen[rel]; ok {
				continue
			}
			seen[rel] = struct{}{}
			configs = append(configs, rel)
		}
	}
	sort.Strings(configs)
	return configs
}
