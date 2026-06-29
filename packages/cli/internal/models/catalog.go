package models

import (
	"regexp"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// ClaudeCodeCatalog uses aliases (opus/sonnet/haiku) so the emitted file
// auto-tracks Anthropic's current model in each tier without churn here.
var ClaudeCodeCatalog = Catalog{
	Format:   FormatAlias,
	Frontier: []string{"opus"},
	Balanced: []string{"sonnet"},
	Speed:    []string{"haiku"},
}

// OpenCodeCatalog spans every provider an OpenCode user might authenticate
// against, in preference order. Anthropic is denied by provider; Claude
// models routed via any other provider are denied by name pattern. Resolve
// further filters by ConfiguredProviders so that a user with only `codex
// login` doesn't get an ollama-cloud reference written.
//
// Note: an `opencode/*` provider prefix used to appear here for a
// "bundled-mix" route, but real-world OpenCode configs (verified against
// `~/.config/opencode/agents/`) use only `openai/`, `google/`,
// `ollama-cloud/`, and `github-copilot/`. The `opencode/*` entries were
// invented and have been removed (#199 Bug 1).
var OpenCodeCatalog = Catalog{
	Format: FormatProviderSlug,
	Frontier: []string{
		"openai/gpt-5.5",
		"github-copilot/gpt-5.5",
		"google/gemini-3.1-pro-preview",
		"ollama-cloud/gpt-oss:120b",
	},
	Balanced: []string{
		"openai/gpt-5.4-mini",
		"github-copilot/gpt-5.4-mini",
		"ollama-cloud/kimi-k2.6:cloud",
		"ollama-cloud/minimax-m2.7",
		"ollama-cloud/glm-4.7",
		"google/gemini-3-flash-preview",
	},
	Speed: []string{
		"google/gemini-3-flash-preview",
		"openai/gpt-5.4-nano",
		"github-copilot/gpt-5-mini",
		"ollama-cloud/gpt-oss:20b",
	},
	DenyProviders:     []string{"anthropic"},
	DenyNamePatterns:  []*regexp.Regexp{regexp.MustCompile(`(?i)claude`)},
	RequireConfigured: true,
}

// CopilotCatalog uses Copilot's symbolic IDs (note dotted form: claude-opus-4.7
// vs Anthropic's claude-opus-4-7). Claude models are allowed here because
// Copilot's licence covers them.
var CopilotCatalog = Catalog{
	Format:   FormatSymbolic,
	Frontier: []string{"claude-opus-4.7", "gpt-5.5", "gemini-3.1-pro-preview"},
	Balanced: []string{"claude-sonnet-4.6", "gpt-5.4-mini"},
	Speed:    []string{"claude-haiku-4.5", "gpt-5-mini"},
}

// CatalogFor returns the Catalog for a given target tool. Used by the
// validation hook in cmd/compile.go and by adapter agent writers.
func CatalogFor(t types.ToolId) Catalog {
	switch t {
	case types.ToolIdClaudeCode:
		return ClaudeCodeCatalog
	case types.ToolIdOpenCode:
		return OpenCodeCatalog
	case types.ToolIdCopilot:
		return CopilotCatalog
	case types.ToolIdPi, types.ToolIdOmp, types.ToolIdKiro, types.ToolIdAntigravity, types.ToolIdCodex:
		return OpenCodeCatalog
	default:
		return Catalog{}
	}
}
