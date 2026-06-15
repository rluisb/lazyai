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
	// Full includes every catalog server.
	if len(full) < 11 {
		t.Fatalf("full count = %d, want at least 11", len(full))
	}
	if full[0] != "atlassian" || full[len(full)-1] != "ripgrep" {
		t.Fatalf("full ordering = %#v, want sorted catalog IDs", full)
	}
}

func TestDefaultMcpSelectionPreservesExplicitSelection(t *testing.T) {
	t.Parallel()

	selected := defaultMcpSelection([]string{"memory"}, McpPresetMinimal)
	if want := []string{"memory"}; !reflect.DeepEqual(selected, want) {
		t.Fatalf("selected = %#v, want %#v", selected, want)
	}
}

func TestAcliCliToolCatalogVisibility(t *testing.T) {
	t.Parallel()

	catalog, err := loadMcpCatalog()
	if err != nil {
		t.Fatalf("loadMcpCatalog: %v", err)
	}

	tool, ok := catalog.CliTools["acli"]
	if !ok {
		t.Fatalf("catalog.CliTools missing acli")
	}
	if want := "https://developer.atlassian.com/cloud/acli/guides/how-to-get-started/"; tool.InstallHint != want {
		t.Fatalf("acli install hint = %q, want %q", tool.InstallHint, want)
	}
	if _, ok := catalog.Servers["acli"]; ok {
		t.Fatalf("catalog.Servers must not include acli")
	}

	found := false
	for _, opt := range cliToolOptionsFromCatalog(catalog) {
		if opt.Value == "acli" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("cliToolOptionsFromCatalog() missing acli")
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
