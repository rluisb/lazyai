package adapter

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
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
// file. Output frontmatter contains name, description, and — for read-only
// agents — a disallowedTools line that denies Edit, Write, and Bash. The source
// body is preserved verbatim after the vibe-lab managed marker. Baseline agents
// carry no LazyAI tier metadata, so this path parses generic frontmatter instead
// of frontmatter.ParseAgentSpec.
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

	grants, grantErr := frontmatter.ParseAgentToolGrants(source)
	if grantErr != nil {
		return nil, fmt.Errorf("claude adapter: %w", grantErr)
	}
	disallowed := claudeDisallowedTools(grants)

	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("name: ")
	b.WriteString(name)
	b.WriteByte('\n')
	b.WriteString("description: ")
	b.WriteString(yamlDoubleQuote(description))
	b.WriteByte('\n')
	if len(disallowed) > 0 {
		b.WriteString("disallowedTools: ")
		b.WriteString(strings.Join(disallowed, " "))
		b.WriteByte('\n')
	}
	b.WriteString("---\n\n")
	b.WriteString(managedAgentMarker("claude", name))
	b.WriteString("\n\n")
	b.Write(body)
	return []byte(b.String()), nil
}

// claudeDisallowedTools returns the Claude Code disallowedTools list for the
// given canonical grants. Returns nil (no restriction) for nil grants (legacy
// unrestricted) or full-capability grants. Returns ["Edit", "Write", "Bash"]
// for read-only agents (grants containing only read and/or search tokens).
func claudeDisallowedTools(grants []frontmatter.AgentToolGrant) []string {
	if len(grants) == 0 {
		// nil or empty: unrestricted legacy — emit nothing.
		return nil
	}
	for _, g := range grants {
		switch g {
		case frontmatter.AgentToolEdit, frontmatter.AgentToolShell,
			frontmatter.AgentToolWeb, frontmatter.AgentToolMCP, frontmatter.AgentToolSpawn:
			// Has at least one write/exec/spawn capability — not read-only.
			return nil
		}
	}
	// Only read and/or search grants present: deny destructive Claude built-ins.
	return []string{"Edit", "Write", "Bash"}
}

// opencodePermissionForGrants derives the OpenCode permission map from
// canonical tool grants. Returns nil when grants are nil (legacy/unrestricted)
// or when the agent holds write or shell capability (no restriction needed
// — OpenCode defaults to allowing all tools).
//
// Read-only agents — those whose grants contain neither edit nor shell —
// receive { "bash": "deny", "edit": "deny" } to prevent working-tree
// mutations or shell execution.
func opencodePermissionForGrants(grants []frontmatter.AgentToolGrant) map[string]string {
	if grants == nil {
		return nil
	}
	for _, g := range grants {
		if g == frontmatter.AgentToolEdit || g == frontmatter.AgentToolShell {
			return nil
		}
	}
	return map[string]string{
		"bash": "deny",
		"edit": "deny",
	}
}

// RewriteAgentForOpenCode transforms a library agent into the baseline
// OpenCode agent shape for emitted .opencode/agents files.
//
// OpenCode emits only the quoted description and the managed marker. Source
// frontmatter is parsed generically so baseline agents without a LazyAI tier
// are accepted. Canonical agents with only read/search grants receive
// permission: { bash: deny, edit: deny } so they cannot mutate the working
// tree or execute shell commands.
func RewriteAgentForOpenCode(source []byte, ctx *AdapterContext, mode string) ([]byte, error) {
	_ = mode
	_ = ctx
	grants, err := frontmatter.ParseAgentToolGrants(source)
	if err != nil {
		return nil, fmt.Errorf("opencode adapter: parse tool grants: %w", err)
	}
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
		Permission:    opencodePermissionForGrants(grants),
	}), nil
}

// codexAgentTOML is the serialized shape of a Codex custom subagent file
// (.codex/agents/<name>.toml). Codex requires name, description, and
// developer_instructions; see https://developers.openai.com/codex/subagents.
type codexAgentTOML struct {
	Name                  string `toml:"name"`
	Description           string `toml:"description"`
	DeveloperInstructions string `toml:"developer_instructions"`
}

// RewriteAgentForCodex transforms a canonical library agent (markdown with
// name/description frontmatter) into a Codex custom subagent TOML file. The
// markdown body becomes developer_instructions. Source frontmatter is parsed
// generically so baseline agents without a LazyAI tier are accepted.
func RewriteAgentForCodex(source []byte, ctx *AdapterContext) ([]byte, error) {
	_ = ctx
	fm, body, err := frontmatter.ExtractFrontmatter(source)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(frontmatter.ExtractField(fm, "name"))
	if name == "" {
		return nil, fmt.Errorf("codex adapter: agent source missing name")
	}
	description := strings.TrimSpace(frontmatter.ExtractField(fm, "description"))
	instructions := strings.TrimSpace(string(stripModelSection(body)))
	if instructions == "" {
		instructions = description
	}
	var buf bytes.Buffer
	buf.WriteString("# Generated by LazyAI — Codex custom subagent. Edit canonical/agents/" + name + ".md and recompile.\n")
	if err := toml.NewEncoder(&buf).Encode(codexAgentTOML{
		Name:                  name,
		Description:           description,
		DeveloperInstructions: instructions,
	}); err != nil {
		return nil, fmt.Errorf("codex adapter: encode agent %q: %w", name, err)
	}
	return buf.Bytes(), nil
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

// ompToolsForGrants maps canonical capability grants to the OMP-native tool
// name allowlist. Canonical grant → OMP tool name(s):
//
//	read   → "read"
//	edit   → "edit", "write"  (OMP has both as distinct write operations)
//	shell  → "bash"
//	search → "search"
//	web    → "web_search"
//	mcp    → (skip — no generic OMP MCP token; server-specific mcp__<srv>_<tool>)
//	spawn  → "task"
//
// Nil/empty grants (no tools: field in source) returns the full default set,
// preserving unrestricted legacy behaviour for agents without a capability
// declaration.
func ompToolsForGrants(grants []frontmatter.AgentToolGrant) []string {
	if len(grants) == 0 {
		return []string{"read", "search", "bash", "edit", "write", "web_search", "task"}
	}
	var tools []string
	for _, g := range grants {
		switch g {
		case frontmatter.AgentToolRead:
			tools = append(tools, "read")
		case frontmatter.AgentToolSearch:
			tools = append(tools, "search")
		case frontmatter.AgentToolEdit:
			tools = append(tools, "edit", "write")
		case frontmatter.AgentToolShell:
			tools = append(tools, "bash")
		case frontmatter.AgentToolWeb:
			tools = append(tools, "web_search")
		case frontmatter.AgentToolMCP:
			// No generic OMP MCP token; server-specific names (mcp__<srv>_<tool>) are
			// configured outside agent frontmatter.
		case frontmatter.AgentToolSpawn:
			tools = append(tools, "task")
		}
	}
	return tools
}

// ompIsReadOnly reports whether grants contain only read/search tokens, meaning
// the agent has no mutation capability. Returns false for nil/empty grants
// (unrestricted legacy agents).
func ompIsReadOnly(grants []frontmatter.AgentToolGrant) bool {
	if len(grants) == 0 {
		return false
	}
	for _, g := range grants {
		if g != frontmatter.AgentToolRead && g != frontmatter.AgentToolSearch {
			return false
		}
	}
	return true
}

// ompThinkingLevel derives the OMP thinkingLevel value from capability grants
// and agent name:
//
//	Read-only grants (only read/search) → "low"
//	name == "planner"                  → "high"
//	anything else                      → "auto"
func ompThinkingLevel(name string, grants []frontmatter.AgentToolGrant) string {
	if ompIsReadOnly(grants) {
		return "low"
	}
	if strings.ToLower(name) == "planner" {
		return "high"
	}
	return "auto"
}

// extractStringList returns the string slice stored under key in a parsed
// frontmatter map. Returns nil when the key is absent, the value is not a
// YAML sequence, or the sequence contains no non-empty string items.
func extractStringList(fm map[string]any, key string) []string {
	val, ok := fm[key]
	if !ok || val == nil {
		return nil
	}
	items, ok := val.([]any)
	if !ok {
		return nil
	}
	var result []string
	for _, item := range items {
		s, ok := item.(string)
		if ok && s != "" {
			result = append(result, s)
		}
	}
	return result
}

// RewriteAgentForOMP transforms a library agent into an OMP-native subagent
// file.
//
// Output frontmatter:
//   - name, description (required)
//   - tools: OMP allowlist derived from canonical capability grants (ParseAgentToolGrants)
//   - thinkingLevel: "low" for read-only, "high" for planner, "auto" otherwise
//   - autoloadSkills: from canonical skills:, omitted when empty
//
// LazyAI-only fields (role, mode, temperature, steps) are dropped.
// The vibe-lab managed marker is inserted; the body is preserved verbatim.
func RewriteAgentForOMP(source []byte, ctx *AdapterContext) ([]byte, error) {
	_ = ctx
	fm, body, err := frontmatter.ExtractFrontmatter(source)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(frontmatter.ExtractField(fm, "name"))
	if name == "" {
		return nil, fmt.Errorf("omp adapter: agent source missing name")
	}
	description := strings.TrimSpace(frontmatter.ExtractField(fm, "description"))

	grants, err := frontmatter.ParseAgentToolGrants(source)
	if err != nil {
		return nil, fmt.Errorf("omp adapter: parse tool grants for %q: %w", name, err)
	}

	tools := ompToolsForGrants(grants)
	thinkingLvl := ompThinkingLevel(name, grants)
	skills := extractStringList(fm, "skills")

	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("name: ")
	b.WriteString(name)
	b.WriteByte('\n')
	b.WriteString("description: ")
	b.WriteString(yamlDoubleQuote(description))
	b.WriteByte('\n')
	// Emit tools as an inline YAML sequence of quoted strings.
	b.WriteString("tools: [")
	for i, t := range tools {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteByte('"')
		b.WriteString(t)
		b.WriteByte('"')
	}
	b.WriteString("]\n")
	b.WriteString("thinkingLevel: ")
	b.WriteString(thinkingLvl)
	b.WriteByte('\n')
	if len(skills) > 0 {
		b.WriteString("autoloadSkills: [")
		for i, s := range skills {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteByte('"')
			b.WriteString(s)
			b.WriteByte('"')
		}
		b.WriteString("]\n")
	}
	b.WriteString("---\n\n")
	b.WriteString(managedAgentMarker("omp", name))
	b.WriteString("\n\n")
	b.Write(trimLeadingNewlines(body))
	return []byte(b.String()), nil
}
