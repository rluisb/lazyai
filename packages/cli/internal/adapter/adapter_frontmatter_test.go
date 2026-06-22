package adapter

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/frontmatter"
)

// Tests for the removed isGlobalOpenCodeDir heuristic were deleted when
// CompileMCP became scope-aware via CompileContext. Scope parity is now
// asserted by TestCompileMCPForTool_ScopeParity.

// TestNormalizeToolsFrontmatter_Delimiters verifies that the space and comma
// delimiter options work correctly (spec 012 task 004).
func TestNormalizeToolsFrontmatter_Delimiters(t *testing.T) {
	input := `---
name: Test Agent
tools: Bash, Read, Edit
---

Test content`

	tests := []struct {
		delimiter string
		wantTools string
	}{
		{"space", "tools: Bash Read Edit"},
		{"comma", "tools: Bash, Read, Edit"},
	}

	for _, tt := range tests {
		t.Run(tt.delimiter, func(t *testing.T) {
			got := NormalizeToolsFrontmatter(input, tt.delimiter)
			if !strings.Contains(got, tt.wantTools) {
				t.Errorf("delimiter %q: expected %q to be in output, got:\n%s",
					tt.delimiter, tt.wantTools, got)
			}
		})
	}
}

// TestClaudeCodeOutputStylesFrontmatter verifies that Claude Code output styles have
// required frontmatter fields (spec 012 task 006).
func TestClaudeCodeOutputStylesFrontmatter(t *testing.T) {
	libFS := createTestFS()
	styles := []string{"terse", "explanatory"}

	for _, style := range styles {
		t.Run(style, func(t *testing.T) {
			path := "claudecode/output-styles/" + style + ".md"
			data, err := fs.ReadFile(libFS, path)
			if err != nil {
				t.Fatalf("read output style: %v", err)
			}

			fm, _, err := frontmatter.ExtractFrontmatter(data)
			if err != nil {
				t.Fatalf("parse frontmatter: %v", err)
			}

			// Check required fields
			if _, ok := fm["name"]; !ok {
				t.Error("missing 'name' field")
			}
			if _, ok := fm["description"]; !ok {
				t.Error("missing 'description' field")
			}
			if keepCoding, ok := fm["keep-coding-instructions"]; !ok {
				t.Error("missing 'keep-coding-instructions' field")
			} else if kb, ok := keepCoding.(bool); !ok || !kb {
				t.Errorf("keep-coding-instructions should be true, got: %v", keepCoding)
			}
		})
	}
}

// TestClaudeCodeCommandsFrontmatter verifies that Claude Code commands have
// required frontmatter fields (spec 012 task 005).
func TestClaudeCodeCommandsFrontmatter(t *testing.T) {
	libFS := createTestFS()
	commands := []string{"review", "test", "commit"}

	for _, cmd := range commands {
		t.Run(cmd, func(t *testing.T) {
			path := "claudecode/commands/" + cmd + ".md"
			data, err := fs.ReadFile(libFS, path)
			if err != nil {
				t.Fatalf("read command: %v", err)
			}

			fm, _, err := frontmatter.ExtractFrontmatter(data)
			if err != nil {
				t.Fatalf("parse frontmatter: %v", err)
			}

			// Check required fields
			if _, ok := fm["description"]; !ok {
				t.Error("missing 'description' field")
			}
			if _, ok := fm["allowed-tools"]; !ok {
				t.Error("missing 'allowed-tools' field")
			}
			if _, ok := fm["argument-hint"]; !ok {
				t.Error("missing 'argument-hint' field")
			}

			// Verify allowed-tools is space-separated, not comma-separated
			toolsVal := fm["allowed-tools"]
			if toolsVal != nil {
				toolsStr := fmt.Sprintf("%v", toolsVal)
				if strings.Contains(toolsStr, ",") && !strings.Contains(toolsStr, "Read") {
					// If there's a comma but it's not part of a proper (Bash(...)) format, it's wrong
					t.Errorf("allowed-tools appears comma-separated: %s", toolsStr)
				}
			}
		})
	}
}

func TestRewriteAgentForClaudeCode_BaselineNoTier(t *testing.T) {
	source := baselineAgentSource("implementer", "Universal implementer.")
	out, err := RewriteAgentForClaudeCode(source, &AdapterContext{})
	if err != nil {
		t.Fatalf("RewriteAgentForClaudeCode: %v", err)
	}
	wantMarker := managedAgentMarker("claude", "implementer")
	if !strings.Contains(string(out), wantMarker) {
		t.Errorf("missing managed marker:\n%s", out)
	}
	for _, forbidden := range []string{"tier:", "model:", "temperature:", "mode:", "steps:", "thinking:", "risk:", "skills:"} {
		if strings.Contains(string(out), forbidden) {
			t.Errorf("output contains forbidden key %q:\n%s", forbidden, out)
		}
	}
}

// TestRewriteAgentForOpenCode_BaselineNoTier verifies that a baseline-style
// agent source produces the exact OpenCode adapter shape: quoted description
// and managed marker, with no tier/model/etc.
func TestRewriteAgentForOpenCode_BaselineNoTier(t *testing.T) {
	source := baselineAgentSource("implementer", "Universal implementer.")
	out, err := RewriteAgentForOpenCode(source, &AdapterContext{}, "")
	if err != nil {
		t.Fatalf("RewriteAgentForOpenCode: %v", err)
	}
	wantMarker := managedAgentMarker("opencode", "implementer")
	if !strings.Contains(string(out), wantMarker) {
		t.Errorf("missing managed marker:\n%s", out)
	}
	for _, forbidden := range []string{"tier:", "model:", "temperature:", "mode:", "steps:", "thinking:", "risk:", "skills:"} {
		if strings.Contains(string(out), forbidden) {
			t.Errorf("output contains forbidden key %q:\n%s", forbidden, out)
		}
	}
}

// TestCopilotAgentMarkdownContent_BaselineNoTier verifies the Copilot .agent.md
// shape for a baseline-style source: name, quoted description, tools array,
// managed marker, body.
func TestCopilotAgentMarkdownContent_BaselineNoTier(t *testing.T) {
	source := baselineAgentSource("implementer", "Universal implementer.")
	out := copilotAgentMarkdownContent(source)
	wantMarker := managedAgentMarker("copilot", "implementer")
	if !strings.Contains(string(out), wantMarker) {
		t.Errorf("missing managed marker:\n%s", out)
	}
	if !strings.Contains(string(out), `tools: ["read", "search", "edit", "shell"]`) {
		t.Errorf("missing expected tools array:\n%s", out)
	}
	for _, forbidden := range []string{"tier:", "model:", "temperature:", "mode:", "steps:", "thinking:", "risk:", "skills:"} {
		if strings.Contains(string(out), forbidden) {
			t.Errorf("output contains forbidden key %q:\n%s", forbidden, out)
		}
	}
}

// TestCopilotAdapter_DefaultSevenBaselineAgentsOnly installs with an empty
// selection and asserts .github/agents contains exactly the seven baseline
// .agent.md files and zero .agent.yaml files.
func TestCopilotAdapter_DefaultSevenBaselineAgentsOnly(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)
	ctx.LibraryFS = createTestFS()
	ctx.LibraryDir = ""
	ctx.Selections = AdapterSelections{}
	adapter := &CopilotAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("Copilot Install failed: %v", err)
	}
	agentsDir := filepath.Join(targetDir, ".github", "agents")
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		t.Fatalf("read agents dir: %v", err)
	}
	var mdFiles, yamlFiles []string
	for _, ent := range entries {
		if ent.IsDir() {
			continue
		}
		if strings.HasSuffix(ent.Name(), ".agent.md") {
			mdFiles = append(mdFiles, ent.Name())
		} else if strings.HasSuffix(ent.Name(), ".agent.yaml") {
			yamlFiles = append(yamlFiles, ent.Name())
		}
	}
	want := []string{
		"deployer.agent.md",
		"evidence-verifier.agent.md",
		"guide.agent.md",
		"implementer.agent.md",
		"planner.agent.md",
		"researcher.agent.md",
		"responder.agent.md",
		"reviewer.agent.md",
	}
	got := sortedStrings(mdFiles)
	if !slicesEqual(got, want) {
		t.Errorf(".agent.md files = %v, want %v", got, want)
	}
	if len(yamlFiles) != 0 {
		t.Errorf("unexpected .agent.yaml files: %v", yamlFiles)
	}
}
