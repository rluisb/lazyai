package adapter

import (
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// TestEveryRegisteredAdapterReportsCapabilities ensures the capability model
// is populated for all 7 V2 targets: every adapter must support MCP and carry
// a recognized support level.
func TestEveryRegisteredAdapterReportsCapabilities(t *testing.T) {
	reg := NewRegistry()
	ids := reg.List()
	if len(ids) != 7 {
		t.Fatalf("registry has %d adapters, want 7 (V2 targets): %v", len(ids), ids)
	}
	for _, id := range ids {
		a, err := reg.Get(id)
		if err != nil {
			t.Fatalf("get %s: %v", id, err)
		}
		cap := a.Capabilities()
		switch cap.Support {
		case SupportStable, SupportBeta, SupportExperimental, SupportDeprecated:
		default:
			t.Errorf("%s: unrecognized support level %q", id, cap.Support)
		}
		if !cap.MCP {
			t.Errorf("%s: every V2 target must support MCP", id)
		}
		if !cap.RootInstructions {
			t.Errorf("%s: every V2 target must emit root instructions", id)
		}
	}
}

// TestBetaAdaptersAreOmpAndAntigravity pins EC-006 / B.4: exactly OMP and
// Antigravity are below stable until their docs snapshots are captured.
func TestBetaAdaptersAreOmpAndAntigravity(t *testing.T) {
	reg := NewRegistry()
	beta := map[types.ToolId]bool{}
	for _, id := range reg.List() {
		a, _ := reg.Get(id)
		if a.Capabilities().IsBeta() {
			beta[id] = true
		}
	}
	want := map[types.ToolId]bool{
		types.ToolIdOmp:         true,
		types.ToolIdAntigravity: true,
	}
	if len(beta) != len(want) {
		t.Fatalf("beta adapters = %v, want %v", beta, want)
	}
	for id := range want {
		if !beta[id] {
			t.Errorf("expected %s to be beta", id)
		}
	}
}

// TestKiroCapabilitiesMatchMatrix spot-checks the verified Kiro surface: the
// adapter emits agents, skills, prompts, hooks, MCP, and permissions, but does
// not claim specs or steering support.
func TestKiroCapabilitiesMatchMatrix(t *testing.T) {
	cap := (&KiroAdapter{}).Capabilities()
	if cap.Specs || cap.Steering {
		t.Error("Kiro must not declare Specs or Steering")
	}
	if !cap.Agents || !cap.Skills || !cap.PromptTemplates {
		t.Error("Kiro must declare Agents, Skills, and PromptTemplates")
	}
	if cap.Support != SupportStable {
		t.Errorf("Kiro support = %q, want stable", cap.Support)
	}
}

// TestClaudeEmitsRootInstructions guards FR-012: Claude Code must declare root
// instructions (CLAUDE.md) as a supported surface.
func TestClaudeEmitsRootInstructions(t *testing.T) {
	cap := (&ClaudeCodeAdapter{}).Capabilities()
	if !cap.RootInstructions {
		t.Error("Claude Code must declare RootInstructions (CLAUDE.md)")
	}
	if cap.Support != SupportStable {
		t.Errorf("Claude Code support = %q, want stable", cap.Support)
	}
}
