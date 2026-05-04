package scout

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDetectLanguagesCountsAndExcludes(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "src/app.ts", "export const app = true")
	writeTestFile(t, dir, "src/page.tsx", "export default function Page() { return null }")
	writeTestFile(t, dir, "cmd/server/main.go", "package main")
	writeTestFile(t, dir, "README.md", "# Example")
	writeTestFile(t, dir, "node_modules/ignored.js", "console.log('ignored')")
	writeTestFile(t, dir, ".git/ignored.go", "package ignored")

	got := DetectLanguages(dir, DefaultExcludedDirs)
	if len(got) == 0 {
		t.Fatalf("DetectLanguages returned no languages")
	}
	if got[0].Name != "TypeScript" || got[0].FileCount != 2 {
		t.Fatalf("primary language = %#v, want TypeScript with 2 files", got[0])
	}

	byName := languageEntriesByName(got)
	if byName["JavaScript"].FileCount != 0 {
		t.Fatalf("excluded node_modules JavaScript was counted: %#v", byName["JavaScript"])
	}
	if byName["Go"].FileCount != 1 {
		t.Fatalf("Go file count = %d, want 1", byName["Go"].FileCount)
	}
	if byName["Markdown"].FileCount != 1 {
		t.Fatalf("Markdown file count = %d, want 1", byName["Markdown"].FileCount)
	}
	if !reflect.DeepEqual(byName["TypeScript"].Extensions, []string{".ts", ".tsx"}) {
		t.Fatalf("TypeScript extensions = %#v, want [.ts .tsx]", byName["TypeScript"].Extensions)
	}
}

func languageEntriesByName(entries []LanguageEntry) map[string]LanguageEntry {
	byName := make(map[string]LanguageEntry, len(entries))
	for _, entry := range entries {
		byName[entry.Name] = entry
	}
	return byName
}

func writeTestFile(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%s): %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}
