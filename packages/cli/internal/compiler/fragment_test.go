package compiler

import "testing"

func TestFragmentContext_ConstitutionFieldResolution(t *testing.T) {
	coverageThreshold := 80
	ctx := FragmentContext{
		Constitution: &ConstitutionContext{
			ProjectOverview: "Test overview",
			Conventions: ConstitutionConventions{
				Naming: "camelCase",
			},
			CoverageThreshold: &coverageThreshold,
		},
	}

	content := "{{PROJECT_OVERVIEW}}|{{NAMING_CONVENTIONS}}|{{COVERAGE_THRESHOLD}}"
	result := NewFragmentResolver("").Resolve(content, ctx)
	expected := "Test overview|camelCase|80"

	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestFragmentContext_FallbackMarkers(t *testing.T) {
	content := "{{PROJECT_OVERVIEW}}|{{COVERAGE_THRESHOLD}}"
	result := NewFragmentResolver("").Resolve(content, FragmentContext{})
	expected := "[YOUR_PROJECT_OVERVIEW]|80"

	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestFragmentContext_CodebaseMapRows(t *testing.T) {
	ctx := FragmentContext{
		Constitution: &ConstitutionContext{
			CodebaseMap: []CodebaseMapEntry{
				{Path: "cmd"},
				{Path: "node_modules"},
				{Path: "internal/compiler"},
				{Path: "dist"},
			},
		},
	}

	result := NewFragmentResolver("").Resolve("{{CODEBASE_MAP}}", ctx)
	expected := "| cmd | [WHAT_IT_DOES] |\n| internal/compiler | [WHAT_IT_DOES] |"

	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestFragmentContext_AdversarialDesignConditional(t *testing.T) {
	enabled := true
	disabled := false
	content := "before{{#if features.adversarialDesign}} adversarial{{/if}} after"

	enabledResult := NewFragmentResolver("").Resolve(content, FragmentContext{
		Features: &FeatureFlags{AdversarialDesign: &enabled},
	})
	if enabledResult != "before adversarial after" {
		t.Fatalf("expected enabled conditional content, got %q", enabledResult)
	}

	disabledResult := NewFragmentResolver("").Resolve(content, FragmentContext{
		Features: &FeatureFlags{AdversarialDesign: &disabled},
	})
	if disabledResult != "before after" {
		t.Fatalf("expected disabled conditional to be removed, got %q", disabledResult)
	}
}

func TestFragmentContext_NestedConditionals(t *testing.T) {
	enabled := true
	disabled := false
	content := "before {{#if features.adversarialDesign}}outer {{#if features.qualityGates}}inner{{/if}} done{{/if}} after"

	result := NewFragmentResolver("").Resolve(content, FragmentContext{
		Features: &FeatureFlags{
			AdversarialDesign: &enabled,
			QualityGates:      &enabled,
		},
	})
	if result != "before outer inner done after" {
		t.Fatalf("expected nested conditionals to resolve, got %q", result)
	}

	result = NewFragmentResolver("").Resolve(content, FragmentContext{
		Features: &FeatureFlags{
			AdversarialDesign: &enabled,
			QualityGates:      &disabled,
		},
	})
	if result != "before outer  done after" {
		t.Fatalf("expected disabled nested conditional to be removed, got %q", result)
	}

	result = NewFragmentResolver("").Resolve(content, FragmentContext{
		Features: &FeatureFlags{
			AdversarialDesign: &disabled,
			QualityGates:      &enabled,
		},
	})
	if result != "before  after" {
		t.Fatalf("expected disabled outer conditional to remove nested body, got %q", result)
	}
}
