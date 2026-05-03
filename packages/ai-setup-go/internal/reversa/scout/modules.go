package scout

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// DetectModules detects top-level source-bearing directories.
func DetectModules(targetDir string, excludedDirs []string) []string {
	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return nil
	}
	excluded := excludedDirSet(excludedDirs)
	modules := make([]string, 0)
	for _, entry := range entries {
		name := entry.Name()
		if !entry.IsDir() || strings.HasPrefix(name, ".") || isExcludedDir(name, excluded) {
			continue
		}
		if directoryContainsSource(filepath.Join(targetDir, name), excluded) {
			modules = append(modules, name)
		}
	}
	sort.Strings(modules)
	return modules
}

func directoryContainsSource(dir string, excluded map[string]struct{}) bool {
	sourceExtensions := map[string]struct{}{
		".ts": {}, ".tsx": {}, ".js": {}, ".jsx": {}, ".go": {}, ".py": {}, ".rb": {}, ".java": {}, ".rs": {}, ".cs": {}, ".swift": {}, ".kt": {}, ".c": {}, ".h": {}, ".cpp": {}, ".cc": {}, ".scala": {}, ".ex": {}, ".exs": {}, ".elm": {}, ".vue": {}, ".svelte": {}, ".php": {}, ".dart": {},
	}
	found := false
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || found {
			return nil
		}
		if d.IsDir() {
			if path != dir && isExcludedDir(d.Name(), excluded) {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		if _, ok := sourceExtensions[strings.ToLower(filepath.Ext(d.Name()))]; ok {
			found = true
		}
		return nil
	})
	return found
}
