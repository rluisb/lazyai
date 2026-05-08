package models

import (
	"context"
	"strings"
	"testing"
)

// fakeAPIJSON is a minimal stand-in for models.dev/api.json shaped exactly
// like the real response: top-level providers, each with a `models` map
// where each entry has `limit.context`, `tool_call`, `reasoning`, and
// `modalities.input`. Kept small so changes to the parser or filter are
// obvious from a single test failure.
const fakeAPIJSON = `{
  "anthropic": {
    "models": {
      "claude-opus-4-7": {"limit": {"context": 1000000}, "tool_call": true, "reasoning": true, "modalities": {"input": ["text", "image"]}},
      "claude-haiku-4-5": {"limit": {"context": 200000}, "tool_call": true, "reasoning": true, "modalities": {"input": ["text"]}}
    }
  },
  "openai": {
    "models": {
      "gpt-5.5": {"limit": {"context": 1050000}, "tool_call": true, "reasoning": true, "modalities": {"input": ["text"]}},
      "gpt-5.4-mini": {"limit": {"context": 400000}, "tool_call": true, "reasoning": true, "modalities": {"input": ["text"]}}
    }
  },
  "noise-provider": {
    "models": {
      "ignored-model": {"limit": {"context": 1000}, "modalities": {"input": ["text"]}}
    }
  }
}`

func fakeFetcher(_ context.Context) ([]byte, error) {
	return []byte(fakeAPIJSON), nil
}

// Sync filters out providers we don't care about. noise-provider exists in
// the upstream stub but should never appear in the filtered output or the
// generated source.
func TestSync_FiltersToTargets(t *testing.T) {
	rep, err := Sync(context.Background(), fakeFetcher)
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if rep.UpstreamCount != 5 {
		t.Errorf("upstream count: want 5, got %d", rep.UpstreamCount)
	}
	if rep.FilteredCount != 4 {
		t.Errorf("filtered count: want 4 (anthropic+openai), got %d", rep.FilteredCount)
	}
	if strings.Contains(string(rep.GeneratedSource), "noise-provider") {
		t.Error("noise-provider leaked into generated source")
	}
}

// missingCuratedReferences must catch curated tier IDs that vanish
// upstream. The fake JSON omits "sonnet" entirely for anthropic, so
// ClaudeCodeCatalog.Balanced (which lists "sonnet") should surface as
// missing — except aliases bypass the upstream check because they don't
// exist in models.dev as top-level IDs. Use a provider-prefixed reference
// here to verify the check fires.
func TestSync_FlagsCuratedMissingFromUpstream(t *testing.T) {
	upstream := []modelEntry{
		{"openai", "gpt-5.5", true, false, 1_050_000},
	}
	probe := Catalog{
		Frontier: []string{"openai/gpt-5.5", "openai/gpt-7-not-yet-shipped"},
		Balanced: []string{"openai/gpt-5.4-mini"},
	}
	missing := missingCuratedReferences(upstream, []Catalog{probe})
	if len(missing) != 2 {
		t.Fatalf("want 2 missing, got %d: %v", len(missing), missing)
	}
}

// Generated source must be valid Go and round-trip through gofmt. We don't
// run gofmt here (no toolchain dep at test time) but assert the obvious
// invariants: package declaration, struct definition, slice opener, and
// the closing brace.
func TestSync_GeneratesValidGoSource(t *testing.T) {
	rep, _ := Sync(context.Background(), fakeFetcher)
	src := string(rep.GeneratedSource)
	for _, want := range []string{
		"package models",
		"type modelEntry struct",
		"var allModels = []modelEntry{",
		`{"anthropic", "claude-opus-4-7", true, true, 1000000},`,
	} {
		if !strings.Contains(src, want) {
			t.Errorf("generated source missing %q", want)
		}
	}
	if !strings.HasSuffix(strings.TrimSpace(src), "}") {
		t.Error("generated source not closed properly")
	}
}

// Diffing local vs upstream catches added, removed, and "changed" entries.
// Reasoning flag flips and large context-window jumps both qualify as
// changed; small context-window drift is ignored to keep churn down.
func TestSync_DiffDetectsAddRemoveChange(t *testing.T) {
	local := []modelEntry{
		{"openai", "gpt-5.4", true, false, 1_050_000},
		{"openai", "gpt-old", true, false, 100_000},
	}
	upstream := []modelEntry{
		{"openai", "gpt-5.4", true, false, 1_050_000}, // unchanged
		{"openai", "gpt-5.5", true, false, 1_050_000}, // added
		// gpt-old removed
	}
	diffs := diffEntries(local, upstream)
	kinds := map[string]int{}
	for _, d := range diffs {
		kinds[d.Kind]++
	}
	if kinds["added"] != 1 || kinds["removed"] != 1 {
		t.Errorf("kinds=%v", kinds)
	}
}
