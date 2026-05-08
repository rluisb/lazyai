// Package models maps abstract agent capability tiers (frontier, balanced,
// speed) to concrete model identifiers per target tool. The CLI's three
// adapters (claude-code, opencode, copilot) each declare a Catalog that
// describes which model IDs are eligible for each tier on that target,
// plus deny rules (e.g., OpenCode rejects Anthropic-routed models entirely
// and any model name matching /claude/i regardless of provider). Resolve
// performs the per-target lookup at compile time.
//
// The agent-side input is an AgentSpec parsed from the agent's source
// frontmatter (tier, temperature, thinking, risk, multimodal). The
// per-target output is a ResolvedModel — the model field to write into the
// emitted agent file (or omit, for targets that pick model per-conversation
// instead of per-agent), plus a fallback chain that downstream code may
// surface as a YAML comment for documentation.
package models

import "regexp"

// Tier classifies an agent by capability requirement. Frontier = strongest
// reasoning available; Balanced = solid reasoning at moderate cost; Speed =
// fast, cheap, adequate for constrained tasks.
type Tier string

const (
	TierFrontier Tier = "frontier"
	TierBalanced Tier = "balanced"
	TierSpeed    Tier = "speed"
)

// Thinking is an abstract budget hint that adapters translate to per-target
// fields (e.g., OpenCode's reasoningEffort, thinkingBudget). Targets that
// don't honour reasoning budgets drop the field on emit.
type Thinking string

const (
	ThinkingHigh   Thinking = "high"
	ThinkingMedium Thinking = "medium"
	ThinkingLow    Thinking = "low"
	ThinkingNone   Thinking = "none"
)

// AgentSpec is the per-agent input to Resolve. It is parsed from the source
// frontmatter and is the same across all targets for a given agent.
type AgentSpec struct {
	Name        string
	Tier        Tier
	Temperature float64
	Thinking    Thinking
	Risk        int  // 1-5; risk≥4 enforces a tier floor (see enforceRiskFloor)
	Multimodal  bool // when true, only multimodal candidates are eligible
}

// ModelFormat governs how the resolved model identifier is serialised into
// the target's agent file. Different tools have incompatible conventions:
// Claude Code expects an alias ("opus"/"sonnet"/"haiku"); OpenCode expects
// "provider/model"; Copilot expects a symbolic name from its own catalog.
type ModelFormat int

const (
	FormatAlias ModelFormat = iota
	FormatProviderSlug
	FormatSymbolic
)

// Catalog declares which model IDs are eligible per tier on one target tool,
// plus deny rules and provider-configured filtering policy.
type Catalog struct {
	Format ModelFormat

	// Per-tier candidate lists in preference order. The first eligible entry
	// becomes the primary model; the rest is the fallback chain.
	Frontier []string
	Balanced []string
	Speed    []string

	// DenyProviders blocks any candidate whose "provider/" prefix matches an
	// entry. OpenCode uses this to refuse anthropic-routed Claude models.
	DenyProviders []string

	// DenyNamePatterns blocks any candidate whose full ID matches a regex.
	// OpenCode uses this to refuse Claude-named models from non-Anthropic
	// providers (e.g., opencode/claude-* or github-copilot/claude-*).
	DenyNamePatterns []*regexp.Regexp

	// RequireConfigured restricts candidates to those whose provider has been
	// authenticated by the user. OpenCode sets this true; Claude Code and
	// Copilot leave it false because they each have a single implicit
	// provider tied to the tool itself.
	RequireConfigured bool
}

// ResolvedModel is Resolve's output for one (agent, target) pair.
type ResolvedModel struct {
	// Field is the model identifier to write into the target's agent file.
	Field string

	// FallbackChain is the remaining eligible candidates, in preference
	// order. None of the supported CLIs read fallback chains at runtime;
	// adapters surface this as a YAML comment for documentation only.
	FallbackChain []string

	// Warnings are non-fatal advisories (e.g., risk-floor promotion fired,
	// only one candidate remained after filtering).
	Warnings []string
}

// ResolveCtx carries per-call inputs to Resolve.
type ResolveCtx struct {
	Catalog Catalog

	// ConfiguredProviders is the set of provider IDs the user has
	// authenticated for this target (e.g., ["openai", "ollama-cloud"] for
	// OpenCode). Ignored when Catalog.RequireConfigured is false.
	ConfiguredProviders []string
}
