package plugin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/library"
)

func TestNormalizeTargetSupportsPhaseETargets(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want BundleTarget
	}{
		{"claude", BundleTargetClaude},
		{"claude-code", BundleTargetClaude},
		{"copilot", BundleTargetCopilotCLI},
		{"copilot-cli", BundleTargetCopilotCLI},
		{"omp", BundleTargetOmp},
		{"pi", BundleTargetPi},
	} {
		got, err := NormalizeTarget(tc.in)
		if err != nil {
			t.Fatalf("NormalizeTarget(%q): %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("NormalizeTarget(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestBuildTarget_CopilotCLIShape(t *testing.T) {
	outDir := t.TempDir()
	result, err := BuildTarget(library.GetLibraryFS(), outDir, "0.1.0", BundleTargetCopilotCLI)
	if err != nil {
		t.Fatalf("BuildTarget copilot-cli: %v", err)
	}
	if result.FileCount == 0 {
		t.Fatal("expected files to be emitted")
	}
	for _, rel := range []string{
		"plugin.json",
		"agents/guide.agent.md",
		"skills/implement/SKILL.md",
		"hooks.json",
		"hooks/block-destructive-shell.sh",
		".mcp.json",
	} {
		if _, err := os.Stat(filepath.Join(outDir, rel)); err != nil {
			t.Fatalf("expected %s: %v", rel, err)
		}
	}
	data, err := os.ReadFile(filepath.Join(outDir, "hooks.json"))
	if err != nil {
		t.Fatalf("read hooks.json: %v", err)
	}
	if strings.Contains(string(data), ".github/hooks/") {
		t.Fatalf("hooks.json still references project-local .github/hooks paths: %s", data)
	}
}

func TestBuildTarget_OmpShape(t *testing.T) {
	outDir := t.TempDir()
	_, err := BuildTarget(library.GetLibraryFS(), outDir, "0.1.0", BundleTargetOmp)
	if err != nil {
		t.Fatalf("BuildTarget omp: %v", err)
	}
	for _, rel := range []string{
		"skills/implement/SKILL.md",
		"commands/handoff.md",
		"hooks/pre/block-destructive-shell.ts",
		"mcp.json",
	} {
		if _, err := os.Stat(filepath.Join(outDir, rel)); err != nil {
			t.Fatalf("expected %s: %v", rel, err)
		}
	}
}

func TestBuildTarget_PiShape(t *testing.T) {
	outDir := t.TempDir()
	_, err := BuildTarget(library.GetLibraryFS(), outDir, "0.1.0", BundleTargetPi)
	if err != nil {
		t.Fatalf("BuildTarget pi: %v", err)
	}
	for _, rel := range []string{
		"agents/guide.md",
		"skills/implement/SKILL.md",
		"prompts/plan.md",
		"extensions/block-destructive-shell.ts",
	} {
		if _, err := os.Stat(filepath.Join(outDir, rel)); err != nil {
			t.Fatalf("expected %s: %v", rel, err)
		}
	}
}
