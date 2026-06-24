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

// TestBetaAdapterIsAntigravityOnly pins EC-006 / B.4 after #486: OMP was promoted
// to stable once its emitted surfaces were verified against the authoritative
// omp:// docs, so Antigravity/Gemini is the only adapter left below stable
// (global-skills path + root-instructions gaps remain). See
// docs/adapters/snapshots/beta-adapter-verification-2026-06.md.
func TestBetaAdapterIsAntigravityOnly(t *testing.T) {
	reg := NewRegistry()
	beta := map[types.ToolId]bool{}
	for _, id := range reg.List() {
		a, _ := reg.Get(id)
		if a.Capabilities().IsBeta() {
			beta[id] = true
		}
	}
	want := map[types.ToolId]bool{
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

// TestOmpCapabilitiesAreStableAndVerified pins #486: OMP is stable and declares
// every surface verified against the authoritative omp:// docs.
func TestOmpCapabilitiesAreStableAndVerified(t *testing.T) {
	cap := (&OmpAdapter{}).Capabilities()
	if cap.Support != SupportStable {
		t.Errorf("OMP support = %q, want stable", cap.Support)
	}
	if cap.IsBeta() {
		t.Error("OMP must not be beta after #486 verification")
	}
	for name, ok := range map[string]bool{
		"RootInstructions": cap.RootInstructions,
		"Agents":           cap.Agents,
		"Skills":           cap.Skills,
		"Hooks":            cap.Hooks,
		"Commands":         cap.Commands,
		"MCP":              cap.MCP,
	} {
		if !ok {
			t.Errorf("OMP must declare verified surface %s", name)
		}
	}
}

// TestKiroCapabilitiesMatchMatrix spot-checks the verified Kiro surface: the
// adapter emits agents, skills, prompts, MCP, and permissions, but does not
// claim specs, steering, or hooks support. Hooks are instruction-only (no
// runtime .kiro/hooks files emitted).
func TestKiroCapabilitiesMatchMatrix(t *testing.T) {
	cap := (&KiroAdapter{}).Capabilities()
	if cap.Specs || cap.Steering {
		t.Error("Kiro must not declare Specs or Steering")
	}
	if cap.Hooks {
		t.Error("Kiro must not declare Hooks — hooks are instruction-only, no .kiro/hooks emitted")
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
