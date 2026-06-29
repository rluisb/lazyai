package frontmatter

import (
	"strings"
	"testing"
)

func TestParseAgentToolGrants_FullGrants(t *testing.T) {
	src := []byte(`---
name: Omnipotent
tier: frontier
temperature: 0.5
thinking: high
risk: 3
tools:
  - read
  - edit
  - shell
  - search
  - web
  - mcp
  - spawn
---

# body
`)
	got, err := ParseAgentToolGrants(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []AgentToolGrant{AgentToolRead, AgentToolEdit, AgentToolShell, AgentToolSearch, AgentToolWeb, AgentToolMCP, AgentToolSpawn}
	if len(got) != len(want) {
		t.Fatalf("len: got %d want %d — %v", len(got), len(want), got)
	}
	for i, g := range got {
		if g != want[i] {
			t.Errorf("index %d: got %q want %q", i, g, want[i])
		}
	}
}

func TestParseAgentToolGrants_ReadOnlyGrant(t *testing.T) {
	src := []byte(`---
name: Reader
tier: speed
temperature: 0.0
thinking: none
risk: 1
tools:
  - read
---
`)
	got, err := ParseAgentToolGrants(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0] != AgentToolRead {
		t.Errorf("got %v, want [read]", got)
	}
}

func TestParseAgentToolGrants_OrderPreserved(t *testing.T) {
	// A non-alphabetical order to verify preservation.
	src := []byte(`---
name: Ordered
tier: balanced
temperature: 0.3
thinking: medium
risk: 2
tools: spawn mcp web shell
---
`)
	got, err := ParseAgentToolGrants(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []AgentToolGrant{AgentToolSpawn, AgentToolMCP, AgentToolWeb, AgentToolShell}
	if len(got) != len(want) {
		t.Fatalf("len: got %d want %d", len(got), len(want))
	}
	for i, g := range got {
		if g != want[i] {
			t.Errorf("index %d: got %q want %q", i, g, want[i])
		}
	}
}

func TestParseAgentToolGrants_NoFrontmatter(t *testing.T) {
	src := []byte("just a plain markdown body with no frontmatter at all")
	got, err := ParseAgentToolGrants(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil grants, got %v", got)
	}
}

func TestParseAgentToolGrants_MissingToolsField(t *testing.T) {
	src := []byte(`---
name: NoTools
tier: frontier
temperature: 0.5
thinking: high
risk: 4
---
`)
	got, err := ParseAgentToolGrants(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil grants for missing tools field, got %v", got)
	}
}

func TestParseAgentToolGrants_EmptyToolsField(t *testing.T) {
	src := []byte(`---
name: EmptyTools
tier: balanced
temperature: 0.0
thinking: none
risk: 1
tools:
---
`)
	got, err := ParseAgentToolGrants(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil grants for empty tools field, got %v", got)
	}
}

func TestParseAgentToolGrants_InvalidToken(t *testing.T) {
	src := []byte(`---
name: Bad
tier: speed
temperature: 0.1
thinking: low
risk: 2
tools: read memory qmd
---
`)
	_, err := ParseAgentToolGrants(src)
	if err == nil {
		t.Fatal("expected error for unknown token, got nil")
	}
	if !strings.Contains(err.Error(), "memory") {
		t.Errorf("error should name the offending token, got: %v", err)
	}
}

func TestParseAgentToolGrants_InvalidTokenAlone(t *testing.T) {
	src := []byte(`---
name: Bad
tier: speed
temperature: 0.0
thinking: none
risk: 1
tools: bogus
---
`)
	_, err := ParseAgentToolGrants(src)
	if err == nil {
		t.Fatal("expected error for unknown token")
	}
	if !strings.Contains(err.Error(), "bogus") {
		t.Errorf("error should contain token name, got: %v", err)
	}
}

// TestParseAgentToolGrants_LegacyToolsIgnoredByParseAgentSpec verifies that
// the existing ParseAgentSpec continues to silently ignore the tools field —
// the legacy "memory qmd" fixture used in agent_spec_test.go must not break.
func TestParseAgentToolGrants_LegacyToolsIgnoredByParseAgentSpec(t *testing.T) {
	src := []byte(`---
name: Planner
tier: frontier
temperature: 0.5
thinking: high
risk: 5
tools: memory qmd
---

# Planner Agent
`)
	// ParseAgentSpec must still succeed regardless of invalid tool tokens.
	spec, err := ParseAgentSpec(src)
	if err != nil {
		t.Fatalf("ParseAgentSpec must not error on unknown tools: %v", err)
	}
	if spec.Name != "Planner" || spec.Tier != "frontier" {
		t.Errorf("unexpected spec: %+v", spec)
	}

	// ParseAgentToolGrants correctly rejects the same content.
	_, gErr := ParseAgentToolGrants(src)
	if gErr == nil {
		t.Error("ParseAgentToolGrants should reject unknown token 'memory'")
	}
}
