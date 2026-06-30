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

func TestClaudeCode_InstalledCanonicalAgentsHaveRequiredFields(t *testing.T) {
	cases := []struct {
		name      string
		scope     types.SetupScope
		root      func(target, home string) string
		selection []types.AgentId
		expected  []string
	}{
		{
			name:      "default_plus_selected",
			scope:     types.SetupScopeProject,
			root:      func(t, _ string) string { return filepath.Join(t, ".claude") },
			selection: []types.AgentId{types.AgentIdResearcher},
			expected:  []string{defaultAgentID, string(types.AgentIdResearcher)},
		},
		{
			name:      "all_when_unset_selection",
			scope:     types.SetupScopeProject,
			root:      func(t, _ string) string { return filepath.Join(t, ".claude") },
			selection: nil,
			expected:  []string{"guide", "implementer", "researcher", "deployer", "responder", "planner", "reviewer", "evidence-verifier"},
		},
		{
			name:      "all_scopes_selected_default",
			scope:     types.SetupScopeWorkspace,
			root:      func(t, _ string) string { return filepath.Join(t, ".claude") },
			selection: []types.AgentId{types.AgentIdResearcher},
			expected:  []string{defaultAgentID, string(types.AgentIdResearcher)},
		},
		{
			name:      "all_scopes_selected_global_default",
			scope:     types.SetupScopeGlobal,
			root:      func(_, h string) string { return filepath.Join(h, ".claude") },
			selection: []types.AgentId{types.AgentIdResearcher},
			expected:  []string{defaultAgentID, string(types.AgentIdResearcher)},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx, target, home := newScopeTestContext(t, c.scope)
			ctx.Selections.Agents = c.selection

			if _, err := (&ClaudeCodeAdapter{}).Install(ctx); err != nil {
				t.Fatalf("Install: %v", err)
			}

			root := c.root(target, home)
			agentsDir := filepath.Join(root, "agents")

			seen := map[string]bool{}
			for _, id := range c.expected {
				if seen[id] {
					continue
				}
				seen[id] = true

				agentPath := filepath.Join(agentsDir, id+".md")
				if !files.FileExists(agentPath) {
					t.Errorf("expected agent %q at %q", id, agentPath)
					continue
				}

				content, err := files.ReadFile(agentPath)
				if err != nil {
					t.Errorf("read agent %q: %v", agentPath, err)
					continue
				}

				fm, _, err := frontmatter.ExtractFrontmatter(content)
				if err != nil {
					t.Errorf("agent %q parse frontmatter: %v", id, err)
					continue
				}

				for _, field := range []string{"name", "description"} {
					v, ok := fm[field]
					if !ok {
						t.Errorf("agent %q missing required '%s' field", id, field)
						continue
					}
					if strings.TrimSpace(v.(string)) == "" {
						t.Errorf("agent %q has empty '%s' field", id, field)
					}
				}
			}
		})
	}
}

// TestRewriteAgentForClaudeCode_ToolGrants verifies that read-only canonical
// agents emit disallowedTools: Edit Write Bash in Claude Code frontmatter, that
// full-access agents emit no restriction keys at all, and that tool names use
// PascalCase with space separation (no commas).
func TestRewriteAgentForClaudeCode_ToolGrants(t *testing.T) {
	ctx := &AdapterContext{LibraryFS: createTestFS()}

	readOnlySource := func(name string) []byte {
		return []byte("---\nname: " + name + "\ndescription: A read-only agent.\ntools:\n  - read\n  - search\n---\n\n# System Prompt\n\nYou are " + name + ".")
	}
	fullAccessSource := func(name string) []byte {
		return []byte("---\nname: " + name + "\ndescription: A full-access agent.\ntools:\n  - read\n  - edit\n  - shell\n  - search\n  - web\n  - mcp\n  - spawn\n---\n\n# System Prompt\n\nYou are " + name + ".")
	}
	legacySource := func(name string) []byte {
		return []byte("---\nname: " + name + "\ndescription: A legacy agent with no tools field.\ntier: balanced\n---\n\n# System Prompt\n\nYou are " + name + ".")
	}

	t.Run("read_only_agents_have_disallowedTools", func(t *testing.T) {
		for _, name := range []string{"researcher", "reviewer", "evidence-verifier"} {
			t.Run(name, func(t *testing.T) {
				out, err := RewriteAgentForClaudeCode(readOnlySource(name), ctx)
				if err != nil {
					t.Fatalf("RewriteAgentForClaudeCode: %v", err)
				}

				raw := string(out)

				// Must contain disallowedTools key
				if !strings.Contains(raw, "disallowedTools:") {
					t.Errorf("agent %q: expected disallowedTools in frontmatter; got:\n%s", name, raw)
				}

				fm, _, err := frontmatter.ExtractFrontmatter(out)
				if err != nil {
					t.Fatalf("agent %q: parse frontmatter: %v", name, err)
				}

				toolsVal, hasTools := fm["disallowedTools"]
				if !hasTools {
					t.Fatalf("agent %q: disallowedTools key missing from parsed frontmatter", name)
				}
				toolsStr, _ := toolsVal.(string)

				// Must deny Edit, Write, and Bash
				for _, denied := range []string{"Edit", "Write", "Bash"} {
					if !strings.Contains(toolsStr, denied) {
						t.Errorf("agent %q: disallowedTools missing %q; got %q", name, denied, toolsStr)
					}
				}

				// PascalCase: must not contain lowercase variants
				for _, bad := range []string{"edit", "write", "bash"} {
					if strings.Contains(toolsStr, bad) {
						t.Errorf("agent %q: disallowedTools has lowercase %q; must be PascalCase; got %q", name, bad, toolsStr)
					}
				}

				// Whitespace-separated: must not contain comma-space
				if strings.Contains(toolsStr, ", ") {
					t.Errorf("agent %q: disallowedTools uses comma-space separator; must be whitespace; got %q", name, toolsStr)
				}

				// Must not also emit a tools: allowlist (would be redundant/conflicting)
				if _, hasAllowlist := fm["tools"]; hasAllowlist {
					t.Errorf("agent %q: unexpected tools allowlist in Claude frontmatter for read-only agent", name)
				}
			})
		}
	})

	t.Run("full_access_agents_are_unrestricted", func(t *testing.T) {
		for _, name := range []string{"implementer", "deployer"} {
			t.Run(name, func(t *testing.T) {
				out, err := RewriteAgentForClaudeCode(fullAccessSource(name), ctx)
				if err != nil {
					t.Fatalf("RewriteAgentForClaudeCode: %v", err)
				}

				raw := string(out)

				if strings.Contains(raw, "disallowedTools:") {
					t.Errorf("agent %q: full-access agent must not emit disallowedTools; got:\n%s", name, raw)
				}
				if strings.Contains(raw, "tools:") {
					t.Errorf("agent %q: full-access agent must not emit tools allowlist; got:\n%s", name, raw)
				}
			})
		}
	})

	t.Run("legacy_agents_are_unrestricted", func(t *testing.T) {
		out, err := RewriteAgentForClaudeCode(legacySource("guide"), ctx)
		if err != nil {
			t.Fatalf("RewriteAgentForClaudeCode: %v", err)
		}
		raw := string(out)
		if strings.Contains(raw, "disallowedTools:") {
			t.Errorf("legacy agent: must not emit disallowedTools when no tools field; got:\n%s", raw)
		}
	})

	t.Run("whitespace_separation_not_comma", func(t *testing.T) {
		out, err := RewriteAgentForClaudeCode(readOnlySource("researcher"), ctx)
		if err != nil {
			t.Fatalf("RewriteAgentForClaudeCode: %v", err)
		}
		raw := string(out)
		if strings.Contains(raw, ", ") {
			t.Errorf("disallowedTools uses comma-space; must be space-separated; got:\n%s", raw)
		}
	})
}

// TestClaudeDisallowedTools_Helper exercises claudeDisallowedTools directly for
// the edge cases that the integration tests do not exhaustively cover.
func TestClaudeDisallowedTools_Helper(t *testing.T) {
	cases := []struct {
		name   string
		grants []frontmatter.AgentToolGrant
		want   []string
	}{
		{"nil_grants_unrestricted", nil, nil},
		{"empty_grants_unrestricted", []frontmatter.AgentToolGrant{}, nil},
		{"read_only", []frontmatter.AgentToolGrant{frontmatter.AgentToolRead, frontmatter.AgentToolSearch}, []string{"Edit", "Write", "Bash"}},
		{"read_only_just_read", []frontmatter.AgentToolGrant{frontmatter.AgentToolRead}, []string{"Edit", "Write", "Bash"}},
		{"read_only_just_search", []frontmatter.AgentToolGrant{frontmatter.AgentToolSearch}, []string{"Edit", "Write", "Bash"}},
		{"full_set_unrestricted", []frontmatter.AgentToolGrant{frontmatter.AgentToolRead, frontmatter.AgentToolEdit, frontmatter.AgentToolShell, frontmatter.AgentToolSearch, frontmatter.AgentToolWeb, frontmatter.AgentToolMCP, frontmatter.AgentToolSpawn}, nil},
		{"edit_present_unrestricted", []frontmatter.AgentToolGrant{frontmatter.AgentToolRead, frontmatter.AgentToolEdit}, nil},
		{"shell_present_unrestricted", []frontmatter.AgentToolGrant{frontmatter.AgentToolRead, frontmatter.AgentToolShell}, nil},
		{"web_present_unrestricted", []frontmatter.AgentToolGrant{frontmatter.AgentToolWeb}, nil},
		{"mcp_present_unrestricted", []frontmatter.AgentToolGrant{frontmatter.AgentToolMCP}, nil},
		{"spawn_present_unrestricted", []frontmatter.AgentToolGrant{frontmatter.AgentToolSpawn}, nil},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := claudeDisallowedTools(c.grants)
			if !slicesEqual(got, c.want) {
				t.Errorf("claudeDisallowedTools(%v) = %v, want %v", c.grants, got, c.want)
			}
		})
	}
}
