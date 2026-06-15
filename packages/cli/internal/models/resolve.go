package models

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

var (
	// ErrNoEligibleModel fires when every candidate in the relevant tier was
	// filtered out (deny rules, configured-provider filter, multimodal
	// filter). Adapters surface this as a compile error rather than silently
	// emitting an empty model field.
	ErrNoEligibleModel = errors.New("no eligible model for tier on this target")

	// ErrTierUndefined fires when the agent's source frontmatter is missing
	// the tier annotation. The 8 library agents all carry it; user-authored
	// agents that lack it must be migrated before they will compile.
	ErrTierUndefined = errors.New("agent has no tier set")

	// ErrMultimodalUnmet is a specialisation of ErrNoEligibleModel for
	// vision-required agents whose target catalog has no multimodal entries
	// in the relevant tier.
	ErrMultimodalUnmet = errors.New("agent requires multimodal but no multimodal candidate available")
)

// Resolve picks one model identifier for the agent on the target described
// by ctx.Catalog. The returned ResolvedModel.Field is what the caller writes
// into the per-target agent file; FallbackChain is documentation-only.
func Resolve(spec AgentSpec, ctx ResolveCtx) (ResolvedModel, error) {
	if spec.Tier == "" {
		return ResolvedModel{}, fmt.Errorf("%w: agent=%s", ErrTierUndefined, spec.Name)
	}

	tier, promoted := enforceRiskFloor(spec.Tier, spec.Risk)

	candidates := pick(tier, ctx.Catalog)
	candidates = filterDenied(candidates, ctx.Catalog)
	if ctx.Catalog.RequireConfigured {
		candidates = filterUnconfigured(candidates, ctx.ConfiguredProviders)
	}
	if spec.Multimodal {
		candidates = filterMultimodal(candidates)
		if len(candidates) == 0 {
			return ResolvedModel{}, fmt.Errorf("%w: agent=%s tier=%s", ErrMultimodalUnmet, spec.Name, tier)
		}
	}
	if len(candidates) == 0 {
		return ResolvedModel{}, fmt.Errorf("%w: agent=%s tier=%s", ErrNoEligibleModel, spec.Name, tier)
	}

	out := ResolvedModel{
		Field:         candidates[0],
		FallbackChain: candidates[1:],
	}
	if promoted {
		out.Warnings = append(out.Warnings,
			fmt.Sprintf("risk=%d promoted tier %s -> %s", spec.Risk, spec.Tier, tier))
	}
	return out, nil
}

// enforceRiskFloor implements the risk-tier rule: risk=5 forces Frontier
// regardless of declared tier, and risk≥4 cannot be Speed. The primary-agent
// is intentionally Tier=Balanced + Risk=5 (router roles benefit from
// constrained pattern-matching), which the floor leaves promoted.
//
// Rationale: declared tier expresses *role*; risk expresses *blast radius*.
// A role-driven Speed pick is wrong when the blast radius is high.
func enforceRiskFloor(t Tier, risk int) (Tier, bool) {
	switch {
	case risk == 5 && t != TierFrontier:
		return TierFrontier, true
	case risk >= 4 && t == TierSpeed:
		return TierBalanced, true
	}
	return t, false
}

func pick(t Tier, c Catalog) []string {
	switch t {
	case TierFrontier:
		return slices.Clone(c.Frontier)
	case TierBalanced:
		return slices.Clone(c.Balanced)
	case TierSpeed:
		return slices.Clone(c.Speed)
	}
	return nil
}

func filterDenied(in []string, c Catalog) []string {
	out := make([]string, 0, len(in))
	for _, m := range in {
		prov, _ := splitProviderModel(m)
		if slices.Contains(c.DenyProviders, prov) {
			continue
		}
		denied := false
		for _, re := range c.DenyNamePatterns {
			if re.MatchString(m) {
				denied = true
				break
			}
		}
		if !denied {
			out = append(out, m)
		}
	}
	return out
}

func filterUnconfigured(in, configured []string) []string {
	out := make([]string, 0, len(in))
	for _, m := range in {
		prov, _ := splitProviderModel(m)
		if prov == "" || slices.Contains(configured, prov) {
			out = append(out, m)
		}
	}
	return out
}

func filterMultimodal(in []string) []string {
	out := make([]string, 0, len(in))
	for _, m := range in {
		if isMultimodal(m) {
			out = append(out, m)
		}
	}
	return out
}

func isMultimodal(modelField string) bool {
	prov, id := splitProviderModel(modelField)
	for _, e := range allModels {
		if (prov == "" || e.Provider == prov) && e.ID == id && e.Multimodal {
			return true
		}
	}
	return false
}

// splitProviderModel separates "provider/id" forms; bare aliases (Claude
// Code's "opus") and Copilot's symbolic names ("claude-opus-4.7") return an
// empty provider so they bypass the provider-deny / configured-provider
// filters that don't apply to those formats.
func splitProviderModel(m string) (provider, id string) {
	prov, rest, ok := strings.Cut(m, "/")
	if !ok {
		return "", m
	}
	return prov, rest
}
