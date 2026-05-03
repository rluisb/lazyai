package scout

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

// DetectLanguages walks the project tree and counts source files by language.
func DetectLanguages(targetDir string, excludedDirs []string) []LanguageEntry {
	exts := map[string]string{
		".ts":     "TypeScript",
		".tsx":    "TypeScript",
		".js":     "JavaScript",
		".go":     "Go",
		".py":     "Python",
		".rb":     "Ruby",
		".java":   "Java",
		".rs":     "Rust",
		".cs":     "C#",
		".swift":  "Swift",
		".kt":     "Kotlin",
		".c":      "C/C++",
		".h":      "C/C++",
		".cpp":    "C++",
		".cc":     "C++",
		".scala":  "Scala",
		".ex":     "Elixir",
		".exs":    "Elixir",
		".elm":    "Elm",
		".vue":    "Vue",
		".svelte": "Svelte",
		".css":    "CSS",
		".scss":   "SCSS",
		".html":   "HTML",
		".md":     "Markdown",
		".sql":    "SQL",
		".proto":  "Protobuf",
		".php":    "PHP",
		".dart":   "Dart",
	}
	counts := map[string]int{}
	extensions := map[string]map[string]struct{}{}
	excluded := excludedDirSet(excludedDirs)

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
		if !d.Type().IsRegular() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(d.Name()))
		name, ok := exts[ext]
		if !ok {
			return nil
		}
		counts[name]++
		if extensions[name] == nil {
			extensions[name] = map[string]struct{}{}
		}
		extensions[name][ext] = struct{}{}
		return nil
	})

	entries := make([]LanguageEntry, 0, len(counts))
	for name, count := range counts {
		if count == 0 {
			continue
		}
		extsForLanguage := make([]string, 0, len(extensions[name]))
		for ext := range extensions[name] {
			extsForLanguage = append(extsForLanguage, ext)
		}
		sort.Strings(extsForLanguage)
		entries = append(entries, LanguageEntry{Name: name, Extensions: extsForLanguage, FileCount: count})
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].FileCount == entries[j].FileCount {
			return entries[i].Name < entries[j].Name
		}
		return entries[i].FileCount > entries[j].FileCount
	})
	return entries
}
