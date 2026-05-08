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
// models routed via any other provider (e.g., opencode/claude-*) are denied
// by name pattern. Resolve further filters by ConfiguredProviders so that a
// user with only `codex login` doesn't get an ollama-cloud reference written.
var OpenCodeCatalog = Catalog{
	Format: FormatProviderSlug,
	Frontier: []string{
		"openai/gpt-5.5",
		"github-copilot/gpt-5.5",
		"google/gemini-3.1-pro-preview",
		"opencode/gpt-5.5",
		"ollama-cloud/gpt-oss:120b",
	},
	Balanced: []string{
		"openai/gpt-5.4-mini",
		"github-copilot/gpt-5.4-mini",
		"ollama-cloud/minimax-m2.7",
		"opencode/glm-4.7",
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
	}
	return Catalog{}
}
