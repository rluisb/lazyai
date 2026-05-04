package adapter

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/frontmatter"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// TestClaudeCodeFrontmatterSchemas verifies that all Claude Code artifacts
// emitted by Install conform to their required frontmatter schemas (spec 012 task 011).
func TestClaudeCodeFrontmatterSchemas(t *testing.T) {
	cases := []struct {
		name  string
		scope types.SetupScope
		root  func(target, home string) string
	}{
		{"project", types.SetupScopeProject, func(t, _ string) string { return filepath.Join(t, ".claude") }},
		{"workspace", types.SetupScopeWorkspace, func(t, _ string) string { return filepath.Join(t, ".claude") }},
		{"global", types.SetupScopeGlobal, func(_, h string) string { return filepath.Join(h, ".claude") }},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx, target, home := newScopeTestContext(t, c.scope)

			if _, err := (&ClaudeCodeAdapter{}).Install(ctx); err != nil {
				t.Fatalf("Install: %v", err)
			}

			claudeDir := c.root(target, home)

			// Test agents
			validateAgentsSchemas(t, filepath.Join(claudeDir, "agents"))

			// Test commands
			validateCommandsSchemas(t, filepath.Join(claudeDir, "commands"))

			// Test output-styles
			validateOutputStylesSchemas(t, filepath.Join(claudeDir, "output-styles"))
		})
	}
}

// validateAgentsSchemas checks agent frontmatter for tool delimiters and other constraints.
func validateAgentsSchemas(t *testing.T, dir string) {
	if !files.DirExists(dir) {
		return
	}

	for _, f := range files.ListDir(dir) {
		// Skip tool context file
		if f == "CLAUDE.md" {
			continue
		}

		path := filepath.Join(dir, f)
		if files.IsDirectory(path) {
			continue
		}

		content, err := files.ReadFile(path)
		if err != nil {
			t.Errorf("read agent %q: %v", f, err)
			continue
		}

		fm, _, parseErr := frontmatter.ExtractFrontmatter(content)
		if parseErr != nil {
			t.Errorf("agent %q: parse error: %v", f, parseErr)
			continue
		}

		// If tools present, they must be whitespace-separated for Claude (spec 012 task 004)
		if toolsVal, ok := fm["tools"]; ok {
			toolsStr := toolsVal.(string)
			// Quick check: no comma-space (space-separated, not comma-separated)
			if strings.Contains(toolsStr, ", ") {
				t.Errorf("agent %q: tools appear comma-separated (should be whitespace): %s", f, toolsStr)
			}
		}
	}
}

// validateCommandsSchemas checks command frontmatter for required fields.
func validateCommandsSchemas(t *testing.T, dir string) {
	if !files.DirExists(dir) {
		return
	}

	for _, f := range files.ListDir(dir) {
		path := filepath.Join(dir, f)
		if files.IsDirectory(path) {
			continue
		}

		content, err := files.ReadFile(path)
		if err != nil {
			t.Errorf("read command %q: %v", f, err)
			continue
		}

		fm, _, parseErr := frontmatter.ExtractFrontmatter(content)
		if parseErr != nil {
			t.Errorf("command %q: parse error: %v", f, parseErr)
			continue
		}

		// Required: description, argument-hint, allowed-tools
		required := []string{"description", "argument-hint", "allowed-tools"}
		for _, field := range required {
			if _, ok := fm[field]; !ok {
				t.Errorf("command %q: missing '%s' field", f, field)
			}
		}
	}
}

// validateOutputStylesSchemas checks output-style frontmatter for required fields.
func validateOutputStylesSchemas(t *testing.T, dir string) {
	if !files.DirExists(dir) {
		return
	}

	for _, f := range files.ListDir(dir) {
		path := filepath.Join(dir, f)
		if files.IsDirectory(path) {
			continue
		}

		content, err := files.ReadFile(path)
		if err != nil {
			t.Errorf("read output-style %q: %v", f, err)
			continue
		}

		fm, _, parseErr := frontmatter.ExtractFrontmatter(content)
		if parseErr != nil {
			t.Errorf("output-style %q: parse error: %v", f, parseErr)
			continue
		}

		// Required: name, description
		required := []string{"name", "description"}
		for _, field := range required {
			if _, ok := fm[field]; !ok {
				t.Errorf("output-style %q: missing '%s' field", f, field)
			}
		}
	}
}
