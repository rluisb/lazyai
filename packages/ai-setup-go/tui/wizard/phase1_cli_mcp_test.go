package wizard

import (
	"reflect"
	"testing"
)

func TestDefaultMcpServersForPreset(t *testing.T) {
	t.Parallel()

	minimal := defaultMcpServersForPreset(McpPresetMinimal)
	if want := []string{"filesystem", "ripgrep"}; !reflect.DeepEqual(minimal, want) {
		t.Fatalf("minimal = %#v, want %#v", minimal, want)
	}

	// Recommended now includes codegraph, qmd, graphify, and obsidian (Spec 022 / E3).
	// The exact count may drift as the catalog evolves; the invariant is that memory
	// and filesystem are included.
	recommended := defaultMcpServersForPreset(McpPresetRecommended)
	if len(recommended) < 4 {
		t.Fatalf("recommended count = %d, want at least 4", len(recommended))
	}
	hasMemory := false
	hasFilesystem := false
	for _, s := range recommended {
		if s == "memory" {
			hasMemory = true
		}
		if s == "filesystem" {
			hasFilesystem = true
		}
	}
	if !hasMemory || !hasFilesystem {
		t.Fatalf("recommended = %#v, must include 'memory' and 'filesystem'", recommended)
	}

	full := defaultMcpServersForPreset(McpPresetFull)
	// Full includes all catalog servers (14 after Spec 022 / E3).
	if len(full) < 14 {
		t.Fatalf("full count = %d, want at least 14 (includes graphify + obsidian from Spec 022/E3)", len(full))
	}
	if full[0] != "atlassian" || full[len(full)-1] != "ripgrep" {
		t.Fatalf("full ordering = %#v, want sorted catalog IDs", full)
	}
}

func TestDefaultMcpSelectionPreservesExplicitSelection(t *testing.T) {
	t.Parallel()

	selected := defaultMcpSelection([]string{"orchestrator"}, McpPresetMinimal)
	if want := []string{"orchestrator"}; !reflect.DeepEqual(selected, want) {
		t.Fatalf("selected = %#v, want %#v", selected, want)
	}
}

func TestNormalizeMcpPresetDefaultsToRecommended(t *testing.T) {
	t.Parallel()

	if got := normalizeMcpPreset(""); got != McpPresetRecommended {
		t.Fatalf("normalizeMcpPreset(empty) = %q, want %q", got, McpPresetRecommended)
	}
	if got := normalizeMcpPreset("unexpected"); got != McpPresetRecommended {
		t.Fatalf("normalizeMcpPreset(unexpected) = %q, want %q", got, McpPresetRecommended)
	}
}
