package preset

import (
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

func TestDefaultPresetForScope(t *testing.T) {
	t.Parallel()

	tests := []struct {
		scope types.SetupScope
		want  types.PresetLevel
	}{
		{types.SetupScopeGlobal, types.PresetLevelMinimal},
		{types.SetupScopeProject, types.PresetLevelStandard},
		{types.SetupScopeWorkspace, types.PresetLevelStandard},
		{types.SetupScope("unknown"), types.PresetLevelStandard},
	}

	for _, tt := range tests {
		got := DefaultPresetForScope(tt.scope)
		if got != tt.want {
			t.Errorf("DefaultPresetForScope(%q) = %q, want %q", tt.scope, got, tt.want)
		}
	}
}

func TestResolvePreset(t *testing.T) {
	t.Parallel()

	t.Run("minimal returns correct flags", func(t *testing.T) {
		t.Parallel()
		f := ResolvePreset(types.PresetLevelMinimal)
		if f == nil {
			t.Fatal("nil flags")
		}
		if f.QualityGates != true {
			t.Error("QualityGates should be true for minimal")
		}
		if f.RPIWorkflow != false {
			t.Error("RPIWorkflow should be false for minimal")
		}
		if f.AgentHarness != false {
			t.Error("AgentHarness should be false for minimal")
		}
		if f.AdversarialDesign != false {
			t.Error("AdversarialDesign should be false for minimal")
		}
	})

	t.Run("standard returns correct flags", func(t *testing.T) {
		t.Parallel()
		f := ResolvePreset(types.PresetLevelStandard)
		if f == nil {
			t.Fatal("nil flags")
		}
		if f.QualityGates != true {
			t.Error("QualityGates should be true for standard")
		}
		if f.RPIWorkflow != true {
			t.Error("RPIWorkflow should be true for standard")
		}
		if f.ChainOfThought != true {
			t.Error("ChainOfThought should be true for standard")
		}
		if f.BugResolution != true {
			t.Error("BugResolution should be true for standard")
		}
		if f.AgentHarness != false {
			t.Error("AgentHarness should be false for standard")
		}
		if f.AdversarialDesign != true {
			t.Error("AdversarialDesign should be true for standard")
		}
	})

	t.Run("full returns all true", func(t *testing.T) {
		t.Parallel()
		f := ResolvePreset(types.PresetLevelFull)
		if f == nil {
			t.Fatal("nil flags")
		}
		if !f.ContextEngineering || !f.RPIWorkflow || !f.ChainOfThought ||
			!f.TreeOfThoughts || !f.ADREnforcement || !f.QualityGates ||
			!f.AgentHarness || !f.BugResolution || !f.PivotHandling || !f.AdversarialDesign {
			t.Errorf("Full preset should have all flags true: %+v", f)
		}
	})

	t.Run("custom returns nil", func(t *testing.T) {
		t.Parallel()
		f := ResolvePreset(types.PresetLevelCustom)
		if f != nil {
			t.Errorf("Custom preset should return nil, got %+v", f)
		}
	})
}

func TestSpecsDirsForPreset(t *testing.T) {
	t.Parallel()

	minimal := SpecsDirsForPreset(types.PresetLevelMinimal)
	if len(minimal) != 2 {
		t.Errorf("Minimal specs dirs = %d, want 2", len(minimal))
	}

	standard := SpecsDirsForPreset(types.PresetLevelStandard)
	if len(standard) != 7 {
		t.Errorf("Standard specs dirs = %d, want 7", len(standard))
	}

	full := SpecsDirsForPreset(types.PresetLevelFull)
	if len(full) != 10 {
		t.Errorf("Full specs dirs = %d, want 10", len(full))
	}

	custom := SpecsDirsForPreset(types.PresetLevelCustom)
	if len(custom) != 10 {
		t.Errorf("Custom specs dirs = %d, want 10", len(custom))
	}

	unknown := SpecsDirsForPreset(types.PresetLevel("unknown"))
	if len(unknown) != 0 {
		t.Errorf("Unknown preset specs dirs = %d, want 0", len(unknown))
	}
}

func TestTemplatesForPreset(t *testing.T) {
	t.Parallel()

	minimal := TemplatesForPreset(types.PresetLevelMinimal)
	if len(minimal) != 0 {
		t.Errorf("Minimal templates = %d, want 0", len(minimal))
	}

	standard := TemplatesForPreset(types.PresetLevelStandard)
	if len(standard) != 7 {
		t.Errorf("Standard templates = %d, want 7", len(standard))
	}

	full := TemplatesForPreset(types.PresetLevelFull)
	if len(full) != 10 {
		t.Errorf("Full templates = %d, want 10", len(full))
	}
}

func TestRulesForPreset(t *testing.T) {
	t.Parallel()

	minimal := RulesForPreset(types.PresetLevelMinimal)
	if len(minimal) != 0 {
		t.Errorf("Minimal rules = %d, want 0", len(minimal))
	}

	standard := RulesForPreset(types.PresetLevelStandard)
	if len(standard) != 5 {
		t.Errorf("Standard rules = %d, want 5", len(standard))
	}

	full := RulesForPreset(types.PresetLevelFull)
	if len(full) != 9 {
		t.Errorf("Full rules = %d, want 9", len(full))
	}

	unknown := RulesForPreset(types.PresetLevel("unknown"))
	if len(unknown) != 0 {
		t.Errorf("Unknown rules = %d, want 0", len(unknown))
	}
}
