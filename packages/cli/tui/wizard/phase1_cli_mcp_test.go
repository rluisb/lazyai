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

	// Recommended now matches the shipped catalog exactly: five retained servers,
	// all enabled by default and sorted by ID.
	recommended := defaultMcpServersForPreset(McpPresetRecommended)
	if want := []string{"ai-memory", "codegraph", "filesystem", "obsidian", "ripgrep"}; !reflect.DeepEqual(recommended, want) {
		t.Fatalf("recommended = %#v, want %#v", recommended, want)
	}

	full := defaultMcpServersForPreset(McpPresetFull)
	if !reflect.DeepEqual(full, recommended) {
		t.Fatalf("full = %#v, want %#v", full, recommended)
	}
}

func TestDefaultMcpSelectionPreservesExplicitSelection(t *testing.T) {
	t.Parallel()

	selected := defaultMcpSelection([]string{"ai-memory"}, McpPresetMinimal)
	if want := []string{"ai-memory"}; !reflect.DeepEqual(selected, want) {
		t.Fatalf("selected = %#v, want %#v", selected, want)
	}
}

func TestCatalogToolEntries(t *testing.T) {
	t.Parallel()

	catalog, err := loadMcpCatalog()
	if err != nil {
		t.Fatalf("loadMcpCatalog: %v", err)
	}

	if want := 5; len(catalog.Servers) != want {
		t.Fatalf("catalog.Servers count = %d, want %d", len(catalog.Servers), want)
	}
	if want := 5; len(catalog.CliTools) != want {
		t.Fatalf("catalog.CliTools count = %d, want %d", len(catalog.CliTools), want)
	}
	if _, ok := catalog.CliTools["ai-memory"]; !ok {
		t.Fatalf("catalog.CliTools missing ai-memory")
	}
	if _, ok := catalog.CliTools["ai-jail"]; !ok {
		t.Fatalf("catalog.CliTools missing ai-jail")
	}
	if _, ok := catalog.Servers["ai-memory"]; !ok {
		t.Fatalf("catalog.Servers missing ai-memory")
	}
	if !catalog.Servers["ai-memory"].Enabled {
		t.Fatalf("catalog.Servers ai-memory must be enabled by default")
	}
	for _, removed := range []string{"memory", "memoria", "qmd", "graphify", "context7", "github", "playwright", "atlassian", "fetch", "acli", "rtk"} {
		if _, ok := catalog.Servers[removed]; ok {
			t.Fatalf("catalog.Servers must not include removed entry %q", removed)
		}
		if _, ok := catalog.CliTools[removed]; ok {
			t.Fatalf("catalog.CliTools must not include removed entry %q", removed)
		}
	}
	if _, ok := catalog.Servers["ai-jail"]; ok {
		t.Fatalf("catalog.Servers must not include ai-jail")
	}

	found := map[string]bool{}
	for _, opt := range cliToolOptionsFromCatalog(catalog) {
		found[opt.Value] = true
	}
	for _, name := range []string{"ai-jail", "ai-memory", "codegraph", "gh", "ob"} {
		if !found[name] {
			t.Fatalf("cliToolOptionsFromCatalog() missing %s", name)
		}
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
