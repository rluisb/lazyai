package adapter

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/rluisb/lazyai/packages/cli/internal/frontmatter"
)

func TestBuildOpenCodeAgentFrontmatter_InheritsNameAsDescription(t *testing.T) {
	src := []byte("---\nname: Builder\nmodel: sonnet\n---\n\n# Builder\n\nBody line.")
	got := BuildOpenCodeAgentFrontmatter(src, OpenCodeAgentOpts{})

	fm, body, err := frontmatter.ExtractFrontmatter(got)
	if err != nil {
		t.Fatalf("emitted frontmatter does not parse: %v\n---\n%s", err, got)
	}
	if fm["description"] != "Builder agent" {
		t.Errorf("description: want %q, got %v", "Builder agent", fm["description"])
	}
	// Default mode changed from "all" to "subagent" in #199 Bug 1 fix —
	// matches what real-world OpenCode subagent configs use. Orchestrator
	// agents pass `mode: "primary"` explicitly to override.
	if fm["mode"] != "subagent" {
		t.Errorf("mode: want %q, got %v", "subagent", fm["mode"])
	}
	if !strings.Contains(string(body), "# Builder") {
		t.Errorf("body was dropped:\n%s", body)
	}
}

func TestBuildOpenCodeAgentFrontmatter_ExplicitOptsOverrideSource(t *testing.T) {
	src := []byte("---\nname: Builder\n---\n\nBody.")
	got := BuildOpenCodeAgentFrontmatter(src, OpenCodeAgentOpts{
		Description: "Custom description",
		Mode:        "primary",
		Model:       "anthropic/claude-sonnet-4-5",
		Tools: map[string]bool{
			"bash": true,
			"read": true,
		},
		Permission: map[string]string{
			"edit": "allow",
			"bash": "ask",
		},
	})

	fm, _, err := frontmatter.ExtractFrontmatter(got)
	if err != nil {
		t.Fatalf("frontmatter does not parse: %v\n%s", err, got)
	}
	if fm["description"] != "Custom description" {
		t.Errorf("description mismatch: %v", fm["description"])
	}
	if fm["mode"] != "primary" {
		t.Errorf("mode mismatch: %v", fm["mode"])
	}
	if fm["model"] != "anthropic/claude-sonnet-4-5" {
		t.Errorf("model mismatch: %v", fm["model"])
	}
	tools, ok := fm["tools"].(map[string]any)
	if !ok {
		t.Fatalf("tools is not a map: %T %v", fm["tools"], fm["tools"])
	}
	if tools["bash"] != true || tools["read"] != true {
		t.Errorf("tools map wrong: %v", tools)
	}
	perm, ok := fm["permission"].(map[string]any)
	if !ok {
		t.Fatalf("permission is not a map: %T", fm["permission"])
	}
	if perm["edit"] != "allow" || perm["bash"] != "ask" {
		t.Errorf("permission map wrong: %v", perm)
	}
}

func TestBuildOpenCodeAgentFrontmatter_DropsSourceExtraKeys(t *testing.T) {
	// Source has a legacy `tools` line (MCP-server-name format) that must
	// NOT leak into the emitted frontmatter — opencode's tools schema is a
	// different keyspace.
	src := []byte("---\nname: Scout\nmodel: sonnet\ntools: filesystem ripgrep memory\n---\n\nBody.")
	got := BuildOpenCodeAgentFrontmatter(src, OpenCodeAgentOpts{})

	fm, _, err := frontmatter.ExtractFrontmatter(got)
	if err != nil {
		t.Fatalf("parse: %v\n%s", err, got)
	}
	if _, present := fm["tools"]; present {
		t.Errorf("tools key leaked from source; emitted frontmatter:\n%s", got)
	}
	if _, present := fm["model"]; present {
		t.Errorf("model key leaked from source (source models are not opencode-format); frontmatter:\n%s", got)
	}
	if _, present := fm["name"]; present {
		t.Errorf("name key leaked from source (opencode derives name from filename); frontmatter:\n%s", got)
	}
}

func TestBuildOpenCodeAgentFrontmatter_DeterministicOrder(t *testing.T) {
	// Same inputs must yield byte-identical output — critical because
	// installers compute file hashes for tracking and diffing.
	src := []byte("---\nname: X\n---\nbody")
	opts := OpenCodeAgentOpts{
		Tools:      map[string]bool{"c": true, "a": true, "b": false},
		Permission: map[string]string{"edit": "ask", "bash": "ask"},
	}
	a := string(BuildOpenCodeAgentFrontmatter(src, opts))
	b := string(BuildOpenCodeAgentFrontmatter(src, opts))
	if a != b {
		t.Fatalf("non-deterministic output:\n--- run 1 ---\n%s\n--- run 2 ---\n%s", a, b)
	}
	// Verify tools are emitted in sorted order.
	aIdx := strings.Index(a, "  a: true")
	bIdx := strings.Index(a, "  b: false")
	cIdx := strings.Index(a, "  c: true")
	if !(aIdx < bIdx && bIdx < cIdx) {
		t.Errorf("tools keys not in sorted order:\n%s", a)
	}
}

func TestBuildOpenCodeAgentFrontmatter_NoSourceFrontmatter(t *testing.T) {
	src := []byte("# Just a body\n\nNo frontmatter at all.")
	got := BuildOpenCodeAgentFrontmatter(src, OpenCodeAgentOpts{})

	fm, body, err := frontmatter.ExtractFrontmatter(got)
	if err != nil {
		t.Fatalf("parse: %v\n%s", err, got)
	}
	if fm["description"] != "Agent" {
		t.Errorf("fallback description: got %v", fm["description"])
	}
	if !strings.Contains(string(body), "Just a body") {
		t.Errorf("body lost:\n%s", body)
	}
}

func TestBuildOpenCodeAgentFrontmatter_DescriptionEscapesQuotes(t *testing.T) {
	// Defensive: callers may hand us a description with embedded quotes.
	got := BuildOpenCodeAgentFrontmatter([]byte("body"), OpenCodeAgentOpts{
		Description: `Contains "quotes" and \backslashes\`,
	})
	// The whole frontmatter block must still be parseable YAML.
	block := extractFrontmatterBlock(t, got)
	var parsed map[string]any
	if err := yaml.Unmarshal([]byte(block), &parsed); err != nil {
		t.Fatalf("escape produced invalid YAML: %v\n%s", err, block)
	}
	want := `Contains "quotes" and \backslashes\`
	if parsed["description"] != want {
		t.Errorf("description round-trip: want %q, got %q", want, parsed["description"])
	}
}

// extractFrontmatterBlock returns just the YAML text between the `---`
// fences (without the fences themselves).
func extractFrontmatterBlock(t *testing.T, content []byte) string {
	t.Helper()
	s := string(content)
	if !strings.HasPrefix(s, "---\n") {
		t.Fatalf("no leading frontmatter fence: %q", s[:min(50, len(s))])
	}
	rest := s[4:]
	end := strings.Index(rest, "\n---")
	if end == -1 {
		t.Fatalf("no closing fence: %q", rest)
	}
	return rest[:end]
}
