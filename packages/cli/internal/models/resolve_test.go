package models

import (
	"errors"
	"strings"
	"testing"
)

// The deny rules on OpenCode are load-bearing: the user has explicitly stated
// "no Claude on OpenCode ever". This test guards both layers — provider
// (anthropic/*) and name pattern (anything matching /claude/i) — across
// every tier. If a future catalog edit accidentally introduces a Claude
// model under any provider, this fails before users see it.
func TestResolve_OpenCodeNeverReturnsClaude(t *testing.T) {
	configured := []string{"openai", "github-copilot", "google", "ollama-cloud", "opencode"}
	for _, tier := range []Tier{TierFrontier, TierBalanced, TierSpeed} {
		spec := AgentSpec{Name: "probe", Tier: tier, Risk: 3}
		got, err := Resolve(spec, ResolveCtx{
			Catalog:             OpenCodeCatalog,
			ConfiguredProviders: configured,
		})
		if err != nil {
			t.Fatalf("tier=%s err=%v", tier, err)
		}
		all := append([]string{got.Field}, got.FallbackChain...)
		for _, m := range all {
			if strings.Contains(strings.ToLower(m), "claude") {
				t.Errorf("tier=%s leaked claude model: %s (full chain: %v)", tier, m, all)
			}
		}
	}
}

// risk=5 promotes any non-frontier tier to frontier. The primary-agent is the
// canonical case: declared Balanced (router-shaped) but Risk 5 (wrong route
// cascades), so we want the catalog's frontier pick — without changing the
// spec.Tier the agent author wrote.
func TestResolve_RiskFloorPromotesToFrontier(t *testing.T) {
	spec := AgentSpec{Name: "router", Tier: TierSpeed, Risk: 5}
	got, err := Resolve(spec, ResolveCtx{Catalog: ClaudeCodeCatalog})
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if got.Field != "opus" {
		t.Errorf("risk=5 should promote to frontier (opus), got %s", got.Field)
	}
	if len(got.Warnings) == 0 {
		t.Error("expected promotion warning, got none")
	}
}

// risk≥4 cannot stay on Speed. Ops-style work (B-tier, risk 4) is unaffected,
// but a mistakenly-declared Speed tier on a high-risk agent gets bumped to
// Balanced rather than silently emitting a small model.
func TestResolve_RiskFloorPromotesSpeedToBalancedAtRiskFour(t *testing.T) {
	spec := AgentSpec{Name: "ops", Tier: TierSpeed, Risk: 4}
	got, _ := Resolve(spec, ResolveCtx{Catalog: ClaudeCodeCatalog})
	if got.Field != "sonnet" {
		t.Errorf("risk=4 speed should promote to balanced (sonnet), got %s", got.Field)
	}
}

// OpenCode's RequireConfigured filter must drop candidates whose provider
// the user hasn't authenticated with. With only ollama-cloud configured,
// every other provider's candidates fall away.
func TestResolve_OpenCodeRespectsConfiguredProviders(t *testing.T) {
	spec := AgentSpec{Name: "scout", Tier: TierBalanced, Risk: 2}
	got, err := Resolve(spec, ResolveCtx{
		Catalog:             OpenCodeCatalog,
		ConfiguredProviders: []string{"ollama-cloud"},
	})
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if !strings.HasPrefix(got.Field, "ollama-cloud/") {
		t.Errorf("only ollama-cloud configured, got %s", got.Field)
	}
}

// When no configured provider has any catalog candidate, Resolve must fail
// with ErrNoEligibleModel rather than picking a random fallback.
func TestResolve_OpenCodeFailsWhenNoConfiguredProviderHasCandidate(t *testing.T) {
	spec := AgentSpec{Name: "scout", Tier: TierFrontier, Risk: 5}
	_, err := Resolve(spec, ResolveCtx{
		Catalog:             OpenCodeCatalog,
		ConfiguredProviders: []string{"anthropic"}, // denied by catalog anyway
	})
	if !errors.Is(err, ErrNoEligibleModel) {
		t.Errorf("expected ErrNoEligibleModel, got %v", err)
	}
}

// Vision-required agents can only resolve to candidates flagged Multimodal
// in catalog_gen.go. On Copilot's frontier tier, that's the Gemini pro
// preview — gpt-5.5 is text-only and claude-opus-4.7 isn't flagged
// multimodal in Copilot's catalog.
func TestResolve_MultimodalRestrictsCandidates(t *testing.T) {
	spec := AgentSpec{Name: "vision", Tier: TierFrontier, Risk: 3, Multimodal: true}
	got, err := Resolve(spec, ResolveCtx{Catalog: CopilotCatalog})
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if got.Field != "gemini-3.1-pro-preview" {
		t.Errorf("multimodal frontier on copilot should be gemini, got %s", got.Field)
	}
}

// Agents missing a tier annotation must fail — silent defaulting would mask
// migration regressions.
func TestResolve_MissingTierIsAnError(t *testing.T) {
	spec := AgentSpec{Name: "stale", Risk: 3}
	_, err := Resolve(spec, ResolveCtx{Catalog: ClaudeCodeCatalog})
	if !errors.Is(err, ErrTierUndefined) {
		t.Errorf("expected ErrTierUndefined, got %v", err)
	}
}

// Claude Code aliases must round-trip unchanged — no provider prefix added,
// no symbolic rewrite. The whole point of FormatAlias is that Anthropic
// resolves "opus" -> latest-opus on its end.
func TestResolve_ClaudeCodeEmitsAliases(t *testing.T) {
	for tier, want := range map[Tier]string{
		TierFrontier: "opus",
		TierBalanced: "sonnet",
		TierSpeed:    "haiku",
	} {
		got, _ := Resolve(AgentSpec{Name: "x", Tier: tier, Risk: 2}, ResolveCtx{Catalog: ClaudeCodeCatalog})
		if got.Field != want {
			t.Errorf("tier=%s want=%s got=%s", tier, want, got.Field)
		}
	}
}

// FallbackChain is the candidates after the primary, in preference order.
// Adapters render this as a YAML comment; users want to see what the
// secondary picks would have been.
func TestResolve_FallbackChainIsRemainingCandidates(t *testing.T) {
	spec := AgentSpec{Name: "planner", Tier: TierFrontier, Risk: 5}
	got, _ := Resolve(spec, ResolveCtx{
		Catalog:             OpenCodeCatalog,
		ConfiguredProviders: []string{"openai", "google", "ollama-cloud"},
	})
	if len(got.FallbackChain) == 0 {
		t.Fatal("expected non-empty fallback chain on frontier tier")
	}
	for _, m := range got.FallbackChain {
		if m == got.Field {
			t.Errorf("primary %s leaked into fallback chain", m)
		}
	}
}
