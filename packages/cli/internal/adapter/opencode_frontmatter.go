// OpenCode agent frontmatter emitter.
//
// Library agents ship with a minimal source frontmatter (name, model, tools
// as MCP-server names). Other adapters route agents through
// StripFrontmatterAndInjectModel, which drops the frontmatter entirely and
// replaces it with HTML comments — fine for Claude/Gemini, but opencode
// requires a real YAML frontmatter block with its own schema:
//
//	---
//	name: <string>                  # optional
//	model: <provider/model>         # optional
//	description: <string>
//	reasoningEffort: high | medium | low | minimal   # optional
//	textVerbosity: high | medium | low               # optional
//	mode: primary | subagent | all
//	temperature: <float>            # optional
//	steps: <int>                    # optional
//	tools:                          # optional — opencode tool gates (write/edit/bash)
//	  bash: true
//	  edit: false
//	permission:                     # optional
//	  edit: ask
//	  bash: ask
//	---
//
// BuildOpenCodeAgentFrontmatter produces that block in the canonical key
// order observed in real-world `~/.config/opencode/agents/` configs (#199
// Bug 1). Empty / zero-value fields are omitted entirely so the rendered
// YAML stays minimal.

package adapter

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rluisb/lazyai/packages/cli/internal/frontmatter"
)

// OpenCodeAgentOpts holds per-agent frontmatter overrides. All fields are
// optional: zero-value fields fall back to source-inherited values or
// opencode's own defaults. Field shapes match the canonical OpenCode agent
// frontmatter (#199 Bug 1).
type OpenCodeAgentOpts struct {
	// Name is the agent's identifier. Emitted as `name:`. Canonical OpenCode
	// configs include this; we mirror that. If empty, key is omitted.
	Name string
	// Description is the agent's one-line summary. If empty, derived from
	// the source frontmatter's `name` field (or defaults to "Agent").
	Description string
	// Model is a provider/model identifier (e.g., "openai/gpt-5.5",
	// "ollama-cloud/kimi-k2.6:cloud"). If empty, the key is omitted and
	// opencode uses its configured default.
	Model string
	// Mode is opencode's primary/subagent/all selector. Defaults to
	// "subagent" — matches what real-world OpenCode subagent configs use.
	// Pass "primary" explicitly for the orchestrator agent.
	Mode string
	// Temperature is the agent's sampling temperature (0.0–1.0+). Emitted
	// only when non-zero. Defaults are inherited from opencode otherwise.
	Temperature float64
	// ReasoningEffort maps to OpenCode's `reasoningEffort:` field —
	// "high"/"medium"/"low"/"minimal". Emitted only when set.
	ReasoningEffort string
	// TextVerbosity maps to OpenCode's `textVerbosity:` field —
	// "high"/"medium"/"low". Emitted only when set.
	TextVerbosity string
	// Steps is the per-form iteration cap (e.g., 16 for planning, 25 for
	// implementation). Emitted only when non-zero.
	Steps int
	// Tools is the per-tool allow map (e.g., {"bash": true, "edit": true}).
	// Note: these are OpenCode's BUILT-IN tool gates (write/edit/bash), NOT
	// MCP server names. If nil or empty, the key is omitted and opencode
	// enables all tools by default. MCP servers belong in `.mcp.json`,
	// not here. (#199 Bug 1)
	Tools map[string]bool
	// Permission maps opencode permission names (edit, bash, ...) to one of
	// "ask" | "allow" | "deny", or to a nested allow/deny structure. If nil
	// or empty, the key is omitted.
	Permission map[string]string
}

// BuildOpenCodeAgentFrontmatter returns source rewritten with an
// opencode-schema-valid YAML frontmatter block. The body is preserved
// verbatim (with a single leading blank line between frontmatter and body).
//
// Key emit order mirrors the canonical configs at
// `~/.config/opencode/agents/`:
//
//	name → model → description → reasoningEffort → textVerbosity →
//	mode → temperature → steps → tools (if non-empty) → permission (if non-empty)
func BuildOpenCodeAgentFrontmatter(source []byte, opts OpenCodeAgentOpts) []byte {
	srcFm, body, _ := frontmatter.ExtractFrontmatter(source)

	description := opts.Description
	if description == "" {
		description = inheritedDescription(srcFm)
	}

	mode := opts.Mode
	if mode == "" {
		mode = "subagent"
	}

	var b strings.Builder
	b.WriteString("---\n")

	if opts.Name != "" {
		fmt.Fprintf(&b, "name: %s\n", opts.Name)
	}
	if opts.Model != "" {
		fmt.Fprintf(&b, "model: %s\n", opts.Model)
	}
	b.WriteString("description: ")
	b.WriteString(yamlDoubleQuote(description))
	b.WriteByte('\n')

	if opts.ReasoningEffort != "" {
		fmt.Fprintf(&b, "reasoningEffort: %s\n", opts.ReasoningEffort)
	}
	if opts.TextVerbosity != "" {
		fmt.Fprintf(&b, "textVerbosity: %s\n", opts.TextVerbosity)
	}

	b.WriteString("mode: ")
	b.WriteString(mode)
	b.WriteByte('\n')

	if opts.Temperature != 0 {
		fmt.Fprintf(&b, "temperature: %s\n", trimFloat(opts.Temperature))
	}
	if opts.Steps != 0 {
		fmt.Fprintf(&b, "steps: %d\n", opts.Steps)
	}

	if len(opts.Tools) > 0 {
		b.WriteString("tools:\n")
		for _, k := range sortedStringKeys(opts.Tools) {
			fmt.Fprintf(&b, "  %s: %t\n", k, opts.Tools[k])
		}
	}

	if len(opts.Permission) > 0 {
		b.WriteString("permission:\n")
		for _, k := range sortedStringKeys(opts.Permission) {
			fmt.Fprintf(&b, "  %s: %s\n", k, opts.Permission[k])
		}
	}

	b.WriteString("---\n\n")
	b.WriteString(strings.TrimLeft(string(body), "\n"))
	return []byte(b.String())
}

// inheritedDescription derives a description from the source frontmatter's
// `name` field when no explicit description was provided. Falls back to a
// generic "Agent" label when the source is frontmatter-less.
func inheritedDescription(srcFm map[string]any) string {
	if name, ok := srcFm["name"].(string); ok {
		if trimmed := strings.TrimSpace(name); trimmed != "" {
			return trimmed + " agent"
		}
	}
	if desc, ok := srcFm["description"].(string); ok {
		if trimmed := strings.TrimSpace(desc); trimmed != "" {
			return trimmed
		}
	}
	return "Agent"
}

// yamlDoubleQuote returns a YAML double-quoted scalar. Defensive: the
// description field may contain colons, apostrophes, or other characters
// that would otherwise need escaping in plain form.
func yamlDoubleQuote(s string) string {
	escaped := strings.ReplaceAll(s, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	return `"` + escaped + `"`
}

// sortedStringKeys returns a map's keys in lexical order so YAML output is
// byte-stable across runs (Go's map iteration is randomized).
func sortedStringKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// trimFloat formats a float without trailing zeros (e.g. 0.5 not 0.500000).
func trimFloat(f float64) string {
	s := fmt.Sprintf("%g", f)
	return s
}
