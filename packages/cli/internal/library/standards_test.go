package library

import (
	"io/fs"
	"slices"
	"strings"
	"testing"
)

func TestStarterStandardsHaveExpectedFilesAndFrontmatter(t *testing.T) {
	t.Parallel()

	libFS := GetLibraryFS()
	entries, err := fs.ReadDir(libFS, "standards/starter")
	if err != nil {
		t.Fatalf("read standards/starter: %v", err)
	}

	got := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		got = append(got, entry.Name())
	}
	slices.Sort(got)

	want := []string{
		"agent-security.md",
		"context-loading.md",
		"error-handling.md",
		"test-patterns.md",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("starter standard files = %v, want %v", got, want)
	}

	for _, name := range want {
		path := "standards/starter/" + name
		contentBytes, err := fs.ReadFile(libFS, path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}

		assertStarterStandardConventions(t, path, string(contentBytes))
	}
}

func assertStarterStandardConventions(t *testing.T, path string, content string) {
	t.Helper()

	if !strings.HasPrefix(content, "# Standard: ") {
		t.Errorf("%s: missing standard title", path)
	}
	if strings.Contains(content, "[Standard Name]") || strings.Contains(content, "[team]") || strings.Contains(content, "[pointer]") {
		t.Errorf("%s: contains unresolved standard-template placeholders", path)
	}

	frontmatter, _, ok := strings.Cut(content, "\n---\n")
	if !ok {
		t.Fatalf("%s: missing frontmatter/body separator", path)
	}

	requiredFrontmatter := []string{
		"**Category:**",
		"**Scope:**",
		"**Date:**",
		"**Owner:**",
		"**Status:**",
		"**Constitution article(s):**",
		"> **Purpose.**",
	}
	for _, required := range requiredFrontmatter {
		if !strings.Contains(frontmatter, required) {
			t.Errorf("%s: missing frontmatter field %q", path, required)
		}
	}

	requiredSections := []string{
		"## Scope Cascade",
		"## Rule",
		"**Trigger:**",
		"## Rationale",
		"## Examples",
		"**Compliant:**",
		"**Non-compliant:**",
		"## Enforcement",
		"## Exceptions",
		"## Workspace Awareness",
		"## Related",
		"## Memory",
	}
	for _, required := range requiredSections {
		if !strings.Contains(content, required) {
			t.Errorf("%s: missing standard section %q", path, required)
		}
	}
}
