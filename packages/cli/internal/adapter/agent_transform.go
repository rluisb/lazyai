package adapter

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rluisb/lazyai/packages/cli/internal/auth"
	"github.com/rluisb/lazyai/packages/cli/internal/frontmatter"
	"github.com/rluisb/lazyai/packages/cli/internal/models"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// agentSpecToModelsSpec converts the package-local AgentSpecRaw into the
// models.AgentSpec shape Resolve consumes. Kept as a thin shim because the
// frontmatter package can't import models (would cycle through here).
func agentSpecToModelsSpec(raw frontmatter.AgentSpecRaw) models.AgentSpec {
	return models.AgentSpec{
		Name:        raw.Name,
		Tier:        models.Tier(raw.Tier),
		Temperature: raw.Temperature,
		Thinking:    models.Thinking(raw.Thinking),
		Risk:        raw.Risk,
		Multimodal:  raw.Multimodal,
	}
}

// resolveCtxFor returns a ResolveCtx suitable for the named target. If the
// adapter context did not pre-populate ConfiguredProviders (wizard not run,
// upgrade path), we fall back to a live auth probe so the resolver still
// has a sensible answer for OpenCode's RequireConfigured filter.
//
// For the OpenCode target, `opencode` is treated as a meta-provider: a user
// who has successfully run `opencode auth list` can reach openai / google /
// ollama-cloud / github-copilot models through OpenCode's bundled UI
// without separately authenticating each provider's CLI. Without that
// expansion, an OpenCode-only user would hit `ErrNoEligibleModel` after
// the catalog's invented `opencode/*` entries were removed in #199 Bug 1.
func resolveCtxFor(tool types.ToolId, ctx *AdapterContext) models.ResolveCtx {
	cat := models.CatalogFor(tool)
	rc := models.ResolveCtx{Catalog: cat}
	if !cat.RequireConfigured {
		return rc
	}
	if len(ctx.ConfiguredProviders) > 0 {
		rc.ConfiguredProviders = ctx.ConfiguredProviders
	} else {
		probeCtx, cancel := context.WithCancel(context.Background())
		defer cancel()
		for _, p := range auth.DetectAll(probeCtx) {
			rc.ConfiguredProviders = append(rc.ConfiguredProviders, string(p))
		}
	}
	if tool == types.ToolIdOpenCode {
		rc.ConfiguredProviders = expandOpenCodeMetaProvider(rc.ConfiguredProviders)
	}
	return rc
}

// opencodeBundledProviders lists the upstream providers OpenCode's bundled
// UI can route through. Authenticating `opencode auth list` is treated as
// implicit access to any of these via OpenCode's mediation.
var opencodeBundledProviders = []string{"openai", "google", "ollama-cloud", "github-copilot"}

// expandOpenCodeMetaProvider treats the `opencode` provider ID as a
// meta-provider: when present in the configured set, ensure the bundled
// providers are present too so the resolver doesn't filter every candidate
// out for an OpenCode-only user.
func expandOpenCodeMetaProvider(configured []string) []string {
	hasOpenCode := false
	have := make(map[string]bool, len(configured))
	for _, p := range configured {
		have[p] = true
		if p == "opencode" {
			hasOpenCode = true
		}
	}
	if !hasOpenCode {
		return configured
	}
	for _, p := range opencodeBundledProviders {
		if !have[p] {
			configured = append(configured, p)
			have[p] = true
		}
	}
	return configured
}

// RewriteAgentForClaudeCode transforms a library agent into a Claude Code agent
// file. Output frontmatter contains only name and description; the source body is
// preserved verbatim after the vibe-lab managed marker. Baseline agents carry
// no LazyAI tier metadata, so this path parses generic frontmatter instead of
// frontmatter.ParseAgentSpec.
func RewriteAgentForClaudeCode(source []byte, ctx *AdapterContext) ([]byte, error) {
	_ = ctx
	fm, body, err := frontmatter.ExtractFrontmatter(source)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(frontmatter.ExtractField(fm, "name"))
	if name == "" {
		return nil, fmt.Errorf("claude adapter: agent source missing name")
	}
	description := strings.TrimSpace(frontmatter.ExtractField(fm, "description"))
	body = trimLeadingNewlines(body)

	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("name: ")
	b.WriteString(name)
	b.WriteByte('\n')
	b.WriteString("description: ")
	b.WriteString(yamlDoubleQuote(description))
	b.WriteString("\n---\n\n")
	b.WriteString(managedAgentMarker("claude", name))
	b.WriteString("\n\n")
	b.Write(body)
	return []byte(b.String()), nil
}

// RewriteAgentForOpenCode transforms a library agent into the baseline
// OpenCode agent shape for emitted .opencode/agents files.
//
// OpenCode emits only the quoted description and the managed marker. Source
// frontmatter is parsed generically so baseline agents without a LazyAI tier
// are accepted.
func RewriteAgentForOpenCode(source []byte, ctx *AdapterContext, mode string) ([]byte, error) {
	_ = mode
	_ = ctx
	fm, body, err := frontmatter.ExtractFrontmatter(source)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(frontmatter.ExtractField(fm, "name"))
	description := strings.TrimSpace(frontmatter.ExtractField(fm, "description"))
	if description == "" {
		description = inheritedDescription(fm)
	}
	cleaned := append([]byte{'\n'}, body...)
	return BuildOpenCodeAgentFrontmatter(cleaned, OpenCodeAgentOpts{
		Description:   description,
		ManagedMarker: managedAgentMarker("opencode", name),
	}), nil
}

// modelSectionRe matches the Claude-centric `## Model\n<paragraph>\n\n`
// editorial section inserted in the source `library/agents/*.md` files.
// It exists to give human readers context about Claude tier selection;
// for non-Claude targets the resolved provider/model contradicts that
// commentary, so the section is stripped (#199 Bug 1).
var modelSectionRe = regexp.MustCompile(`(?m)^## Model\n[\s\S]+?(?:\n\n|\z)`)

// stripModelSection removes the `## Model\n…\n\n` paragraph from a markdown
// document while preserving any frontmatter and the rest of the body
// verbatim. Safe to call when no such section exists (returns input
// unchanged).
func stripModelSection(source []byte) []byte {
	return modelSectionRe.ReplaceAll(source, []byte(""))
}

// opencodeReasoningEffort maps the source `thinking:` annotation to
// OpenCode's `reasoningEffort` enum. "none" maps to omit (returns "").
func opencodeReasoningEffort(thinking string) string {
	switch strings.ToLower(thinking) {
	case "high":
		return "high"
	case "medium":
		return "medium"
	case "low":
		return "low"
	case "minimal":
		return "minimal"
	}
	return ""
}

// opencodeTextVerbosity derives `textVerbosity` from the source `risk:`
// annotation. High-risk roles (planning, review) prefer terse output;
// lower-risk roles get the medium default.
func opencodeTextVerbosity(risk int) string {
	if risk >= 4 {
		return "low"
	}
	return "medium"
}

// opencodeStepsFor returns a per-tier max-iteration cap. Values mirror the
// canonical configs at `~/.config/opencode/agents/`: planner=16, researcher=10.
// Frontier roles get more steps; speed roles get fewer.
func opencodeStepsFor(tier string) int {
	switch strings.ToLower(tier) {
	case "frontier":
		return 16
	case "balanced":
		return 20
	case "speed":
		return 10
	}
	return 0
}

// prependFallbackComment inserts a fallback-chain comment between the
// closing frontmatter delimiter and the body. None of the supported CLIs
// read it; it serves humans reviewing the compiled file.
func prependFallbackComment(content []byte, chain []string) []byte {
	if len(chain) == 0 {
		return content
	}
	s := string(content)
	const delim = "---\n"
	// Locate the *closing* delimiter — the second occurrence of "---\n".
	first := strings.Index(s, delim)
	if first < 0 {
		return content
	}
	closeIdx := strings.Index(s[first+len(delim):], delim)
	if closeIdx < 0 {
		return content
	}
	insertAt := first + len(delim) + closeIdx + len(delim)
	comment := fmt.Sprintf("# fallback-chain: %s\n", strings.Join(chain, " -> "))
	return []byte(s[:insertAt] + comment + s[insertAt:])
}

// formatFloat trims trailing zeros so 0.10 emits as "0.1" and 0.0 as "0".
// Stable output makes adapter-output tests easier to write.
func formatFloat(f float64) string {
	s := fmt.Sprintf("%g", f)
	return s
}

func trimLeadingNewlines(b []byte) []byte {
	i := 0
	for i < len(b) && b[i] == '\n' {
		i++
	}
	return b[i:]
}

// copilotAgentNameRe matches the top-level `name:` field of a Copilot
// .agent.yaml file. We use a tolerant string scan rather than full YAML
// parsing because Copilot agents include long inline `prompt: |` blocks
// where adding a YAML dependency for a 2-line lookup is overkill.
var copilotAgentNameRe = regexp.MustCompile(`(?m)^name:\s*(\S+)\s*$`)

// copilotAgentModelRe matches the top-level `model:` line so RewriteCopilotAgent
// can replace it without parsing or re-emitting the rest of the file.
var copilotAgentModelRe = regexp.MustCompile(`(?m)^model:\s*\S+\s*$`)

// RewriteCopilotAgent updates the model: line of a hand-authored Copilot
// .agent.yaml using the tier declared in the corresponding library agent
// markdown. The yaml's body and prompt are preserved verbatim — only the
// model line is touched.
//
// Lookup: the yaml's `name:` field maps to `canonical/agents/<name>.md`. If no
// matching markdown exists, the function returns the input unchanged so the
// existing yaml stays authoritative.
// hand-authored model pin remains in effect.
func RewriteCopilotAgent(content []byte, ctx *AdapterContext) ([]byte, error) {
	nameMatch := copilotAgentNameRe.FindSubmatch(content)
	if nameMatch == nil {
		return content, nil
	}
	name := strings.TrimSpace(string(nameMatch[1]))
	srcMd, ok := loadLibraryAgentMd(ctx, name)
	if !ok {
		return content, nil
	}
	raw, err := frontmatter.ParseAgentSpec(srcMd)
	if err != nil {
		return content, nil
	}
	rc := resolveCtxFor(types.ToolIdCopilot, ctx)
	res, err := models.Resolve(agentSpecToModelsSpec(raw), rc)
	if err != nil {
		return nil, fmt.Errorf("copilot resolve %s: %w", name, err)
	}
	replacement := fmt.Appendf(nil, "model: %s", res.Field)
	return copilotAgentModelRe.ReplaceAll(content, replacement), nil
}

// loadLibraryAgentMd reads the source-of-truth markdown for an agent, used
// by RewriteCopilotAgent to locate tier annotations. Tries the embedded FS
// first, falls back to disk under LibraryDir. Returns ("", false) when no
// matching agent exists at either location.
func loadLibraryAgentMd(ctx *AdapterContext, name string) ([]byte, bool) {
	rel := filepath.ToSlash(filepath.Join("canonical/agents", name+".md"))
	if ctx.LibraryFS != nil {
		if data, err := fs.ReadFile(ctx.LibraryFS, rel); err == nil {
			return data, true
		}
	}
	if ctx.LibraryDir != "" {
		path := filepath.Join(ctx.LibraryDir, "canonical", "agents", name+".md")
		if data, err := os.ReadFile(path); err == nil {
			return data, true
		}
	}
	return nil, false
}
