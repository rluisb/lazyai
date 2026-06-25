package adapter

import (
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// TestEveryRegisteredAdapterReportsCapabilities ensures the capability model
// is populated for all 7 V2 targets: every adapter must carry a recognized
// support level and emit root instructions. MCP is expected for every adapter
// except Pi, whose CompileMCP is an intentional no-op (Pi has no native MCP
// surface; see issue #531).
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
		if !cap.RootInstructions {
			t.Errorf("%s: every V2 target must emit root instructions", id)
		}
		if id != types.ToolIdPi && !cap.MCP {
			t.Errorf("%s: every V2 target except Pi must support MCP", id)
		}
		if id == types.ToolIdPi && cap.MCP {
			t.Error("Pi must not declare MCP: CompileMCP is a no-op (issue #531)")
		}
	}
}

// TestNoBetaAdaptersRemain pins EC-006 / B.4 after #486: both formerly-beta
// adapters were promoted to stable once their emitted surfaces were verified
// against host docs (OMP via the authoritative omp:// docs; Antigravity once the
// global-skills-path and root-instructions gaps were closed and pinned). No
// registered adapter is below stable. See
// docs/adapters/snapshots/beta-adapter-verification-2026-06.md.
func TestNoBetaAdaptersRemain(t *testing.T) {
	reg := NewRegistry()
	var beta []types.ToolId
	for _, id := range reg.List() {
		a, _ := reg.Get(id)
		if a.Capabilities().IsBeta() {
			beta = append(beta, id)
		}
	}
	if len(beta) != 0 {
		t.Fatalf("expected no beta adapters after #486, got %v", beta)
	}
}

// TestAntigravityCapabilitiesAreStableAndVerified pins #486: Antigravity is
// stable and declares its verified emit surfaces once the two beta gaps closed.
func TestAntigravityCapabilitiesAreStableAndVerified(t *testing.T) {
	cap := (&AntigravityAdapter{}).Capabilities()
	if cap.Support != SupportStable {
		t.Errorf("Antigravity support = %q, want stable", cap.Support)
	}
	if cap.IsBeta() {
		t.Error("Antigravity must not be beta after #486 gaps closed")
	}
	for name, ok := range map[string]bool{
		"RootInstructions": cap.RootInstructions,
		"Skills":           cap.Skills,
		"Hooks":            cap.Hooks,
		"MCP":              cap.MCP,
	} {
		if !ok {
			t.Errorf("Antigravity must declare verified surface %s", name)
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
// adapter emits agents, skills, prompts, hooks, MCP, and permissions, but does
// not claim specs or steering support.
func TestKiroCapabilitiesMatchMatrix(t *testing.T) {
	cap := (&KiroAdapter{}).Capabilities()
	if cap.Specs || cap.Steering {
		t.Error("Kiro must not declare Specs or Steering")
	}
	if !cap.Hooks {
		t.Error("Kiro must declare Hooks — native .kiro/hooks/*.json files are emitted")
	}
	if !cap.Agents || !cap.Skills || !cap.PromptTemplates {
		t.Error("Kiro must declare Agents, Skills, and PromptTemplates")
	}
	if cap.Support != SupportStable {
		t.Errorf("Kiro support = %q, want stable", cap.Support)
	}
}

// TestPiCapabilitiesMatchEmittedSurfaces pins Pi capability metadata to the
// adapter's current emitted surfaces. Declared: root instructions (AGENTS.md),
// agents (.pi/agents), skills, hooks (as .pi/extensions), prompts, plugins
// (host-support), compaction (host-support), and GlobalConfig (.pi/settings.json
// plus ~/.pi/agent/settings.json). Not declared: MCP (CompileMCP is a no-op).
func TestPiCapabilitiesMatchEmittedSurfaces(t *testing.T) {
	cap := (&PiAdapter{}).Capabilities()
	if cap.Support != SupportStable {
		t.Errorf("Pi support = %q, want stable", cap.Support)
	}
	if !cap.RootInstructions {
		t.Error("Pi must declare RootInstructions (AGENTS.md)")
	}
	if !cap.Agents {
		t.Error("Pi must declare Agents — adapter installs .pi/agents/<name>.md")
	}
	if !cap.Skills || !cap.Hooks || !cap.PromptTemplates {
		t.Error("Pi must declare Skills, Hooks (as extensions), and PromptTemplates")
	}
	if cap.MCP {
		t.Error("Pi must not declare MCP: CompileMCP is a no-op (issue #531)")
	}
	if !cap.GlobalConfig {
		t.Error("Pi must declare GlobalConfig: adapter emits .pi/settings.json and ~/.pi/agent/settings.json (issue #532)")
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
