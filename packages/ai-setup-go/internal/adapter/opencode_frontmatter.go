// OpenCode agent frontmatter emitter.
//
// Library agents ship with a minimal source frontmatter (name, model, tools
// as MCP-server names). Other adapters route agents through
// StripFrontmatterAndInjectModel, which drops the frontmatter entirely and
// replaces it with HTML comments — fine for Claude/Gemini, but opencode
// requires a real YAML frontmatter block with its own schema:
//
//	---
//	description: <string>
//	mode: primary | subagent | all
//	tools:
//	  bash: true
//	  read: true
//	model: <provider/model>        # optional
//	permission:                    # optional
//	  edit: ask
//	  bash: ask
//	---
//
// BuildOpenCodeAgentFrontmatter produces that block. It strips any existing
// frontmatter from the source, inherits `description` from the source's
// `name` when opts.Description is empty, and emits stable (sorted-key) YAML
// so tests can assert on exact output.

package adapter

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ricardoborges-teachable/ai-setup/internal/frontmatter"
)

// OpenCodeAgentOpts holds per-agent frontmatter overrides. All fields are
// optional: zero-value fields fall back to source-inherited values or
// opencode's own defaults.
type OpenCodeAgentOpts struct {
	// Description is the agent's one-line summary. If empty, it is derived
	// from the source frontmatter's `name` field (or defaults to "Agent").
	Description string
	// Mode is opencode's primary/subagent/all selector. Defaults to "all".
	Mode string
	// Tools is the per-tool allow map (e.g., {"bash": true, "read": true}).
	// If nil or empty, the key is omitted and opencode enables all tools.
	Tools map[string]bool
	// Model is a provider/model identifier (e.g., "anthropic/claude-sonnet-4-5").
	// If empty, the key is omitted and opencode uses its configured default.
	Model string
	// Permission maps opencode permission names (edit, bash, ...) to one of
	// "ask" | "allow" | "deny". If nil or empty, the key is omitted.
	Permission map[string]string
}

// BuildOpenCodeAgentFrontmatter returns source rewritten with an
// opencode-schema-valid YAML frontmatter block. The body is preserved
// verbatim (with a single leading blank line between frontmatter and body).
func BuildOpenCodeAgentFrontmatter(source []byte, opts OpenCodeAgentOpts) []byte {
	srcFm, body, _ := frontmatter.ExtractFrontmatter(source)

	description := opts.Description
	if description == "" {
		description = inheritedDescription(srcFm)
	}

	mode := opts.Mode
	if mode == "" {
		mode = "all"
	}

	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("description: ")
	b.WriteString(yamlDoubleQuote(description))
	b.WriteByte('\n')
	b.WriteString("mode: ")
	b.WriteString(mode)
	b.WriteByte('\n')

	if len(opts.Tools) > 0 {
		b.WriteString("tools:\n")
		for _, k := range sortedStringKeys(opts.Tools) {
			fmt.Fprintf(&b, "  %s: %t\n", k, opts.Tools[k])
		}
	}

	if opts.Model != "" {
		fmt.Fprintf(&b, "model: %s\n", opts.Model)
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
