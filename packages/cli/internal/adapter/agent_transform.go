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
func resolveCtxFor(tool types.ToolId, ctx *AdapterContext) models.ResolveCtx {
	cat := models.CatalogFor(tool)
	rc := models.ResolveCtx{Catalog: cat}
	if !cat.RequireConfigured {
		return rc
	}
	if len(ctx.ConfiguredProviders) > 0 {
		rc.ConfiguredProviders = ctx.ConfiguredProviders
		return rc
	}
	probeCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for _, p := range auth.DetectAll(probeCtx) {
		rc.ConfiguredProviders = append(rc.ConfiguredProviders, string(p))
	}
	return rc
}

// RewriteAgentForClaudeCode transforms a library agent (.md with tier-based
// frontmatter) into a Claude Code agent file. The output frontmatter
// contains only the keys Claude Code honours: name, description, model,
// temperature, plus the existing tools field passed through. Agent body is
// preserved verbatim.
//
// The "model" field is the resolved alias (opus/sonnet/haiku); Claude Code
// expands aliases to the latest version on its end so we don't need to pin
// a dated ID here.
func RewriteAgentForClaudeCode(source []byte, ctx *AdapterContext) ([]byte, error) {
	return rewriteAgentForTarget(source, types.ToolIdClaudeCode, ctx, claudeCodeFrontmatter)
}

// RewriteAgentForOpenCode runs the existing BuildOpenCodeAgentFrontmatter
// pipeline but populates the previously-empty Model slot with the result of
// Resolve against the OpenCode catalog (which excludes Claude on every
// provider via DenyProviders + DenyNamePatterns).
func RewriteAgentForOpenCode(source []byte, ctx *AdapterContext, mode string) ([]byte, error) {
	raw, err := frontmatter.ParseAgentSpec(source)
	if err != nil {
		return nil, err
	}
	rc := resolveCtxFor(types.ToolIdOpenCode, ctx)
	res, err := models.Resolve(agentSpecToModelsSpec(raw), rc)
	if err != nil {
		return nil, fmt.Errorf("opencode resolve %s: %w", raw.Name, err)
	}
	out := BuildOpenCodeAgentFrontmatter(source, OpenCodeAgentOpts{
		Mode:  mode,
		Model: res.Field,
	})
	return prependFallbackComment(out, res.FallbackChain), nil
}

// rewriteAgentForTarget is the shared core for all targets that emit a YAML
// frontmatter block. The emit callback is responsible for the per-target
// field set (Claude Code drops `thinking`/`risk`; OpenCode keeps a richer
// subset; Copilot uses a different file format and bypasses this helper).
func rewriteAgentForTarget(
	source []byte,
	tool types.ToolId,
	ctx *AdapterContext,
	emit func(spec frontmatter.AgentSpecRaw, modelField string, fallback []string, body []byte) []byte,
) ([]byte, error) {
	raw, err := frontmatter.ParseAgentSpec(source)
	if err != nil {
		return nil, err
	}
	rc := resolveCtxFor(tool, ctx)
	res, err := models.Resolve(agentSpecToModelsSpec(raw), rc)
	if err != nil {
		return nil, fmt.Errorf("%s resolve %s: %w", tool, raw.Name, err)
	}
	_, body, err := frontmatter.ExtractFrontmatter(source)
	if err != nil {
		return nil, err
	}
	return emit(raw, res.Field, res.FallbackChain, body), nil
}

// claudeCodeFrontmatter writes the minimal frontmatter Claude Code's agent
// loader honours. Tier/thinking/risk are intentionally dropped — Claude
// Code ignores them, and leaving them in source-of-truth shape would
// confuse readers into thinking they're load-bearing on this target.
func claudeCodeFrontmatter(spec frontmatter.AgentSpecRaw, modelField string, fallback []string, body []byte) []byte {
	var b strings.Builder
	b.WriteString("---\n")
	if spec.Name != "" {
		fmt.Fprintf(&b, "name: %s\n", spec.Name)
	}
	fmt.Fprintf(&b, "model: %s\n", modelField)
	fmt.Fprintf(&b, "temperature: %s\n", formatFloat(spec.Temperature))
	b.WriteString("---\n")
	if len(fallback) > 0 {
		fmt.Fprintf(&b, "<!-- fallback-chain: %s -->\n", strings.Join(fallback, " -> "))
	}
	b.WriteString("\n")
	b.Write(trimLeadingNewlines(body))
	return []byte(b.String())
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
// Lookup: the yaml's `name:` field maps to library/agents/<name>.md. If no
// matching markdown exists (e.g., a Copilot-only agent whose tier we can't
// derive), the function returns the input unchanged so the existing
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
	rel := filepath.ToSlash(filepath.Join("agents", name+".md"))
	if ctx.LibraryFS != nil {
		if data, err := fs.ReadFile(ctx.LibraryFS, rel); err == nil {
			return data, true
		}
	}
	if ctx.LibraryDir != "" {
		path := filepath.Join(ctx.LibraryDir, "agents", name+".md")
		if data, err := os.ReadFile(path); err == nil {
			return data, true
		}
	}
	return nil, false
}
