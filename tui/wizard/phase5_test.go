package wizard

import "testing"

func TestRunPhase5NonInteractiveDefaults(t *testing.T) {
	t.Parallel()

	result, action, err := RunPhase5(nil, true)
	if err != nil {
		t.Fatalf("RunPhase5: %v", err)
	}
	if action != PhaseContinue {
		t.Fatalf("action = %v, want %v", action, PhaseContinue)
	}
	if result.MemoryPath != "specs/memory" {
		t.Fatalf("MemoryPath = %q, want specs/memory", result.MemoryPath)
	}
	if result.QmdIndexPath != ".qmd-index" {
		t.Fatalf("QmdIndexPath = %q, want .qmd-index", result.QmdIndexPath)
	}
	if result.CodegraphDataPath != ".codegraph" {
		t.Fatalf("CodegraphDataPath = %q, want .codegraph", result.CodegraphDataPath)
	}
}
