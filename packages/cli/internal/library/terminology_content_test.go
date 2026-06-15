package library

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSpeckitClarifyIncludesTerminologyGuidance(t *testing.T) {
	t.Parallel()

	content := readLibraryFile(t, "skills/speckit-clarify.md")

	required := []string{
		"domain vocabulary",
		"KNOWLEDGE_MAP.md",
		"human",
		"ambiguous terms",
	}
	assertContainsAll(t, "skills/speckit-clarify.md", content, required)
}

func TestKnowledgeMapIncludesLightweightTerminologySection(t *testing.T) {
	t.Parallel()

	content := readRootKnowledgeMap(t)

	required := []string{
		"## Terminology",
		"Accepted domain terms",
		"Vocabulary source of truth",
	}
	assertContainsAll(t, "KNOWLEDGE_MAP.md", content, required)
}

func TestTerminologyGuidanceDoesNotIntroduceRuntimeInfrastructure(t *testing.T) {
	t.Parallel()

	contents := map[string]string{
		"skills/speckit-clarify.md": readLibraryFile(t, "skills/speckit-clarify.md"),
		"KNOWLEDGE_MAP.md":          readRootKnowledgeMap(t),
	}
	for path, content := range contents {
		content := strings.ToLower(content)
		forbidden := []string{
			"terminology database",
			"glossary engine",
			"new top-level skill",
			"runtime terminology service",
		}
		for _, phrase := range forbidden {
			if strings.Contains(content, phrase) {
				t.Errorf("%s should not introduce terminology infrastructure with phrase %q", path, phrase)
			}
		}
	}
}

func readRootKnowledgeMap(t *testing.T) string {
	t.Helper()

	libraryDir, err := FindLibraryDir()
	if err == nil && libraryDir != "" {
		repoRoot := filepath.Dir(filepath.Dir(filepath.Dir(libraryDir)))
		path := filepath.Join(repoRoot, "KNOWLEDGE_MAP.md")
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		return string(content)
	}

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	for {
		path := filepath.Join(dir, "KNOWLEDGE_MAP.md")
		content, err := os.ReadFile(path)
		if err == nil {
			return string(content)
		}
		if !os.IsNotExist(err) {
			t.Fatalf("read %s: %v", path, err)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("KNOWLEDGE_MAP.md not found walking upward from test working directory")
		}
		dir = parent
	}
}
