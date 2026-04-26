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

	recommended := defaultMcpServersForPreset(McpPresetRecommended)
	if want := []string{"filesystem", "memoria", "memory", "ripgrep"}; !reflect.DeepEqual(recommended, want) {
		t.Fatalf("recommended = %#v, want %#v", recommended, want)
	}

	full := defaultMcpServersForPreset(McpPresetFull)
	if len(full) != 12 {
		t.Fatalf("full count = %d, want 12", len(full))
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
