package generator

import (
	"strings"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

func TestHookGeneratorType(t *testing.T) {
	g := &HookGenerator{}
	if g.Type() != types.ArtifactTypeHook {
		t.Fatalf("HookGenerator.Type() = %q, want %q", g.Type(), types.ArtifactTypeHook)
	}
}

func TestHookGeneratorPromptQuestions(t *testing.T) {
	g := &HookGenerator{}
	qs := g.GetPromptQuestions()
	if len(qs) == 0 {
		t.Fatal("expected at least one prompt question")
	}

	keys := map[string]bool{}
	for _, q := range qs {
		keys[q.Key] = true
	}
	for _, want := range []string{"purpose", "events", "denied", "allowed"} {
		if !keys[want] {
			t.Errorf("missing prompt question key %q", want)
		}
	}
}

func TestHookGeneratorGenerateDefault(t *testing.T) {
	g := &HookGenerator{}
	files, err := g.Generate(GeneratorConfig{
		Name:      "test-hook",
		TargetDir: "/placeholder",
		Answers:   map[string]string{},
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("Generate returned %d files, want 1", len(files))
	}
	if want := "library/hooks/test-hook.md"; files[0].Path != want {
		t.Errorf("Path = %q, want %q", files[0].Path, want)
	}
	if !strings.Contains(files[0].Content, "# Test Hook Policy") {
		t.Errorf("content missing title, got:\n%s", files[0].Content)
	}
	if !strings.Contains(files[0].Content, "PreToolUse") {
		t.Errorf("content missing default event PreToolUse, got:\n%s", files[0].Content)
	}
}

func TestHookGeneratorGenerateWithAnswers(t *testing.T) {
	g := &HookGenerator{}
	files, err := g.Generate(GeneratorConfig{
		Name:        "block-dangerous-cmd",
		Description: "Block dangerous shell commands.",
		TargetDir:   "/placeholder",
		Answers: map[string]string{
			"events":  "PreToolUse, PostToolUse",
			"denied":  "rm -rf /, mkfs",
			"allowed": "git clean",
		},
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("Generate returned %d files, want 1", len(files))
	}
	c := files[0].Content
	if !strings.Contains(c, "Block Dangerous Cmd") {
		t.Errorf("content missing title")
	}
	if !strings.Contains(c, "Block dangerous shell commands.") {
		t.Errorf("content missing description")
	}
	if !strings.Contains(c, "`PreToolUse`") {
		t.Errorf("content missing PreToolUse event")
	}
	if !strings.Contains(c, "`PostToolUse`") {
		t.Errorf("content missing PostToolUse event")
	}
	if !strings.Contains(c, "rm -rf /") {
		t.Errorf("content missing denied entry")
	}
	if !strings.Contains(c, "mkfs") {
		t.Errorf("content missing denied entry mkfs")
	}
	if !strings.Contains(c, "git clean") {
		t.Errorf("content missing allowed entry")
	}
}

func TestHookGeneratorGenerateSlug(t *testing.T) {
	g := &HookGenerator{}
	files, err := g.Generate(GeneratorConfig{
		Name:      "  MY HOOK  ",
		TargetDir: "/placeholder",
		Answers:   map[string]string{},
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.HasSuffix(files[0].Path, ".md") {
		t.Errorf("Path should end with .md, got %q", files[0].Path)
	}
}

func TestHookGeneratorRegisteredInRegistry(t *testing.T) {
	r := NewRegistry()
	g, err := r.Get(types.ArtifactTypeHook)
	if err != nil {
		t.Fatalf("HookGenerator not registered: %v", err)
	}
	if g.Type() != types.ArtifactTypeHook {
		t.Fatalf("registered generator Type() = %q, want %q", g.Type(), types.ArtifactTypeHook)
	}
}
