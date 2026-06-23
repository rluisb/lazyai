package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

func TestToolOverrideMapCoversSupportedTools(t *testing.T) {
	for _, tool := range types.SupportedToolIDs {
		override, ok := ToolOverrideMap[string(tool)]
		if !ok {
			t.Fatalf("ToolOverrideMap missing %s", tool)
		}
		if override.Description == "" {
			t.Fatalf("ToolOverrideMap[%s] missing description", tool)
		}
	}
}

func TestFragmentResolver_VariableInterpolation(t *testing.T) {
	resolver := NewFragmentResolver("")

	content := "Project: {{PROJECT_NAME}}, Lang: {{PRIMARY_LANGUAGE}}"
	ctx := FragmentContext{
		ProjectName:     "my-app",
		PrimaryLanguage: "Go",
	}

	result := resolver.Resolve(content, ctx)
	if result != "Project: my-app, Lang: Go" {
		t.Errorf("expected 'Project: my-app, Lang: Go', got %q", result)
	}
	_ = resolver // ensure resolver is used
}

func TestFragmentResolver_DefaultVariable(t *testing.T) {
	resolver := NewFragmentResolver("")

	content := "Lang: {{PRIMARY_LANGUAGE}}"
	ctx := FragmentContext{} // No PrimaryLanguage set

	result := resolver.Resolve(content, ctx)
	if result != "Lang: TypeScript" {
		t.Errorf("expected default 'TypeScript', got %q", result)
	}
}

func TestFragmentResolver_Conditional(t *testing.T) {
	yes := true
	no := false

	tests := []struct {
		name     string
		content  string
		features *FeatureFlags
		expected string
	}{
		{
			name:     "enabled feature",
			content:  "{{#if features.rpiWorkflow}}RPI enabled{{/if}}",
			features: &FeatureFlags{RPIWorkflow: &yes},
			expected: "RPI enabled",
		},
		{
			name:     "disabled feature",
			content:  "{{#if features.rpiWorkflow}}RPI enabled{{/if}}",
			features: &FeatureFlags{RPIWorkflow: &no},
			expected: "",
		},
		{
			name:     "nil features",
			content:  "{{#if features.rpiWorkflow}}RPI enabled{{/if}}",
			features: nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := FragmentContext{Features: tt.features}
			result := NewFragmentResolver("").Resolve(tt.content, ctx)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFragmentResolver_IncludeFromFS(t *testing.T) {
	libFS := fstest.MapFS{
		"fragments/test-fragment.md": &fstest.MapFile{
			Data: []byte("Fragment content here"),
		},
	}

	resolver := NewFragmentResolver("", libFS)
	content := "Before {{#include fragments/test-fragment.md}} After"
	ctx := FragmentContext{}

	result := resolver.Resolve(content, ctx)
	expected := "Before Fragment content here After"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFragmentResolver_IncludeFromDisk(t *testing.T) {
	// Create a temp library with a fragment.
	libDir := t.TempDir()
	fragDir := filepath.Join(libDir, "fragments")
	if err := os.MkdirAll(fragDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fragDir, "test-fragment.md"), []byte("Disk fragment content"), 0o644); err != nil {
		t.Fatal(err)
	}

	resolver := NewFragmentResolver(libDir) // No libFS = disk mode
	content := "Before {{#include fragments/test-fragment.md}} After"
	ctx := FragmentContext{}

	result := resolver.Resolve(content, ctx)
	expected := "Before Disk fragment content After"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestFragmentResolver_IncludeNotFound(t *testing.T) {
	resolver := NewFragmentResolver("")
	content := "Before {{#include fragments/missing.md}} After"
	ctx := FragmentContext{}

	result := resolver.Resolve(content, ctx)
	if result == "" {
		t.Fatal("result should not be empty")
	}
	// Missing fragment should produce a comment.
	if result != "Before <!-- Fragment not found: fragments/missing.md --> After" {
		t.Errorf("expected fragment-not-found comment, got %q", result)
	}
}

// --- TemplateCompiler tests ---

func TestTemplateCompiler_CompileFromFS(t *testing.T) {
	libFS := fstest.MapFS{
		"tool-templates/shared/root.template.md": &fstest.MapFile{
			Data: []byte("# {{PROJECT_NAME}}\n\nRoot template for {{TOOL_DESCRIPTION}}"),
		},
		"tool-templates/opencode/settings.template.json": &fstest.MapFile{
			Data: []byte(`{"project":"{{PROJECT_NAME}}"}`),
		},
		"fragments/quality-gates.md": &fstest.MapFile{
			Data: []byte("## Quality Gates"),
		},
	}

	compiler := NewTemplateCompiler(CompilerConfig{
		LibraryDir: "",
		LibraryFS:  libFS,
		OutputDir:  t.TempDir(),
		Tool:       "opencode",
		Context: FragmentContext{
			ProjectName:     "my-project",
			ToolDescription: "OpenCode integration",
		},
	})

	output, err := compiler.Compile()
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	if output.Tool != "opencode" {
		t.Errorf("expected tool 'opencode', got %q", output.Tool)
	}
	if len(output.Files) == 0 {
		t.Fatal("expected at least one compiled file")
	}

	// Check that the root template was compiled and contains the project name.
	foundRoot := false
	for _, f := range output.Files {
		if f.RelativePath == "root.md" {
			foundRoot = true
			if !contains(f.Content, "my-project") {
				t.Errorf("root content should contain 'my-project', got %q", f.Content)
			}
		}
	}
	if !foundRoot {
		t.Error("root.md was not compiled")
	}
}

func TestTemplateCompiler_CompileFromDisk(t *testing.T) {
	libDir := t.TempDir()
	sharedDir := filepath.Join(libDir, "tool-templates", "shared")
	if err := os.MkdirAll(sharedDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sharedDir, "root.template.md"), []byte("# {{PROJECT_NAME}}\n\nRoot."), 0o644); err != nil {
		t.Fatal(err)
	}

	compiler := NewTemplateCompiler(CompilerConfig{
		LibraryDir: libDir,
		LibraryFS:  nil, // disk mode
		OutputDir:  t.TempDir(),
		Tool:       "opencode",
		Context: FragmentContext{
			ProjectName: "disk-project",
		},
	})

	output, err := compiler.Compile()
	if err != nil {
		t.Fatalf("Compile from disk failed: %v", err)
	}

	if len(output.Files) == 0 {
		t.Fatal("expected at least one compiled file from disk")
	}

	foundRoot := false
	for _, f := range output.Files {
		if f.RelativePath == "root.md" {
			foundRoot = true
			if !contains(f.Content, "disk-project") {
				t.Errorf("root should contain 'disk-project', got %q", f.Content)
			}
		}
	}
	if !foundRoot {
		t.Error("root.md was not found in compiled output")
	}
}

func TestCompileForTools(t *testing.T) {
	libFS := fstest.MapFS{
		"tool-templates/shared/root.template.md": &fstest.MapFile{
			Data: []byte("# {{PROJECT_NAME}}\n\nShared root."),
		},
		"fragments/quality-gates.md": &fstest.MapFile{
			Data: []byte("## Quality Gates"),
		},
	}

	outputDir := t.TempDir()
	results, err := CompileForTools(
		[]string{"opencode", "claude-code"},
		"",
		libFS,
		outputDir,
		FragmentContext{ProjectName: "multi-tool-project"},
	)
	if err != nil {
		t.Fatalf("CompileForTools failed: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 tool results, got %d", len(results))
	}
	for _, tool := range []string{"opencode", "claude-code"} {
		if results[tool] == nil {
			t.Errorf("no result for tool %q", tool)
		}
	}
}

// TestValidateAgentResolutions_ToleratesBaselineNoTier verifies that canonical
// agents copied from the vibe-lab baseline (which carry name/description but no
// LazyAI tier metadata) do not produce validation issues for the missing tier.
func TestValidateAgentResolutions_ToleratesBaselineNoTier(t *testing.T) {
	libFS := fstest.MapFS{
		"canonical/agents/implementer.md": &fstest.MapFile{
			Data: []byte("---\nname: implementer\ndescription: Test implementer.\n---\n\nBody."),
		},
	}
	issues, err := ValidateAgentResolutions(libFS, []types.ToolId{types.ToolIdClaudeCode}, []string{"openai"})
	if err != nil {
		t.Fatalf("ValidateAgentResolutions: %v", err)
	}
	for _, issue := range issues {
		if strings.Contains(issue.Err.Error(), "missing required field: tier") {
			t.Errorf("unexpected missing-tier issue: %v", issue)
		}
	}
}

// TestValidateAgentResolutions_StillReportsMalformedFrontmatter verifies that
// genuinely malformed frontmatter is still reported even while missing-tier is
// tolerated for baseline agents.
func TestValidateAgentResolutions_StillReportsMalformedFrontmatter(t *testing.T) {
	libFS := fstest.MapFS{
		"canonical/agents/broken.md": &fstest.MapFile{
			Data: []byte("---\nname: broken\ndescription: ok\n  unclosed: [\n---\n\nBody."),
		},
	}
	issues, err := ValidateAgentResolutions(libFS, []types.ToolId{types.ToolIdClaudeCode}, []string{"openai"})
	if err != nil {
		t.Fatalf("ValidateAgentResolutions: %v", err)
	}
	if len(issues) == 0 {
		t.Fatal("expected malformed-frontmatter issue, got none")
	}
}

// --- Helper functions ---

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
