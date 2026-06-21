package adapter

import (
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/frontmatter"
	"github.com/rluisb/lazyai/packages/cli/internal/library"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// TestRewriteAgentForClaudeCode_EmitsDescription locks the contract that
// the Claude transform preserves the source `description:` field. Without
// this guard, dropping the description emit in claudeCodeFrontmatter
// silently regresses to the pre-#208 behaviour where Claude Code rejects
// installed agents.
func TestRewriteAgentForClaudeCode_EmitsDescription(t *testing.T) {
	source := []byte("---\nname: Builder\ndescription: Coordinates feature builds.\ntier: balanced\ntemperature: 0.7\nthinking: low\nrisk: 3\n---\n\nbody\n")
	ctx := &AdapterContext{LibraryFS: createTestFS()}
	out, err := RewriteAgentForClaudeCode(source, ctx)
	if err != nil {
		t.Fatalf("RewriteAgentForClaudeCode: %v", err)
	}
	fm, _, err := frontmatter.ExtractFrontmatter(out)
	if err != nil {
		t.Fatalf("extract frontmatter: %v", err)
	}
	if fm["description"] != "Coordinates feature builds." {
		t.Errorf("description mismatch: got %v", fm["description"])
	}
	if _, ok := fm["model"]; ok {
		t.Errorf("Claude frontmatter should not include model; got %v", fm["model"])
	}
	if _, ok := fm["temperature"]; ok {
		t.Errorf("Claude frontmatter should not include temperature; got %v", fm["temperature"])
	}
}

// TestLibraryCanonicalAgentsHaveDescription asserts every active canonical
// source agent in the embedded library declares a non-empty `description:`
// field. Claude Code rejects agents that omit it, so missing this field on
// any canonical agent produces a parse error on fresh install.
func TestLibraryCanonicalAgentsHaveDescription(t *testing.T) {
	libFS := library.GetLibraryFS()
	if libFS == nil {
		t.Fatal("library.GetLibraryFS returned nil")
	}
	entries, err := fs.ReadDir(libFS, "canonical/agents")
	if err != nil {
		t.Fatalf("read canonical agents dir: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("library canonical agents directory is empty")
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		data, err := fs.ReadFile(libFS, "canonical/agents/"+e.Name())
		if err != nil {
			t.Errorf("read %q: %v", e.Name(), err)
			continue
		}
		fm, _, err := frontmatter.ExtractFrontmatter(data)
		if err != nil {
			t.Errorf("%s: parse frontmatter: %v", e.Name(), err)
			continue
		}
		desc, _ := fm["description"].(string)
		if strings.TrimSpace(desc) == "" {
			t.Errorf("%s: missing required 'description:' frontmatter (Claude Code rejects agents without it)", e.Name())
		}
	}
}

// TestLibrarySkillsFrontmatterParses asserts every source skill frontmatter
// block is valid YAML. Pi, Claude Code, and other adapters copy these files
// into tool-native SKILL.md surfaces, so invalid source YAML breaks agent
// startup after install. This catches unquoted flow-sequence scalars like
// specs/{NNN-slug}/ or [T001-T005].
func TestLibrarySkillsFrontmatterParses(t *testing.T) {
	libFS := library.GetLibraryFS()
	if libFS == nil {
		t.Fatal("library.GetLibraryFS returned nil")
	}

	if err := fs.WalkDir(libFS, "skills", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			t.Errorf("walk %q: %v", path, err)
			return nil
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		data, err := fs.ReadFile(libFS, path)
		if err != nil {
			t.Errorf("read %q: %v", path, err)
			return nil
		}
		if _, _, err := frontmatter.ExtractFrontmatter(data); err != nil {
			t.Errorf("%s: parse frontmatter: %v", path, err)
		}
		return nil
	}); err != nil {
		t.Fatalf("walk skills: %v", err)
	}
}

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

// validateAgentsSchemas checks agent frontmatter for required Claude Code
// fields and tool delimiter conventions. Claude Code rejects an agent file
// without `name` and `description`, so missing either is a parse error on
// fresh install (#208).
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

		// Required: name, description (Claude Code rejects agents without these — #208)
		for _, field := range []string{"name", "description"} {
			v, ok := fm[field]
			if !ok {
				t.Errorf("agent %q: missing required '%s' field", f, field)
				continue
			}
			if s, _ := v.(string); strings.TrimSpace(s) == "" {
				t.Errorf("agent %q: '%s' field is empty", f, field)
			}
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
