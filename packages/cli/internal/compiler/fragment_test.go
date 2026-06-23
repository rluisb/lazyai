package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

func TestFragmentContext_PathTraversal(t *testing.T) {
	// Create a temp library directory with a real fragment and a secret file
	// outside of it to prove traversal is blocked.
	libDir := t.TempDir()
	secretDir := t.TempDir()
	secretPath := filepath.Join(secretDir, "secret.txt")
	if err := os.WriteFile(secretPath, []byte("TOPSECRET"), 0o644); err != nil {
		t.Fatalf("setup: write secret: %v", err)
	}

	// Place a legitimate fragment so we know the resolver is wired correctly.
	legitPath := filepath.Join(libDir, "fragments", "ok.xml")
	if err := os.MkdirAll(filepath.Dir(legitPath), 0o755); err != nil {
		t.Fatalf("setup: mkdir: %v", err)
	}
	if err := os.WriteFile(legitPath, []byte("<ok/>"), 0o644); err != nil {
		t.Fatalf("setup: write legit: %v", err)
	}

	// Compute a traversal path from libDir to the secret file.
	rel, err := filepath.Rel(filepath.Join(libDir, "fragments"), secretPath)
	if err != nil {
		t.Fatalf("setup: rel: %v", err)
	}
	traversalInclude := "{{#include fragments/" + rel + "}}"

	// libFS is nil → disk fallback, where the guard must fire.
	r := NewFragmentResolver(libDir)
	result := r.Resolve(traversalInclude, FragmentContext{})

	expected := "<!-- Fragment not found: fragments/" + rel + " -->"
	if result != expected {
		t.Fatalf("path traversal not blocked:\nexpected %q\n  actual %q", expected, result)
	}
	if strings.Contains(result, "TOPSECRET") {
		t.Fatalf("path traversal leaked secret file contents: %q", result)
	}

	// Sanity check: a legitimate include still resolves.
	legitResult := r.Resolve("{{#include fragments/ok.xml}}", FragmentContext{})
	if legitResult != "<ok/>" {
		t.Fatalf("legitimate include broken: expected %q, got %q", "<ok/>", legitResult)
	}
}
